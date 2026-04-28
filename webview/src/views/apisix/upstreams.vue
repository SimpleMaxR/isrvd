<script lang="ts">
import { Component, Inject, Ref, Vue, toNative } from 'vue-facing-decorator'

import type { AppActions } from '@/store/state'
import { APP_ACTIONS_KEY } from '@/store/state'

import { normalizeUpstreamNodes, normalizeUpstreamType } from '@/helper/apisix'
import api from '@/service/api'
import type { ApisixUpstream } from '@/service/types'

import UpstreamEditModal from '@/views/apisix/widget/upstream-edit-modal.vue'

@Component({
    components: { UpstreamEditModal }
})
class Upstreams extends Vue {
    @Inject({ from: APP_ACTIONS_KEY }) readonly actions!: AppActions

    @Ref readonly editModalRef!: InstanceType<typeof UpstreamEditModal>

    upstreams: ApisixUpstream[] = []
    loading = false
    searchText = ''

    get filteredUpstreams() {
        if (!this.searchText) return this.upstreams
        const s = this.searchText.toLowerCase()
        return this.upstreams.filter((u: ApisixUpstream) =>
            (u.name || '').toLowerCase().includes(s) ||
            (u.id || '').toLowerCase().includes(s) ||
            (u.desc || '').toLowerCase().includes(s) ||
            (u.type || '').toLowerCase().includes(s)
        )
    }

    getNodes(upstream: ApisixUpstream) {
        return normalizeUpstreamNodes(upstream)
    }

    getNodeSummary(upstream: ApisixUpstream) {
        const nodes = this.getNodes(upstream)
        if (nodes.length === 0) return '无节点'
        const first = nodes[0]
        const firstLabel = `${first.host || '-'}:${first.port || '-'}`
        if (nodes.length === 1) return firstLabel
        return `${firstLabel} 等 ${nodes.length} 个节点`
    }

    getTypeBadgeClass(type: string) {
        const map: Record<string, string> = {
            roundrobin: 'bg-indigo-50 text-indigo-700',
            least_conn: 'bg-emerald-50 text-emerald-700',
            ewma: 'bg-amber-50 text-amber-700',
            chash: 'bg-violet-50 text-violet-700',
        }
        return map[normalizeUpstreamType(type)] || 'bg-slate-100 text-slate-600'
    }

    async loadUpstreams() {
        this.loading = true
        try {
            const res = await api.apisixListUpstreams()
            this.upstreams = (res.payload || []).sort((a: ApisixUpstream, b: ApisixUpstream) =>
                (a.name || '').localeCompare(b.name || '')
            )
        } catch {
            this.actions.showNotification('error', '加载上游列表失败')
        }
        this.loading = false
    }

    openCreateModal() {
        this.editModalRef?.show(null)
    }

    openEditModal(upstream: ApisixUpstream) {
        this.editModalRef?.show(upstream)
    }

    deleteUpstream(upstream: ApisixUpstream) {
        const id = upstream.id
        if (!id) return
        this.actions.showConfirm({
            title: '删除上游',
            message: `确定要删除上游 <strong class="text-slate-900">${upstream.name || id}</strong> 吗？请确认没有路由正在引用此上游，否则相关路由将无法正常转发。`,
            icon: 'fa-trash',
            iconColor: 'red',
            confirmText: '确认删除',
            danger: true,
            onConfirm: async () => {
                await api.apisixDeleteUpstream(id)
                this.actions.showNotification('success', '删除成功')
                this.loadUpstreams()
            }
        })
    }

    mounted() {
        this.loadUpstreams()
    }
}

export default toNative(Upstreams)
</script>

<template>
  <div>
    <div class="card mb-4 overflow-hidden">
      <div class="bg-slate-50 border-b border-slate-200 rounded-t-2xl px-4 md:px-6 py-3">
        <div class="hidden md:flex items-center justify-between">
          <div class="flex items-center gap-3">
            <div class="w-9 h-9 rounded-lg bg-emerald-500 flex items-center justify-center shadow-sm"><i class="fas fa-diagram-project text-white"></i></div>
            <div>
              <h1 class="text-lg font-semibold text-slate-800">上游管理</h1>
              <p class="text-xs text-slate-500">管理 APISIX 上游对象，供路由引用以实现集中维护</p>
            </div>
          </div>
          <div class="flex items-center gap-2">
            <div class="relative">
              <input v-model="searchText" type="text" placeholder="搜索上游名称、类型或描述..." class="pl-8 pr-3 py-1.5 text-xs border border-slate-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-emerald-500 focus:border-transparent w-64 bg-white" />
              <i class="fas fa-search absolute left-2.5 top-1/2 -translate-y-1/2 text-slate-400 text-xs"></i>
            </div>
            <button @click="loadUpstreams()" class="px-3 py-1.5 rounded-lg bg-white border border-slate-200 hover:bg-slate-50 text-slate-700 text-xs font-medium flex items-center gap-1.5 transition-colors"><i class="fas fa-rotate"></i>刷新</button>
            <button v-if="actions.hasPerm('apisix', true)" @click="openCreateModal()" class="px-3 py-1.5 rounded-lg bg-emerald-500 hover:bg-emerald-600 text-white text-xs font-medium flex items-center gap-1.5 transition-colors shadow-sm"><i class="fas fa-plus"></i>创建</button>
          </div>
        </div>

        <div class="flex md:hidden items-center justify-between">
          <div class="flex items-center gap-3 min-w-0 flex-1">
            <div class="w-9 h-9 rounded-lg bg-emerald-500 flex items-center justify-center flex-shrink-0 shadow-sm"><i class="fas fa-diagram-project text-white"></i></div>
            <div class="min-w-0">
              <h1 class="text-lg font-semibold text-slate-800 truncate">上游管理</h1>
              <p class="text-xs text-slate-500 truncate">管理 APISIX 上游对象</p>
            </div>
          </div>
          <div class="flex items-center gap-1 flex-shrink-0">
            <button @click="loadUpstreams()" class="w-9 h-9 rounded-lg bg-white border border-slate-200 hover:bg-slate-50 flex items-center justify-center text-slate-600 transition-colors" title="刷新"><i class="fas fa-rotate text-sm"></i></button>
            <button v-if="actions.hasPerm('apisix', true)" @click="openCreateModal()" class="w-9 h-9 rounded-lg bg-emerald-500 hover:bg-emerald-600 flex items-center justify-center text-white transition-colors" title="创建"><i class="fas fa-plus text-sm"></i></button>
          </div>
        </div>
      </div>

      <div class="md:hidden px-4 py-2 border-b border-slate-100 bg-white">
        <div class="relative">
          <input v-model="searchText" type="text" placeholder="搜索上游名称、类型..." class="w-full pl-8 pr-3 py-1.5 text-xs border border-slate-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-emerald-500 focus:border-transparent" />
          <i class="fas fa-search absolute left-2.5 top-1/2 -translate-y-1/2 text-slate-400 text-xs"></i>
        </div>
      </div>

      <div v-if="loading" class="flex flex-col items-center justify-center py-20">
        <div class="w-12 h-12 spinner mb-3"></div>
        <p class="text-slate-500">加载中...</p>
      </div>

      <div v-else-if="filteredUpstreams.length === 0" class="flex flex-col items-center justify-center py-20">
        <div class="w-20 h-20 rounded-full bg-slate-100 flex items-center justify-center mb-4"><i class="fas fa-diagram-project text-4xl text-slate-300"></i></div>
        <p class="text-slate-600 font-medium mb-1">暂无上游</p>
        <p class="text-sm text-slate-400">点击「创建」添加新的上游对象</p>
      </div>

      <div v-else class="space-y-3">
        <div class="hidden md:block overflow-x-auto">
          <table class="w-full border-collapse">
            <thead>
              <tr class="bg-slate-50 border-b border-slate-200">
                <th class="text-left px-4 py-3 text-xs font-semibold text-slate-600 uppercase tracking-wider">名称</th>
                <th class="text-left px-4 py-3 text-xs font-semibold text-slate-600 uppercase tracking-wider">负载均衡</th>
                <th class="text-left px-4 py-3 text-xs font-semibold text-slate-600 uppercase tracking-wider">节点</th>
                <th class="text-left px-4 py-3 text-xs font-semibold text-slate-600 uppercase tracking-wider">节点数</th>
                <th class="w-28 px-4 py-3 text-right text-xs font-semibold text-slate-600 uppercase tracking-wider">操作</th>
              </tr>
            </thead>
            <tbody class="bg-white divide-y divide-slate-100">
              <tr v-for="upstream in filteredUpstreams" :key="upstream.id" class="hover:bg-slate-50 transition-colors">
                <td class="px-4 py-3 max-w-[260px]">
                  <div class="flex items-center gap-2 min-w-0">
                    <div class="w-8 h-8 rounded-lg bg-emerald-400 flex items-center justify-center flex-shrink-0"><i class="fas fa-diagram-project text-white text-sm"></i></div>
                    <div class="min-w-0">
                      <span class="font-medium text-slate-800 truncate block">{{ upstream.name || upstream.id }}</span>
                      <span v-if="upstream.desc" class="text-xs text-slate-400 truncate block mt-0.5">{{ upstream.desc }}</span>
                    </div>
                  </div>
                </td>
                <td class="px-4 py-3 align-top">
                  <span :class="['inline-flex items-center px-2.5 py-1 rounded-full text-xs font-medium', getTypeBadgeClass(upstream.type)]">{{ upstream.type || 'roundrobin' }}</span>
                </td>
                <td class="px-4 py-3 align-top"><code class="text-xs bg-slate-100 px-2 py-1 rounded text-slate-700 break-all">{{ getNodeSummary(upstream) }}</code></td>
                <td class="px-4 py-3 align-top">
                  <span class="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs font-medium bg-slate-100 text-slate-600"><i class="fas fa-circle-nodes text-[10px]"></i>{{ getNodes(upstream).length }}</span>
                </td>
                <td class="px-4 py-3 align-top">
                  <div class="flex justify-end items-center gap-1">
                    <button v-if="actions.hasPerm('apisix', true)" @click="openEditModal(upstream)" class="btn-icon text-indigo-600 hover:bg-indigo-50" title="编辑"><i class="fas fa-pen text-xs"></i></button>
                    <button v-if="actions.hasPerm('apisix', true)" @click="deleteUpstream(upstream)" class="btn-icon text-red-600 hover:bg-red-50" title="删除"><i class="fas fa-trash text-xs"></i></button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>

        <div class="md:hidden space-y-3 p-4">
          <div v-for="upstream in filteredUpstreams" :key="upstream.id" class="rounded-xl border border-slate-200 bg-white p-4 transition-all hover:shadow-sm">
            <div class="flex items-center justify-between mb-3">
              <div class="flex items-center gap-3 min-w-0 flex-1">
                <div class="w-10 h-10 rounded-lg bg-emerald-400 flex items-center justify-center flex-shrink-0 shadow-sm"><i class="fas fa-diagram-project text-white text-base"></i></div>
                <div class="min-w-0">
                  <div class="font-medium text-sm text-slate-800 truncate">{{ upstream.name || upstream.id }}</div>
                  <div v-if="upstream.desc" class="text-xs text-slate-400 mt-0.5 truncate">{{ upstream.desc }}</div>
                </div>
              </div>
              <span :class="['inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium flex-shrink-0', getTypeBadgeClass(upstream.type)]">{{ upstream.type || 'roundrobin' }}</span>
            </div>

            <div class="flex items-start gap-2 mb-3">
              <span class="text-xs text-slate-400 flex-shrink-0 mt-0.5">节点</span>
              <code class="text-xs bg-slate-100 px-2 py-1 rounded text-slate-700 break-all">{{ getNodeSummary(upstream) }}</code>
            </div>

            <div class="flex flex-wrap gap-1 pt-2 border-t border-slate-100">
              <button v-if="actions.hasPerm('apisix', true)" @click="openEditModal(upstream)" class="btn-icon text-indigo-600 hover:bg-indigo-50" title="编辑"><i class="fas fa-pen text-xs"></i><span class="text-xs ml-1">编辑</span></button>
              <button v-if="actions.hasPerm('apisix', true)" @click="deleteUpstream(upstream)" class="btn-icon text-red-600 hover:bg-red-50" title="删除"><i class="fas fa-trash text-xs"></i><span class="text-xs ml-1">删除</span></button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <UpstreamEditModal ref="editModalRef" @success="loadUpstreams" />
  </div>
</template>
