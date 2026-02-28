
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
        <n-form label-placement="left" label-width="120px">
           <n-form-item label="抓取时间">
             <div class="time-list">
               <div v-for="(_time, index) in config.cron_times" :key="index" class="time-item">
                 <span class="index-label">{{ index + 1 }}</span>
                 <n-time-picker 
                    v-model:formatted-value="config.cron_times[index]" 
                    value-format="HH:mm" 
                    format="HH:mm" 
                    placeholder="选择时间" 
                    :actions="['confirm']"
                 />
                 <n-button circle type="error" size="small" @click="removeTime(index)" v-if="config.cron_times.length > 1">
                   -
                 </n-button>
               </div>
               <n-button dashed type="primary" @click="addTime" class="add-time-btn">
                 + 添加时间
               </n-button>
             </div>
           </n-form-item>
           
           <n-form-item label="RSS 最大条目数">
             <n-input-number v-model:value="config.rss_max_items" :min="1" />
             <span class="tip">单次最大数量</span>
           </n-form-item>
           
           <n-form-item label="数据保留天数">
             <n-input-number v-model:value="config.retention_days" :min="1" />
           </n-form-item>
           
           <div class="form-actions">
             <n-button type="primary" @click="saveConfig" :loading="saving">保存</n-button>
           </div>
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
import { ref, onMounted, h, reactive } from 'vue'
import { NTabs, NTabPane, NDataTable, NButton, NTag, NForm, NFormItem, NInputNumber, NModal, NTimePicker, useMessage } from 'naive-ui'
import axios from 'axios'

const accounts = ref([])
const loading = ref(false)
const showQrModal = ref(false)
const qrCodeUrl = ref('')
const message = useMessage()

// Config State
const config = reactive({
  cron_times: ['08:00'],
  rss_max_items: 20,
  retention_days: 60
})
const saving = ref(false)

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

const fetchConfig = async () => {
  try {
    const res = await axios.get(`${API_BASE}/system/config`)
    if (res.data) {
        config.cron_times = res.data.cron_times || ['08:00']
        config.rss_max_items = res.data.rss_max_items || 20
        config.retention_days = res.data.retention_days || 60
    }
  } catch (e) {
    message.error('加载配置失败')
  }
}

const saveConfig = async () => {
  saving.value = true
  try {
    await axios.post(`${API_BASE}/system/config`, config)
    message.success('保存成功')
  } catch (e) {
    message.error('保存失败')
  } finally {
    saving.value = false
  }
}

const addTime = () => {
  config.cron_times.push('12:00')
}

const removeTime = (index: number) => {
  config.cron_times.splice(index, 1)
}

const startLogin = async () => {
    showQrModal.value = true
    try {
        const res = await axios.get(`${API_BASE}/auth/qrcode`)
        qrCodeUrl.value = "https://api.qrserver.com/v1/create-qr-code/?size=200x200&data=" + encodeURIComponent(res.data.qrcode_url)
    } catch (e) {
        message.error('获取二维码失败')
    }
}

onMounted(() => {
  fetchAccounts()
  fetchConfig()
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
.time-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.time-item {
  display: flex;
  align-items: center;
  gap: 10px;
}
.index-label {
  font-weight: bold;
  width: 20px;
  text-align: right;
  color: #666;
}
.tip {
    margin-left: 10px;
    color: #999;
    font-size: 12px;
}
.form-actions {
    margin-top: 20px;
    padding-left: 120px;
}
</style>
