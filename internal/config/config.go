package config

import (
	"log"

	"github.com/spf13/viper"
)

func Load() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("/app/config")
	viper.AddConfigPath(".")
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.token", "")
	viper.SetDefault("rss.max_item_count", 20)
	viper.SetDefault("rss.keep_old_count", 50)
	viper.SetDefault("rss.enc_feed_id", false)
	viper.SetDefault("rss.static", false)
	viper.SetDefault("rss.proxy_disable_img", false)
	viper.SetDefault("scheduler.times", []string{"07:00", "12:00", "20:00"})

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Config file not found, using defaults: %v", err)
	}

	// Override with environment variables
	viper.AutomaticEnv()

	return nil
}

func GetString(key string) string {
	return viper.GetString(key)
}

func GetInt(key string) int {
	return viper.GetInt(key)
}

func GetBool(key string) bool {
	return viper.GetBool(key)
}

func GetStringSlice(key string) []string {
	return viper.GetStringSlice(key)
}
