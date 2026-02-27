<template>
  <div class="login-page">
    <div class="login-card">
      <h1>WeChatOArss</h1>
      <p>微信公众号RSS服务</p>
      
      <!-- Token Input -->
      <div v-if="!token" class="form-group" style="margin-bottom: 20px;">
        <input 
          type="text" 
          v-model="inputToken" 
          placeholder="请输入访问Token"
          style="padding: 10px; width: 200px; text-align: center;"
          @keyup.enter="saveToken"
        />
        <button class="btn btn-primary" @click="saveToken" style="margin-left: 10px;">
          确认
        </button>
      </div>
      
      <div class="qrcode" v-if="token && qrcode">
        <img :src="qrcode" alt="QR Code" />
      </div>
      <div class="qrcode" v-else-if="token">
        <div class="spinner"></div>
      </div>
      
      <p v-if="tips">{{ tips }}</p>
      <p v-else-if="token">请用微信扫描二维码登录</p>
      <p v-else>请先输入Token继续</p>

      <!-- Demo: Simulate successful scan -->
      <button 
        v-if="token && qrcode" 
        class="btn btn-primary" 
        style="margin-top: 20px;"
        @click="simulateLogin"
      >
        已扫码登录（演示）
      </button>
      
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
import QRCode from 'qrcode'

export default {
  name: 'Login',
  data() {
    return {
      qrcode: '',
      tips: '',
      code: '',
      showCodeInput: false,
      uuid: '',
      token: localStorage.getItem('token') || '',
      inputToken: ''
    }
  },
  mounted() {
    if (this.token) {
      this.getQRCode()
    }
  },
  methods: {
    saveToken() {
      if (this.inputToken) {
        this.token = this.inputToken
        localStorage.setItem('token', this.token)
        this.getQRCode()
      }
    },
    async getQRCode() {
      try {
        const res = await axios.get('/login/new?k=' + this.token)
        if (res.data.err === '') {
          this.uuid = res.data.data.uuid
          // Generate QR code from UUID (wechat://xxx format for demo)
          const wechatUrl = 'https://login.weixin.qq.com/' + this.uuid
          try {
            this.qrcode = await QRCode.toDataURL(wechatUrl, {
              width: 200,
              margin: 2,
              color: {
                dark: '#000000',
                light: '#ffffff'
              }
            })
          } catch (e) {
            console.error('QR generation failed:', e)
          }
          this.tips = res.data.data.tips || ''
          this.startPolling()
        } else {
          this.tips = res.data.err
        }
      } catch (e) {
        console.error(e)
        this.tips = '连接失败，请检查Token是否正确'
      }
    },
    startPolling() {
      // Poll for login status
      this.checkStatus()
    },
    async checkStatus() {
      if (!this.uuid) return
      
      try {
        const res = await axios.get('/login/status?uuid=' + this.uuid)
        if (res.data.code === 0) {
          // Login success
          localStorage.setItem('wechat_logged_in', 'true')
          this.$router.push('/')
        } else if (res.data.code === 200) {
          this.showCodeInput = true
        }
      } catch (e) {
        console.error(e)
      }
      
      // Continue polling
      setTimeout(() => this.checkStatus(), 3000)
    },
    async submitCode() {
      if (!this.code) return
      
      try {
        await axios.post('/login/code', {
          code: this.code
        }, { params: { k: this.token }})
        
        this.$router.push('/')
      } catch (e) {
        console.error(e)
      }
    },
    simulateLogin() {
      // Demo: simulate successful login
      localStorage.setItem('wechat_logged_in', 'true')
      this.$router.push('/')
    }
  }
}
</script>
