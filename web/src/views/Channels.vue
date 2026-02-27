<template>
  <div class="channels-page">
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
      <div class="card">
        <div class="card-header">
          <h2 class="card-title">公众号管理</h2>
          <button class="btn btn-primary" @click="showAddModal = true">
            + 添加公众号
          </button>
        </div>

        <div class="search-box" style="margin-bottom: 20px; max-width: 300px;">
          <input 
            type="text" 
            v-model="searchKeyword" 
            placeholder="搜索公众号..."
            @keyup.enter="fetchChannels"
          />
        </div>

        <div v-if="loading" class="loading">
          <div class="spinner"></div>
        </div>

        <div v-else-if="channels.length === 0" class="empty">
          <p>暂无订阅的公众号</p>
          <button class="btn btn-primary" @click="showAddModal = true" style="margin-top: 16px;">
            添加第一个公众号
          </button>
        </div>

        <div v-else class="channel-list">
          <div v-for="channel in channels" :key="channel.id" class="channel-item">
            <div class="channel-header">
              <div class="channel-avatar">{{ getFirstChar(channel.name) }}</div>
              <div class="channel-info">
                <h3>{{ channel.name }}</h3>
                <p>{{ channel.description || '暂无简介' }}</p>
              </div>
            </div>
            <div class="channel-stats">
              <span>文章: {{ channel.articleCount }}</span>
              <span>更新: {{ formatDate(channel.lastUpdate) }}</span>
              <span>状态: {{ channel.status === 'active' ? '正常' : '已暂停' }}</span>
            </div>
            <div class="channel-actions">
              <button class="btn btn-secondary btn-sm" @click="copyRSS(channel.link)">
                复制RSS
              </button>
              <button class="btn btn-secondary btn-sm" @click="togglePause(channel)">
                {{ channel.status === 'active' ? '暂停' : '恢复' }}
              </button>
              <button class="btn btn-danger btn-sm" @click="deleteChannel(channel)">
                删除
              </button>
            </div>
          </div>
        </div>
      </div>
    </main>

    <!-- Add Modal -->
    <div v-if="showAddModal" class="modal-overlay" @click.self="showAddModal = false">
      <div class="modal">
        <div class="modal-header">
          <h2>添加公众号</h2>
          <button class="modal-close" @click="showAddModal = false">&times;</button>
        </div>
        
        <div class="form-group">
          <label>通过文章链接添加</label>
          <input 
            type="text" 
            v-model="articleUrl" 
            placeholder="粘贴微信公众号文章链接..."
          />
        </div>
        
        <div style="text-align: center; margin: 20px 0;">
          <span>或者</span>
        </div>
        
        <div class="form-group">
          <label>通过名称搜索</label>
          <input 
            type="text" 
            v-model="searchName" 
            placeholder="输入公众号名称搜索..."
            @keyup.enter="searchChannel"
          />
        </div>
        
        <div v-if="searchResults.length > 0" style="margin-top: 16px;">
          <div 
            v-for="result in searchResults" 
            :key="result.biz_id"
            class="channel-item"
            style="cursor: pointer;"
            @click="addChannel(result.biz_id)"
          >
            <div class="channel-header">
              <div class="channel-avatar">{{ getFirstChar(result.name) }}</div>
              <div class="channel-info">
                <h3>{{ result.name }}</h3>
                <p>{{ result.description }}</p>
              </div>
            </div>
          </div>
        </div>
        
        <div style="margin-top: 20px; text-align: right;">
          <button class="btn btn-secondary" @click="showAddModal = false">取消</button>
          <button class="btn btn-primary" @click="addByUrl">确认添加</button>
        </div>
      </div>
    </div>

    <footer class="footer">
      WeChatOArss v1.0.0 - 微信公众号RSS服务
    </footer>
  </div>
</template>

<script>
import axios from 'axios'
import dayjs from 'dayjs'

export default {
  name: 'Channels',
  data() {
    return {
      channels: [],
      loading: true,
      searchKeyword: '',
      showAddModal: false,
      articleUrl: '',
      searchName: '',
      searchResults: [],
      token: localStorage.getItem('token') || ''
    }
  },
  mounted() {
    this.fetchChannels()
  },
  methods: {
    async fetchChannels() {
      this.loading = true
      try {
        const params = { k: this.token }
        if (this.searchKeyword) params.name = this.searchKeyword
        
        const res = await axios.get('/api/list', { params })
        this.channels = res.data.data || []
      } catch (e) {
        console.error(e)
      } finally {
        this.loading = false
      }
    },
    async addByUrl() {
      if (!this.articleUrl) return
      
      try {
        await axios.get('/api/addurl', { 
          params: { url: this.articleUrl, k: this.token }
        })
        this.showAddModal = false
        this.articleUrl = ''
        this.fetchChannels()
      } catch (e) {
        alert('添加失败: ' + e.response?.data?.err || e.message)
      }
    },
    async searchChannel() {
      if (!this.searchName) return
      
      // Mock search results for demo
      this.searchResults = [
        { biz_id: 'Mzkz' + Math.random().toString(36).substr(2, 8), name: this.searchName + '公众号', description: '搜索结果' }
      ]
    },
    async addChannel(bizId) {
      try {
        await axios.get('/api/add/' + bizId, { params: { k: this.token }})
        this.showAddModal = false
        this.searchResults = []
        this.fetchChannels()
      } catch (e) {
        alert('添加失败')
      }
    },
    async togglePause(channel) {
      try {
        const status = channel.status === 'active' ? 'true' : 'false'
        await axios.get('/api/pause/' + channel.biz_id, { 
          params: { status: status, k: this.token }
        })
        channel.status = channel.status === 'active' ? 'paused' : 'active'
      } catch (e) {
        alert('操作失败')
      }
    },
    async deleteChannel(channel) {
      if (!confirm('确定要删除这个公众号吗？')) return
      
      try {
        await axios.delete('/api/del/' + channel.biz_id, { 
          params: { k: this.token }
        })
        this.fetchChannels()
      } catch (e) {
        alert('删除失败')
      }
    },
    copyRSS(link) {
      navigator.clipboard.writeText(link)
      alert('RSS链接已复制')
    },
    getFirstChar(name) {
      return name ? name.charAt(0).toUpperCase() : 'W'
    },
    formatDate(dateStr) {
      if (!dateStr) return '未知'
      return dayjs(dateStr).format('MM-DD HH:mm')
    }
  }
}
</script>
