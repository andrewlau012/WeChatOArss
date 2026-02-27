
<template>
  <div class="settings-page">
    <div class="header">
      <router-link to="/">← 返回首页</router-link>
      <h2>系统设置</h2>
    </div>
    
    <n-tabs type="line" animated>
      <n-tab-pane name="accounts" tab="账号管理">
        <div class="toolbar">
           <n-button type="primary" @click="startLogin">添加微信读书账号 (扫码)</n-button>
        </div>
        <n-data-table :columns="columns" :data="accounts" :loading="loading" />
      </n-tab-pane>
      <n-tab-pane name="config" tab="抓取配置">
        <n-form>
           <n-form-item label="Cron 表达式">
             <n-input value="15 8 * * *, 30 12 * * *" />
           </n-form-item>
           <n-form-item label="RSS 最大条目数">
             <n-input-number :value="20" />
           </n-form-item>
           <n-form-item label="数据保留天数">
             <n-input-number :value="60" />
           </n-form-item>
           <n-button type="primary">保存</n-button>
        </n-form>
      </n-tab-pane>
    </n-tabs>

    <!-- QR Code Modal -->
    <n-modal v-model:show="showQrModal" preset="dialog" title="微信扫码登录">
      <div style="text-align: center;">
         <div v-if="qrCodeUrl">
            <img :src="qrCodeUrl" style="width: 200px; height: 200px;" />
            <p>请使用微信扫描上方二维码</p>
         </div>
         <div v-else>正在获取二维码...</div>
      </div>
    </n-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, h } from 'vue'
import { NTabs, NTabPane, NDataTable, NButton, NTag, NForm, NFormItem, NInput, NInputNumber, NModal, useMessage } from 'naive-ui'
import axios from 'axios'

const accounts = ref([])
const loading = ref(false)
const showQrModal = ref(false)
const qrCodeUrl = ref('')
const message = useMessage()

const API_BASE = 'http://localhost:8000/api/v1'

const columns = [
  { title: 'VID', key: 'vid' },
  { title: '昵称', key: 'name' },
  { 
    title: '状态', 
    key: 'status',
    render(row: any) {
      return h(NTag, { type: row.status === 'active' ? 'success' : 'error' }, { default: () => row.status })
    }
  },
  { title: '最后活跃', key: 'last_active' }
]

const fetchAccounts = async () => {
  loading.value = true
  try {
    const res = await axios.get(`${API_BASE}/auth/accounts`)
    accounts.value = res.data
  } catch (e) {
    message.error('加载账号失败')
  } finally {
    loading.value = false
  }
}

const startLogin = async () => {
    showQrModal.value = true
    try {
        const res = await axios.get(`${API_BASE}/auth/qrcode`)
        // Assuming the backend returns a QR image URL or we render it
        // For now, let's assume it returns a URL to a QR image
        // In the mock, it returns a login URL string, we might need a qrcode generator lib in frontend if it's just a string
        // But let's assume backend gives an image url for simplicity or we use a placeholder
        qrCodeUrl.value = "https://api.qrserver.com/v1/create-qr-code/?size=200x200&data=" + encodeURIComponent(res.data.qrcode_url)
        
        // Start polling status...
    } catch (e) {
        message.error('获取二维码失败')
    }
}

onMounted(() => {
  fetchAccounts()
})
</script>

<style scoped>
.settings-page {
  padding: 20px;
  max-width: 1000px;
  margin: 0 auto;
}
.header {
  margin-bottom: 20px;
}
.toolbar {
  margin-bottom: 16px;
}
</style>
