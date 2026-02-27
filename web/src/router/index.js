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
    component: Home,
    meta: { requiresAuth: true }
  },
  {
    path: '/login',
    name: 'Login',
    component: Login
  },
  {
    path: '/channels',
    name: 'Channels',
    component: Channels,
    meta: { requiresAuth: true }
  },
  {
    path: '/channel/:id',
    name: 'ChannelDetail',
    component: ChannelDetail,
    meta: { requiresAuth: true }
  },
  {
    path: '/article/:id',
    name: 'Article',
    component: Article,
    meta: { requiresAuth: true }
  },
  {
    path: '/settings',
    name: 'Settings',
    component: Settings,
    meta: { requiresAuth: true }
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

// Navigation guard
router.beforeEach((to, from, next) => {
  const token = localStorage.getItem('token')
  
  // If route requires auth and no token, redirect to login
  if (to.meta.requiresAuth && !token) {
    next({ name: 'Login' })
  } 
  // If already logged in and trying to access login, redirect to home
  else if (to.name === 'Login' && token) {
    next({ name: 'Home' })
  }
  else {
    next()
  }
})

export default router
