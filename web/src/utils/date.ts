import dayjs from 'dayjs'
import relativeTime from 'dayjs/plugin/relativeTime'
import 'dayjs/locale/zh-cn'

// 配置 dayjs
dayjs.extend(relativeTime)
dayjs.locale('zh-cn')

/**
 * 格式化日期时间
 */
export const formatDateTime = (date: string | Date, format = 'YYYY-MM-DD HH:mm:ss') => {
  return dayjs(date).format(format)
}

/**
 * 格式化相对时间（如：2分钟前）
 */
export const formatDistanceToNow = (date: string | Date) => {
  return dayjs(date).fromNow()
}

/**
 * 格式化日期
 */
export const formatDate = (date: string | Date) => {
  return dayjs(date).format('YYYY-MM-DD')
}

/**
 * 格式化时间
 */
export const formatTime = (date: string | Date) => {
  return dayjs(date).format('HH:mm:ss')
}

/**
 * 获取时间段
 */
export const getTimeRange = (hours: number) => {
  const end = dayjs()
  const start = end.subtract(hours, 'hour')
  
  return {
    start: start.toISOString(),
    end: end.toISOString()
  }
}

/**
 * 检查日期是否为今天
 */
export const isToday = (date: string | Date) => {
  return dayjs(date).isSame(dayjs(), 'day')
}

/**
 * 检查日期是否为昨天
 */
export const isYesterday = (date: string | Date) => {
  return dayjs(date).isSame(dayjs().subtract(1, 'day'), 'day')
}