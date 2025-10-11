import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/login',
      name: 'Login',
      component: () => import('../views/Login.vue'),
      meta: { title: '登录', requiresAuth: false }
    },
    {
      path: '/',
      name: 'Dashboard',
      component: () => import('../views/Dashboard.vue'),
      meta: { title: '总览', requiresAuth: true }
    },
    {
      path: '/clusters',
      name: 'Clusters',
      component: () => import('../views/Clusters.vue'),
      meta: { title: '集群管理', requiresAuth: true }
    },
    {
      path: '/pods',
      name: 'Pods',
      component: () => import('../views/Pods.vue'),
      meta: { title: 'Pod监控', requiresAuth: true }
    },
    {
      path: '/analysis',
      name: 'Analysis',
      component: () => import('../views/Analysis.vue'),
      meta: { title: '资源分析', requiresAuth: true }
    },
    {
      path: '/history',
      name: 'History',
      component: () => import('../views/History.vue'),
      meta: { title: '历史数据', requiresAuth: true }
    },
    {
      path: '/schedule',
      name: 'Schedule',
      component: () => import('../views/Schedule.vue'),
      meta: { title: '调度管理', requiresAuth: true }
    },
    {
      path: '/alerts',
      name: 'Alerts',
      component: () => import('../views/Alerts.vue'),
      meta: { title: '系统告警', requiresAuth: true }
    },
    {
      path: '/users',
      name: 'Users',
      component: () => import('../views/Users.vue'),
      meta: { title: '用户管理', requiresAuth: true, requiresAdmin: true }
    },
    {
      path: '/roles',
      name: 'Roles',
      component: () => import('../views/Roles.vue'),
      meta: { title: '角色管理', requiresAuth: true, requiresAdmin: true }
    },
    {
      path: '/apis',
      name: 'Apis',
      component: () => import('../views/Apis.vue'),
      meta: { title: 'API管理', requiresAuth: true, requiresAdmin: true }
    },
    {
      path: '/audit',
      name: 'AuditLogs',
      component: () => import('../views/AuditLogs.vue'),
      meta: { title: '审计日志', requiresAuth: true, requiresAdmin: true }
    }
  ]
})

// 路由守卫
router.beforeEach((to, from, next) => {
  const authStore = useAuthStore()

  // 设置页面标题
  document.title = `${to.meta.title} - ClusterResourceInsight`

  // 不需要认证的页面直接放行
  if (to.meta.requiresAuth === false) {
    // 如果已登录且访问登录页,重定向到首页
    if (to.path === '/login' && authStore.isLoggedIn) {
      next('/')
      return
    }
    next()
    return
  }

  // 需要认证的页面检查登录状态
  if (!authStore.isLoggedIn) {
    next({
      path: '/login',
      query: { redirect: to.fullPath } // 保存目标路径,登录后跳转
    })
    return
  }

  // 检查管理员权限
  if (to.meta.requiresAdmin && !authStore.isAdmin) {
    alert('权限不足,需要管理员权限')
    next(false) // 阻止导航
    return
  }

  // 通过所有检查,允许访问
  next()
})

export default router