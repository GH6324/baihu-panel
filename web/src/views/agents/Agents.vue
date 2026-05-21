<script setup lang="ts">
import { ref, onMounted, computed, onUnmounted } from 'vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter, DialogDescription } from '@/components/ui/dialog'
import { AlertDialog, AlertDialogAction, AlertDialogCancel, AlertDialogContent, AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogTitle } from '@/components/ui/alert-dialog'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import {
  RefreshCw, Trash2, Pencil, Copy, Server, Search, Download, RotateCw,
  Plus, Ticket, ListTodo, Eye, Wifi, WifiOff,
  Zap, ZapOff, Check, X, MoreHorizontal
} from 'lucide-vue-next'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
  DropdownMenuSeparator,
} from '@/components/ui/dropdown-menu'
import { api, type Agent, type AgentToken } from '@/api'
import { toast } from 'vue-sonner'
import { useRouter } from 'vue-router'
import { AGENT_STATUS } from '@/constants'
import StatusDot from '@/components/StatusDot.vue'

const router = useRouter()

const agents = ref<Agent[]>([])
const tokens = ref<AgentToken[]>([])
const loading = ref(false)
const searchQuery = ref('')
const activeTab = ref('agents')
const agentVersion = ref('')
const platforms = ref<{ os: string; arch: string; filename: string }[]>([])
const showEditDialog = ref(false)
const showDeleteDialog = ref(false)
const showDownloadDialog = ref(false)
const showTokenDialog = ref(false)
const showEditTokenDialog = ref(false)
const showDetailDialog = ref(false)
const customScheduler = ref(false)
const formData = ref({
  name: '',
  description: '',
  scheduler_config: {
    worker_count: 1,
    queue_size: 100,
    rate_interval: 200,
    verbose: false,
    strict_queue: false
  }
})
const tokenForm = ref({ remark: '', max_uses: 0, expires_at: '' })
const editingToken = ref<AgentToken | null>(null)
const editTokenForm = ref({ remark: '', max_uses: 0, expires_at: '' })
const editingAgent = ref<Agent | null>(null)
const deletingAgent = ref<Agent | null>(null)
const viewingAgent = ref<Agent | null>(null)
let refreshTimer: ReturnType<typeof setInterval> | null = null

const filteredAgents = computed(() => {
  if (!searchQuery.value) return agents.value
  const q = searchQuery.value.toLowerCase()
  return agents.value.filter(a =>
    a.name.toLowerCase().includes(q) ||
    a.hostname?.toLowerCase().includes(q) ||
    a.ip?.toLowerCase().includes(q)
  )
})

function isOnline(agent: Agent): boolean {
  return agent.status === AGENT_STATUS.ONLINE
}

async function loadAgents() {
  loading.value = true
  try {
    const [agentList, versionInfo, tokenList] = await Promise.all([
      api.agents.list(),
      api.agents.getVersion(),
      api.agents.listTokens()
    ])
    agents.value = agentList
    agentVersion.value = versionInfo.version || ''
    platforms.value = versionInfo.platforms || []
    tokens.value = tokenList
  } catch {
    toast.error('加载失败')
  } finally {
    loading.value = false
  }
}

function viewDetail(agent: Agent) {
  ;(document.activeElement as HTMLElement)?.blur()
  viewingAgent.value = agent
  showDetailDialog.value = true
}

function openEditDialog(agent: Agent) {
  ;(document.activeElement as HTMLElement)?.blur()
  editingAgent.value = agent
  customScheduler.value = !!agent.scheduler_config
  formData.value = {
    name: agent.name,
    description: agent.description,
    scheduler_config: agent.scheduler_config ? {
      worker_count: agent.scheduler_config.worker_count,
      queue_size: agent.scheduler_config.queue_size,
      rate_interval: agent.scheduler_config.rate_interval,
      verbose: agent.scheduler_config.verbose,
      strict_queue: agent.scheduler_config.strict_queue
    } : {
      worker_count: 1,
      queue_size: 100,
      rate_interval: 200,
      verbose: false,
      strict_queue: false
    }
  }
  showEditDialog.value = true
}

async function updateAgent() {
  if (!editingAgent.value || !formData.value.name.trim()) return
  try {
    const payload = {
      name: formData.value.name,
      description: formData.value.description,
      enabled: editingAgent.value.enabled,
      scheduler_config: customScheduler.value ? formData.value.scheduler_config : null
    }
    await api.agents.update(editingAgent.value.id, payload)
    showEditDialog.value = false
    await loadAgents()
    toast.success('更新成功')
  } catch (e: unknown) {
    toast.error((e as Error).message || '更新失败')
  }
}

async function toggleEnabled(agent: Agent) {
  try {
    const newEnabled = !agent.enabled
    await api.agents.update(agent.id, {
      name: agent.name,
      description: agent.description,
      enabled: newEnabled,
      scheduler_config: agent.scheduler_config
    })
    await loadAgents()
    toast.success(`${agent.name} 已${newEnabled ? '启用' : '禁用'}`)
  } catch (e: unknown) {
    toast.error((e as Error).message || '操作失败')
  }
}

function confirmDelete(agent: Agent) {
  ;(document.activeElement as HTMLElement)?.blur()
  deletingAgent.value = agent
  showDeleteDialog.value = true
}

function openDownloadDialog() {
  ;(document.activeElement as HTMLElement)?.blur()
  showDownloadDialog.value = true
}

function openTokenDialog() {
  ;(document.activeElement as HTMLElement)?.blur()
  showTokenDialog.value = true
}

async function deleteAgent() {
  if (!deletingAgent.value) return
  try {
    await api.agents.delete(deletingAgent.value.id)
    showDeleteDialog.value = false
    await loadAgents()
    toast.success('删除成功')
  } catch (e: unknown) {
    toast.error((e as Error).message || '删除失败')
  }
}

async function forceUpdate(agent: Agent) {
  try {
    await api.agents.forceUpdate(agent.id)
    toast.success('已标记强制更新')
  } catch (e: unknown) {
    toast.error((e as Error).message || '操作失败')
  }
}

function viewTasks(agent: Agent) {
  router.push({ path: '/tasks', query: { agent_id: String(agent.id) } })
}

function copyToken(token: string) {
  navigator.clipboard.writeText(token)
  toast.success('已复制')
}

async function createToken() {
  try {
    let expiresAt = tokenForm.value.expires_at
    if (expiresAt) {
      // 适配后端格式: 2006-01-02 15:04:05
      expiresAt = expiresAt.replace('T', ' ') + ':00'
    }
    await api.agents.createToken({
      remark: tokenForm.value.remark,
      max_uses: tokenForm.value.max_uses,
      expires_at: expiresAt || undefined
    })
    showTokenDialog.value = false
    tokenForm.value = { remark: '', max_uses: 0, expires_at: '' }
    await loadAgents()
    toast.success('创建成功')
  } catch (e: unknown) {
    toast.error((e as Error).message || '创建失败')
  }
}

async function deleteToken(id: string) {
  try {
    await api.agents.deleteToken(id)
    await loadAgents()
    toast.success('删除成功')
  } catch (e: unknown) {
    toast.error((e as Error).message || '删除失败')
  }
}

function openEditToken(token: AgentToken) {
  ;(document.activeElement as HTMLElement)?.blur()
  editingToken.value = token
  // expires_at 后端格式 "2006-01-02 15:04:05"，转回 datetime-local 格式
  const rawExpires = token.expires_at?.replace(' ', 'T').slice(0, 16) || ''
  editTokenForm.value = { remark: token.remark || '', max_uses: token.max_uses, expires_at: rawExpires }
  showEditTokenDialog.value = true
}

async function updateToken() {
  if (!editingToken.value) return
  try {
    let expiresAt = editTokenForm.value.expires_at
    if (expiresAt) {
      expiresAt = expiresAt.replace('T', ' ') + ':00'
    }
    await api.agents.updateToken(editingToken.value.id, {
      remark: editTokenForm.value.remark,
      max_uses: editTokenForm.value.max_uses,
      expires_at: expiresAt || undefined
    })
    showEditTokenDialog.value = false
    await loadAgents()
    toast.success('更新成功')
  } catch (e: unknown) {
    toast.error((e as Error).message || '更新失败')
  }
}

function isTokenExpired(token: AgentToken) {
  if (!token.expires_at) return false
  // 将 "YYYY-MM-DD HH:mm:ss" 转换为 ISO 格式 "YYYY-MM-DDTHH:mm:ss" 以提高浏览器兼容性
  const dateStr = token.expires_at.replace(' ', 'T')
  return new Date(dateStr) < new Date()
}

function isTokenExhausted(token: AgentToken) {
  return token.max_uses > 0 && token.used_count >= token.max_uses
}

function downloadAgent(os: string, arch: string) {
  window.open(api.agents.downloadUrl(os, arch), '_blank')
}

function getPlatformLabel(os: string, arch: string) {
  const osLabels: Record<string, string> = { linux: 'Linux', windows: 'Windows', darwin: 'macOS' }
  const archLabels: Record<string, string> = { amd64: 'x64', arm64: 'ARM64', '386': 'x86' }
  return `${osLabels[os] || os} ${archLabels[arch] || arch}`
}

onMounted(() => {
  loadAgents()
  refreshTimer = setInterval(loadAgents, 10000)
})

onUnmounted(() => {
  if (refreshTimer) clearInterval(refreshTimer)
})
</script>


<template>
  <Tabs v-model="activeTab" class="space-y-6">
    <div class="flex flex-col md:flex-row md:items-center justify-between gap-4">
      <div class="flex flex-col shrink-0">
        <h2 class="text-xl sm:text-2xl font-bold tracking-tight">Agent 管理</h2>
        <p class="text-muted-foreground text-xs mt-0.5 ml-0.5">管理远程执行代理</p>
      </div>

      <div class="flex flex-row items-center flex-wrap gap-2 w-full md:w-auto md:ml-auto md:justify-end">
        <!-- 搜索与操作 -->
        <div class="flex flex-row items-center gap-2 w-full sm:flex-1 md:flex-none md:w-auto text-sm">
          <div class="relative flex-1 md:flex-none md:w-[200px] group">
            <Search class="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground group-focus-within:text-primary transition-colors" />
            <Input v-model="searchQuery" placeholder="搜索 Agent..." class="h-9 pl-9 w-full bg-muted/20 border-muted-foreground/10 focus:bg-background text-sm" />
          </div>
          
          <Button variant="outline" size="icon" class="h-9 w-9 shrink-0" @click="loadAgents" :disabled="loading" title="刷新">
            <RefreshCw class="h-4 w-4" :class="{ 'animate-spin': loading }" />
          </Button>

          <Button variant="outline" class="h-9 px-3 shrink-0 shadow-sm" @click="openDownloadDialog">
            <Download class="h-4 w-4 md:mr-2" /> <span class="hidden md:inline">下载 Agent</span>
          </Button>
        </div>

        <TabsList class="h-9 p-0.5 bg-muted/20 border border-border/40 rounded-lg w-full sm:w-auto">
          <TabsTrigger value="agents" class="px-3 h-8 text-xs gap-1.5 font-medium transition-all flex-1 sm:flex-none">
            <Server class="w-3.5 h-3.5 opacity-70" />
            <span>列表</span>
          </TabsTrigger>
          <TabsTrigger value="regcodes" class="px-3 h-8 text-xs gap-1.5 font-medium transition-all flex-1 sm:flex-none">
            <Ticket class="w-3.5 h-3.5 opacity-70" />
            <span>令牌</span>
          </TabsTrigger>
        </TabsList>
      </div>
    </div>

    <TabsContent value="agents" class="mt-0">
        <div class="rounded-lg border bg-card overflow-hidden">
          <!-- ========== 1. 大屏布局 (Large >= 1280px) ========== -->
          <div class="hidden xl:block">
            <!-- 表头 -->
            <div class="flex items-center gap-4 px-4 py-1.5 border-b bg-muted/20 text-xs text-muted-foreground font-medium">
              <span class="w-12 shrink-0 pl-1">序号</span>
              <span class="w-48 shrink-0">名称</span>
              <span class="w-32 shrink-0">IP 地址</span>
              <span class="w-32 shrink-0">主机名</span>
              <span class="w-28 shrink-0">版本</span>
              <span class="flex-1 min-w-0">心跳时间</span>
              <span class="w-24 shrink-0 text-center">操作</span>
            </div>
            <!-- 列表 -->
            <div class="divide-y text-sm">
              <div v-if="filteredAgents.length === 0" class="text-center py-12 text-muted-foreground">
                <Server class="h-8 w-8 mx-auto mb-2 opacity-50" />
                {{ searchQuery ? '无匹配结果' : '暂无 Agent' }}
              </div>
              <div v-for="(agent, index) in filteredAgents" :key="`large-${agent.id}`"
                class="flex items-center gap-2 px-4 py-1.5 hover:bg-muted/30 transition-colors">
                <StatusDot :state="isOnline(agent) ? 'online' : 'offline'" :title="isOnline(agent) ? '在线' : '离线'" />
                <div class="w-12 shrink-0 pl-1 text-muted-foreground tabular-nums text-sm">#{{ filteredAgents.length - index }}</div>
                <div class="w-48 shrink-0 flex flex-col justify-center gap-0.5 overflow-hidden">
                  <span class="font-medium truncate cursor-pointer hover:text-primary transition-colors" @click="viewDetail(agent)">{{ agent.name }}</span>
                  <div v-if="agent.description" class="text-[10px] text-muted-foreground truncate">{{ agent.description }}</div>
                </div>
                <span class="w-32 shrink-0 text-xs text-muted-foreground truncate">{{ agent.ip || '-' }}</span>
                <span class="w-32 shrink-0 text-xs text-muted-foreground truncate">{{ agent.hostname || '-' }}</span>
                <span class="w-28 shrink-0 text-xs text-muted-foreground truncate">{{ agent.version || '-' }}</span>
                <span class="flex-1 min-w-0 text-[11px] text-muted-foreground tabular-nums truncate">
                  {{ agent.last_seen || '-' }}
                </span>
                <span class="w-24 shrink-0 flex justify-center items-center">
                  <span class="cursor-pointer group mr-1" @click="toggleEnabled(agent)" :title="agent.enabled ? '点击禁用' : '点击启用'">
                    <div v-if="agent.enabled" class="h-6 w-6 rounded-md bg-green-500/5 flex items-center justify-center group-hover:bg-green-500/10">
                      <Zap class="h-3 w-3 text-green-500 fill-green-500" />
                    </div>
                    <div v-else class="h-6 w-6 rounded-md bg-muted flex items-center justify-center group-hover:bg-muted/80">
                      <ZapOff class="h-3 w-3 text-muted-foreground" />
                    </div>
                  </span>
                  <Button variant="ghost" size="icon" class="h-6 w-6" @click="viewDetail(agent)" title="详情"><Eye class="h-3 w-3" /></Button>
                  <Button variant="ghost" size="icon" class="h-6 w-6" @click="viewTasks(agent)" title="查看任务"><ListTodo class="h-3 w-3" /></Button>
                  <Button variant="ghost" size="icon" class="h-6 w-6" @click="openEditDialog(agent)" title="编辑"><Pencil class="h-3 w-3" /></Button>
                  
                  <DropdownMenu>
                    <DropdownMenuTrigger as-child>
                      <Button variant="ghost" size="icon" class="h-6 w-6"><MoreHorizontal class="h-3 w-3" /></Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end" class="w-32">
                      <DropdownMenuItem @click="forceUpdate(agent)">
                        <RotateCw class="h-3.5 w-3.5 mr-2" />
                        <span>强制更新</span>
                      </DropdownMenuItem>
                      <DropdownMenuSeparator />
                      <DropdownMenuItem class="text-destructive focus:text-destructive" @click="confirmDelete(agent)">
                        <Trash2 class="h-3.5 w-3.5 mr-2" />
                        <span>删除 Agent</span>
                      </DropdownMenuItem>
                    </DropdownMenuContent>
                  </DropdownMenu>
                </span>
              </div>
            </div>
          </div>

          <!-- ========== 2. 中屏布局 (Medium 640px - 1280px) ========== -->
          <div class="hidden sm:block xl:hidden">
            <!-- 表头 -->
            <div class="flex items-center gap-4 px-4 py-1.5 border-b bg-muted/20 text-xs text-muted-foreground font-medium">
              <span class="w-12 shrink-0 pl-1">序号</span>
              <span class="w-48 shrink-0">名称</span>
              <span class="flex-1 min-w-0">IP 地址</span>
              <span class="w-24 shrink-0 text-center">操作</span>
            </div>
            <!-- 列表 -->
            <div class="divide-y text-sm">
              <div v-for="(agent, index) in filteredAgents" :key="`medium-${agent.id}`"
                class="flex items-center gap-2 px-4 py-2.5 hover:bg-muted/30 transition-colors">
                <StatusDot :state="isOnline(agent) ? 'online' : 'offline'" :title="isOnline(agent) ? '在线' : '离线'" />
                <div class="w-12 shrink-0 pl-1 text-muted-foreground tabular-nums text-xs">#{{ filteredAgents.length - index }}</div>
                <div class="w-48 shrink-0 flex flex-col justify-center gap-0.5 overflow-hidden">
                  <span class="font-medium truncate">{{ agent.name }}</span>
                  <div v-if="agent.description" class="text-[10px] text-muted-foreground truncate">{{ agent.description }}</div>
                </div>
                <span class="flex-1 min-w-0 text-xs text-muted-foreground truncate">{{ agent.ip || '-' }}</span>
                <div class="w-24 shrink-0 flex justify-center">
                  <Button variant="ghost" size="icon" class="h-6 w-6" @click="viewDetail(agent)"><Eye class="h-3 w-3" /></Button>
                  <Button variant="ghost" size="icon" class="h-6 w-6" @click="viewTasks(agent)"><ListTodo class="h-3 w-3" /></Button>
                  <Button variant="ghost" size="icon" class="h-6 w-6" @click="openEditDialog(agent)"><Pencil class="h-3 w-3" /></Button>
                  <DropdownMenu>
                    <DropdownMenuTrigger as-child>
                      <Button variant="ghost" size="icon" class="h-6 w-6"><MoreHorizontal class="h-3 w-3" /></Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end">
                      <DropdownMenuItem @click="forceUpdate(agent)">
                        <RotateCw class="h-3.5 w-3.5 mr-2" />更新
                      </DropdownMenuItem>
                      <DropdownMenuSeparator />
                      <DropdownMenuItem class="text-destructive" @click="confirmDelete(agent)">
                        <Trash2 class="h-3.5 w-3.5 mr-2" />删除
                      </DropdownMenuItem>
                    </DropdownMenuContent>
                  </DropdownMenu>
                </div>
              </div>
            </div>
          </div>

          <!-- ========== 3. 小屏布局 (Small < 640px) ========== -->
          <div class="divide-y sm:hidden">
            <div v-if="filteredAgents.length === 0" class="text-sm text-muted-foreground text-center py-12">暂无 Agent</div>
            <div v-for="(agent, index) in filteredAgents" :key="`small-${agent.id}`" class="p-3 hover:bg-muted/50 transition-colors">
              <div class="flex items-start justify-between mb-3 border-b border-border/40 pb-2">
                <div class="flex items-center gap-2 flex-1 min-w-0 pr-2">
                  <StatusDot :state="isOnline(agent) ? 'online' : 'offline'" :title="isOnline(agent) ? '在线' : '离线'" />
                  <span class="text-xs text-muted-foreground tabular-nums flex-shrink-0">#{{ filteredAgents.length - index }}</span>
                  <div class="flex items-center gap-1.5 min-w-0 flex-1">
                    <span class="font-bold text-sm truncate" @click="viewDetail(agent)">{{ agent.name }}</span>
                  </div>
                </div>
                <span @click="toggleEnabled(agent)" class="cursor-pointer">
                  <div v-if="agent.enabled" class="h-6 w-6 rounded-md bg-green-500/10 flex items-center justify-center">
                    <Zap class="h-3.5 w-3.5 text-green-500 fill-green-500" />
                  </div>
                  <div v-else class="h-6 w-6 rounded-md bg-muted flex items-center justify-center">
                    <ZapOff class="h-3.5 w-3.5 text-muted-foreground" />
                  </div>
                </span>
              </div>
              <!-- 详情信息 -->
              <div class="space-y-1.5 text-xs text-muted-foreground mb-3 px-1">
                <div class="flex items-center gap-3">
                  <span class="w-10 shrink-0 font-medium opacity-70">IP:</span>
                  <span class="flex-1 truncate text-foreground">{{ agent.ip || '-' }}</span>
                </div>
                <div class="flex items-center gap-3">
                  <span class="w-10 shrink-0 font-medium opacity-70">主机:</span>
                  <span class="flex-1 truncate">{{ agent.hostname || '-' }}</span>
                </div>
                <div v-if="agent.description" class="flex items-start gap-3">
                  <span class="w-10 shrink-0 font-medium mt-0.5 opacity-70">描述:</span>
                  <span class="flex-1 text-[11px] line-clamp-1">{{ agent.description }}</span>
                </div>
              </div>
              <div class="grid grid-cols-4 items-center pt-2 mt-2 border-t border-border/40 -mx-1">
                <Button variant="ghost" class="h-9 px-0 text-xs gap-1.5 hover:bg-primary/5 rounded-none" @click="viewDetail(agent)">
                  <Eye class="h-3.5 w-3.5" />详情
                </Button>
                <Button variant="ghost" class="h-9 px-0 text-xs gap-1.5 hover:bg-primary/5 rounded-none border-l border-border/10" @click="viewTasks(agent)">
                  <ListTodo class="h-3.5 w-3.5" />任务
                </Button>
                <Button variant="ghost" class="h-9 px-0 text-xs gap-1.5 hover:bg-primary/5 rounded-none border-l border-border/10" @click="openEditDialog(agent)">
                  <Pencil class="h-3.5 w-3.5" />编辑
                </Button>
                <DropdownMenu>
                  <DropdownMenuTrigger as-child>
                    <Button variant="ghost" class="h-9 px-0 text-xs gap-1.5 hover:bg-primary/5 rounded-none border-l border-border/10 w-full">
                      <MoreHorizontal class="h-3.5 w-3.5" />更多
                    </Button>
                  </DropdownMenuTrigger>
                  <DropdownMenuContent align="end" class="w-40">
                    <DropdownMenuItem @click="forceUpdate(agent)">
                      <RotateCw class="h-4 w-4 mr-2" />更新 Agent
                    </DropdownMenuItem>
                    <DropdownMenuSeparator />
                    <DropdownMenuItem class="text-destructive" @click="confirmDelete(agent)">
                      <Trash2 class="h-4 w-4 mr-2" />删除 Agent
                    </DropdownMenuItem>
                  </DropdownMenuContent>
                </DropdownMenu>
              </div>
            </div>
          </div>
        </div>
      </TabsContent>

      <TabsContent value="regcodes" class="mt-0">
        <div class="rounded-lg border bg-card overflow-hidden">
          <!-- 表头 -->
          <div class="flex items-center gap-4 px-4 py-1.5 border-b bg-muted/20 text-xs text-muted-foreground font-medium">
            <span class="w-8 shrink-0">状态</span>
            <span class="flex-1 min-w-0">令牌</span>
            <span class="w-32 shrink-0 hidden sm:block">备注</span>
            <span class="w-16 shrink-0 text-center hidden sm:block">次数</span>
            <span class="w-32 shrink-0 hidden md:block">过期时间</span>
            <span class="w-24 shrink-0 flex justify-end">
              <Button size="sm" class="h-6 px-2 text-[10px]" @click="openTokenDialog">
                <Plus class="h-3 w-3 mr-1" />生成
              </Button>
            </span>
          </div>
          <!-- 数据行 -->
          <div class="divide-y text-sm">
            <div v-if="tokens.length === 0" class="text-center py-12 text-muted-foreground">
              <Ticket class="h-8 w-8 mx-auto mb-2 opacity-50" />暂无令牌
            </div>
            <div v-for="token in tokens" :key="token.id"
              class="flex items-center gap-4 px-4 py-1.5 hover:bg-muted/30 transition-colors">
              <!-- 状态 -->
              <span class="w-8 shrink-0 flex justify-center">
                <div v-if="!isTokenExpired(token) && !isTokenExhausted(token)"
                  class="h-5 w-5 rounded-full bg-green-500/10 flex items-center justify-center">
                  <Check class="h-3 w-3 text-green-500 stroke-[3]" />
                </div>
                <div v-else class="h-5 w-5 rounded-full bg-red-500/10 flex items-center justify-center">
                  <X class="h-3 w-3 text-red-500 stroke-[3]" />
                </div>
              </span>
              <!-- Token -->
              <code class="flex-1 min-w-0 font-mono text-xs bg-muted/40 px-2 py-0.5 rounded truncate text-muted-foreground">{{ token.token }}</code>
              <!-- 备注 -->
              <span class="w-32 shrink-0 text-xs text-muted-foreground truncate hidden sm:block">{{ token.remark || '-' }}</span>
              <!-- 使用次数 -->
              <span class="w-16 shrink-0 text-xs text-muted-foreground text-center hidden sm:block tabular-nums">
                {{ token.used_count }}/{{ token.max_uses === 0 ? '∞' : token.max_uses }}
              </span>
              <!-- 过期时间 -->
              <span class="w-32 shrink-0 text-[11px] text-muted-foreground truncate hidden md:block tabular-nums">
                {{ token.expires_at || '永不过期' }}
              </span>
              <!-- 操作 -->
              <span class="w-24 shrink-0 flex justify-end items-center">
                <Button variant="ghost" size="icon" class="h-6 w-6" @click="copyToken(token.token)" title="复制">
                  <Copy class="h-3 w-3" />
                </Button>
                <Button variant="ghost" size="icon" class="h-6 w-6" @click="openEditToken(token)" title="编辑">
                  <Pencil class="h-3 w-3" />
                </Button>
                <Button variant="ghost" size="icon" class="h-6 w-6 text-destructive" @click="deleteToken(token.id)" title="删除">
                  <Trash2 class="h-3 w-3" />
                </Button>
              </span>
            </div>
          </div>
        </div>
      </TabsContent>

    <!-- 详情对话框 -->
    <Dialog v-model:open="showDetailDialog">
      <DialogContent class="sm:max-w-md md:max-w-lg" @openAutoFocus.prevent>
        <DialogHeader>
          <DialogTitle>Agent 详情</DialogTitle>
          <DialogDescription class="sr-only">显示 Agent 的详细配置和状态信息</DialogDescription>
        </DialogHeader>
        <div v-if="viewingAgent" class="space-y-3">
          <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
            <div class="flex items-center justify-between sm:block">
              <Label class="text-muted-foreground text-xs">ID</Label>
              <div class="text-sm font-medium">#{{ viewingAgent.id }}</div>
            </div>
            <div class="flex items-center justify-between sm:block">
              <Label class="text-muted-foreground text-xs">名称</Label>
              <div class="text-sm font-medium">{{ viewingAgent.name }}</div>
            </div>
            <div class="flex items-center justify-between sm:block">
              <Label class="text-muted-foreground text-xs">IP 地址</Label>
              <div class="text-sm">{{ viewingAgent.ip || '-' }}</div>
            </div>
            <div class="flex items-center justify-between sm:block">
              <Label class="text-muted-foreground text-xs">主机名</Label>
              <div class="text-sm">{{ viewingAgent.hostname || '-' }}</div>
            </div>
            <div class="flex items-center justify-between sm:block">
              <Label class="text-muted-foreground text-xs">操作系统</Label>
              <div class="text-sm">{{ viewingAgent.os || '-' }}</div>
            </div>
            <div class="flex items-center justify-between sm:block">
              <Label class="text-muted-foreground text-xs">架构</Label>
              <div class="text-sm">{{ viewingAgent.arch || '-' }}</div>
            </div>
            <div class="flex items-center justify-between sm:block">
              <Label class="text-muted-foreground text-xs">版本</Label>
              <div class="text-sm">{{ viewingAgent.version || '-' }}</div>
            </div>
            <div class="flex items-center justify-between sm:block">
              <Label class="text-muted-foreground text-xs">构建时间</Label>
              <div class="text-sm">{{ viewingAgent.build_time || '-' }}</div>
            </div>
            <div class="flex items-center justify-between sm:block">
              <Label class="text-muted-foreground text-xs">在线状态</Label>
              <div class="flex items-center gap-2">
                <Wifi v-if="isOnline(viewingAgent)" class="h-4 w-4 text-green-500" />
                <WifiOff v-else class="h-4 w-4 text-muted-foreground" />
                <span class="text-sm">{{ isOnline(viewingAgent) ? '在线' : '离线' }}</span>
              </div>
            </div>
            <div class="flex items-center justify-between sm:block">
              <Label class="text-muted-foreground text-xs">启用状态</Label>
              <div class="text-sm">{{ viewingAgent.enabled ? '已启用' : '已禁用' }}</div>
            </div>
            <div class="flex items-center justify-between sm:block">
              <Label class="text-muted-foreground text-xs">最后心跳</Label>
              <div class="text-sm">{{ viewingAgent.last_seen || '-' }}</div>
            </div>
            <div class="flex items-center justify-between sm:block">
              <Label class="text-muted-foreground text-xs">注册时间</Label>
              <div class="text-sm">{{ viewingAgent.created_at || '-' }}</div>
            </div>
            <div class="flex items-center justify-between sm:block">
              <Label class="text-muted-foreground text-xs">任务排队</Label>
              <div class="text-sm">
                {{ viewingAgent.scheduler_config ? `自定义 (并发: ${viewingAgent.scheduler_config.worker_count}, 队列: ${viewingAgent.scheduler_config.queue_size}, 严格排队: ${viewingAgent.scheduler_config.strict_queue ? '是' : '否'})` : '继承全局' }}
              </div>
            </div>
          </div>
          <div v-if="viewingAgent.description" class="pt-2 border-t">
            <Label class="text-muted-foreground text-xs">描述</Label>
            <div class="text-sm mt-1">{{ viewingAgent.description }}</div>
          </div>
        </div>
      </DialogContent>
    </Dialog>

    <!-- 编辑对话框 -->
    <Dialog v-model:open="showEditDialog">
      <DialogContent @openAutoFocus.prevent>
        <DialogHeader>
          <DialogTitle>编辑 Agent</DialogTitle>
          <DialogDescription class="sr-only">修改 Agent 的名称和描述信息</DialogDescription>
        </DialogHeader>
        <div class="space-y-4">
          <div class="space-y-1.5">
            <Label class="text-xs font-medium text-foreground">名称</Label>
            <Input v-model="formData.name" placeholder="Agent 名称" class="h-9" />
          </div>
          <div class="space-y-1.5">
            <Label class="text-xs font-medium text-foreground">描述</Label>
            <Input v-model="formData.description" placeholder="描述信息（可选）" class="h-9" />
          </div>
          <div class="flex items-center justify-between rounded-lg border p-3 shadow-sm">
            <div class="space-y-0.5">
              <Label class="text-sm font-medium">自定义调度配置</Label>
              <div class="text-xs text-muted-foreground">开启后可独立配置该 Agent 的并发限制与任务排队参数</div>
            </div>
            <Switch v-model="customScheduler" />
          </div>
          <div v-if="customScheduler" class="p-4 rounded-lg border border-border bg-muted/20 space-y-4 mt-2">
            <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <div class="space-y-1.5">
                <Label class="text-xs font-medium text-foreground">并发限制数 (Workers)</Label>
                <Input type="number" v-model.number="formData.scheduler_config.worker_count" :min="1" class="h-9" />
                <p class="text-[10px] text-muted-foreground">同一时间最大并行任务数</p>
              </div>
              <div class="space-y-1.5">
                <Label class="text-xs font-medium text-foreground">最大队列数 (Queue Size)</Label>
                <Input type="number" v-model.number="formData.scheduler_config.queue_size" :min="1" class="h-9" />
                <p class="text-[10px] text-muted-foreground">并发满时等待排队的任务数</p>
              </div>
            </div>
            <div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
              <div class="space-y-1.5">
                <Label class="text-xs font-medium text-foreground">调度频率限制 (Rate Interval)</Label>
                <div class="relative">
                  <Input type="number" v-model.number="formData.scheduler_config.rate_interval" :min="0" class="h-9 pr-10" />
                  <span class="absolute right-3 top-1/2 -translate-y-1/2 text-xs text-muted-foreground">ms</span>
                </div>
                <p class="text-[10px] text-muted-foreground">两次调度启动的最小间隔时间</p>
              </div>
              <div class="space-y-1.5">
                <Label class="text-xs font-medium text-foreground">执行降级策略</Label>
                <div class="flex items-center justify-between rounded-md border border-input bg-card px-3 h-9 shadow-sm">
                  <span class="text-xs text-muted-foreground">队列满时拒绝执行</span>
                  <Switch v-model="formData.scheduler_config.strict_queue" class="scale-90" />
                </div>
                <p class="text-[10px] text-muted-foreground">开启后拒绝执行；关闭则同步直接执行</p>
              </div>
            </div>
            <div class="rounded-md bg-yellow-500/10 border border-yellow-500/20 p-2.5 text-[10px] text-yellow-600 dark:text-yellow-400 leading-relaxed">
              <strong>提示：</strong>开启严格排队后，服务端将不再拦截此 Agent 上已运行任务的并行触发，而是交由 Agent 本地队列调度。若要严格保证任务不并行，请将 <strong>并发限制数 (Workers)</strong> 设为 1。
            </div>
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" @click="showEditDialog = false">取消</Button>
          <Button @click="updateAgent">保存</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- 删除确认对话框 -->
    <AlertDialog v-model:open="showDeleteDialog">
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>确认删除</AlertDialogTitle>
          <AlertDialogDescription>
            确定要删除 Agent "{{ deletingAgent?.name }}" 吗？此操作无法撤销。
          </AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>取消</AlertDialogCancel>
          <AlertDialogAction class="bg-destructive text-white hover:bg-destructive/90" @click="deleteAgent">删除
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>

    <!-- 下载对话框 -->
    <Dialog v-model:open="showDownloadDialog">
      <DialogContent class="sm:max-w-lg" @openAutoFocus.prevent>
        <DialogHeader>
          <DialogTitle>下载 Agent</DialogTitle>
          <DialogDescription>当前版本: {{ agentVersion }}</DialogDescription>
        </DialogHeader>
        <div class="space-y-4">
          <div class="bg-blue-500/10 text-blue-600 dark:text-blue-400 p-3 rounded-md text-sm border border-blue-500/20">
            <p class="font-medium mb-1">💡 下载说明：</p>
            <ul class="list-disc list-inside space-y-1 text-xs opacity-90">
              <li>若主程序为 <strong class="font-semibold">Docker 部署</strong>：支持直接在此处下载包含配置的自动打包程序。</li>
              <li>若主程序为 <strong class="font-semibold">单文件二进制部署</strong>：面板无法直接提供完整打包下载，请前往 <a href="https://github.com/engigu/baihu-panel/releases" target="_blank" class="underline font-medium hover:text-blue-500 transition-colors">GitHub Releases</a> 手动下载对应的 Agent。</li>
            </ul>
          </div>
          <div class="space-y-2">
            <div v-for="platform in platforms" :key="`${platform.os}-${platform.arch}`"
              class="flex items-center justify-between p-3 border rounded-lg hover:bg-muted/50 transition-colors">
              <span class="font-medium">{{ getPlatformLabel(platform.os, platform.arch) }}</span>
              <Button size="sm" @click="downloadAgent(platform.os, platform.arch)">
                <Download class="h-4 w-4 mr-1.5" />下载
              </Button>
            </div>
          </div>
          <div class="border-t pt-4">
            <h4 class="font-medium mb-2">使用说明</h4>
            <ol class="text-sm text-muted-foreground space-y-1.5 list-decimal list-inside">
              <li>下载对应平台的 Agent 压缩包并解压</li>
              <li>复制 <code class="bg-muted px-1.5 py-0.5 rounded text-foreground">config.example.ini</code> 为 <code
                  class="bg-muted px-1.5 py-0.5 rounded text-foreground">config.ini</code></li>
              <li>编辑 <code class="bg-muted px-1.5 py-0.5 rounded text-foreground">config.ini</code>，填写服务器地址和注册令牌</li>
              <li>运行 <code class="bg-muted px-1.5 py-0.5 rounded text-foreground">./baihu-agent start</code> 启动（后台运行）
              </li>
            </ol>
            <div class="mt-3 text-sm text-muted-foreground">
              <p class="font-medium text-foreground mb-1.5">常用命令：</p>
              <div class="space-y-1">
                <div><code class="bg-muted px-1.5 py-0.5 rounded text-foreground text-xs">baihu-agent start</code> <span
                    class="text-xs">- 后台启动</span></div>
                <div><code class="bg-muted px-1.5 py-0.5 rounded text-foreground text-xs">baihu-agent stop</code> <span
                    class="text-xs">- 停止运行</span></div>
                <div><code class="bg-muted px-1.5 py-0.5 rounded text-foreground text-xs">baihu-agent status</code>
                  <span class="text-xs">- 查看状态</span>
                </div>
                <div><code class="bg-muted px-1.5 py-0.5 rounded text-foreground text-xs">baihu-agent logs</code> <span
                    class="text-xs">- 查看日志</span></div>
                <div><code class="bg-muted px-1.5 py-0.5 rounded text-foreground text-xs">baihu-agent run</code> <span
                    class="text-xs">- 前台运行</span></div>
              </div>
            </div>
          </div>
        </div>
      </DialogContent>
    </Dialog>

    <!-- 创建令牌对话框 -->
    <Dialog v-model:open="showTokenDialog">
      <DialogContent @openAutoFocus.prevent>
        <DialogHeader>
          <DialogTitle>生成令牌</DialogTitle>
          <DialogDescription class="sr-only">创建一个新的注册令牌，用于 Agent 认证</DialogDescription>
        </DialogHeader>
        <div class="space-y-4">
          <div>
            <Label>备注</Label>
            <Input v-model="tokenForm.remark" placeholder="备注信息（可选）" />
          </div>
          <div>
            <Label>最大使用次数</Label>
            <Input v-model.number="tokenForm.max_uses" type="number" placeholder="0 表示无限制" />
          </div>
          <div>
            <Label>过期时间</Label>
            <Input v-model="tokenForm.expires_at" type="datetime-local" />
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" @click="showTokenDialog = false">取消</Button>
          <Button @click="createToken">生成</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>

    <!-- 编辑令牌对话框 -->
    <Dialog v-model:open="showEditTokenDialog">
      <DialogContent @openAutoFocus.prevent>
        <DialogHeader>
          <DialogTitle>编辑令牌</DialogTitle>
          <DialogDescription class="sr-only">修改令牌的备注、使用次数和过期时间</DialogDescription>
        </DialogHeader>
        <div class="space-y-4">
          <div>
            <Label>备注</Label>
            <Input v-model="editTokenForm.remark" placeholder="备注信息（可选）" />
          </div>
          <div>
            <Label>最大使用次数</Label>
            <Input v-model.number="editTokenForm.max_uses" type="number" placeholder="0 表示无限制" />
          </div>
          <div>
            <Label>过期时间</Label>
            <Input v-model="editTokenForm.expires_at" type="datetime-local" />
          </div>
        </div>
        <DialogFooter>
          <Button variant="outline" @click="showEditTokenDialog = false">取消</Button>
          <Button @click="updateToken">保存</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </Tabs>
</template>
