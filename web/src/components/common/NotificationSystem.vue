<template>
  <div class="fixed top-20 right-6 z-50 space-y-3 max-w-sm">
    <TransitionGroup name="notification" tag="div">
      <div
        v-for="notification in systemStore.notifications"
        :key="notification.id"
        class="notification-card"
        :class="notificationClasses(notification.type)"
        @click="systemStore.removeNotification(notification.id)"
      >
        <!-- 图标 -->
        <div class="flex-shrink-0">
          <component 
            :is="getIcon(notification.type)" 
            class="w-5 h-5"
            :class="iconClasses(notification.type)"
          />
        </div>
        
        <!-- 内容 -->
        <div class="flex-1 min-w-0">
          <h4 class="text-sm font-medium text-white">
            {{ notification.title }}
          </h4>
          <p class="text-sm text-gray-300 mt-1">
            {{ notification.message }}
          </p>
          <p class="text-xs text-gray-400 mt-2">
            {{ formatDistanceToNow(notification.timestamp) }}
          </p>
        </div>
        
        <!-- 关闭按钮 -->
        <button
          @click.stop="systemStore.removeNotification(notification.id)"
          class="flex-shrink-0 ml-3 p-1 rounded-lg hover:bg-white/10 transition-colors"
        >
          <X class="w-4 h-4 text-gray-400 hover:text-white" />
        </button>
      </div>
    </TransitionGroup>
  </div>
</template>

<script setup lang="ts">
import { CheckCircle, AlertTriangle, XCircle, Info, X } from 'lucide-vue-next'
import { useSystemStore } from '../../stores/system'
import { formatDistanceToNow } from '../../utils/date'
import type { Notification } from '../../types'

const systemStore = useSystemStore()

const getIcon = (type: Notification['type']) => {
  const icons = {
    success: CheckCircle,
    warning: AlertTriangle,
    error: XCircle,
    info: Info
  }
  return icons[type]
}

const notificationClasses = (type: Notification['type']) => {
  const classes = {
    success: 'border-success-500/50 bg-success-500/10',
    warning: 'border-warning-500/50 bg-warning-500/10',
    error: 'border-danger-500/50 bg-danger-500/10',
    info: 'border-primary-500/50 bg-primary-500/10'
  }
  return classes[type]
}

const iconClasses = (type: Notification['type']) => {
  const classes = {
    success: 'text-success-400',
    warning: 'text-warning-400',
    error: 'text-danger-400',
    info: 'text-primary-400'
  }
  return classes[type]
}
</script>

<style scoped>
.notification-card {
  @apply flex items-start space-x-3 p-4 rounded-xl border backdrop-blur-sm;
  @apply cursor-pointer hover:bg-white/5 transition-all duration-300;
  box-shadow: 
    0 10px 15px -3px rgba(0, 0, 0, 0.3),
    0 4px 6px -2px rgba(0, 0, 0, 0.2);
}

.notification-enter-active,
.notification-leave-active {
  transition: all 0.3s ease-out;
}

.notification-enter-from {
  opacity: 0;
  transform: translateX(100%);
}

.notification-leave-to {
  opacity: 0;
  transform: translateX(100%);
}

.notification-move {
  transition: transform 0.3s ease-out;
}
</style>