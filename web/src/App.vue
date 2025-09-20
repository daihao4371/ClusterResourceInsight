<template>
  <div class="min-h-screen bg-dark-950">
    <!-- 导航栏 -->
    <NavBar />
    
    <!-- 主内容区域 -->
    <main class="flex">
      <!-- 侧边栏 -->
      <SideBar />
      
      <!-- 页面内容 -->
      <div class="flex-1 ml-64">
        <div class="p-6">
          <RouterView v-slot="{ Component }">
            <Transition name="page" mode="out-in">
              <component :is="Component" />
            </Transition>
          </RouterView>
        </div>
      </div>
    </main>
    
    <!-- 通知系统 -->
    <NotificationSystem />
  </div>
</template>

<script setup lang="ts">
import { RouterView } from 'vue-router'
import NavBar from './components/layout/NavBar.vue'
import SideBar from './components/layout/SideBar.vue'
import NotificationSystem from './components/common/NotificationSystem.vue'
</script>

<style scoped>
/* 页面切换动画 */
.page-enter-active,
.page-leave-active {
  transition: all 0.3s ease-out;
}

.page-enter-from {
  opacity: 0;
  transform: translateY(10px);
}

.page-leave-to {
  opacity: 0;
  transform: translateY(-10px);
}
</style>