import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import './style.css'
import axios from 'axios'

// 创建应用实例
const app = createApp(App)
const pinia = createPinia()

// 使用插件
app.use(pinia)
app.use(router)

// 配置Axios拦截器
import { useAuthStore } from './stores/auth'

// 请求拦截器 - 自动添加Authorization头
axios.interceptors.request.use(
  (config) => {
    // 直接从localStorage读取token,避免Pinia store初始化时机问题
    const token = localStorage.getItem('token')

    // 如果有token且不是登录接口,添加Authorization头
    if (token && !config.url?.includes('/auth/login')) {
      config.headers.Authorization = `Bearer ${token}`
    }

    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// 响应拦截器 - 处理401/403错误
axios.interceptors.response.use(
  (response) => {
    return response
  },
  (error) => {
    const authStore = useAuthStore(pinia)

    if (error.response) {
      const status = error.response.status

      // 401: 未登录或Token失效
      if (status === 401) {
        authStore.logout()
        router.push('/login')
        return Promise.reject(new Error('未登录或登录已过期,请重新登录'))
      }

      // 403: 权限不足
      if (status === 403) {
        alert('权限不足,无法访问')
        return Promise.reject(new Error('权限不足'))
      }
    }

    return Promise.reject(error)
  }
)

// 挂载应用
app.mount('#app')

// 移除加载动画
setTimeout(() => {
  const loader = document.querySelector('.loader')
  if (loader) {
    loader.style.transition = 'opacity 0.3s ease-out'
    loader.style.opacity = '0'
    setTimeout(() => loader.remove(), 300)
  }
}, 500)