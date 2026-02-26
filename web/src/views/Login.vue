<template>
  <div class="login-page">
    <div class="login-card">
      <h1>WeChatOArss</h1>
      <p>微信公众号RSS服务</p>
      
      <div class="qrcode" v-if="qrcode">
        <img :src="qrcode" alt="QR Code" />
      </div>
      <div class="qrcode" v-else>
        <div class="spinner"></div>
      </div>
      
      <p v-if="tips">{{ tips }}</p>
      <p v-else>请用微信扫描二维码登录</p>
      
      <div v-if="showCodeInput" style="margin-top: 20px;">
        <input 
          type="text" 
          v-model="code" 
          placeholder="请输入验证码"
          style="padding: 10px; width: 200px; text-align: center;"
          @keyup.enter="submitCode"
        />
        <button class="btn btn-primary" @click="submitCode" style="margin-left: 10px;">
          确认
        </button>
      </div>
      
      <div style="margin-top: 30px; font-size: 12px; color: #999;">
        <p>登录后即可添加公众号订阅</p>
        <p>您的登录信息将安全存储在本地</p>
      </div>
    </div>
  </div>
</template>

<script>
import axios from 'axios'

export default {
  name: 'Login',
  data() {
    return {
      qrcode: '',
      tips: '',
      code: '',
      showCodeInput: false,
      uuid: '',
      token: localStorage.getItem('token') || ''
    }
  },
  mounted() {
    this.getQRCode()
    this.checkStatus()
  },
  methods: {
    async getQRCode() {
      try {
        const res = await axios.get('/api/login/new?k=' + this.token)
        if (res.data.err === '') {
          this.qrcode = res.data.data.qrcode
          this.uuid = res.data.data.uuid
        }
      } catch (e) {
        console.error(e)
      }
    },
    async checkStatus() {
      if (!this.uuid) return
      
      try {
        const res = await axios.get('/api/login/status?uuid=' + this.uuid)
        if (res.data.code === 0) {
          // Login success
          localStorage.setItem('token', this.token)
          this.$router.push('/')
        } else if (res.data.code === 200) {
          this.showCodeInput = true
        }
      } catch (e) {
        console.error(e)
      }
      
      setTimeout(() => this.checkStatus(), 2000)
    },
    async submitCode() {
      if (!this.code) return
      
      try {
        await axios.post('/api/login/code', {
          code: this.code
        }, { params: { k: this.token }})
        
        this.$router.push('/')
      } catch (e) {
        console.error(e)
      }
    }
  }
}
</script>
