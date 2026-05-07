import type { RepoConfig, Task } from '@/api'

/**
 * 仓库命令解析结果
 */
export interface ParsedRepoResult {
  repoConfig: Partial<RepoConfig>
  task: Partial<Task>
}

/**
 * 解析带引号的命令行字符串为参数数组
 */
export function parseArgs(command: string): string[] {
  const args: string[] = []
  const regex = /[^\s"']+|"([^"]*)"|'([^']*)'/g
  let match
  while ((match = regex.exec(command)) !== null) {
    args.push(match[1] || match[2] || match[0])
  }
  return args
}

/**
 * 解析 Baihu 格式的导入命令
 */
export function parseBaihuCommand(command: string): ParsedRepoResult | null {
  const s = command.trim()
  if (!s) return null

  const args = parseArgs(s)
  let i = 0
  // 跳过开头的 'baihu' 或 'reposync'
  if (args[i] === 'baihu') i++
  if (args[i] === 'reposync') i++

  const repoConfig: Partial<RepoConfig> = {}
  const task: Partial<Task> = {}
  let hasValidField = false

  for (; i < args.length; i++) {
    const arg = args[i]
    if (!arg || !arg.startsWith('--')) continue

    const value = args[i + 1]
    if (value === undefined || value.startsWith('--')) continue

    i++ // 跳过已处理的 value
    hasValidField = true

    switch (arg) {
      case '--source-type':
        repoConfig.source_type = value
        break
      case '--source-url':
        repoConfig.source_url = value
        // 如果 URL 存在，尝试生成默认名称
        if (value) {
          try {
            const urlPaths = value.split('/')
            const name = urlPaths[urlPaths.length - 1]?.replace('.git', '') || '未命名仓库'
            task.name = '同步 ' + name
          } catch { /* ignore */ }
        }
        break
      case '--target-path':
        // 处理 $SCRIPTS_DIR$ 占位符
        if (value.startsWith('$SCRIPTS_DIR$/')) {
          repoConfig.target_path = value.replace('$SCRIPTS_DIR$/', '')
        } else if (value === '$SCRIPTS_DIR$') {
          repoConfig.target_path = ''
        } else {
          repoConfig.target_path = value
        }
        break
      case '--branch':
        repoConfig.branch = value
        break
      case '--path':
        repoConfig.sparse_path = value
        break
      case '--single-file':
        repoConfig.single_file = value === 'true'
        break
      case '--proxy-url':
        repoConfig.proxy_url = value
        repoConfig.proxy = 'custom'
        break
      case '--auth-token':
        repoConfig.auth_token = value
        break
      case '--whitelist-paths':
        repoConfig.whitelist_paths = value
        break
      case '--blacklist':
        repoConfig.blacklist = value
        break
      case '--dependence':
        repoConfig.dependence = value
        break
      case '--extensions':
        repoConfig.extensions = value
        break
      case '--task-timeout':
        task.timeout = parseInt(value) || 30
        break
      case '--task-langs':
        try {
          const langs = JSON.parse(value)
          if (Array.isArray(langs)) {
            task.languages = langs.map(l => ({
              name: l.name || '',
              version: l.version || ''
            }))
          }
        } catch (e) {
          console.error('Parse task-langs failed', e)
        }
        break
    }
  }

  if (!hasValidField) return null

  // 设置默认规则
  repoConfig.auto_add_cron = true
  repoConfig.commenttotask = 'true'

  return { repoConfig, task }
}

/**
 * 解析 青龙 (Qinglong) 格式的导入命令
 */
export function parseQlCommand(command: string): ParsedRepoResult | null {
  const s = command.trim()
  if (!s || !s.startsWith('ql repo')) return null

  const args = parseArgs(s)
  const repoConfig: Partial<RepoConfig> = {}
  const task: Partial<Task> = {}

  if (args[2]) {
    repoConfig.source_url = args[2]
    repoConfig.source_type = 'git'
    // 生成名称
    try {
      const urlPaths = args[2].split('/')
      const name = urlPaths.length > 0 ? urlPaths[urlPaths.length - 1]?.replace('.git', '') : '未命名仓库'
      task.name = '同步 ' + (name || '未命名仓库')
    } catch {
      task.name = '同步 未命名仓库'
    }
  }

  if (args[3]) repoConfig.whitelist_paths = args[3]
  if (args[4]) repoConfig.blacklist = args[4]
  if (args[5]) repoConfig.dependence = args[5]
  if (args[6]) repoConfig.branch = args[6]
  if (args[7]) repoConfig.extensions = args[7]

  repoConfig.auto_add_cron = true
  repoConfig.commenttotask = 'true'
  repoConfig.repo_source = 'ql'

  return { repoConfig, task }
}
