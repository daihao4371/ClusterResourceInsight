import { createRouter, createWebHistory } from 'vue-router'
import Home from '@/views/Home.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      name: 'home',
      component: Home
    },
    {
      path: '/clusters',
      name: 'clusters',
      component: () => import('@/views/Clusters.vue')
    },
    {
      path: '/analysis',
      name: 'analysis',
      component: () => import('@/views/Analysis.vue')
    },
    {
      path: '/pods',
      name: 'pods',
      component: () => import('@/views/Pods.vue')
    },
    {
      path: '/schedule',
      name: 'schedule',
      component: () => import('@/views/Schedule.vue')
    },
    {
      path: '/history',
      name: 'history',
      component: () => import('@/views/History.vue')
    }
  ]
})

export default router