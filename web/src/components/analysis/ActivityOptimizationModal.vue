<template>
  <div v-if="show" class="fixed inset-0 flex items-center justify-center z-50" style="background: rgba(0, 0, 0, 0.5);" @click="$emit('close')">
    <div class="rounded-lg p-6 w-full max-w-md m-4 shadow-2xl" style="background: var(--bg-secondary); border: 1px solid var(--border-color);" @click.stop>
      <div class="flex items-center justify-between mb-6">
        <h3 class="text-lg font-semibold" style="color: var(--text-primary);">活动优化配置</h3>
        <button @click="$emit('close')" class="transition-colors hover:opacity-75" style="color: var(--text-secondary);">
          <X class="w-5 h-5" />
        </button>
      </div>
      
      <div v-if="config" class="space-y-4">
        <!-- 去重配置 -->
        <div>
          <label class="block text-sm font-medium mb-2" style="color: var(--text-secondary);">去重时间窗口（分钟）</label>
          <input 
            v-model.number="localConfig.deduplication_window"
            type="number"
            min="10"
            max="120"
            class="w-full input-field"
            placeholder="30"
          />
          <div class="text-xs mt-1" style="color: var(--text-muted);">在此时间窗口内的重复活动将被去重</div>
        </div>
        
        <!-- 最大重复次数 -->
        <div>
          <label class="block text-sm font-medium mb-2" style="color: var(--text-secondary);">最大重复次数</label>
          <input 
            v-model.number="localConfig.max_duplicate_count"
            type="number"
            min="1"
            max="10"
            class="w-full input-field"
            placeholder="3"
          />
          <div class="text-xs mt-1" style="color: var(--text-muted);">超过此数量的重复活动将被删除</div>
        </div>
        
        <!-- 聚合阈值 -->
        <div>
          <label class="block text-sm font-medium mb-2" style="color: var(--text-secondary);">聚合阈值</label>
          <input 
            v-model.number="localConfig.aggregation_threshold"
            type="number"
            min="3"
            max="20"
            class="w-full input-field"
            placeholder="5"
          />
          <div class="text-xs mt-1" style="color: var(--text-muted);">达到此数量的相似活动将被聚合</div>
        </div>
        
        <!-- 数据保留天数 -->
        <div>
          <label class="block text-sm font-medium mb-2" style="color: var(--text-secondary);">数据保留天数</label>
          <input 
            v-model.number="localConfig.cleanup_retention_days"
            type="number"
            min="1"
            max="30"
            class="w-full input-field"
            placeholder="7"
          />
          <div class="text-xs mt-1" style="color: var(--text-muted);">超过此天数的活动数据将被清理</div>
        </div>
        
        <!-- 开关配置 -->
        <div class="space-y-3">
          <div class="flex items-center justify-between">
            <span class="text-sm" style="color: var(--text-secondary);">启用聚合</span>
            <button 
              @click="localConfig.enable_aggregation = !localConfig.enable_aggregation"
              class="w-12 h-6 rounded-full transition-colors"
              :style="{ background: localConfig.enable_aggregation ? 'var(--accent-primary)' : 'var(--bg-tertiary)' }"
            >
              <div 
                class="w-5 h-5 rounded-full transition-transform"
                :class="localConfig.enable_aggregation ? 'translate-x-6' : 'translate-x-1'"
                style="background: var(--text-primary);"
              ></div>
            </button>
          </div>
          
          <div class="flex items-center justify-between">
            <span class="text-sm" style="color: var(--text-secondary);">启用噪音过滤</span>
            <button 
              @click="localConfig.noise_filter_enabled = !localConfig.noise_filter_enabled"
              class="w-12 h-6 rounded-full transition-colors"
              :style="{ background: localConfig.noise_filter_enabled ? 'var(--accent-primary)' : 'var(--bg-tertiary)' }"
            >
              <div 
                class="w-5 h-5 rounded-full transition-transform"
                :class="localConfig.noise_filter_enabled ? 'translate-x-6' : 'translate-x-1'"
                style="background: var(--text-primary);"
              ></div>
            </button>
          </div>
          
          <div class="flex items-center justify-between">
            <span class="text-sm" style="color: var(--text-secondary);">自动清理</span>
            <button 
              @click="localConfig.auto_cleanup_enabled = !localConfig.auto_cleanup_enabled"
              class="w-12 h-6 rounded-full transition-colors"
              :style="{ background: localConfig.auto_cleanup_enabled ? 'var(--accent-primary)' : 'var(--bg-tertiary)' }"
            >
              <div 
                class="w-5 h-5 rounded-full transition-transform"
                :class="localConfig.auto_cleanup_enabled ? 'translate-x-6' : 'translate-x-1'"
                style="background: var(--text-primary);"
              ></div>
            </button>
          </div>
        </div>
      </div>
      
      <div class="flex space-x-3 mt-6">
        <button 
          @click="handleSave"
          :disabled="loading"
          class="flex-1 btn-primary"
        >
          <Save :class="['w-4 h-4 mr-2', { 'animate-spin': loading }]" />
          {{ loading ? '保存中...' : '保存配置' }}
        </button>
        <button 
          @click="$emit('close')"
          class="flex-1 btn-secondary"
        >
          取消
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { X, Save } from 'lucide-vue-next'

// 定义优化配置接口类型
interface OptimizationConfig {
  deduplication_window: number
  max_duplicate_count: number
  aggregation_threshold: number
  cleanup_retention_days: number
  enable_aggregation: boolean
  noise_filter_enabled: boolean
  auto_cleanup_enabled: boolean
}

// 定义props接口
interface Props {
  show: boolean
  config?: OptimizationConfig | null
  loading?: boolean
}

// 定义事件接口
interface Emits {
  close: []
  save: [config: OptimizationConfig]
}

// 接收props和定义emits
const props = defineProps<Props>()
const emit = defineEmits<Emits>()

// 本地配置副本，避免直接修改props
const localConfig = ref<OptimizationConfig>({
  deduplication_window: 30,
  max_duplicate_count: 3,
  aggregation_threshold: 5,
  cleanup_retention_days: 7,
  enable_aggregation: true,
  noise_filter_enabled: true,
  auto_cleanup_enabled: true
})

// 监听config变化，同步到本地副本
watch(() => props.config, (newConfig) => {
  if (newConfig) {
    localConfig.value = { ...newConfig }
  }
}, { immediate: true })

// 处理保存操作
const handleSave = () => {
  emit('save', localConfig.value)
}
</script>