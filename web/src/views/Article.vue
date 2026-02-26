<template>
  <div class="article-page">
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

      <div v-if="loading" class="loading">
        <div class="spinner"></div>
      </div>

      <div v-else-if="article" class="article-detail">
        <div class="meta">
          <span>{{ article.biz_name }}</span>
          <span>{{ formatDate(article.created) }}</span>
        </div>
        
        <h1>{{ article.title }}</h1>
        
        <div class="article-content" v-html="article.content"></div>
        
        <div style="margin-top: 30px; padding-top: 20px; border-top: 1px solid #eee;">
          <a :href="article.link" target="_blank" class="btn btn-secondary">
            在微信中查看原文
          </a>
        </div>
      </div>

      <div v-else class="empty">
        <p>文章不存在</p>
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
  name: 'Article',
  data() {
    return {
      article: null,
      loading: true,
      token: localStorage.getItem('token') || ''
    }
  },
  mounted() {
    this.fetchArticle()
  },
  methods: {
    async fetchArticle() {
      this.loading = true
      try {
        const res = await axios.get('/api/article/' + this.$route.params.id, {
          params: { k: this.token }
        })
        this.article = res.data.data
      } catch (e) {
        console.error(e)
      } finally {
        this.loading = false
      }
    },
    formatDate(dateStr) {
      if (!dateStr) return ''
      return dayjs(dateStr).format('YYYY-MM-DD HH:mm')
    }
  }
}
</script>
