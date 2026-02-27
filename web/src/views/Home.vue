<template>
  <div class="home-page">
    <header class="header">
      <div class="header-content">
        <router-link to="/" class="logo">
          <svg viewBox="0 0 24 24" fill="currentColor">
            <path d="M8.691 2.188C3.891 2.188 0 5.476 0 9.53c0 2.212 1.17 4.203 3.002 5.55a.59.59 0 0 1 .213.665l-.39 1.48c-.019.07-.048.141-.048.213 0 .163.13.295.29.295a.326.326 0 0 0 .167-.054l1.903-1.114a.864.864 0 0 1 .717-.098 10.16 10.16 0 0 0 2.837.403c.276 0 .543-.027.811-.05-.857-2.578.157-4.972 1.932-6.446 1.703-1.415 3.882-1.98 5.853-1.838-.576-3.583-4.196-6.348-8.596-6.348z"/>
          </svg>
          WeChatOArss
        </router-link>
        
        <div class="search-box">
          <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="11" cy="11" r="8"/>
            <path d="M21 21l-4.35-4.35"/>
          </svg>
          <input 
            type="text" 
            v-model="searchKeyword" 
            placeholder="搜索文章..."
            @keyup.enter="handleSearch"
          />
        </div>

        <nav class="nav">
          <router-link to="/">首页</router-link>
          <router-link to="/channels">公众号</router-link>
          <router-link to="/settings">设置</router-link>
          <router-link to="/about">关于</router-link>
        </nav>
      </div>
    </header>

    <main class="container">
      <div class="filter-bar">
        <button 
          class="btn" 
          :class="{ active: filter === 'all' }"
          @click="setFilter('all')"
        >
          全部
        </button>
        <button 
          class="btn" 
          :class="{ active: filter === 'today' }"
          @click="setFilter('today')"
        >
          今天
        </button>
        <button 
          class="btn" 
          :class="{ active: filter === 'yesterday' }"
          @click="setFilter('yesterday')"
        >
          昨天
        </button>
        <button 
          class="btn" 
          :class="{ active: filter === 'week' }"
          @click="setFilter('week')"
        >
          本周
        </button>
      </div>

      <div v-if="loading" class="loading">
        <div class="spinner"></div>
      </div>

      <div v-else-if="articles.length === 0" class="empty">
        <svg viewBox="0 0 24 24" fill="currentColor">
          <path d="M19 3H5c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h14c1.1 0 2-.9 2-2V5c0-1.1-.9-2-2-2zm0 16H5V5h14v14z"/>
          <path d="M7 7h10v2H7zm0 4h10v2H7zm0 4h7v2H7z"/>
        </svg>
        <p>暂无文章</p>
        <p>请先添加公众号订阅</p>
        <router-link to="/channels" class="btn btn-primary" style="margin-top: 16px;">
          添加公众号
        </router-link>
      </div>

      <div v-else class="article-list">
        <div 
          v-for="article in articles" 
          :key="article.id"
          class="article-item"
          @click="goToArticle(article.id)"
        >
          <div class="article-header">
            <div class="article-avatar">{{ getFirstChar(article.biz_name) }}</div>
            <div>
              <div class="article-meta">{{ article.biz_name }}</div>
            </div>
          </div>
          <h3 class="article-title">{{ article.title }}</h3>
          <p class="article-desc">{{ article.desc }}</p>
          <div class="article-time">{{ formatDate(article.created) }}</div>
        </div>
      </div>

      <div v-if="articles.length > 0" class="pagination">
        <button @click="prevPage" :disabled="page <= 1">上一页</button>
        <span>{{ page }} / {{ totalPages }}</span>
        <button @click="nextPage" :disabled="page >= totalPages">下一页</button>
      </div>
    </main>

    <footer class="footer">
      WeChatOArss v1.0.0 - 微信公众号RSS服务
    </footer>
  </div>
</template>

<script>
import axios from 'axios'
import dayjs from 'dayjs'

const API_BASE = ''

export default {
  name: 'Home',
  data() {
    return {
      articles: [],
      loading: true,
      page: 1,
      size: 20,
      total: 0,
      filter: 'all',
      searchKeyword: '',
      token: localStorage.getItem('token') || ''
    }
  },
  computed: {
    totalPages() {
      return Math.ceil(this.total / this.size) || 1
    }
  },
  mounted() {
    this.fetchArticles()
  },
  methods: {
    async fetchArticles() {
      this.loading = true
      try {
        let after = ''
        if (this.filter === 'today') {
          after = dayjs().format('YYYYMMDD')
        } else if (this.filter === 'yesterday') {
          after = dayjs().subtract(1, 'day').format('YYYYMMDD')
        } else if (this.filter === 'week') {
          after = dayjs().subtract(7, 'day').format('YYYYMMDD')
        }

        const params = {
          page: this.page,
          size: this.size,
          content: 0,
          k: this.token
        }
        if (after) params.after = after
        
        const res = await axios.get(API_BASE + '/api/query', { params })
        this.articles = res.data.data || []
        this.total = res.data.meta?.total || 0
      } catch (e) {
        console.error(e)
      } finally {
        this.loading = false
      }
    },
    setFilter(filter) {
      this.filter = filter
      this.page = 1
      this.fetchArticles()
    },
    handleSearch() {
      this.page = 1
      this.fetchArticles()
    },
    prevPage() {
      if (this.page > 1) {
        this.page--
        this.fetchArticles()
      }
    },
    nextPage() {
      if (this.page < this.totalPages) {
        this.page++
        this.fetchArticles()
      }
    },
    goToArticle(id) {
      this.$router.push('/article/' + id)
    },
    getFirstChar(name) {
      return name ? name.charAt(0).toUpperCase() : 'W'
    },
    formatDate(dateStr) {
      if (!dateStr) return ''
      return dayjs(dateStr).format('YYYY-MM-DD HH:mm')
    }
  }
}
</script>
