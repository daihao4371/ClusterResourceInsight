import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import router from './router'
import './style.css'

// 创建应用实例
const app = createApp(App)

// 使用插件
app.use(createPinia())
app.use(router)

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