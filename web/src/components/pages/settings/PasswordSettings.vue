<script setup lang="ts">
import { ref } from 'vue'
import { usersApi } from '@/api/users'
import { notifyError, notifySuccess } from '@/util/uiFeedback'

const passwordForm = ref({
  newPassword: '',
  currentPassword: ''
})

const resetPassword = async () => {
  const isDemoMode = (import.meta as any).env.VITE_RUN_MODE === 'demo'
  
  if (isDemoMode) {
    notifyError('演示模式下无法重置密码')
    return
  }

  try {
    const postData = { new_passwd: passwordForm.value.newPassword, old_passwd: passwordForm.value.currentPassword }
    const rsp = await usersApi.updatePassword(postData)
    if (rsp.data.code == 200) {
      notifySuccess(rsp.data.msg)
    } else {
      notifyError(rsp.data.msg || '密码重置失败')
    }
  } catch (error) {
    notifyError('密码重置失败，请稍后重试')
  }
}
</script>

<script lang="ts">
export default {
  name: 'PasswordSettings'
}
</script>

<template>
  <el-card shadow="never">
    <template #header>
      <div>
        <div class="text-base font-semibold">重置密码</div>
        <div class="mt-1 text-sm text-muted-foreground">更改您的登录密码</div>
      </div>
    </template>
    <el-form label-position="top" class="space-y-4" @submit.prevent="resetPassword">
      <el-form-item label="旧密码">
        <el-input
          id="current-password"
          type="password"
          v-model="passwordForm.currentPassword"
          placeholder="请输入旧密码"
          show-password
        />
      </el-form-item>
      <el-form-item label="新密码">
        <el-input
          id="new-password"
          type="password"
          v-model="passwordForm.newPassword"
          placeholder="请输入新密码"
          show-password
        />
      </el-form-item>
      <el-button type="primary" native-type="submit" class="w-full sm:w-auto">
        重置密码
      </el-button>
    </el-form>
  </el-card>
</template>
