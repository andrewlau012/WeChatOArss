<template>
  <div class="about-page">
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
      <div class="card" style="text-align: center; padding: 40px;">
        <div style="font-size: 48px; margin-bottom: 20px;">
          🖥️
        </div>
        
        <h1 style="font-size: 28px; margin-bottom: 10px;">WeChatOArss</h1>
        <p style="color: #666; margin-bottom: 30px;">微信公众号RSS服务</p>
        
        <p style="font-size: 18px; margin-bottom: 40px;">
          版本: {{ version }}
        </p>
        
        <p style="color: #666; max-width: 500px; margin: 0 auto 40px;">
          一个长期稳定可用的微信公众号RSS服务，支持本地部署、私有云部署，让你可以使用RSS阅读器订阅喜欢的公众号。
        </p>

        <div style="border-top: 1px solid #eee; padding-top: 30px;">
          <h3 style="margin-bottom: 20px;">功能特点</h3>
          <div style="text-align: left; max-width: 400px; margin: 0 auto;">
            <p style="margin-bottom: 10px;">✓ 微信扫码登录</p>
            <p style="margin-bottom: 10px;">✓ 公众号搜索订阅</p>
            <p style="margin-bottom: 10px;">✓ RSS标准输出</p>
            <p style="margin-bottom: 10px;">✓ Web阅读界面</p>
            <p style="margin-bottom: 10px;">✓ Docker一键部署</p>
            <p style="margin-bottom: 10px;">✓ 数据本地存储</p>
          </div>
        </div>

        <div style="border-top: 1px solid #eee; padding-top: 30px; margin-top: 30px;">
          <p style="color: #999; font-size: 12px;">
            © 2024 WeChatOArss. All rights reserved.
          </p>
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
  name: 'About',
  data() {
    return {
      version: 'v1.0.0'
    }
  },
  mounted() {
    this.fetchVersion()
  },
  methods: {
    async fetchVersion() {
      try {
        const res = await axios.get('/version')
        this.version = res.data.version || 'v1.0.0'
      } catch (e) {
        console.error(e)
      }
    }
  }
}
</script>
