import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      name: 'Dashboard',
      component: () => import('../views/Dashboard.vue'),
      meta: { title: '总览' }
    },
    {
      path: '/clusters',
      name: 'Clusters',
      component: () => import('../views/Clusters.vue'),
      meta: { title: '集群管理' }
    },
    {
      path: '/pods',
      name: 'Pods',
      component: () => import('../views/Pods.vue'),
      meta: { title: 'Pod监控' }
    },
    {
      path: '/analysis',
      name: 'Analysis',
      component: () => import('../views/Analysis.vue'),
      meta: { title: '资源分析' }
    },
    {
      path: '/history',
      name: 'History',
      component: () => import('../views/History.vue'),
      meta: { title: '历史数据' }
    },
    {
      path: '/schedule',
      name: 'Schedule',
      component: () => import('../views/Schedule.vue'),
      meta: { title: '调度管理' }
    },
    {
      path: '/alerts',
      name: 'Alerts',
      component: () => import('../views/Alerts.vue'),
      meta: { title: '系统告警' }
    }
  ]
})

// 路由守卫
router.beforeEach((to, from, next) => {
  // 设置页面标题
  document.title = `${to.meta.title} - ClusterResourceInsight`
  next()
})

export default router