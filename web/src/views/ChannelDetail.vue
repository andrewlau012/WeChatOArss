<template>
  <div class="channel-detail-page">
    <header class="header">
      <div class="header-content">
        <router-link to="/" class="logo">
          <svg viewBox="0 0 24 24" fill="currentColor">
            <path d="M8.691 2.188C3.891 2.188 0 5.476 0 9.53c0 2.212 1.17 4.203 3.002 5.55a.59.59 0 0 1 .213.665l-.39 1.48c-.019.07-.048.141-.048.213 0 .163.13.295.29.295a.326.326 0 0 0 .167-.054l1.903-1.114a.864.864 0 0 1 .717-.098 10.16 10.16 0 0 0 2.837.403c.276 0 .543-.027.811-.05-.857-2.578.157-4.972 1.932-6.446 1.703-1.415 3.882-1.98 5.853-1.838-.576-3.583-4.196-6.348-8.596-6.348z"/>
          </svg>
          WeChatOArss
        </router-link>
        
        <nav class="nav">
          <router-link to="/">首页</router-link>
          <router-link to="/channels">公众号</router-link>
          <router-link to="/settings">设置</router-link>
          <router-link to="/about">关于</router-link>
        </nav>
      </div>
    </header>

    <main class="container">
      <div class="back-link" @click="$router.back()">
        ← 返回
      </div>

      <div class="card" v-if="channel">
        <div class="channel-header">
          <div class="channel-avatar" style="width: 64px; height: 64px; font-size: 24px;">
            {{ getFirstChar(channel.name) }}
          </div>
          <div class="channel-info">
            <h2>{{ channel.name }}</h2>
            <p>{{ channel.description || '暂无简介' }}</p>
            <div class="channel-stats" style="margin-top: 8px;">
              <span>文章: {{ channel.articleCount }}</span>
              <span>更新: {{ formatDate(channel.lastUpdate) }}</span>
            </div>
          </div>
        </div>
      </div>

      <div v-if="loading" class="loading">
        <div class="spinner"></div>
      </div>

      <div v-else-if="articles.length === 0" class="empty">
        <p>暂无文章</p>
      </div>

      <div v-else class="article-list">
        <div 
          v-for="article in articles" 
          :key="article.id"
          class="article-item"
          @click="goToArticle(article.id)"
        >
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

export default {
  name: 'ChannelDetail',
  data() {
    return {
      channel: null,
      articles: [],
      loading: true,
      page: 1,
      size: 20,
      total: 0,
      token: localStorage.getItem('token') || ''
    }
  },
  computed: {
    totalPages() {
      return Math.ceil(this.total / this.size) || 1
    }
  },
  mounted() {
    this.fetchChannel()
    this.fetchArticles()
  },
  methods: {
    async fetchChannel() {
      try {
        const res = await axios.get('/api/list', { 
          params: { k: this.token, size: 1000 }
        })
        const channels = res.data.data || []
        this.channel = channels.find(c => c.biz_id === this.$route.params.id)
      } catch (e) {
        console.error(e)
      }
    },
    async fetchArticles() {
      this.loading = true
      try {
        const res = await axios.get('/api/query', { 
          params: { 
            bid: this.$route.params.id,
            page: this.page,
            size: this.size,
            content: 0,
            k: this.token
          }
        })
        this.articles = res.data.data || []
        this.total = res.data.meta?.total || 0
      } catch (e) {
        console.error(e)
      } finally {
        this.loading = false
      }
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
