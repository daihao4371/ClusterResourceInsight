<template>
  <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
    <!-- CPU资源分布 -->
    <div class="glass-card p-6">
      <h3 class="text-lg font-semibold mb-4">CPU资源分布</h3>
      <div class="h-64 flex items-center justify-center text-gray-500">
        <BarChart3 class="w-12 h-12 mr-3" />
        图表组件开发中...
      </div>
    </div>
    
    <!-- 内存资源分布 -->
    <div class="glass-card p-6">
      <h3 class="text-lg font-semibold mb-4">内存资源分布</h3>
      <div class="h-64 flex items-center justify-center text-gray-500">
        <PieChart class="w-12 h-12 mr-3" />
        图表组件开发中...
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { BarChart3, PieChart } from 'lucide-vue-next'
</script>