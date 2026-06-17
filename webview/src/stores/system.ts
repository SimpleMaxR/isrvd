import { defineStore } from 'pinia'
import { ref, reactive } from 'vue'

import type { BootstrapData, LinkConfig } from '@/service/types'

interface ServiceAvailability {
    agent: boolean
    apisix: boolean
    caddy: boolean
    docker: boolean
    swarm: boolean
    compose: boolean
}

/**
 * 系统 Store
 *
 * 持有服务可用性、版本信息、系统配置等运行时状态。
 * 数据通过 apply(data) 由 portal 统一注入，不直接发起网络请求。
 */
export const useSystemStore = defineStore('system', () => {
    // ─── 状态定义 ───

    const initialized = ref(false)
    const initError = ref<string | null>(null)
    const serviceAvailability = reactive<ServiceAvailability>({
        agent: false,
        apisix: false,
        caddy: false,
        docker: false,
        swarm: false,
        compose: false,
    })
    const toolbarLinks = ref<LinkConfig[]>([])
    const maxUploadSize = ref<number>(104857600) // 默认 100MB

    // ─── 操作定义 ───

    // apply 将 bootstrap 响应中的 probe 和 config 写入 store
    function apply(data: BootstrapData) {
        const { probe, config } = data

        if (probe) {
            Object.assign(serviceAvailability, {
                agent: probe.agent || false,
                apisix: probe.apisix || false,
                caddy: probe.caddy || false,
                docker: probe.docker || false,
                swarm: probe.swarm || false,
                compose: probe.compose || false,
            })
        }

        if (config) {
            if (typeof config.maxUploadSize === 'number') {
                maxUploadSize.value = config.maxUploadSize
            }
            toolbarLinks.value = config.links || []
        }
    }

    function hasPerm(module: string, founder: boolean, permissions: string[]): boolean {
        const checkAvailability = (seg: string): boolean => {
            const key = seg as keyof ServiceAvailability
            return !(key in serviceAvailability && !serviceAvailability[key])
        }

        if (module.includes(' ')) {
            // 旧格式 "METHOD /api/path"：检查服务可用性 + 精确匹配权限
            const path = module.split(' ')[1]
            const seg = path?.match(/^\/api\/([^/]+)/)?.[1]
            if (seg && !checkAvailability(seg)) return false
            return founder || permissions.includes(module)
        }

        // 模块名匹配（如 "docker"）：检查服务可用性 + 模糊匹配权限
        if (!checkAvailability(module)) return false
        if (founder) return true
        return permissions.some(key => {
            // 兼容旧 "METHOD /api/path" 格式
            if (key.includes(' ')) {
                const path = key.split(' ')[1]
                return !!(path && (path.startsWith(`/api/${module}/`) || path === `/api/${module}`))
            }
            // 稳定权限 ID 格式（如 "docker.container.list"）
            return key.startsWith(module + '.')
        })
    }

    return {
        // 状态
        initialized,
        initError,
        serviceAvailability,
        toolbarLinks,
        maxUploadSize,
        // 操作
        apply,
        hasPerm,
    }
})

// ─── 类型导出 ───
export type SystemStore = ReturnType<typeof useSystemStore>
