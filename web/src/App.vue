<template>
  <div id="app">
    <el-container>
      <!-- 顶部导航栏 -->
      <el-header>
        <div class="header-content">
          <div class="logo">
            <el-icon><Monitor /></el-icon>
            <span>ClusterResourceInsight</span>
          </div>
          <el-menu
            mode="horizontal"
            :default-active="$route.path"
            router
            class="main-menu"
          >
            <el-menu-item index="/">
              <el-icon><HomeFilled /></el-icon>
              <span>首页</span>
            </el-menu-item>
            <el-menu-item index="/clusters">
              <el-icon><Setting /></el-icon>
              <span>集群管理</span>
            </el-menu-item>
            <el-menu-item index="/analysis">
              <el-icon><DataAnalysis /></el-icon>
              <span>资源分析</span>
            </el-menu-item>
            <el-menu-item index="/pods">
              <el-icon><Box /></el-icon>
              <span>Pod管理</span>
            </el-menu-item>
            <el-menu-item index="/schedule">
              <el-icon><Clock /></el-icon>
              <span>调度任务</span>
            </el-menu-item>
            <el-menu-item index="/history">
              <el-icon><Histogram /></el-icon>
              <span>历史数据</span>
            </el-menu-item>
          </el-menu>
        </div>
      </el-header>

      <!-- 主内容区 -->
      <el-main>
        <router-view />
      </el-main>
    </el-container>
  </div>
</template>

<script setup lang="ts">
import { onMounted } from 'vue'
import { 
  Monitor, 
  HomeFilled, 
  Setting, 
  DataAnalysis, 
  Box, 
  Clock, 
  Histogram 
} from '@element-plus/icons-vue'

onMounted(() => {
  console.log('ClusterResourceInsight 应用已启动')
})
</script>

<style scoped>
.el-header {
  background-color: #545c64;
  color: #fff;
  padding: 0;
  height: 60px;
}

.header-content {
  display: flex;
  align-items: center;
  height: 100%;
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 20px;
}

.logo {
  display: flex;
  align-items: center;
  color: #fff;
  font-size: 20px;
  font-weight: bold;
  margin-right: 40px;
}

.logo .el-icon {
  margin-right: 8px;
  font-size: 24px;
}

.main-menu {
  flex: 1;
  background-color: transparent;
  border-bottom: none;
}

.main-menu .el-menu-item {
  color: #fff;
  border-bottom: 2px solid transparent;
}

.main-menu .el-menu-item:hover,
.main-menu .el-menu-item.is-active {
  background-color: rgba(255, 255, 255, 0.1);
  border-bottom-color: #409eff;
  color: #fff;
}

.el-main {
  padding: 20px;
  background-color: #f5f5f5;
  min-height: calc(100vh - 60px);
}
</style>