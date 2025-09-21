<template>
  <div class="theme-switcher">
    <!-- 主题切换按钮 -->
    <div class="relative">
      <button
        @click="toggleDropdown"
        class="theme-btn"
        :class="{ 'active': showDropdown }"
        :title="`当前主题: ${getThemeDisplayName(themeMode)}`"
      >
        <component :is="currentIcon" class="w-5 h-5" />
        <ChevronDown class="w-3 h-3 transition-transform" :class="{ 'rotate-180': showDropdown }" />
      </button>

      <!-- 主题选择下拉菜单 -->
      <transition
        enter-active-class="transition ease-out duration-200"
        enter-from-class="transform opacity-0 scale-95"
        enter-to-class="transform opacity-100 scale-100"
        leave-active-class="transition ease-in duration-150"
        leave-from-class="transform opacity-100 scale-100"
        leave-to-class="transform opacity-0 scale-95"
      >
        <div
          v-if="showDropdown"
          class="theme-dropdown"
          @click.stop
        >
          <div class="dropdown-header">
            <span class="text-sm font-medium text-secondary">主题设置</span>
          </div>
          
          <div class="dropdown-options">
            <button
              v-for="mode in themeOptions"
              :key="mode"
              @click="selectTheme(mode)"
              class="theme-option"
              :class="{ 'active': themeMode === mode }"
            >
              <component :is="getThemeIcon(mode)" class="w-4 h-4" />
              <span>{{ getThemeDisplayName(mode) }}</span>
              <div v-if="themeMode === mode" class="active-indicator">
                <Check class="w-3 h-3" />
              </div>
            </button>
          </div>
          
          <!-- 当前系统偏好提示 -->
          <div v-if="themeMode === 'auto'" class="dropdown-footer">
            <div class="system-preference">
              <component :is="systemPrefersDark ? 'Moon' : 'Sun'" class="w-3 h-3" />
              <span class="text-xs text-muted">
                系统偏好: {{ systemPrefersDark ? '深色' : '浅色' }}
              </span>
            </div>
          </div>
        </div>
      </transition>
    </div>

    <!-- 点击外部关闭下拉菜单 -->
    <div
      v-if="showDropdown"
      class="dropdown-overlay"
      @click="closeDropdown"
    ></div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { ChevronDown, Check, Sun, Moon, Monitor } from 'lucide-vue-next'
import { useTheme, type ThemeMode } from '../../composables/useTheme'

// 使用主题管理
const { 
  themeMode, 
  currentTheme, 
  systemPrefersDark, 
  setTheme, 
  getThemeDisplayName, 
  getThemeIcon 
} = useTheme()

// 下拉菜单状态
const showDropdown = ref(false)

// 主题选项配置
const themeOptions: ThemeMode[] = ['auto', 'light', 'dark']

// 图标映射
const iconMap = {
  Monitor,
  Sun,
  Moon
}

// 当前显示的图标
const currentIcon = computed(() => {
  const iconName = getThemeIcon(themeMode.value)
  return iconMap[iconName as keyof typeof iconMap]
})

// 切换下拉菜单
const toggleDropdown = () => {
  showDropdown.value = !showDropdown.value
}

// 关闭下拉菜单
const closeDropdown = () => {
  showDropdown.value = false
}

// 选择主题
const selectTheme = (mode: ThemeMode) => {
  setTheme(mode)
  closeDropdown()
}
</script>

<style scoped>
/* 主题切换器样式 */
.theme-switcher {
  position: relative;
  display: inline-block;
}

/* 主题切换按钮 */
.theme-btn {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0.5rem 0.75rem;
  border-radius: 0.5rem;
  transition: all 0.2s ease;
  background-color: var(--bg-secondary);
  border: 1px solid var(--border-color);
  color: var(--text-primary);
}

.theme-btn:hover {
  background-color: var(--bg-tertiary);
  border-color: var(--accent-primary);
}

.theme-btn:focus {
  outline: none;
  box-shadow: 0 0 0 2px rgba(59, 130, 246, 0.2);
}

.theme-btn.active {
  background-color: var(--bg-tertiary);
  border-color: var(--accent-primary);
}

/* 下拉菜单容器 */
.theme-dropdown {
  position: absolute;
  right: 0;
  margin-top: 0.5rem;
  width: 12rem;
  border-radius: 0.75rem;
  box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05);
  z-index: 50;
  background-color: var(--bg-secondary);
  border: 1px solid var(--border-color);
  backdrop-filter: blur(8px);
}

/* 下拉菜单头部 */
.dropdown-header {
  padding: 0.75rem 1rem;
  border-bottom: 1px solid var(--border-color);
  color: var(--text-secondary);
  font-size: 0.875rem;
  font-weight: 500;
}

/* 下拉菜单选项容器 */
.dropdown-options {
  padding: 0.5rem 0;
}

/* 主题选项 */
.theme-option {
  width: 100%;
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.625rem 1rem;
  text-align: left;
  transition: background-color 0.2s ease;
  color: var(--text-primary);
  background: none;
  border: none;
}

.theme-option:hover {
  background-color: var(--bg-tertiary);
}

.theme-option.active {
  background-color: rgba(59, 130, 246, 0.2);
  color: var(--accent-primary);
}

.theme-option.active .active-indicator {
  margin-left: auto;
  color: var(--accent-primary);
}

/* 下拉菜单底部 */
.dropdown-footer {
  padding: 0.75rem 1rem;
  border-top: 1px solid var(--border-color);
}

/* 系统偏好提示 */
.system-preference {
  display: flex;
  align-items: center;
  gap: 0.5rem;
  color: var(--text-muted);
  font-size: 0.75rem;
}

/* 遮罩层 */
.dropdown-overlay {
  position: fixed;
  inset: 0;
  z-index: 40;
}
</style>