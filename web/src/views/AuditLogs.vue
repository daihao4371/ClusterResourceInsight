<template>
  <div class="h-full flex flex-col bg-gray-50 dark:bg-gray-900">
    <!-- 页面头部 -->
    <div class="bg-white dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700 px-6 py-4">
      <div class="flex items-center justify-between">
        <h1 class="text-2xl font-bold text-gray-900 dark:text-white">审计日志</h1>
        <div class="flex gap-2">
          <button @click="openStatsDialog"
            class="px-4 py-2 bg-cyan-600 text-white rounded-lg hover:bg-cyan-700 transition-colors">
            <BarChart3 class="w-4 h-4 inline-block mr-1" />
            统计分析
          </button>
          <button @click="showCleanupDialog = true"
            class="px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700 transition-colors">
            <Trash2 class="w-4 h-4 inline-block mr-1" />
            清理日志
          </button>
          <button @click="handleExport" :disabled="exporting"
            class="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors disabled:opacity-50">
            <Download class="w-4 h-4 inline-block mr-1" />
            {{ exporting ? '导出中...' : '导出日志' }}
          </button>
        </div>
      </div>
    </div>

    <!-- 统计卡片 -->
    <div v-if="statistics" class="bg-white dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700 px-6 py-4">
      <div class="grid grid-cols-6 gap-4">
        <div class="bg-blue-50 dark:bg-blue-900/30 rounded-lg p-4 border border-blue-100 dark:border-blue-800">
          <div class="text-sm text-blue-600 dark:text-blue-400 font-medium">总记录数</div>
          <div class="text-2xl font-bold text-blue-700 dark:text-blue-300 mt-1">{{ statistics.total }}</div>
        </div>
        <div class="bg-green-50 dark:bg-green-900/30 rounded-lg p-4 border border-green-100 dark:border-green-800">
          <div class="text-sm text-green-600 dark:text-green-400 font-medium">成功请求</div>
          <div class="text-2xl font-bold text-green-700 dark:text-green-300 mt-1">{{ statistics.total - statistics.error_count }}</div>
        </div>
        <div class="bg-red-50 dark:bg-red-900/30 rounded-lg p-4 border border-red-100 dark:border-red-800">
          <div class="text-sm text-red-600 dark:text-red-400 font-medium">错误请求</div>
          <div class="text-2xl font-bold text-red-700 dark:text-red-300 mt-1">{{ statistics.error_count }}</div>
        </div>
        <div class="bg-yellow-50 dark:bg-yellow-900/30 rounded-lg p-4 border border-yellow-100 dark:border-yellow-800">
          <div class="text-sm text-yellow-600 dark:text-yellow-400 font-medium">慢请求</div>
          <div class="text-2xl font-bold text-yellow-700 dark:text-yellow-300 mt-1">{{ statistics.slow_request_count }}</div>
        </div>
        <div class="bg-indigo-50 dark:bg-indigo-900/30 rounded-lg p-4 border border-indigo-100 dark:border-indigo-800">
          <div class="text-sm text-indigo-600 dark:text-indigo-400 font-medium">平均耗时</div>
          <div class="text-2xl font-bold text-indigo-700 dark:text-indigo-300 mt-1">{{ Math.round(statistics.avg_duration) }}ms</div>
        </div>
        <div class="bg-blue-50 dark:bg-blue-900/30 rounded-lg p-4 border border-blue-100 dark:border-blue-800">
          <div class="text-sm text-blue-600 dark:text-blue-400 font-medium">活跃用户</div>
          <div class="text-2xl font-bold text-blue-700 dark:text-blue-300 mt-1">{{ statistics.by_user?.length || 0 }}</div>
        </div>
      </div>
    </div>

    <!-- 高级搜索栏 -->
    <div class="bg-white dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700 px-6 py-3">
      <div class="grid grid-cols-6 gap-3">
        <input v-model="filters.username" @input="handleSearch" type="text" placeholder="用户名"
          class="px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:outline-none focus:border-blue-500 dark:bg-gray-700 dark:text-white" />

        <select v-model="filters.operate_type" @change="handleSearch"
          class="px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:outline-none focus:border-blue-500 dark:bg-gray-700 dark:text-white">
          <option value="">所有操作</option>
          <option value="create">创建</option>
          <option value="update">更新</option>
          <option value="delete">删除</option>
          <option value="query">查询</option>
          <option value="login">登录</option>
          <option value="logout">登出</option>
        </select>

        <select v-model="filters.module" @change="handleSearch"
          class="px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:outline-none focus:border-blue-500 dark:bg-gray-700 dark:text-white">
          <option value="">所有模块</option>
          <option value="user">用户</option>
          <option value="role">角色</option>
          <option value="api">API</option>
          <option value="cluster">集群</option>
          <option value="pod">Pod</option>
          <option value="schedule">调度</option>
          <option value="alert">告警</option>
        </select>

        <select v-model="filters.status_code" @change="handleSearch"
          class="px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:outline-none focus:border-blue-500 dark:bg-gray-700 dark:text-white">
          <option value="">所有状态</option>
          <option value="200">成功(200)</option>
          <option value="400">请求错误(400)</option>
          <option value="401">未授权(401)</option>
          <option value="403">禁止访问(403)</option>
          <option value="404">未找到(404)</option>
          <option value="500">服务器错误(500)</option>
        </select>

        <input v-model="filters.start_time" @change="handleSearch" type="datetime-local"
          class="px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:outline-none focus:border-blue-500 dark:bg-gray-700 dark:text-white" />

        <input v-model="filters.end_time" @change="handleSearch" type="datetime-local"
          class="px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:outline-none focus:border-blue-500 dark:bg-gray-700 dark:text-white" />
      </div>
      <div class="mt-2 flex items-center gap-2">
        <label class="text-sm text-gray-600 dark:text-gray-400">耗时过滤:</label>
        <input v-model.number="filters.min_duration" @input="handleSearch" type="number"
          placeholder="最小(ms)" min="0"
          class="w-28 px-3 py-1 border border-gray-300 dark:border-gray-600 rounded-lg focus:outline-none focus:border-blue-500 dark:bg-gray-700 dark:text-white text-sm" />
        <span class="text-gray-400">-</span>
        <input v-model.number="filters.max_duration" @input="handleSearch" type="number"
          placeholder="最大(ms)" min="0"
          class="w-28 px-3 py-1 border border-gray-300 dark:border-gray-600 rounded-lg focus:outline-none focus:border-blue-500 dark:bg-gray-700 dark:text-white text-sm" />
        <button @click="resetFilters" class="ml-auto px-3 py-1 text-sm text-gray-600 dark:text-gray-400 hover:text-gray-900 dark:hover:text-gray-200 border border-gray-300 dark:border-gray-600 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-700">
          重置筛选
        </button>
      </div>
    </div>

    <!-- 日志列表 -->
    <div class="flex-1 overflow-auto p-6">
      <div v-if="loading" class="text-center py-12">
        <Loader2 class="w-8 h-8 animate-spin text-blue-600 mx-auto" />
        <p class="mt-2 text-gray-500 dark:text-gray-400">加载中...</p>
      </div>

      <div v-else-if="logs.length === 0" class="text-center py-12">
        <FileText class="w-16 h-16 text-gray-400 mx-auto" />
        <p class="mt-4 text-gray-500 dark:text-gray-400">暂无审计日志</p>
      </div>

      <div v-else class="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
        <table class="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
          <thead class="bg-gray-50 dark:bg-gray-900">
            <tr>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">时间</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">用户</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">操作类型</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">模块</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">IP地址</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">请求路径</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">状态</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">耗时</th>
              <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 dark:text-gray-400 uppercase">操作</th>
            </tr>
          </thead>
          <tbody class="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
            <tr v-for="log in logs" :key="log.id" class="hover:bg-gray-50 dark:hover:bg-gray-700">
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500 dark:text-gray-400">
                {{ formatDate(log.created_at) }}
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900 dark:text-white">
                {{ log.username }}
                <div class="text-xs text-gray-500 dark:text-gray-400">{{ log.real_name || '-' }}</div>
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <span :class="getOperateTypeClass(log.operate_type)" class="px-2 py-1 text-xs rounded-full">
                  {{ getOperateTypeName(log.operate_type) }}
                </span>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-700 dark:text-gray-300">
                <span class="px-2 py-1 bg-gray-100 dark:bg-gray-700 rounded text-xs">
                  {{ log.module || '-' }}
                </span>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-700 dark:text-gray-300 font-mono">
                {{ log.ip || '-' }}
              </td>
              <td class="px-6 py-4 text-sm font-mono text-gray-700 dark:text-gray-300">
                <div class="max-w-xs truncate" :title="log.path">
                  {{ log.method }} {{ log.path }}
                </div>
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <span :class="getStatusClass(log.response_code)" class="px-2 py-1 text-xs rounded">
                  {{ log.response_code }}
                </span>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm">
                <span :class="getDurationClass(log.duration)">
                  {{ log.duration }}ms
                </span>
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                <button @click="showDetail(log)" class="text-blue-600 dark:text-blue-400 hover:text-blue-900 dark:hover:text-blue-300">详情</button>
              </td>
            </tr>
          </tbody>
        </table>

        <!-- 分页 -->
        <div class="bg-gray-50 dark:bg-gray-900 px-6 py-3 flex items-center justify-between border-t border-gray-200 dark:border-gray-700">
          <div class="text-sm text-gray-500 dark:text-gray-400">
            共 {{ total }} 条记录，第 {{ currentPage }} / {{ totalPages }} 页
          </div>
          <div class="flex gap-2">
            <button @click="handlePageChange(currentPage - 1)" :disabled="currentPage === 1"
              class="px-3 py-1 border border-gray-300 dark:border-gray-600 rounded hover:bg-gray-100 dark:hover:bg-gray-700 disabled:opacity-50 disabled:cursor-not-allowed dark:text-gray-300">
              上一页
            </button>
            <button @click="handlePageChange(currentPage + 1)" :disabled="currentPage === totalPages"
              class="px-3 py-1 border border-gray-300 dark:border-gray-600 rounded hover:bg-gray-100 dark:hover:bg-gray-700 disabled:opacity-50 disabled:cursor-not-allowed dark:text-gray-300">
              下一页
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- 日志详情弹窗 -->
    <div v-if="showDetailDialog" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div class="bg-white dark:bg-gray-800 rounded-lg shadow-xl w-full max-w-4xl mx-4 max-h-[90vh] overflow-hidden flex flex-col">
        <div class="px-6 py-4 border-b border-gray-200 dark:border-gray-700">
          <h2 class="text-xl font-bold text-gray-900 dark:text-white">审计日志详情</h2>
        </div>
        <div v-if="selectedLog" class="px-6 py-4 flex-1 overflow-y-auto space-y-4">
          <!-- 基本信息 -->
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="text-sm font-medium text-gray-500 dark:text-gray-400">日志ID</label>
              <div class="mt-1 text-sm text-gray-900 dark:text-white">{{ selectedLog.id }}</div>
            </div>
            <div>
              <label class="text-sm font-medium text-gray-500 dark:text-gray-400">操作时间</label>
              <div class="mt-1 text-sm text-gray-900 dark:text-white">{{ formatDate(selectedLog.created_at) }}</div>
            </div>
            <div>
              <label class="text-sm font-medium text-gray-500 dark:text-gray-400">用户信息</label>
              <div class="mt-1 text-sm text-gray-900 dark:text-white">{{ selectedLog.username }} ({{ selectedLog.real_name || '-' }})</div>
            </div>
            <div>
              <label class="text-sm font-medium text-gray-500 dark:text-gray-400">IP地址</label>
              <div class="mt-1 text-sm text-gray-900 dark:text-white font-mono">{{ selectedLog.ip || '-' }}</div>
            </div>
            <div>
              <label class="text-sm font-medium text-gray-500 dark:text-gray-400">操作类型</label>
              <div class="mt-1">
                <span :class="getOperateTypeClass(selectedLog.operate_type)" class="px-2 py-1 text-xs rounded-full">
                  {{ getOperateTypeName(selectedLog.operate_type) }}
                </span>
              </div>
            </div>
            <div>
              <label class="text-sm font-medium text-gray-500 dark:text-gray-400">模块</label>
              <div class="mt-1 text-sm text-gray-900 dark:text-white">{{ selectedLog.module || '-' }}</div>
            </div>
          </div>

          <!-- 请求信息 -->
          <div>
            <label class="text-sm font-medium text-gray-500 dark:text-gray-400">请求路径</label>
            <div class="mt-1 text-sm text-gray-900 dark:text-white font-mono">
              {{ selectedLog.method }} {{ selectedLog.path }}
            </div>
          </div>

          <div>
            <label class="text-sm font-medium text-gray-500 dark:text-gray-400">请求体（已脱敏）</label>
            <pre class="mt-1 text-xs bg-gray-50 dark:bg-gray-900 text-gray-800 dark:text-gray-200 p-3 rounded border border-gray-200 dark:border-gray-700 overflow-x-auto">{{ formatJson(selectedLog.request_body) }}</pre>
          </div>

          <!-- 响应信息 -->
          <div class="grid grid-cols-2 gap-4">
            <div>
              <label class="text-sm font-medium text-gray-500 dark:text-gray-400">响应状态</label>
              <div class="mt-1">
                <span :class="getStatusClass(selectedLog.response_code)" class="px-2 py-1 text-xs rounded">
                  {{ selectedLog.response_code }}
                </span>
              </div>
            </div>
            <div>
              <label class="text-sm font-medium text-gray-500 dark:text-gray-400">执行耗时</label>
              <div class="mt-1">
                <span :class="getDurationClass(selectedLog.duration)" class="text-sm">
                  {{ selectedLog.duration }}ms
                </span>
              </div>
            </div>
          </div>

          <div v-if="selectedLog.error_msg">
            <label class="text-sm font-medium text-red-600 dark:text-red-400">错误信息</label>
            <div class="mt-1 text-sm text-red-700 dark:text-red-300 bg-red-50 dark:bg-red-900/30 p-3 rounded border border-red-200 dark:border-red-800">
              {{ selectedLog.error_msg }}
            </div>
          </div>

          <div v-if="selectedLog.user_agent">
            <label class="text-sm font-medium text-gray-500 dark:text-gray-400">User-Agent</label>
            <div class="mt-1 text-xs text-gray-700 dark:text-gray-300 font-mono bg-gray-50 dark:bg-gray-900 p-2 rounded">
              {{ selectedLog.user_agent }}
            </div>
          </div>
        </div>
        <div class="px-6 py-4 border-t border-gray-200 dark:border-gray-700 flex justify-end">
          <button @click="closeDetailDialog" class="px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-700 dark:text-gray-300">关闭</button>
        </div>
      </div>
    </div>

    <!-- 统计分析弹窗 -->
    <div v-if="showStatsDialog" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div class="bg-white dark:bg-gray-800 rounded-lg shadow-xl w-full max-w-6xl mx-4 max-h-[90vh] overflow-hidden flex flex-col">
        <div class="px-6 py-4 border-b border-gray-200 dark:border-gray-700 flex items-center justify-between">
          <h2 class="text-xl font-bold text-gray-900 dark:text-white">审计统计分析</h2>
          <button @click="showStatsDialog = false" class="text-gray-400 dark:text-gray-500 hover:text-gray-600 dark:hover:text-gray-300">
            <X class="w-5 h-5" />
          </button>
        </div>
        <div v-if="detailedStats" class="px-6 py-4 flex-1 overflow-y-auto space-y-6">
          <!-- 操作类型统计 -->
          <div>
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-3">操作类型分布</h3>
            <div class="grid grid-cols-3 gap-4">
              <div v-for="item in detailedStats.by_operate_type" :key="item.operate_type"
                class="border border-gray-200 dark:border-gray-700 rounded-lg p-4 bg-white dark:bg-gray-900">
                <div class="text-sm text-gray-600 dark:text-gray-400">{{ getOperateTypeName(item.operate_type) }}</div>
                <div class="text-2xl font-bold text-gray-900 dark:text-white mt-1">{{ item.count }}</div>
              </div>
            </div>
          </div>

          <!-- 模块统计 -->
          <div>
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-3">模块访问统计 (Top 10)</h3>
            <div class="grid grid-cols-5 gap-4">
              <div v-for="item in detailedStats.by_module" :key="item.module"
                class="border border-gray-200 dark:border-gray-700 rounded-lg p-4 bg-white dark:bg-gray-900">
                <div class="text-sm text-gray-600 dark:text-gray-400">{{ item.module || '未知' }}</div>
                <div class="text-2xl font-bold text-blue-600 dark:text-blue-400 mt-1">{{ item.count }}</div>
              </div>
            </div>
          </div>

          <!-- 用户活动统计 -->
          <div>
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-3">活跃用户统计 (Top 10)</h3>
            <div class="grid grid-cols-5 gap-4">
              <div v-for="item in detailedStats.by_user" :key="item.username"
                class="border border-gray-200 dark:border-gray-700 rounded-lg p-4 bg-white dark:bg-gray-900">
                <div class="text-sm text-gray-600 dark:text-gray-400 truncate" :title="item.username">{{ item.username }}</div>
                <div class="text-2xl font-bold text-cyan-600 dark:text-cyan-400 mt-1">{{ item.count }}</div>
              </div>
            </div>
          </div>

          <!-- 状态码统计 -->
          <div>
            <h3 class="text-lg font-semibold text-gray-900 dark:text-white mb-3">HTTP状态码分布</h3>
            <div class="grid grid-cols-4 gap-4">
              <div v-for="item in detailedStats.by_status" :key="item.response_code"
                class="border border-gray-200 dark:border-gray-700 rounded-lg p-4 bg-white dark:bg-gray-900">
                <div class="text-sm text-gray-600 dark:text-gray-400">状态码 {{ item.response_code }}</div>
                <div class="text-2xl font-bold mt-1"
                  :class="item.response_code >= 400 ? 'text-red-600 dark:text-red-400' : 'text-green-600 dark:text-green-400'">
                  {{ item.count }}
                </div>
              </div>
            </div>
          </div>
        </div>
        <div class="px-6 py-4 border-t border-gray-200 dark:border-gray-700 flex justify-end">
          <button @click="showStatsDialog = false" class="px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-700 dark:text-gray-300">关闭</button>
        </div>
      </div>
    </div>

    <!-- 清理日志弹窗 -->
    <div v-if="showCleanupDialog" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div class="bg-white dark:bg-gray-800 rounded-lg shadow-xl w-full max-w-md mx-4">
        <div class="px-6 py-4 border-b border-gray-200 dark:border-gray-700">
          <h2 class="text-xl font-bold text-gray-900 dark:text-white">清理旧日志</h2>
        </div>
        <div class="px-6 py-4">
          <p class="text-gray-600 dark:text-gray-400 mb-4">清理指定天数之前的审计日志，此操作不可恢复。</p>
          <div class="mb-4">
            <label class="block text-sm font-medium text-gray-700 dark:text-gray-300 mb-2">保留天数</label>
            <input v-model.number="cleanupDays" type="number" min="1"
              class="w-full px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg focus:outline-none focus:border-blue-500 dark:bg-gray-700 dark:text-white"
              placeholder="输入保留天数（默认90天）">
            <p class="text-sm text-gray-500 dark:text-gray-400 mt-2">将删除 {{ cleanupDays }} 天之前的所有审计日志</p>
          </div>
        </div>
        <div class="px-6 py-4 border-t border-gray-200 dark:border-gray-700 flex justify-end gap-2">
          <button @click="showCleanupDialog = false"
            class="px-4 py-2 border border-gray-300 dark:border-gray-600 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-700 dark:text-gray-300">取消</button>
          <button @click="handleCleanup" :disabled="cleaning"
            class="px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700 disabled:opacity-50">
            {{ cleaning ? '清理中...' : '确认清理' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { Download, Loader2, FileText, BarChart3, Trash2, X } from 'lucide-vue-next'
import { getAuditLogs, exportAuditLogs, getAuditStatistics, cleanupOldLogs } from '../api/audit'
import type { AuditLog, AuditStatistics } from '../types/management'
import dayjs from 'dayjs'

// 状态
const loading = ref(false)
const exporting = ref(false)
const cleaning = ref(false)
const logs = ref<AuditLog[]>([])
const total = ref(0)
const currentPage = ref(1)
const pageSize = ref(20)
const showDetailDialog = ref(false)
const showStatsDialog = ref(false)
const showCleanupDialog = ref(false)
const selectedLog = ref<AuditLog | null>(null)
const statistics = ref<AuditStatistics | null>(null)
const detailedStats = ref<AuditStatistics | null>(null)
const cleanupDays = ref(90)

// 筛选条件
const filters = ref({
  username: '',
  operate_type: '',
  module: '',
  status_code: '',
  start_time: '',
  end_time: '',
  min_duration: undefined as number | undefined,
  max_duration: undefined as number | undefined
})

// 计算属性
const totalPages = computed(() => Math.ceil(total.value / pageSize.value))

// 加载审计日志
const loadLogs = async () => {
  loading.value = true
  try {
    const res = await getAuditLogs({
      page: currentPage.value,
      page_size: pageSize.value,
      username: filters.value.username || undefined,
      operate_type: filters.value.operate_type || undefined,
      module: filters.value.module || undefined,
      status_code: filters.value.status_code ? parseInt(filters.value.status_code) : undefined,
      start_time: filters.value.start_time || undefined,
      end_time: filters.value.end_time || undefined,
      min_duration: filters.value.min_duration,
      max_duration: filters.value.max_duration
    })
    logs.value = res.data.data.list
    total.value = res.data.data.total
  } catch (error) {
    console.error('加载审计日志失败:', error)
    alert('加载失败，请重试')
  } finally {
    loading.value = false
  }
}

// 搜索（防抖处理可选）
const handleSearch = () => {
  currentPage.value = 1
  loadLogs()
}

// 重置筛选
const resetFilters = () => {
  filters.value = {
    username: '',
    operate_type: '',
    module: '',
    status_code: '',
    start_time: '',
    end_time: '',
    min_duration: undefined,
    max_duration: undefined
  }
  handleSearch()
}

// 分页
const handlePageChange = (page: number) => {
  if (page >= 1 && page <= totalPages.value) {
    currentPage.value = page
    loadLogs()
  }
}

// 显示详情
const showDetail = (log: AuditLog) => {
  selectedLog.value = log
  showDetailDialog.value = true
}

const closeDetailDialog = () => {
  showDetailDialog.value = false
  selectedLog.value = null
}

// 导出日志
const handleExport = async () => {
  if (exporting.value) return

  const confirmed = confirm('确定要导出当前筛选条件下的审计日志吗？')
  if (!confirmed) return

  exporting.value = true
  try {
    const res = await exportAuditLogs({
      username: filters.value.username || undefined,
      operate_type: filters.value.operate_type || undefined,
      module: filters.value.module || undefined,
      status_code: filters.value.status_code ? parseInt(filters.value.status_code) : undefined,
      start_time: filters.value.start_time || undefined,
      end_time: filters.value.end_time || undefined,
      min_duration: filters.value.min_duration,
      max_duration: filters.value.max_duration
    })

    // 创建下载链接
    const blob = new Blob([res.data], { type: 'text/csv' })
    const url = window.URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = `audit_logs_${dayjs().format('YYYYMMDD_HHmmss')}.csv`
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    window.URL.revokeObjectURL(url)

    alert('导出成功')
  } catch (error: any) {
    alert(error.response?.data?.msg || '导出失败')
  } finally {
    exporting.value = false
  }
}

// 格式化日期
const formatDate = (date: string) => {
  return dayjs(date).format('YYYY-MM-DD HH:mm:ss')
}

// 格式化JSON
const formatJson = (jsonStr: string) => {
  try {
    const obj = JSON.parse(jsonStr || '{}')
    return JSON.stringify(obj, null, 2)
  } catch {
    return jsonStr || '-'
  }
}

// 操作类型样式
const getOperateTypeClass = (type: string) => {
  const classes = {
    create: 'bg-green-100 dark:bg-green-900 text-green-800 dark:text-green-200',
    update: 'bg-blue-100 dark:bg-blue-900 text-blue-800 dark:text-blue-200',
    delete: 'bg-red-100 dark:bg-red-900 text-red-800 dark:text-red-200',
    query: 'bg-gray-100 dark:bg-gray-700 text-gray-800 dark:text-gray-200',
    login: 'bg-indigo-100 dark:bg-indigo-900 text-indigo-800 dark:text-indigo-200',
    logout: 'bg-yellow-100 dark:bg-yellow-900 text-yellow-800 dark:text-yellow-200'
  }
  return classes[type as keyof typeof classes] || 'bg-gray-100 dark:bg-gray-700 text-gray-800 dark:text-gray-200'
}

const getOperateTypeName = (type: string) => {
  const names: Record<string, string> = {
    create: '创建',
    update: '更新',
    delete: '删除',
    query: '查询',
    login: '登录',
    logout: '登出'
  }
  return names[type] || type
}

// 状态码样式
const getStatusClass = (statusCode: number) => {
  if (statusCode >= 200 && statusCode < 300) {
    return 'bg-green-100 dark:bg-green-900 text-green-800 dark:text-green-200'
  } else if (statusCode >= 400 && statusCode < 500) {
    return 'bg-yellow-100 dark:bg-yellow-900 text-yellow-800 dark:text-yellow-200'
  } else if (statusCode >= 500) {
    return 'bg-red-100 dark:bg-red-900 text-red-800 dark:text-red-200'
  }
  return 'bg-gray-100 dark:bg-gray-700 text-gray-800 dark:text-gray-200'
}

// 耗时样式
const getDurationClass = (duration: number) => {
  if (duration < 100) return 'text-green-600 dark:text-green-400'
  if (duration < 500) return 'text-yellow-600 dark:text-yellow-400'
  if (duration < 1000) return 'text-orange-600 dark:text-orange-400'
  return 'text-red-600 dark:text-red-400 font-semibold'
}

// 加载统计数据
const loadStatistics = async () => {
  try {
    const res = await getAuditStatistics({
      start_time: filters.value.start_time || undefined,
      end_time: filters.value.end_time || undefined
    })
    statistics.value = res.data.data
  } catch (error) {
    console.error('加载统计数据失败:', error)
  }
}

// 打开统计弹窗时加载详细统计
const openStatsDialog = async () => {
  showStatsDialog.value = true
  try {
    const res = await getAuditStatistics({
      start_time: filters.value.start_time || undefined,
      end_time: filters.value.end_time || undefined
    })
    detailedStats.value = res.data.data
  } catch (error) {
    console.error('加载详细统计失败:', error)
    alert('加载统计数据失败')
  }
}

// 清理旧日志
const handleCleanup = async () => {
  if (cleaning.value) return

  const confirmed = confirm(`确定要删除 ${cleanupDays.value} 天之前的所有审计日志吗？此操作不可恢复！`)
  if (!confirmed) return

  cleaning.value = true
  try {
    const res = await cleanupOldLogs(cleanupDays.value)
    alert(`清理成功，共删除 ${res.data.data.deleted_count} 条记录`)
    showCleanupDialog.value = false
    // 重新加载日志和统计
    loadLogs()
    loadStatistics()
  } catch (error: any) {
    alert(error.response?.data?.msg || '清理失败')
  } finally {
    cleaning.value = false
  }
}

// 初始化
onMounted(() => {
  loadLogs()
  loadStatistics()
})
</script>