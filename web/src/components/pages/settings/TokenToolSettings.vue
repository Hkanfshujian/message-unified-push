<script setup lang="ts">
import { reactive } from 'vue'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { TokenEncryption } from '@/util/viewApi'
import { toast } from 'vue-sonner'
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
    toast.error('token解析失败，请检查是否正确')
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
    <div class="rounded-lg border border-slate-300/80 dark:border-slate-700 p-4 space-y-3 bg-white/70 dark:bg-slate-900/30">
      <div class="space-y-6">
        <!-- 自动生成模板ID（独立区域） -->
        <div class="space-y-3">
          <div class="text-sm font-medium text-gray-700">生成模板ID</div>
          <div class="flex items-center gap-2">
            <Input
              v-model="tokenTool.generatedTemplateId"
              placeholder="自动生成的模板ID（TP开头）"
              class="h-8"
              readonly
            />
            <Button type="button" size="sm" variant="outline" @click="autoGenerateTemplateId">
              生成模板ID
            </Button>
          </div>
        </div>

        <!-- 模板ID -> token -->
        <div class="space-y-3">
          <div class="text-sm font-medium text-gray-700">模板ID → token</div>
          <div class="space-y-2">
            <div class="flex items-center gap-2">
              <Input
                v-model="tokenTool.templateIdInput"
                placeholder="输入或粘贴模板ID，例如 TPxxxx"
                class="h-8"
              />
              <Button type="button" size="sm" @click="generateTokenFromTemplateId">
                生成 token
              </Button>
            </div>
            <Input
              v-model="tokenTool.tokenFromId"
              placeholder="生成的 token"
              class="h-8 text-xs"
              readonly
            />
          </div>
        </div>

        <!-- token -> 模板ID -->
        <div class="space-y-3">
          <div class="text-sm font-medium text-gray-700">token → 模板ID</div>
          <div class="space-y-2">
            <div class="flex items-center gap-2">
              <Input
                v-model="tokenTool.tokenInput"
                placeholder="输入 token"
                class="h-8 text-xs"
              />
              <Button type="button" size="sm" @click="decodeTemplateIdFromToken">
                解析模板ID
              </Button>
            </div>
            <Input
              v-model="tokenTool.templateIdFromToken"
              placeholder="解析出的模板ID"
              class="h-8 text-xs"
              readonly
            />
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
