<template>
  <div class="settings-page">
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
        <h2 class="card-title" style="margin-bottom: 20px;">系统设置</h2>
        
        <div class="settings-section">
          <h2>抓取设置</h2>
          <div class="form-group">
            <label>定时抓取时间（每天）</label>
            <input 
              type="text" 
              v-model="config.schedulerTimes" 
              placeholder="07:00,12:00,20:00"
            />
            <p style="font-size: 12px; color: #999; margin-top: 4px;">
              多个时间点用逗号分隔
            </p>
          </div>
          
          <div class="form-group">
            <label>RSS最大文章数</label>
            <input 
              type="number" 
              v-model="config.maxItemCount" 
              min="1"
              max="100"
            />
          </div>
          
          <div class="form-group">
            <label>历史保留文章数</label>
            <input 
              type="number" 
              v-model="config.keepOldCount" 
              min="10"
              max="1000"
            />
          </div>
        </div>

        <div class="settings-section">
          <h2>安全设置</h2>
          <div class="form-group">
            <label>访问Token</label>
            <input 
              type="text" 
              v-model="config.token" 
              placeholder="请输入访问Token"
            />
            <p style="font-size: 12px; color: #999; margin-top: 4px;">
              用于API鉴权，请妥善保管
            </p>
          </div>
        </div>

        <div class="settings-section">
          <h2>RSS设置</h2>
          <div class="form-group">
            <label>
              <input type="checkbox" v-model="config.encFeedId" />
              加密RSS ID
            </label>
            <p style="font-size: 12px; color: #999; margin-top: 4px;">
              开启后RSS地址会被加密，更安全但不够直观
            </p>
          </div>
          
          <div class="form-group">
            <label>
              <input type="checkbox" v-model="config.static" />
              静态RSS文件
            </label>
            <p style="font-size: 12px; color: #999; margin-top: 4px;">
              开启后会生成静态RSS文件，提升性能
            </p>
          </div>
        </div>

        <div style="margin-top: 30px;">
          <button class="btn btn-primary" @click="saveConfig">
            保存配置
          </button>
          <button class="btn btn-secondary" @click="exportOPML" style="margin-left: 10px;">
            导出OPML
          </button>
        </div>
      </div>
    </main>

    <footer class="footer">
      WeChatOArss v1.0.0 - 微信公众号RSS服务
    </footer>
  </div>
</template>

<script>
import axios from 'axios'

export default {
  name: 'Settings',
  data() {
    return {
      config: {
        schedulerTimes: '07:00,12:00,20:00',
        maxItemCount: 20,
        keepOldCount: 50,
        token: '',
        encFeedId: false,
        static: false
      },
      token: localStorage.getItem('token') || ''
    }
  },
  mounted() {
    this.fetchConfig()
  },
  methods: {
    async fetchConfig() {
      try {
        const res = await axios.get('/api/config', {
          params: { k: this.token }
        })
        const data = res.data.data
        if (data) {
          this.config = {
            schedulerTimes: data.schedulerTimes ? data.schedulerTimes.join(',') : '07:00,12:00,20:00',
            maxItemCount: data.maxItemCount || 20,
            keepOldCount: data.keepOldCount || 50,
            token: data.token || '',
            encFeedId: data.encFeedId || false,
            static: data.static || false
          }
        }
      } catch (e) {
        console.error(e)
      }
    },
    async saveConfig() {
      try {
        const params = {
          ...this.config,
          schedulerTimes: this.config.schedulerTimes.split(',').map(t => t.trim()),
          k: this.token
        }
        
        await axios.post('/api/config', params)
        alert('配置保存成功')
      } catch (e) {
        alert('保存失败: ' + (e.response?.data?.err || e.message))
      }
    },
    async exportOPML() {
      try {
        const res = await axios.get('/opml', {
          params: { k: this.token },
          responseType: 'blob'
        })
        
        const url = window.URL.createObjectURL(new Blob([res.data]))
        const link = document.createElement('a')
        link.href = url
        link.setAttribute('download', 'wechatoarss.opml')
        document.body.appendChild(link)
        link.click()
        link.remove()
      } catch (e) {
        alert('导出失败')
      }
    }
  }
}
</script>
