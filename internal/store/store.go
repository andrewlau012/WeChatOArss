package store

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"time"

	_ "github.com/glebarez/sqlite"
	"github.com/spf13/viper"

	"wechatoarss/internal/model"
)

var db *sql.DB

func InitDB() error {
	// Get database path
	dbPath := viper.GetString("database.path")
	if dbPath == "" {
		// Default to data directory
		homeDir, _ := os.UserHomeDir()
		dbPath = filepath.Join(homeDir, "wechatoarss", "data", "wechatoarss.db")
	}

	// Ensure directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	// Open database
	var err error
	db, err = sql.Open("sqlite", dbPath+"?_journal_mode=WAL&_busy_timeout=5000")
	if err != nil {
		return err
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Create tables
	if err := createTables(); err != nil {
		return err
	}

	log.Printf("Database initialized at: %s", dbPath)
	return nil
}

func createTables() error {
	// Accounts table
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS accounts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			cookie TEXT,
			token TEXT,
			available INTEGER DEFAULT 1,
			need_check INTEGER DEFAULT 0,
			wait_time TEXT,
			created_at TEXT DEFAULT (datetime('now')),
			updated_at TEXT DEFAULT (datetime('now'))
		)
	`)
	if err != nil {
		return err
	}

	// Channels table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS channels (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			biz_id TEXT NOT NULL UNIQUE,
			name TEXT NOT NULL,
			description TEXT,
			avatar TEXT,
			link TEXT,
			account_id INTEGER,
			last_update TEXT,
			article_count INTEGER DEFAULT 0,
			status TEXT DEFAULT 'active',
			created_at TEXT DEFAULT (datetime('now')),
			FOREIGN KEY (account_id) REFERENCES accounts(id)
		)
	`)
	if err != nil {
		return err
	}

	// Articles table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS articles (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			biz_id TEXT NOT NULL,
			title TEXT NOT NULL,
			description TEXT,
			content TEXT,
			link TEXT NOT NULL UNIQUE,
			cover TEXT,
			created_at TEXT DEFAULT (datetime('now')),
			published_at TEXT,
			FOREIGN KEY (biz_id) REFERENCES channels(biz_id)
		)
	`)
	if err != nil {
		return err
	}

	// Create indexes
	_, err = db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_articles_biz_id ON articles(biz_id);
		CREATE INDEX IF NOT EXISTS idx_articles_published ON articles(published_at);
		CREATE INDEX IF NOT EXISTS idx_channels_status ON channels(status);
	`)
	if err != nil {
		return err
	}

	return nil
}

func GetDB() *sql.DB {
	return db
}

// Account operations
func CreateAccount(name, cookie, token string) (*model.Account, error) {
	result, err := db.Exec(
		"INSERT INTO accounts (name, cookie, token) VALUES (?, ?, ?)",
		name, cookie, token,
	)
	if err != nil {
		return nil, err
	}

	id, _ := result.LastInsertId()
	return &model.Account{
		ID:        id,
		Name:      name,
		Cookie:    cookie,
		Token:     token,
		Available: true,
		CreatedAt: time.Now(),
	}, nil
}

func GetAccounts() ([]model.Account, error) {
	rows, err := db.Query(`
		SELECT id, name, cookie, token, available, need_check, wait_time, created_at, updated_at 
		FROM accounts
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []model.Account
	for rows.Next() {
		var a model.Account
		var waitTime sql.NullString
		err := rows.Scan(&a.ID, &a.Name, &a.Cookie, &a.Token, &a.Available, &a.NeedCheck, &waitTime, &a.CreatedAt, &a.UpdatedAt)
		if err != nil {
			return nil, err
		}
		if waitTime.Valid {
			a.WaitTime, _ = time.Parse("2006-01-02 15:04:05", waitTime.String)
		}
		accounts = append(accounts, a)
	}
	return accounts, nil
}

func GetAccountByID(id int64) (*model.Account, error) {
	var a model.Account
	var waitTime sql.NullString
	err := db.QueryRow(`
		SELECT id, name, cookie, token, available, need_check, wait_time, created_at, updated_at 
		FROM accounts WHERE id = ?
	`, id).Scan(&a.ID, &a.Name, &a.Cookie, &a.Token, &a.Available, &a.NeedCheck, &waitTime, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return nil, err
	}
	if waitTime.Valid {
		a.WaitTime, _ = time.Parse("2006-01-02 15:04:05", waitTime.String)
	}
	return &a, nil
}

func UpdateAccount(id int64, name, cookie, token string, available bool) error {
	_, err := db.Exec(`
		UPDATE accounts SET name = ?, cookie = ?, token = ?, available = ?, updated_at = datetime('now') WHERE id = ?
	`, name, cookie, token, available, id)
	return err
}

func DeleteAccount(id int64) error {
	_, err := db.Exec("DELETE FROM accounts WHERE id = ?", id)
	return err
}

// Channel operations
func CreateChannel(bizID, name, description, avatar, link string, accountID int64) (*model.Channel, error) {
	result, err := db.Exec(`
		INSERT INTO channels (biz_id, name, description, avatar, link, account_id, last_update)
		VALUES (?, ?, ?, ?, ?, ?, datetime('now'))
	`, bizID, name, description, avatar, link, accountID)
	if err != nil {
		return nil, err
	}

	id, _ := result.LastInsertId()
	return &model.Channel{
		ID:        id,
		BizID:     bizID,
		Name:      name,
		Description: description,
		Avatar:    avatar,
		Link:      link,
		AccountID: accountID,
		Status:    "active",
		LastUpdate: time.Now(),
	}, nil
}

func GetChannels(page, size int, name string) ([]model.Channel, int, error) {
	offset := (page - 1) * size
	var query string
	var countQuery string
	var args []interface{}

	if name != "" {
		query = "SELECT id, biz_id, name, description, avatar, link, account_id, last_update, article_count, status, created_at FROM channels WHERE name LIKE ? LIMIT ? OFFSET ?"
		countQuery = "SELECT COUNT(*) FROM channels WHERE name LIKE ?"
		args = []interface{}{"%" + name + "%", size, offset}
	} else {
		query = "SELECT id, biz_id, name, description, avatar, link, account_id, last_update, article_count, status, created_at FROM channels LIMIT ? OFFSET ?"
		countQuery = "SELECT COUNT(*) FROM channels"
		args = []interface{}{size, offset}
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var channels []model.Channel
	for rows.Next() {
		var c model.Channel
		var lastUpdate, createdAt sql.NullString
		err := rows.Scan(&c.ID, &c.BizID, &c.Name, &c.Description, &c.Avatar, &c.Link, &c.AccountID, &lastUpdate, &c.ArticleCount, &c.Status, &createdAt)
		if err != nil {
			return nil, 0, err
		}
		if lastUpdate.Valid {
			c.LastUpdate, _ = time.Parse("2006-01-02 15:04:05", lastUpdate.String)
		}
		if createdAt.Valid {
			c.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt.String)
		}
		channels = append(channels, c)
	}

	// Get total count
	var total int
	if name != "" {
		db.QueryRow(countQuery, "%"+name+"%").Scan(&total)
	} else {
		db.QueryRow(countQuery).Scan(&total)
	}

	return channels, total, nil
}

func GetChannelByBizID(bizID string) (*model.Channel, error) {
	var c model.Channel
	var lastUpdate, createdAt sql.NullString
	err := db.QueryRow(`
		SELECT id, biz_id, name, description, avatar, link, account_id, last_update, article_count, status, created_at 
		FROM channels WHERE biz_id = ?
	`, bizID).Scan(&c.ID, &c.BizID, &c.Name, &c.Description, &c.Avatar, &c.Link, &c.AccountID, &lastUpdate, &c.ArticleCount, &c.Status, &createdAt)
	if err != nil {
		return nil, err
	}
	if lastUpdate.Valid {
		c.LastUpdate, _ = time.Parse("2006-01-02 15:04:05", lastUpdate.String)
	}
	return &c, nil
}

func GetActiveChannels() ([]model.Channel, error) {
	rows, err := db.Query(`
		SELECT id, biz_id, name, description, avatar, link, account_id, last_update, article_count, status, created_at 
		FROM channels WHERE status = 'active'
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var channels []model.Channel
	for rows.Next() {
		var c model.Channel
		var lastUpdate, createdAt sql.NullString
		err := rows.Scan(&c.ID, &c.BizID, &c.Name, &c.Description, &c.Avatar, &c.Link, &c.AccountID, &lastUpdate, &c.ArticleCount, &c.Status, &createdAt)
		if err != nil {
			return nil, err
		}
		if lastUpdate.Valid {
			c.LastUpdate, _ = time.Parse("2006-01-02 15:04:05", lastUpdate.String)
		}
		channels = append(channels, c)
	}
	return channels, nil
}

func DeleteChannel(bizID string) error {
	_, err := db.Exec("DELETE FROM channels WHERE biz_id = ?", bizID)
	if err != nil {
		return err
	}
	_, err = db.Exec("DELETE FROM articles WHERE biz_id = ?", bizID)
	return err
}

func UpdateChannelStatus(bizID string, status string) error {
	_, err := db.Exec("UPDATE channels SET status = ? WHERE biz_id = ?", status, bizID)
	return err
}

func UpdateChannelArticleCount(bizID string, count int) error {
	_, err := db.Exec("UPDATE channels SET article_count = ?, last_update = datetime('now') WHERE biz_id = ?", count, bizID)
	return err
}

// Article operations
func CreateArticle(bizID, title, description, content, link, cover string, publishedAt time.Time) (*model.Article, error) {
	result, err := db.Exec(`
		INSERT OR IGNORE INTO articles (biz_id, title, description, content, link, cover, published_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, datetime('now'))
	`, bizID, title, description, content, link, cover, publishedAt)
	if err != nil {
		return nil, err
	}

	id, _ := result.LastInsertId()
	if id == 0 {
		return nil, nil // Already exists
	}

	return &model.Article{
		ID:          id,
		BizID:       bizID,
		Title:       title,
		Description: description,
		Content:     content,
		Link:        link,
		Cover:       cover,
		PublishedAt: publishedAt,
		CreatedAt:   time.Now(),
	}, nil
}

func GetArticles(bizID string, before, after string, page, size int, includeContent bool) ([]model.Article, int, error) {
	offset := (page - 1) * size
	var args []interface{}
	query := "SELECT id, biz_id, title, description, content, link, cover, created_at, published_at FROM articles"
	where := ""

	if bizID != "" {
		where = " biz_id = ?"
		args = append(args, bizID)
	}

	if before != "" {
		if where != "" {
			where += " AND"
		}
		where += " published_at < ?"
		t, _ := time.Parse("20060102", before)
		args = append(args, t.Format("2006-01-02"))
	}

	if after != "" {
		if where != "" {
			where += " AND"
		}
		where += " published_at > ?"
		t, _ := time.Parse("20060102", after)
		args = append(args, t.Format("2006-01-02"))
	}

	if where != "" {
		query += " WHERE" + where
	}

	query += " ORDER BY published_at DESC LIMIT ? OFFSET ?"
	args = append(args, size, offset)

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var articles []model.Article
	for rows.Next() {
		var a model.Article
		var createdAt, publishedAt sql.NullString
		var content sql.NullString
		if includeContent {
			err = rows.Scan(&a.ID, &a.BizID, &a.Title, &a.Description, &content, &a.Link, &a.Cover, &createdAt, &publishedAt)
		} else {
			err = rows.Scan(&a.ID, &a.BizID, &a.Title, &a.Description, &content, &a.Link, &a.Cover, &createdAt, &publishedAt)
		}
		if err != nil {
			return nil, 0, err
		}
		if content.Valid {
			a.Content = content.String
		}
		if createdAt.Valid {
			a.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt.String)
		}
		if publishedAt.Valid {
			a.PublishedAt, _ = time.Parse("2006-01-02 15:04:05", publishedAt.String)
		}
		articles = append(articles, a)
	}

	// Get total count
	var total int
	countQuery := "SELECT COUNT(*) FROM articles"
	if bizID != "" {
		countQuery += " WHERE biz_id = ?"
		db.QueryRow(countQuery, bizID).Scan(&total)
	} else {
		db.QueryRow(countQuery).Scan(&total)
	}

	return articles, total, nil
}

func GetArticleByID(id int64) (*model.Article, error) {
	var a model.Article
	var createdAt, publishedAt sql.NullString
	err := db.QueryRow(`
		SELECT id, biz_id, title, description, content, link, cover, created_at, published_at 
		FROM articles WHERE id = ?
	`, id).Scan(&a.ID, &a.BizID, &a.Title, &a.Description, &a.Content, &a.Link, &a.Cover, &createdAt, &publishedAt)
	if err != nil {
		return nil, err
	}
	if createdAt.Valid {
		a.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt.String)
	}
	if publishedAt.Valid {
		a.PublishedAt, _ = time.Parse("2006-01-02 15:04:05", publishedAt.String)
	}

	// Get channel name
	var channelName string
	db.QueryRow("SELECT name FROM channels WHERE biz_id = ?", a.BizID).Scan(&channelName)
	a.ChannelName = channelName

	return &a, nil
}

func SearchArticles(keyword string, page, size int) ([]model.Article, int, error) {
	offset := (page - 1) * size
	rows, err := db.Query(`
		SELECT id, biz_id, title, description, content, link, cover, created_at, published_at 
		FROM articles 
		WHERE title LIKE ? OR description LIKE ? OR content LIKE ?
		ORDER BY published_at DESC 
		LIMIT ? OFFSET ?
	`, "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%", size, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var articles []model.Article
	for rows.Next() {
		var a model.Article
		var createdAt, publishedAt sql.NullString
		err := rows.Scan(&a.ID, &a.BizID, &a.Title, &a.Description, &a.Content, &a.Link, &a.Cover, &createdAt, &publishedAt)
		if err != nil {
			return nil, 0, err
		}
		if createdAt.Valid {
			a.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt.String)
		}
		if publishedAt.Valid {
			a.PublishedAt, _ = time.Parse("2006-01-02 15:04:05", publishedAt.String)
		}
		articles = append(articles, a)
	}

	var total int
	db.QueryRow(`
		SELECT COUNT(*) FROM articles 
		WHERE title LIKE ? OR description LIKE ? OR content LIKE ?
	`, "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%").Scan(&total)

	return articles, total, nil
}
