<template>
  <AppLayout>
    <div class="space-y-6">
      <!-- Loading State -->
      <div v-if="loading" class="flex items-center justify-center py-12">
        <LoadingSpinner />
      </div>

      <template v-else-if="stats">
        <!-- Row 1: Core Stats -->
        <div class="grid grid-cols-2 gap-4 lg:grid-cols-4">
          <!-- Total API Keys -->
          <div class="card p-4">
            <div class="flex items-center gap-3">
              <div class="rounded-lg bg-blue-100 p-2 dark:bg-blue-900/30">
                <Icon name="key" size="md" class="text-blue-600 dark:text-blue-400" :stroke-width="2" />
              </div>
              <div>
                <p class="text-xs font-medium text-gray-500 dark:text-gray-400">
                  {{ t('admin.dashboard.apiKeys') }}
                </p>
                <p class="text-xl font-bold text-gray-900 dark:text-white">
                  {{ stats.total_api_keys }}
                </p>
                <p class="text-xs text-green-600 dark:text-green-400">
                  {{ stats.active_api_keys }} {{ t('common.active') }}
                </p>
              </div>
            </div>
          </div>

          <!-- Service Accounts -->
          <div class="card p-4">
            <div class="flex items-center gap-3">
              <div class="rounded-lg bg-purple-100 p-2 dark:bg-purple-900/30">
                <Icon name="server" size="md" class="text-purple-600 dark:text-purple-400" :stroke-width="2" />
              </div>
              <div>
                <p class="text-xs font-medium text-gray-500 dark:text-gray-400">
                  {{ t('admin.dashboard.accounts') }}
                </p>
                <p class="text-xl font-bold text-gray-900 dark:text-white">
                  {{ stats.total_accounts }}
                </p>
                <p class="text-xs">
                  <span class="text-green-600 dark:text-green-400"
                    >{{ stats.normal_accounts }} {{ t('common.active') }}</span
                  >
                  <span v-if="stats.error_accounts > 0" class="ml-1 text-red-500"
                    >{{ stats.error_accounts }} {{ t('common.error') }}</span
                  >
                </p>
              </div>
            </div>
          </div>

          <!-- Today Requests -->
          <div class="card p-4">
            <div class="flex items-center gap-3">
              <div class="rounded-lg bg-green-100 p-2 dark:bg-green-900/30">
                <Icon name="chart" size="md" class="text-green-600 dark:text-green-400" :stroke-width="2" />
              </div>
              <div>
                <p class="text-xs font-medium text-gray-500 dark:text-gray-400">
                  {{ t('admin.dashboard.todayRequests') }}
                </p>
                <p class="text-xl font-bold text-gray-900 dark:text-white">
                  {{ stats.today_requests }}
                </p>
                <p class="text-xs text-gray-500 dark:text-gray-400">
                  {{ t('common.total') }}: {{ formatNumber(stats.total_requests) }}
                </p>
              </div>
            </div>
          </div>

          <!-- New Users Today -->
          <div class="card p-4">
            <div class="flex items-center gap-3">
              <div class="rounded-lg bg-emerald-100 p-2 dark:bg-emerald-900/30">
                <Icon name="userPlus" size="md" class="text-emerald-600 dark:text-emerald-400" :stroke-width="2" />
              </div>
              <div>
                <p class="text-xs font-medium text-gray-500 dark:text-gray-400">
                  {{ t('admin.dashboard.users') }}
                </p>
                <p class="text-xl font-bold text-emerald-600 dark:text-emerald-400">
                  +{{ stats.today_new_users }}
                </p>
                <p class="text-xs text-gray-500 dark:text-gray-400">
                  {{ t('common.total') }}: {{ formatNumber(stats.total_users) }}
                </p>
              </div>
            </div>
          </div>
        </div>

        <!-- Row 2: Token Stats -->
        <div class="grid grid-cols-2 gap-4 lg:grid-cols-4">
          <!-- Today Tokens -->
          <div class="card p-4">
            <div class="flex items-center gap-3">
              <div class="rounded-lg bg-amber-100 p-2 dark:bg-amber-900/30">
                <Icon name="cube" size="md" class="text-amber-600 dark:text-amber-400" :stroke-width="2" />
              </div>
              <div>
                <p class="text-xs font-medium text-gray-500 dark:text-gray-400">
                  {{ t('admin.dashboard.todayTokens') }}
                </p>
                <p class="text-xl font-bold text-gray-900 dark:text-white">
                  {{ formatTokens(stats.today_tokens) }}
                </p>
                <p class="text-xs">
                  <span
                    class="text-amber-600 dark:text-amber-400"
                    :title="t('admin.dashboard.actual')"
                    >${{ formatCost(stats.today_actual_cost) }}</span
                  >
                  <span
                    class="text-gray-400 dark:text-gray-500"
                    :title="t('admin.dashboard.standard')"
                  >
                    / ${{ formatCost(stats.today_cost) }}</span
                  >
                </p>
              </div>
            </div>
          </div>

          <!-- Total Tokens -->
          <div class="card p-4">
            <div class="flex items-center gap-3">
              <div class="rounded-lg bg-indigo-100 p-2 dark:bg-indigo-900/30">
                <Icon name="database" size="md" class="text-indigo-600 dark:text-indigo-400" :stroke-width="2" />
              </div>
              <div>
                <p class="text-xs font-medium text-gray-500 dark:text-gray-400">
                  {{ t('admin.dashboard.totalTokens') }}
                </p>
                <p class="text-xl font-bold text-gray-900 dark:text-white">
                  {{ formatTokens(stats.total_tokens) }}
                </p>
                <p class="text-xs">
                  <span
                    class="text-indigo-600 dark:text-indigo-400"
                    :title="t('admin.dashboard.actual')"
                    >${{ formatCost(stats.total_actual_cost) }}</span
                  >
                  <span
                    class="text-gray-400 dark:text-gray-500"
                    :title="t('admin.dashboard.standard')"
                  >
                    / ${{ formatCost(stats.total_cost) }}</span
                  >
                </p>
              </div>
            </div>
          </div>

          <!-- Performance (RPM/TPM) -->
          <div class="card p-4">
            <div class="flex items-center gap-3">
              <div class="rounded-lg bg-violet-100 p-2 dark:bg-violet-900/30">
                <Icon name="bolt" size="md" class="text-violet-600 dark:text-violet-400" :stroke-width="2" />
              </div>
              <div class="flex-1">
                <p class="text-xs font-medium text-gray-500 dark:text-gray-400">
                  {{ t('admin.dashboard.performance') }}
                </p>
                <div class="flex items-baseline gap-2">
                  <p class="text-xl font-bold text-gray-900 dark:text-white">
                    {{ formatTokens(stats.rpm) }}
                  </p>
                  <span class="text-xs text-gray-500 dark:text-gray-400">RPM</span>
                </div>
                <div class="flex items-baseline gap-2">
                  <p class="text-sm font-semibold text-violet-600 dark:text-violet-400">
                    {{ formatTokens(stats.tpm) }}
                  </p>
                  <span class="text-xs text-gray-500 dark:text-gray-400">TPM</span>
                </div>
              </div>
            </div>
          </div>

          <!-- Avg Response Time -->
          <div class="card p-4">
            <div class="flex items-center gap-3">
              <div class="rounded-lg bg-rose-100 p-2 dark:bg-rose-900/30">
                <Icon name="clock" size="md" class="text-rose-600 dark:text-rose-400" :stroke-width="2" />
              </div>
              <div>
                <p class="text-xs font-medium text-gray-500 dark:text-gray-400">
                  {{ t('admin.dashboard.avgResponse') }}
                </p>
                <p class="text-xl font-bold text-gray-900 dark:text-white">
                  {{ formatDuration(stats.average_duration_ms) }}
                </p>
                <p class="text-xs text-gray-500 dark:text-gray-400">
                  {{ stats.active_users }} {{ t('admin.dashboard.activeUsers') }}
                </p>
              </div>
            </div>
          </div>
        </div>

        <!-- Runtime Cluster Monitor -->
        <div class="card p-4">
          <div class="mb-4 flex flex-wrap items-center justify-between gap-3">
            <div>
              <h3 class="text-sm font-semibold text-gray-900 dark:text-white">集群运行监控</h3>
              <p class="text-xs text-gray-500 dark:text-gray-400">
                每 5 秒自动刷新，最近 {{ runtimeHistory.length }} 个趋势点
              </p>
            </div>
            <button
              type="button"
              class="inline-flex items-center gap-2 rounded-md bg-slate-900 px-3 py-2 text-xs font-medium text-white transition hover:bg-slate-700 dark:bg-slate-200 dark:text-slate-900 dark:hover:bg-white"
              @click="openOrchestrator"
            >
              <Icon name="externalLink" size="sm" :stroke-width="2" />
              打开批量注册同步后台
            </button>
          </div>

          <div class="grid grid-cols-2 gap-3 lg:grid-cols-5">
            <div class="rounded-lg border border-gray-200 bg-gray-50 p-3 dark:border-gray-700 dark:bg-gray-900/40">
              <p class="text-xs text-gray-500 dark:text-gray-400">活跃连接数</p>
              <p class="mt-1 text-lg font-semibold text-gray-900 dark:text-white">
                {{ formatNumber(latestRuntime.totalActiveConnections) }}
              </p>
            </div>
            <div class="rounded-lg border border-gray-200 bg-gray-50 p-3 dark:border-gray-700 dark:bg-gray-900/40">
              <p class="text-xs text-gray-500 dark:text-gray-400">活跃 API Key 数</p>
              <p class="mt-1 text-lg font-semibold text-gray-900 dark:text-white">
                {{ formatNumber(latestRuntime.activeApiKeys) }}
              </p>
            </div>
            <div class="rounded-lg border border-gray-200 bg-gray-50 p-3 dark:border-gray-700 dark:bg-gray-900/40">
              <p class="text-xs text-gray-500 dark:text-gray-400">CPU 占用（均值）</p>
              <p class="mt-1 text-lg font-semibold text-gray-900 dark:text-white">
                {{ latestRuntime.avgCpu.toFixed(2) }}%
              </p>
            </div>
            <div class="rounded-lg border border-gray-200 bg-gray-50 p-3 dark:border-gray-700 dark:bg-gray-900/40">
              <p class="text-xs text-gray-500 dark:text-gray-400">内存占用（均值）</p>
              <p class="mt-1 text-lg font-semibold text-gray-900 dark:text-white">
                {{ latestRuntime.avgMemory.toFixed(2) }}%
              </p>
            </div>
            <div class="rounded-lg border border-gray-200 bg-gray-50 p-3 dark:border-gray-700 dark:bg-gray-900/40">
              <p class="text-xs text-gray-500 dark:text-gray-400">网络进/出速率</p>
              <p class="mt-1 text-sm font-semibold text-gray-900 dark:text-white">
                {{ formatBytesPerSecond(latestRuntime.rxRateBytesPerSec) }} / {{ formatBytesPerSecond(latestRuntime.txRateBytesPerSec) }}
              </p>
            </div>
          </div>

          <div class="mt-4 grid grid-cols-1 gap-3 lg:grid-cols-3">
            <div
              v-for="node in runtimeNodes"
              :key="node.node || node.node_name"
              class="rounded-lg border p-3"
              :class="node.ok ? 'border-gray-200 dark:border-gray-700' : 'border-red-300 dark:border-red-700'"
            >
              <div class="flex items-center justify-between">
                <p class="text-xs font-semibold text-gray-800 dark:text-gray-200">
                  {{ node.node_name || node.node || 'unknown' }}
                </p>
                <span
                  class="rounded px-2 py-0.5 text-[11px] font-medium"
                  :class="
                    node.ok
                      ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-300'
                      : 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-300'
                  "
                >
                  {{ node.ok ? 'ONLINE' : 'ERROR' }}
                </span>
              </div>
              <div v-if="node.ok" class="mt-2 space-y-1 text-xs text-gray-600 dark:text-gray-300">
                <p>连接：{{ formatNumber(node.active_connections || 0) }}</p>
                <p>CPU：{{ (node.cpu_percent || 0).toFixed(2) }}%</p>
                <p>内存：{{ (node.memory_percent || 0).toFixed(2) }}%</p>
                <p>网络：{{ formatBytes(node.network_rx_bytes || 0) }} / {{ formatBytes(node.network_tx_bytes || 0) }}</p>
              </div>
              <p v-else class="mt-2 text-xs text-red-600 dark:text-red-400">
                {{ node.error || 'collect failed' }}
              </p>
            </div>
          </div>

          <div class="mt-4 grid grid-cols-1 gap-4 xl:grid-cols-3">
            <div class="rounded-lg border border-gray-200 p-3 dark:border-gray-700">
              <h4 class="mb-2 text-xs font-semibold text-gray-700 dark:text-gray-300">连接与活跃 Key 趋势</h4>
              <div class="h-56">
                <Line v-if="runtimeCountChartData" :data="runtimeCountChartData" :options="runtimeCountChartOptions" />
                <div v-else class="flex h-full items-center justify-center text-xs text-gray-500 dark:text-gray-400">
                  暂无趋势数据
                </div>
              </div>
            </div>
            <div class="rounded-lg border border-gray-200 p-3 dark:border-gray-700">
              <h4 class="mb-2 text-xs font-semibold text-gray-700 dark:text-gray-300">CPU / 内存趋势</h4>
              <div class="h-56">
                <Line v-if="runtimeResourceChartData" :data="runtimeResourceChartData" :options="runtimePercentChartOptions" />
                <div v-else class="flex h-full items-center justify-center text-xs text-gray-500 dark:text-gray-400">
                  暂无趋势数据
                </div>
              </div>
            </div>
            <div class="rounded-lg border border-gray-200 p-3 dark:border-gray-700">
              <h4 class="mb-2 text-xs font-semibold text-gray-700 dark:text-gray-300">网络进/出趋势</h4>
              <div class="h-56">
                <Line v-if="runtimeNetworkChartData" :data="runtimeNetworkChartData" :options="runtimeNetworkChartOptions" />
                <div v-else class="flex h-full items-center justify-center text-xs text-gray-500 dark:text-gray-400">
                  暂无趋势数据
                </div>
              </div>
            </div>
          </div>

          <div class="mt-3 flex items-center justify-between text-xs">
            <p v-if="runtimeError" class="text-red-600 dark:text-red-400">{{ runtimeError }}</p>
            <p v-else class="text-gray-500 dark:text-gray-400">
              最近更新时间：{{ runtimeLastUpdated ? new Date(runtimeLastUpdated).toLocaleTimeString() : '-' }}
            </p>
            <p class="text-gray-500 dark:text-gray-400">刷新间隔：5s</p>
          </div>
        </div>

        <!-- Charts Section -->
        <div class="space-y-6">
          <!-- Date Range Filter -->
          <div class="card p-4">
            <div class="flex flex-wrap items-center gap-4">
              <div class="flex items-center gap-2">
                <span class="text-sm font-medium text-gray-700 dark:text-gray-300"
                  >{{ t('admin.dashboard.timeRange') }}:</span
                >
                <DateRangePicker
                  v-model:start-date="startDate"
                  v-model:end-date="endDate"
                  @change="onDateRangeChange"
                />
              </div>
              <div class="ml-auto flex items-center gap-2">
                <span class="text-sm font-medium text-gray-700 dark:text-gray-300"
                  >{{ t('admin.dashboard.granularity') }}:</span
                >
                <div class="w-28">
                  <Select
                    v-model="granularity"
                    :options="granularityOptions"
                    @change="loadChartData"
                  />
                </div>
              </div>
            </div>
          </div>

          <!-- Charts Grid -->
          <div class="grid grid-cols-1 gap-6 lg:grid-cols-2">
            <ModelDistributionChart :model-stats="modelStats" :loading="chartsLoading" />
            <TokenUsageTrend :trend-data="trendData" :loading="chartsLoading" />
          </div>

          <!-- User Usage Trend (Full Width) -->
          <div class="card p-4">
            <h3 class="mb-4 text-sm font-semibold text-gray-900 dark:text-white">
              {{ t('admin.dashboard.recentUsage') }} (Top 12)
            </h3>
            <div class="h-64">
              <div v-if="userTrendLoading" class="flex h-full items-center justify-center">
                <LoadingSpinner size="md" />
              </div>
              <Line v-else-if="userTrendChartData" :data="userTrendChartData" :options="lineOptions" />
              <div
                v-else
                class="flex h-full items-center justify-center text-sm text-gray-500 dark:text-gray-400"
              >
                {{ t('admin.dashboard.noDataAvailable') }}
              </div>
            </div>
          </div>
        </div>
      </template>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onBeforeUnmount } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'

const { t } = useI18n()
import { adminAPI } from '@/api/admin'
import type { DashboardStats, TrendDataPoint, ModelStat, UserUsageTrendPoint } from '@/types'
import type { RuntimeNodeMetric } from '@/api/admin/dashboard'
import AppLayout from '@/components/layout/AppLayout.vue'
import LoadingSpinner from '@/components/common/LoadingSpinner.vue'
import Icon from '@/components/icons/Icon.vue'
import DateRangePicker from '@/components/common/DateRangePicker.vue'
import Select from '@/components/common/Select.vue'
import ModelDistributionChart from '@/components/charts/ModelDistributionChart.vue'
import TokenUsageTrend from '@/components/charts/TokenUsageTrend.vue'

import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  Filler
} from 'chart.js'
import { Line } from 'vue-chartjs'

// Register Chart.js components
ChartJS.register(
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Title,
  Tooltip,
  Legend,
  Filler
)

const appStore = useAppStore()
const stats = ref<DashboardStats | null>(null)
const loading = ref(false)
const chartsLoading = ref(false)
const userTrendLoading = ref(false)

// Chart data
const trendData = ref<TrendDataPoint[]>([])
const modelStats = ref<ModelStat[]>([])
const userTrend = ref<UserUsageTrendPoint[]>([])
let chartLoadSeq = 0
let usersTrendLoadSeq = 0

interface RuntimeTrendPoint {
  timestampMs: number
  label: string
  totalActiveConnections: number
  activeApiKeys: number
  avgCpu: number
  avgMemory: number
  totalRxBytes: number
  totalTxBytes: number
  rxRateBytesPerSec: number
  txRateBytesPerSec: number
}

const RUNTIME_REFRESH_MS = 5000
const RUNTIME_MAX_POINTS = 60

const runtimeNodes = ref<RuntimeNodeMetric[]>([])
const runtimeHistory = ref<RuntimeTrendPoint[]>([])
const runtimeLastUpdated = ref<number | null>(null)
const runtimeError = ref('')
let runtimeLoadSeq = 0
let runtimePollTimer: ReturnType<typeof setInterval> | null = null

// Helper function to format date in local timezone
const formatLocalDate = (date: Date): string => {
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-${String(date.getDate()).padStart(2, '0')}`
}

// Initialize date range immediately
const now = new Date()
const weekAgo = new Date(now)
weekAgo.setDate(weekAgo.getDate() - 6)

// Date range
const granularity = ref<'day' | 'hour'>('day')
const startDate = ref(formatLocalDate(weekAgo))
const endDate = ref(formatLocalDate(now))

// Granularity options for Select component
const granularityOptions = computed(() => [
  { value: 'day', label: t('admin.dashboard.day') },
  { value: 'hour', label: t('admin.dashboard.hour') }
])

// Dark mode detection
const isDarkMode = computed(() => {
  return document.documentElement.classList.contains('dark')
})

// Chart colors
const chartColors = computed(() => ({
  text: isDarkMode.value ? '#e5e7eb' : '#374151',
  grid: isDarkMode.value ? '#374151' : '#e5e7eb'
}))

// Line chart options (for user trend chart)
const lineOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  interaction: {
    intersect: false,
    mode: 'index' as const
  },
  plugins: {
    legend: {
      position: 'top' as const,
      labels: {
        color: chartColors.value.text,
        usePointStyle: true,
        pointStyle: 'circle',
        padding: 15,
        font: {
          size: 11
        }
      }
    },
    tooltip: {
      itemSort: (a: any, b: any) => {
        const aValue = typeof a?.raw === 'number' ? a.raw : Number(a?.parsed?.y ?? 0)
        const bValue = typeof b?.raw === 'number' ? b.raw : Number(b?.parsed?.y ?? 0)
        return bValue - aValue
      },
      callbacks: {
        label: (context: any) => {
          return `${context.dataset.label}: ${formatTokens(context.raw)}`
        }
      }
    }
  },
  scales: {
    x: {
      grid: {
        color: chartColors.value.grid
      },
      ticks: {
        color: chartColors.value.text,
        font: {
          size: 10
        }
      }
    },
    y: {
      grid: {
        color: chartColors.value.grid
      },
      ticks: {
        color: chartColors.value.text,
        font: {
          size: 10
        },
        callback: (value: string | number) => formatTokens(Number(value))
      }
    }
  }
}))

const latestRuntime = computed<RuntimeTrendPoint>(() => {
  return (
    runtimeHistory.value[runtimeHistory.value.length - 1] || {
      timestampMs: 0,
      label: '',
      totalActiveConnections: 0,
      activeApiKeys: 0,
      avgCpu: 0,
      avgMemory: 0,
      totalRxBytes: 0,
      totalTxBytes: 0,
      rxRateBytesPerSec: 0,
      txRateBytesPerSec: 0
    }
  )
})

const runtimeBaseChartOptions = computed(() => ({
  responsive: true,
  maintainAspectRatio: false,
  interaction: {
    intersect: false,
    mode: 'index' as const
  },
  plugins: {
    legend: {
      position: 'top' as const,
      labels: {
        color: chartColors.value.text,
        usePointStyle: true,
        pointStyle: 'circle',
        padding: 12,
        font: {
          size: 10
        }
      }
    }
  },
  scales: {
    x: {
      grid: {
        color: chartColors.value.grid
      },
      ticks: {
        color: chartColors.value.text,
        font: {
          size: 10
        },
        maxRotation: 0,
        minRotation: 0
      }
    }
  }
}))

const runtimeCountChartData = computed(() => {
  if (!runtimeHistory.value.length) return null
  return {
    labels: runtimeHistory.value.map((point) => point.label),
    datasets: [
      {
        label: '活跃连接',
        data: runtimeHistory.value.map((point) => point.totalActiveConnections),
        borderColor: '#2563eb',
        backgroundColor: '#2563eb1f',
        tension: 0.25,
        fill: true
      },
      {
        label: '活跃 API Key',
        data: runtimeHistory.value.map((point) => point.activeApiKeys),
        borderColor: '#0d9488',
        backgroundColor: '#0d94881f',
        tension: 0.25,
        fill: true
      }
    ]
  }
})

const runtimeCountChartOptions = computed(() => ({
  ...runtimeBaseChartOptions.value,
  scales: {
    ...runtimeBaseChartOptions.value.scales,
    y: {
      grid: {
        color: chartColors.value.grid
      },
      ticks: {
        color: chartColors.value.text,
        font: {
          size: 10
        },
        callback: (value: string | number) => formatNumber(Number(value))
      }
    }
  }
}))

const runtimeResourceChartData = computed(() => {
  if (!runtimeHistory.value.length) return null
  return {
    labels: runtimeHistory.value.map((point) => point.label),
    datasets: [
      {
        label: 'CPU%',
        data: runtimeHistory.value.map((point) => Number(point.avgCpu.toFixed(2))),
        borderColor: '#dc2626',
        backgroundColor: '#dc26261f',
        tension: 0.25,
        fill: true
      },
      {
        label: '内存%',
        data: runtimeHistory.value.map((point) => Number(point.avgMemory.toFixed(2))),
        borderColor: '#7c3aed',
        backgroundColor: '#7c3aed1f',
        tension: 0.25,
        fill: true
      }
    ]
  }
})

const runtimePercentChartOptions = computed(() => ({
  ...runtimeBaseChartOptions.value,
  scales: {
    ...runtimeBaseChartOptions.value.scales,
    y: {
      min: 0,
      max: 100,
      grid: {
        color: chartColors.value.grid
      },
      ticks: {
        color: chartColors.value.text,
        font: {
          size: 10
        },
        callback: (value: string | number) => `${Number(value).toFixed(0)}%`
      }
    }
  }
}))

const runtimeNetworkChartData = computed(() => {
  if (!runtimeHistory.value.length) return null
  return {
    labels: runtimeHistory.value.map((point) => point.label),
    datasets: [
      {
        label: '网络入',
        data: runtimeHistory.value.map((point) => Number((point.rxRateBytesPerSec / (1024 * 1024)).toFixed(4))),
        borderColor: '#16a34a',
        backgroundColor: '#16a34a1f',
        tension: 0.25,
        fill: true
      },
      {
        label: '网络出',
        data: runtimeHistory.value.map((point) => Number((point.txRateBytesPerSec / (1024 * 1024)).toFixed(4))),
        borderColor: '#f97316',
        backgroundColor: '#f973161f',
        tension: 0.25,
        fill: true
      }
    ]
  }
})

const runtimeNetworkChartOptions = computed(() => ({
  ...runtimeBaseChartOptions.value,
  scales: {
    ...runtimeBaseChartOptions.value.scales,
    y: {
      grid: {
        color: chartColors.value.grid
      },
      ticks: {
        color: chartColors.value.text,
        font: {
          size: 10
        },
        callback: (value: string | number) => `${Number(value).toFixed(2)} MB/s`
      }
    }
  }
}))

// User trend chart data
const userTrendChartData = computed(() => {
  if (!userTrend.value?.length) return null

  // Extract display name from email (part before @)
  const getDisplayName = (email: string, userId: number): string => {
    if (email && email.includes('@')) {
      return email.split('@')[0]
    }
    return t('admin.redeem.userPrefix', { id: userId })
  }

  // Group by user
  const userGroups = new Map<string, { name: string; data: Map<string, number> }>()
  const allDates = new Set<string>()

  userTrend.value.forEach((point) => {
    allDates.add(point.date)
    const key = getDisplayName(point.email, point.user_id)
    if (!userGroups.has(key)) {
      userGroups.set(key, { name: key, data: new Map() })
    }
    userGroups.get(key)!.data.set(point.date, point.tokens)
  })

  const sortedDates = Array.from(allDates).sort()
  const colors = [
    '#3b82f6',
    '#10b981',
    '#f59e0b',
    '#ef4444',
    '#8b5cf6',
    '#ec4899',
    '#14b8a6',
    '#f97316',
    '#6366f1',
    '#84cc16',
    '#06b6d4',
    '#a855f7'
  ]

  const datasets = Array.from(userGroups.values()).map((group, idx) => ({
    label: group.name,
    data: sortedDates.map((date) => group.data.get(date) || 0),
    borderColor: colors[idx % colors.length],
    backgroundColor: `${colors[idx % colors.length]}20`,
    fill: false,
    tension: 0.3
  }))

  return {
    labels: sortedDates,
    datasets
  }
})

// Format helpers
const formatTokens = (value: number | undefined): string => {
  if (value === undefined || value === null) return '0'
  if (value >= 1_000_000_000) {
    return `${(value / 1_000_000_000).toFixed(2)}B`
  } else if (value >= 1_000_000) {
    return `${(value / 1_000_000).toFixed(2)}M`
  } else if (value >= 1_000) {
    return `${(value / 1_000).toFixed(2)}K`
  }
  return value.toLocaleString()
}

const formatNumber = (value: number): string => {
  return value.toLocaleString()
}

const formatBytes = (bytes: number): string => {
  if (!Number.isFinite(bytes) || bytes <= 0) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let size = bytes
  let idx = 0
  while (size >= 1024 && idx < units.length - 1) {
    size /= 1024
    idx += 1
  }
  return `${size.toFixed(size >= 10 || idx === 0 ? 0 : 2)} ${units[idx]}`
}

const formatBytesPerSecond = (bytesPerSec: number): string => {
  return `${formatBytes(bytesPerSec)}/s`
}

const formatCost = (value: number): string => {
  if (value >= 1000) {
    return (value / 1000).toFixed(2) + 'K'
  } else if (value >= 1) {
    return value.toFixed(2)
  } else if (value >= 0.01) {
    return value.toFixed(3)
  }
  return value.toFixed(4)
}

const formatDuration = (ms: number): string => {
  if (ms >= 1000) {
    return `${(ms / 1000).toFixed(2)}s`
  }
  return `${Math.round(ms)}ms`
}

const openOrchestrator = () => {
  window.open('https://vip.rshan.cc:9443/', '_blank', 'noopener,noreferrer')
}

const buildRuntimePoint = (
  nodes: RuntimeNodeMetric[],
  activeApiKeys: number,
  timestampMs: number
): RuntimeTrendPoint => {
  const validNodes = nodes.filter((node) => node.ok)
  const divisor = validNodes.length > 0 ? validNodes.length : 1

  const totalActiveConnections = validNodes.reduce(
    (sum, node) => sum + (node.active_connections || 0),
    0
  )
  const avgCpu =
    validNodes.reduce((sum, node) => sum + (node.cpu_percent || 0), 0) / divisor
  const avgMemory =
    validNodes.reduce((sum, node) => sum + (node.memory_percent || 0), 0) / divisor
  const totalRxBytes = validNodes.reduce(
    (sum, node) => sum + (node.network_rx_bytes || 0),
    0
  )
  const totalTxBytes = validNodes.reduce(
    (sum, node) => sum + (node.network_tx_bytes || 0),
    0
  )

  return {
    timestampMs,
    label: new Date(timestampMs).toLocaleTimeString(),
    totalActiveConnections,
    activeApiKeys: activeApiKeys || 0,
    avgCpu,
    avgMemory,
    totalRxBytes,
    totalTxBytes,
    rxRateBytesPerSec: 0,
    txRateBytesPerSec: 0
  }
}

const loadRuntimeClusterMetrics = async () => {
  const currentSeq = ++runtimeLoadSeq
  try {
    const response = await adminAPI.dashboard.getRuntimeClusterMetrics()
    if (currentSeq !== runtimeLoadSeq) return

    const nodes = response.nodes || []
    runtimeNodes.value = nodes

    const timestampMs = (response.timestamp || Math.floor(Date.now() / 1000)) * 1000
    const point = buildRuntimePoint(nodes, response.active_api_keys || 0, timestampMs)
    const previous = runtimeHistory.value[runtimeHistory.value.length - 1]
    if (previous) {
      const secondsDiff = Math.max((point.timestampMs - previous.timestampMs) / 1000, 1)
      point.rxRateBytesPerSec = Math.max(point.totalRxBytes - previous.totalRxBytes, 0) / secondsDiff
      point.txRateBytesPerSec = Math.max(point.totalTxBytes - previous.totalTxBytes, 0) / secondsDiff
    }

    runtimeHistory.value = [...runtimeHistory.value, point].slice(-RUNTIME_MAX_POINTS)
    runtimeLastUpdated.value = timestampMs
    runtimeError.value = ''
  } catch (error) {
    if (currentSeq !== runtimeLoadSeq) return
    runtimeError.value = '集群运行指标拉取失败'
    console.error('Error loading runtime cluster metrics:', error)
  }
}

const startRuntimePolling = () => {
  void loadRuntimeClusterMetrics()
  if (runtimePollTimer) {
    clearInterval(runtimePollTimer)
  }
  runtimePollTimer = setInterval(() => {
    void loadRuntimeClusterMetrics()
  }, RUNTIME_REFRESH_MS)
}

const stopRuntimePolling = () => {
  if (runtimePollTimer) {
    clearInterval(runtimePollTimer)
    runtimePollTimer = null
  }
}

// Date range change handler
const onDateRangeChange = (range: {
  startDate: string
  endDate: string
  preset: string | null
}) => {
  // Auto-select granularity based on date range
  const start = new Date(range.startDate)
  const end = new Date(range.endDate)
  const daysDiff = Math.ceil((end.getTime() - start.getTime()) / (1000 * 60 * 60 * 24))

  // If range is 1 day, use hourly granularity
  if (daysDiff <= 1) {
    granularity.value = 'hour'
  } else {
    granularity.value = 'day'
  }

  loadChartData()
}

// Load data
const loadDashboardSnapshot = async (includeStats: boolean) => {
  const currentSeq = ++chartLoadSeq
  if (includeStats && !stats.value) {
    loading.value = true
  }
  chartsLoading.value = true
  try {
    const response = await adminAPI.dashboard.getSnapshotV2({
      start_date: startDate.value,
      end_date: endDate.value,
      granularity: granularity.value,
      include_stats: includeStats,
      include_trend: true,
      include_model_stats: true,
      include_group_stats: false,
      include_users_trend: false
    })
    if (currentSeq !== chartLoadSeq) return
    if (includeStats && response.stats) {
      stats.value = response.stats
    }
    trendData.value = response.trend || []
    modelStats.value = response.models || []
  } catch (error) {
    if (currentSeq !== chartLoadSeq) return
    appStore.showError(t('admin.dashboard.failedToLoad'))
    console.error('Error loading dashboard snapshot:', error)
  } finally {
    if (currentSeq === chartLoadSeq) {
      loading.value = false
      chartsLoading.value = false
    }
  }
}

const loadUsersTrend = async () => {
  const currentSeq = ++usersTrendLoadSeq
  userTrendLoading.value = true
  try {
    const response = await adminAPI.dashboard.getUserUsageTrend({
      start_date: startDate.value,
      end_date: endDate.value,
      granularity: granularity.value,
      limit: 12
    })
    if (currentSeq !== usersTrendLoadSeq) return
    userTrend.value = response.trend || []
  } catch (error) {
    if (currentSeq !== usersTrendLoadSeq) return
    console.error('Error loading users trend:', error)
    userTrend.value = []
  } finally {
    if (currentSeq === usersTrendLoadSeq) {
      userTrendLoading.value = false
    }
  }
}

const loadDashboardStats = async () => {
  await loadDashboardSnapshot(true)
  void loadUsersTrend()
}

const loadChartData = async () => {
  await loadDashboardSnapshot(false)
  void loadUsersTrend()
}

onMounted(() => {
  void loadDashboardStats()
  startRuntimePolling()
})

onBeforeUnmount(() => {
  stopRuntimePolling()
})
</script>

<style scoped>
</style>
