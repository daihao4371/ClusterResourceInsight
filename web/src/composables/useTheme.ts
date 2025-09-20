import { ref, computed, onMounted, watch } from 'vue'

// 主题类型定义
export type ThemeMode = 'auto' | 'light' | 'dark'

// 主题状态管理
const themeMode = ref<ThemeMode>('auto')
const systemPrefersDark = ref(false)

/**
 * 主题管理 composable
 * 支持自动检测系统偏好、手动切换主题、持久化用户选择
 */
export function useTheme() {
  // 计算当前实际使用的主题
  const currentTheme = computed(() => {
    if (themeMode.value === 'auto') {
      return systemPrefersDark.value ? 'dark' : 'light'
    }
    return themeMode.value
  })

  // 检测系统主题偏好
  const detectSystemTheme = () => {
    if (typeof window !== 'undefined') {
      const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)')
      systemPrefersDark.value = mediaQuery.matches
      
      // 监听系统主题变化
      mediaQuery.addEventListener('change', (e) => {
        systemPrefersDark.value = e.matches
      })
    }
  }

  // 应用主题到DOM
  const applyTheme = (theme: ThemeMode) => {
    if (typeof document !== 'undefined') {
      document.documentElement.setAttribute('data-theme', theme)
    }
  }

  // 从localStorage读取用户偏好
  const loadThemeFromStorage = () => {
    if (typeof localStorage !== 'undefined') {
      const stored = localStorage.getItem('theme-mode') as ThemeMode
      if (stored && ['auto', 'light', 'dark'].includes(stored)) {
        themeMode.value = stored
      }
    }
  }

  // 保存主题偏好到localStorage
  const saveThemeToStorage = (theme: ThemeMode) => {
    if (typeof localStorage !== 'undefined') {
      localStorage.setItem('theme-mode', theme)
    }
  }

  // 切换主题
  const setTheme = (theme: ThemeMode) => {
    themeMode.value = theme
    saveThemeToStorage(theme)
  }

  // 循环切换主题
  const toggleTheme = () => {
    const modes: ThemeMode[] = ['auto', 'light', 'dark']
    const currentIndex = modes.indexOf(themeMode.value)
    const nextIndex = (currentIndex + 1) % modes.length
    setTheme(modes[nextIndex])
  }

  // 获取主题显示名称
  const getThemeDisplayName = (theme: ThemeMode) => {
    const names = {
      auto: '跟随系统',
      light: '浅色模式', 
      dark: '深色模式'
    }
    return names[theme]
  }

  // 获取主题图标
  const getThemeIcon = (theme: ThemeMode) => {
    const icons = {
      auto: 'Monitor',
      light: 'Sun',
      dark: 'Moon'
    }
    return icons[theme]
  }

  // 初始化主题系统
  const initTheme = () => {
    detectSystemTheme()
    loadThemeFromStorage()
  }

  // 监听主题变化，应用到DOM
  watch([themeMode, systemPrefersDark], () => {
    applyTheme(themeMode.value)
  }, { immediate: true })

  // 组件挂载时初始化
  onMounted(() => {
    initTheme()
  })

  return {
    // 状态
    themeMode: computed(() => themeMode.value),
    currentTheme,
    systemPrefersDark: computed(() => systemPrefersDark.value),
    
    // 方法
    setTheme,
    toggleTheme,
    getThemeDisplayName,
    getThemeIcon,
    
    // 工具方法
    initTheme
  }
}