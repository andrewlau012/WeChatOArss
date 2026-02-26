import { createRouter, createWebHistory } from 'vue-router'
import Home from '@/views/Home.vue'
import Channels from '@/views/Channels.vue'
import ChannelDetail from '@/views/ChannelDetail.vue'
import Article from '@/views/Article.vue'
import Settings from '@/views/Settings.vue'
import Login from '@/views/Login.vue'
import About from '@/views/About.vue'

const routes = [
  {
    path: '/',
    name: 'Home',
    component: Home
  },
  {
    path: '/login',
    name: 'Login',
    component: Login
  },
  {
    path: '/channels',
    name: 'Channels',
    component: Channels
  },
  {
    path: '/channel/:id',
    name: 'ChannelDetail',
    component: ChannelDetail
  },
  {
    path: '/article/:id',
    name: 'Article',
    component: Article
  },
  {
    path: '/settings',
    name: 'Settings',
    component: Settings
  },
  {
    path: '/about',
    name: 'About',
    component: About
  }
]

const router = createRouter({
  history: createWebHistory(),
  routes
})

export default router
