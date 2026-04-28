<script lang="ts">
import { Component, Inject, toNative, Vue } from 'vue-facing-decorator'

import type { AppActions } from '@/store/state'
import { APP_ACTIONS_KEY } from '@/store/state'

import api from '@/service/api'

import type {
    ApisixUpstream,
    ApisixUpstreamDetail,
    ApisixUpstreamFormNode,
    ApisixUpstreamHashOn,
    ApisixUpstreamType,
    DockerContainerInfo,
} from '@/service/types'

import {
    buildUpstreamPayload,
    HASH_ON_OPTIONS,
    normalizeUpstreamFormNodes,
    normalizeUpstreamType,
    UPSTREAM_TYPE_OPTIONS,
    validateUpstreamNodes
} from '@/helper/apisix'

import BaseModal from '@/component/modal.vue'

import HostSelect from './host-select.vue'
import PortSelect from './port-select.vue'

const defaultFormData = () => ({
    name: '',
    desc: '',
    type: 'roundrobin' as ApisixUpstreamType,
    hash_on: 'vars' as ApisixUpstreamHashOn,
    key: 'remote_addr',
    nodes: [{ host: '', port: '', weight: 1 }] as ApisixUpstreamFormNode[],
})

@Component({
    expose: ['show'],
    components: { BaseModal, HostSelect, PortSelect },
    emits: ['success']
})
class UpstreamEditModal extends Vue {
    @Inject({ from: APP_ACTIONS_KEY }) readonly actions!: AppActions

    isOpen = false
    modalLoading = false
    isEditMode = false
    editingId = ''
    originalUpstream: ApisixUpstreamDetail | null = null

    containers: DockerContainerInfo[] = []

    formData = defaultFormData()

    readonly upstreamTypeOptions = UPSTREAM_TYPE_OPTIONS
    readonly hashOnOptions = HASH_ON_OPTIONS

    get selectedUpstreamTypeOption() {
        return UPSTREAM_TYPE_OPTIONS.find(o => o.value === this.formData.type) ?? UPSTREAM_TYPE_OPTIONS[0]
    }

    get selectedHashOnOption() {
        return HASH_ON_OPTIONS.find(o => o.value === this.formData.hash_on) ?? HASH_ON_OPTIONS[0]
    }

    get validNodeCount() {
        return this.formData.nodes.filter(n => n.host.trim() && String(n.port).trim()).length
    }

    get validationMessage() {
        return validateUpstreamNodes(this.formData.nodes, this.formData.type, this.formData.key)
    }

    resetForm() {
        Object.assign(this.formData, defaultFormData())
        this.editingId = ''
        this.originalUpstream = null
    }

    createNode(): ApisixUpstreamFormNode {
        return { host: '', port: '', weight: 1 }
    }

    addNode() {
        this.formData.nodes = [...this.formData.nodes, this.createNode()]
    }

    removeNode(index: number) {
        const next = this.formData.nodes.filter((_, i) => i !== index)
        this.formData.nodes = next.length > 0 ? next : [this.createNode()]
    }

    getPortsByHost(host: string): string[] {
        return this.containers.find(c => c.name === host.trim())?.ports || []
    }

    updateNode(index: number, field: 'host' | 'port', value: string) {
        const next = [...this.formData.nodes]
        const node = next[index]
        if (!node) return
        node[field] = value
        if (field === 'host' && value.trim()) {
            const port = (this.getPortsByHost(value)[0] || '').split('/')[0].split(':').pop() || ''
            if (port) node.port = port
        }
        this.formData.nodes = next
    }

    async loadContainers() {
        try {
            const ct = await api.listContainers()
            this.containers = (ct.payload || []).filter((c: DockerContainerInfo) => c.state === 'running')
        } catch {}
    }

    async show(upstream: ApisixUpstream | null) {
        await this.loadContainers()
        if (upstream?.id) {
            this.isEditMode = true
            this.resetForm()
            this.editingId = upstream.id
            this.modalLoading = true
            this.isOpen = true
            try {
                const detail = (await api.apisixGetUpstream(upstream.id)).payload
                if (!detail) {
                    this.actions.showNotification('error', '加载上游详情失败')
                    this.isOpen = false
                    this.modalLoading = false
                    return
                }
                this.originalUpstream = detail
                Object.assign(this.formData, {
                    name: detail.name || '',
                    desc: detail.desc || '',
                    type: normalizeUpstreamType(String(detail.type || '')),
                    hash_on: (detail.hash_on as ApisixUpstreamHashOn) || 'vars',
                    key: String(detail.key || 'remote_addr'),
                    nodes: normalizeUpstreamFormNodes(detail),
                })
            } catch {
                this.actions.showNotification('error', '加载上游详情失败')
                this.isOpen = false
            }
            this.modalLoading = false
            return
        }
        this.isEditMode = false
        this.resetForm()
        this.isOpen = true
    }

    async handleConfirm() {
        if (!this.formData.name.trim()) return this.actions.showNotification('error', '上游名称不能为空')
        if (this.validationMessage) return this.actions.showNotification('error', this.validationMessage)

        const payload = buildUpstreamPayload(this.formData, this.originalUpstream)

        this.modalLoading = true
        try {
            if (this.isEditMode) {
                await api.apisixUpdateUpstream(this.editingId, payload)
                this.actions.showNotification('success', '上游更新成功')
            } else {
                await api.apisixCreateUpstream(payload)
                this.actions.showNotification('success', '上游创建成功')
            }
            this.isOpen = false
            this.resetForm()
            this.$emit('success')
        } catch (e: unknown) {
            this.actions.showNotification('error', (e instanceof Error ? e.message : '') || '操作失败')
        }
        this.modalLoading = false
    }
}

export default toNative(UpstreamEditModal)
</script>

<template>
  <BaseModal v-model="isOpen" :title="isEditMode ? '编辑上游' : '创建上游'" :loading="modalLoading">
    <div class="space-y-5 p-1">
      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        <div>
          <label class="block text-xs font-semibold text-slate-500 uppercase tracking-wider mb-1">上游名称</label>
          <input v-model="formData.name" type="text" class="input" placeholder="例如：member-service" />
          <p class="text-xs text-slate-400 mt-1">用于在路由中引用此上游时识别。</p>
        </div>
        <div>
          <label class="block text-xs font-semibold text-slate-500 uppercase tracking-wider mb-1">描述</label>
          <input v-model="formData.desc" type="text" class="input" placeholder="补充上游用途或业务说明" />
        </div>
      </div>

      <div class="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm space-y-4">
        <div>
          <label class="block text-xs font-semibold text-slate-500 uppercase tracking-wider mb-1">负载均衡策略</label>
          <select v-model="formData.type" class="input">
            <option v-for="o in upstreamTypeOptions" :key="o.value" :value="o.value">{{ o.label }}</option>
          </select>
          <p class="text-xs text-slate-400 mt-1">{{ selectedUpstreamTypeOption.desc }}</p>
        </div>

        <div v-if="formData.type === 'chash'" class="rounded-2xl border border-violet-200 bg-violet-50/40 p-4 space-y-3">
          <div class="flex items-center gap-2 mb-1">
            <i class="fas fa-fingerprint text-violet-500 text-sm"></i>
            <span class="text-sm font-semibold text-violet-800">一致性哈希参数</span>
          </div>
          <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
            <div>
              <label class="block text-xs font-semibold text-slate-500 uppercase tracking-wider mb-1">哈希依据（hash_on）</label>
              <select v-model="formData.hash_on" class="input">
                <option v-for="o in hashOnOptions" :key="o.value" :value="o.value">{{ o.label }}</option>
              </select>
              <p class="text-xs text-slate-400 mt-1">决定用什么来计算哈希值</p>
            </div>
            <div>
              <label class="block text-xs font-semibold text-slate-500 uppercase tracking-wider mb-1">哈希键（key）</label>
              <input v-model="formData.key" type="text" class="input" :placeholder="selectedHashOnOption.keyPlaceholder" />
              <p class="text-xs text-slate-400 mt-1">{{ selectedHashOnOption.keyHint }}</p>
            </div>
          </div>
        </div>
      </div>

      <div class="space-y-3">
        <div class="flex flex-col md:flex-row md:items-center md:justify-between gap-3 rounded-2xl bg-white border border-slate-200 p-4 shadow-sm">
          <div class="grid grid-cols-2 gap-3 flex-1">
            <div class="rounded-xl bg-slate-50 border border-slate-200 px-3 py-2">
              <div class="text-[11px] uppercase tracking-wider text-slate-400 font-semibold">节点行数</div>
              <div class="text-lg font-semibold text-slate-800 mt-1">{{ formData.nodes.length }}</div>
            </div>
            <div class="rounded-xl bg-slate-50 border border-slate-200 px-3 py-2">
              <div class="text-[11px] uppercase tracking-wider text-slate-400 font-semibold">有效节点</div>
              <div class="text-lg font-semibold text-indigo-600 mt-1">{{ validNodeCount }}</div>
            </div>
          </div>
          <button type="button" @click="addNode()" class="inline-flex items-center justify-center gap-2 px-4 py-2 rounded-xl bg-indigo-500 hover:bg-indigo-600 text-white text-sm font-medium transition-colors shadow-sm">
            <i class="fas fa-plus"></i>添加节点
          </button>
        </div>

        <div v-for="(node, index) in formData.nodes" :key="`node-${index}`" class="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
          <div class="flex items-center justify-between gap-3 mb-4">
            <div class="flex items-center gap-3 min-w-0">
              <div class="w-9 h-9 rounded-xl bg-indigo-100 text-indigo-700 flex items-center justify-center font-semibold text-sm">{{ index + 1 }}</div>
              <div>
                <div class="text-sm font-semibold text-slate-800">上游节点 {{ index + 1 }}</div>
                <div class="text-xs text-slate-400">支持容器选择或直接手动输入</div>
              </div>
            </div>
            <button type="button" @click="removeNode(index)" class="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-red-50 text-red-600 hover:bg-red-100 text-xs font-medium transition-colors">
              <i class="fas fa-trash"></i>删除
            </button>
          </div>
          <div class="grid grid-cols-1 md:grid-cols-12 gap-4 items-start">
            <div class="md:col-span-5">
              <label class="block text-xs font-semibold text-slate-500 uppercase tracking-wider mb-1">上游主机</label>
              <HostSelect :model-value="node.host" :containers="containers" placeholder="127.0.0.1 或 容器名" @update:modelValue="updateNode(index, 'host', $event)" />
            </div>
            <div class="md:col-span-4">
              <label class="block text-xs font-semibold text-slate-500 uppercase tracking-wider mb-1">上游端口</label>
              <PortSelect :model-value="node.port" :ports="getPortsByHost(node.host)" placeholder="8080" @update:modelValue="updateNode(index, 'port', $event)" />
            </div>
            <div class="md:col-span-3">
              <label class="block text-xs font-semibold text-slate-500 uppercase tracking-wider mb-1">节点权重</label>
              <input v-model.number="node.weight" type="number" min="0" class="input" placeholder="1" />
              <p class="text-xs text-slate-400 mt-1">当前策略 {{ formData.type }} 下的节点权重。</p>
            </div>
          </div>
        </div>

        <div v-if="validationMessage" class="rounded-xl border border-amber-200 bg-amber-50 px-3 py-2 text-xs text-amber-700">
          {{ validationMessage }}
        </div>
      </div>
    </div>

    <template #footer>
      <div class="flex justify-end gap-2">
        <button @click="isOpen = false" class="px-4 py-2 text-sm font-medium text-slate-700 bg-white border border-slate-200 rounded-lg hover:bg-slate-50">取消</button>
        <button @click="handleConfirm()" :disabled="modalLoading" class="px-4 py-2 text-sm font-medium text-white bg-indigo-500 rounded-lg hover:bg-indigo-600 disabled:opacity-50 shadow-sm">
          <i v-if="modalLoading" class="fas fa-spinner fa-spin mr-1"></i>{{ isEditMode ? '保存' : '创建' }}
        </button>
      </div>
    </template>
  </BaseModal>
</template>
