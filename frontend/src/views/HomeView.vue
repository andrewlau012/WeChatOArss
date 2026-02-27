
<template>
  <div class="home-layout">
    <div class="sidebar">
      <div class="sidebar-header">
        <n-input placeholder="搜索公众号" v-model:value="searchKeyword" clearable />
        <n-button type="primary" circle @click="showAddModal = true" class="add-btn">
          +
        </n-button>
      </div>
      <n-list hoverable clickable class="feed-list">
        <n-list-item v-for="feed in filteredFeeds" :key="feed.biz_id" @click="selectFeed(feed)" :class="{ active: currentFeed?.biz_id === feed.biz_id }">
          <n-thing :title="feed.name" :description="feed.status === 'hidden' ? '已隐藏' : '正常'">
             <template #avatar>
                <n-avatar :src="feed.head_img" fallback-src="https://res.wx.qq.com/a/wx_fed/assets/res/NTI4MWUi5.ico" />
             </template>
          </n-thing>
        </n-list-item>
      </n-list>
      <div class="sidebar-footer">
        <n-button block @click="selectFeed(null)" :type="!currentFeed ? 'primary' : 'default'">
          全部文章
        </n-button>
        <router-link to="/settings" style="margin-top: 10px; display: block; text-align: center;">设置</router-link>
      </div>
    </div>
    
    <div class="main-content">
      <div class="content-header">
        <h2>{{ currentFeed ? currentFeed.name : '全部文章' }}</h2>
        <div class="actions">
           <n-button size="small" @click="copyRssLink">RSS 链接</n-button>
           <n-button size="small" @click="refreshFeed">刷新</n-button>
        </div>
      </div>
      
      <n-scrollbar style="flex: 1">
        <div class="article-list">
          <n-card v-for="article in articles" :key="article.article_id" :title="article.title" hoverable @click="openArticle(article.url)" style="margin-bottom: 12px; cursor: pointer;">
             <template #header-extra>
                {{ formatDate(article.pub_date) }}
             </template>
             {{ article.digest }}
             <template #footer>
                {{ article.author || '未知作者' }}
             </template>
          </n-card>
        </div>
      </n-scrollbar>
    </div>

    <!-- Add Feed Modal -->
    <n-modal v-model:show="showAddModal" preset="card" title="添加公众号" style="width: 600px">
      <n-tabs type="segment">
        <n-tab-pane name="link" tab="链接添加">
          <n-input v-model:value="addLinkUrl" placeholder="粘贴任意一篇公众号文章链接 (https://mp.weixin.qq.com/s/...)" type="textarea" :rows="3" />
          <n-button type="primary" block style="margin-top: 12px" @click="addByLink" :loading="adding">解析并添加</n-button>
        </n-tab-pane>
        <n-tab-pane name="search" tab="搜索添加">
          <n-input-group>
            <n-input v-model:value="searchAddKeyword" placeholder="输入公众号名称" />
            <n-button type="primary" @click="searchFeeds" :loading="searching">搜索</n-button>
          </n-input-group>
          <n-list style="margin-top: 12px">
            <n-list-item v-for="item in searchResults" :key="item.biz_id">
              <n-thing :title="item.name">
                <template #avatar><n-avatar :src="item.head_img" /></template>
                <template #action>
                  <n-button size="small" @click="addBySearch(item)">添加</n-button>
                </template>
              </n-thing>
            </n-list-item>
          </n-list>
        </n-tab-pane>
      </n-tabs>
    </n-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { NList, NListItem, NThing, NAvatar, NInput, NButton, NScrollbar, NCard, NModal, NTabs, NTabPane, NInputGroup, useMessage } from 'naive-ui'
import axios from 'axios'

const message = useMessage()

// State
const feeds = ref<any[]>([])
const articles = ref<any[]>([])
const currentFeed = ref<any>(null)
const searchKeyword = ref('')
const showAddModal = ref(false)

// Add Modal State
const addLinkUrl = ref('')
const adding = ref(false)
const searchAddKeyword = ref('')
const searching = ref(false)
const searchResults = ref<any[]>([])

const API_BASE = 'http://localhost:8000/api/v1' // Configurable

// Computed
const filteredFeeds = computed(() => {
  if (!searchKeyword.value) return feeds.value
  return feeds.value.filter(f => f.name.includes(searchKeyword.value))
})

// Methods
const fetchFeeds = async () => {
  try {
    const res = await axios.get(`${API_BASE}/feeds/list`)
    feeds.value = res.data
  } catch (e) {
    message.error('加载公众号列表失败')
  }
}

const selectFeed = async (feed: any) => {
  currentFeed.value = feed
  // Fetch articles for this feed (mock for now, need endpoint)
  // Real endpoint: /api/v1/articles?biz_id=...
  // Since we don't have articles endpoint in spec yet, using RSS endpoint or mock
  // Let's assume we have a way to list articles JSON
  // Temporary: Just clear list
  articles.value = [] 
}

const copyRssLink = () => {
  const bizId = currentFeed.value ? currentFeed.value.biz_id : 'all'
  const url = `${API_BASE}/rss/${bizId}.xml`
  navigator.clipboard.writeText(url)
  message.success('RSS 链接已复制')
}

const refreshFeed = async () => {
  const target = currentFeed.value ? currentFeed.value.biz_id : 'all'
  await axios.post(`${API_BASE}/feeds/refresh`, { target_id: target })
  message.success('已触发刷新任务')
}

const openArticle = (url: string) => {
  window.open(url, '_blank')
}

const formatDate = (dateStr: string) => {
  return new Date(dateStr).toLocaleString()
}

// Add Logic
const addByLink = async () => {
  if (!addLinkUrl.value) return
  adding.value = true
  try {
    await axios.post(`${API_BASE}/feeds/add`, { type: 'link_parse', value: addLinkUrl.value })
    message.success('添加成功')
    addLinkUrl.value = ''
    showAddModal.value = false
    fetchFeeds()
  } catch (e) {
    message.error('添加失败，请检查链接')
  } finally {
    adding.value = false
  }
}

const searchFeeds = async () => {
    if (!searchAddKeyword.value) return
    searching.value = true
    try {
        const res = await axios.get(`${API_BASE}/feeds/search`, { params: { keyword: searchAddKeyword.value } })
        searchResults.value = res.data
    } catch (e) {
        message.error('搜索失败')
    } finally {
        searching.value = false
    }
}

const addBySearch = async (item: any) => {
    // Implement logic
    console.log(item)
    message.info('该功能暂未对接后端')
}

onMounted(() => {
  fetchFeeds()
})
</script>

<style scoped>
.home-layout {
  display: flex;
  height: 100vh;
}
.sidebar {
  width: 300px;
  border-right: 1px solid #eee;
  display: flex;
  flex-direction: column;
  background: #f9f9f9;
}
.sidebar-header {
  padding: 12px;
  display: flex;
  gap: 8px;
}
.add-btn {
  flex-shrink: 0;
}
.feed-list {
  flex: 1;
  overflow-y: auto;
}
.sidebar-footer {
  padding: 12px;
  border-top: 1px solid #eee;
}
.main-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  background: #fff;
}
.content-header {
  padding: 16px;
  border-bottom: 1px solid #eee;
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.article-list {
  padding: 20px;
  max-width: 800px;
  margin: 0 auto;
}
.active {
    background-color: #e6f7ff;
}
</style>
