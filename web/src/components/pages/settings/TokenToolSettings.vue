<script setup lang="ts">
import { reactive } from 'vue'
import { TokenEncryption } from '@/util/viewApi'
import { notifyError } from '@/util/uiFeedback'
import { generateBizUniqueID } from '@/util/uuid'

const tokenTool = reactive({
  generatedTemplateId: '',
  templateIdInput: '',
  tokenFromId: '',
  tokenInput: '',
  templateIdFromToken: ''
})

const autoGenerateTemplateId = () => {
  const id = generateBizUniqueID('TP')
  tokenTool.generatedTemplateId = id
}

const generateTokenFromTemplateId = () => {
  const id = tokenTool.templateIdInput.trim()
  if (!id) {
    tokenTool.tokenFromId = ''
    return
  }
  tokenTool.tokenFromId = TokenEncryption.encryptHex(id, 71)
}

const decodeTemplateIdFromToken = () => {
  const token = tokenTool.tokenInput.trim()
  if (!token) {
    tokenTool.templateIdFromToken = ''
    return
  }
  try {
    tokenTool.templateIdFromToken = TokenEncryption.decryptHex(token, 71)
  } catch (e) {
    tokenTool.templateIdFromToken = ''
    notifyError('token解析失败，请检查是否正确')
  }
}
</script>

<script lang="ts">
export default {
  name: 'TokenToolSettings'
}
</script>

<template>
  <div class="space-y-5">
    <div>
      <div class="text-lg font-semibold">加解密工具</div>
      <div class="text-sm text-muted-foreground">在这里可以方便地进行模板ID与加密token之间的转换，用于对接接口调试。</div>
    </div>
    <el-card shadow="never" class="settings-section-card">
      <div class="space-y-6">
        <div class="space-y-3">
          <div class="text-sm font-medium text-foreground">生成模板ID</div>
          <div class="flex items-center gap-2">
            <el-input
              v-model="tokenTool.generatedTemplateId"
              placeholder="自动生成的模板ID（TP开头）"
              class="h-8"
              readonly
            />
            <el-button native-type="button" @click="autoGenerateTemplateId">
              生成模板ID
            </el-button>
          </div>
        </div>

        <div class="space-y-3">
          <div class="text-sm font-medium text-foreground">模板ID → token</div>
          <div class="space-y-2">
            <div class="flex items-center gap-2">
              <el-input
                v-model="tokenTool.templateIdInput"
                placeholder="输入或粘贴模板ID，例如 TPxxxx"
                class="h-8"
              />
              <el-button type="primary" native-type="button" @click="generateTokenFromTemplateId">
                生成 token
              </el-button>
            </div>
            <el-input
              v-model="tokenTool.tokenFromId"
              placeholder="生成的 token"
              class="h-8 text-xs"
              readonly
            />
          </div>
        </div>

        <div class="space-y-3">
          <div class="text-sm font-medium text-foreground">token → 模板ID</div>
          <div class="space-y-2">
            <div class="flex items-center gap-2">
              <el-input
                v-model="tokenTool.tokenInput"
                placeholder="输入 token"
                class="h-8 text-xs"
              />
              <el-button type="primary" native-type="button" @click="decodeTemplateIdFromToken">
                解析模板ID
              </el-button>
            </div>
            <el-input
              v-model="tokenTool.templateIdFromToken"
              placeholder="解析出的模板ID"
              class="h-8 text-xs"
              readonly
            />
          </div>
        </div>
      </div>
    </el-card>
  </div>
</template>

<style scoped>
:deep(.settings-section-card > .el-card__body) {
  padding: 16px;
}
</style>
