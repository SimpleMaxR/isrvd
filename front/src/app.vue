<script setup>
import { onMounted, provide } from 'vue'

import api from '@/service/api.js'
import { initProvider, APP_STATE_KEY, APP_ACTIONS_KEY } from '@/store/state.js'

import NavigationBar from '@/layout/navigation.vue'
import NotificationManager from '@/layout/notification.vue'

import AuthLogin from '@/layout/login.vue'
import FileManager from '@/layout/file-manager/index.vue'

const { state, actions } = initProvider()

// 提供状态和动作给子组件
provide(APP_STATE_KEY, state)
provide(APP_ACTIONS_KEY, actions)

onMounted(async () => {
  const token = localStorage.getItem('app-token')
  const username = localStorage.getItem('app-username')
  const role = localStorage.getItem('app-role')
  const permissionsStr = localStorage.getItem('app-permissions')

  if (token) {
    // 先恢复本地状态，保证快速响应
    const authData = { token, username, role }
    if (permissionsStr) {
      try {
        authData.permissions = JSON.parse(permissionsStr)
      } catch (e) {
        console.error('Failed to parse permissions', e)
      }
    }
    actions.setAuth(authData)
    
    // 再进行后端验证并更新状态
    try {
      const res = await api.getMe()
      if (res.success) {
        actions.setAuth({ token, ...res.payload })
      } else {
        actions.clearAuth()
      }
    } catch (e) {
      actions.clearAuth()
    }
  }
})
</script>

<template>
  <template v-if="state.username">
    <NavigationBar />
    <div class="container-fluid">
      <FileManager />
    </div>
  </template>

  <AuthLogin v-else />

  <NotificationManager />
</template>
