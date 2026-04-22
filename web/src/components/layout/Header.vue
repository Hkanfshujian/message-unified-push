<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { 
  Search, 
  User, 
  LogOut, 
  Monitor, 
  Moon, 
  Sun 
} from 'lucide-vue-next'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip'

const props = defineProps<{
  userAccount: string
  theme: 'light' | 'dark'
  themePreference: 'light' | 'dark' | 'system'
}>()

const emit = defineEmits<{
  (e: 'toggle-theme'): void
  (e: 'logout'): void
}>()

const isUserMenuOpen = ref(false)
const userMenuRef = ref<HTMLElement | null>(null)

const toggleUserMenu = () => {
  isUserMenuOpen.value = !isUserMenuOpen.value
}

const handleClickOutside = (event: MouseEvent) => {
  if (userMenuRef.value && !userMenuRef.value.contains(event.target as Node)) {
    isUserMenuOpen.value = false
  }
}

onMounted(() => {
  document.addEventListener('click', handleClickOutside)
})

onUnmounted(() => {
  document.removeEventListener('click', handleClickOutside)
})

const themeLabel = () => {
  if (props.themePreference === 'system') return '跟随系统'
  return props.theme === 'dark' ? '深色' : '浅色'
}
</script>

<template>
  <header class="sticky top-0 z-40 w-full border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
    <div class="flex h-16 items-center px-4">
      <!-- 搜索框 (居左) -->
      <div class="flex items-center flex-1 md:w-auto md:flex-none mr-4">
        <div class="relative w-full max-w-sm">
          <Search class="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
          <Input
            type="search"
            placeholder="搜索..."
            class="pl-8 md:w-[300px] lg:w-[400px]"
          />
        </div>
      </div>

      <!-- 右侧区域 -->
      <div class="ml-auto flex items-center space-x-4">
        <!-- 主题切换 -->
        <TooltipProvider>
          <Tooltip>
            <TooltipTrigger as-child>
              <Button variant="ghost" size="icon" @click="emit('toggle-theme')">
                <Monitor v-if="themePreference === 'system'" class="h-5 w-5" />
                <Moon v-else-if="theme === 'dark'" class="h-5 w-5" />
                <Sun v-else class="h-5 w-5" />
                <span class="sr-only">切换主题</span>
              </Button>
            </TooltipTrigger>
            <TooltipContent>
              <p>切换主题 (当前: {{ themeLabel() }})</p>
            </TooltipContent>
          </Tooltip>
        </TooltipProvider>

        <!-- 用户头像 + 下拉菜单 -->
        <div class="relative" ref="userMenuRef">
          <Button 
            variant="ghost" 
            class="relative h-8 w-8 rounded-full"
            @click="toggleUserMenu"
          >
            <div class="h-8 w-8 rounded-full bg-primary/10 flex items-center justify-center border border-border">
              <User class="h-4 w-4 text-primary" />
            </div>
          </Button>

          <!-- 下拉菜单 -->
          <div 
            v-if="isUserMenuOpen"
            class="absolute right-0 mt-2 w-56 rounded-md border bg-popover text-popover-foreground shadow-md outline-none animate-in fade-in-0 zoom-in-95 data-[side=bottom]:slide-in-from-top-2 data-[side=left]:slide-in-from-right-2 data-[side=right]:slide-in-from-left-2 data-[side=top]:slide-in-from-bottom-2"
          >
            <div class="px-2 py-1.5 text-sm font-semibold">
              <div class="flex flex-col space-y-1">
                <p class="text-sm font-medium leading-none">{{ userAccount }}</p>
                <p class="text-xs leading-none text-muted-foreground">
                  已登录用户
                </p>
              </div>
            </div>
            <div class="h-px bg-muted my-1" />
            <button
              @click="emit('logout')"
              class="relative flex w-full cursor-pointer select-none items-center rounded-sm px-2 py-1.5 text-sm outline-none transition-colors hover:bg-accent hover:text-accent-foreground data-[disabled]:pointer-events-none data-[disabled]:opacity-50 text-red-600 focus:text-red-600"
            >
              <LogOut class="mr-2 h-4 w-4" />
              <span>退出登录</span>
            </button>
          </div>
        </div>
      </div>
    </div>
  </header>
</template>
