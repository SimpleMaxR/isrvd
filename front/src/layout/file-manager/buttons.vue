<script setup>
import { inject, ref } from 'vue'

import { APP_STATE_KEY, APP_ACTIONS_KEY } from '@/store/state.js'

import MkdirModal from '@/component/modal/mkdir.vue'
import CreateModal from '@/component/modal/create.vue'
import UploadModal from '@/component/modal/upload.vue'

const state = inject(APP_STATE_KEY)
const actions = inject(APP_ACTIONS_KEY)

const mkdirModalRef = ref(null)
const createModalRef = ref(null)
const uploadModal = ref(null)

const refreshFiles = () => actions.loadFiles()
</script>

<template>
  <div>
    <div class="mb-3 d-flex flex-wrap align-items-center gap-2">
      <template v-if="state.permissions.canWrite">
        <button class="btn btn-success btn-sm" @click="mkdirModalRef.show">
          <i class="fas fa-folder me-1"></i> 新建目录
        </button>
        <button class="btn btn-primary btn-sm" @click="createModalRef.show">
          <i class="fas fa-file me-1"></i> 新建文件
        </button>
        <button class="btn btn-info btn-sm" @click="uploadModal.show">
          <i class="fas fa-upload me-1"></i> 上传文件
        </button>
      </template>
      <button v-if="state.permissions.canRead" class="btn btn-secondary btn-sm" @click="refreshFiles">
        <i class="fas fa-sync-alt me-1"></i> 刷新
      </button>
    </div>

    <!-- 模态框组件 -->
    <MkdirModal ref="mkdirModalRef" />
    <CreateModal ref="createModalRef" />
    <UploadModal ref="uploadModal" />
  </div>
</template>
