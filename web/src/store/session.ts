import { defineStore } from 'pinia'
import { CONSTANT } from '@/constant'
import { logout } from '@/api/client'
import type { AuthSessionState, AuthSource } from '@/types/app'

type SessionStoreState = AuthSessionState

export const useSessionStore = defineStore('session', {
  state: (): SessionStoreState => {
    const token = localStorage.getItem(CONSTANT.STORE_TOKEN_NAME) || ''
    const authSource = (localStorage.getItem(CONSTANT.STORE_AUTH_SOURCE_NAME) || '') as AuthSource

    return {
      token,
      isLogin: token.trim() !== '',
      authSource,
      username: '',
      isLoggingOut: false
    }
  },
  actions: {
    setToken(this: SessionStoreState, token: string) {
      this.token = token
      this.isLogin = token.trim() !== ''
      if (this.isLogin) {
        localStorage.setItem(CONSTANT.STORE_TOKEN_NAME, token)
      } else {
        localStorage.removeItem(CONSTANT.STORE_TOKEN_NAME)
      }
    },
    setAuthSource(this: SessionStoreState, source: AuthSource) {
      this.authSource = source
      if (source) {
        localStorage.setItem(CONSTANT.STORE_AUTH_SOURCE_NAME, source)
      } else {
        localStorage.removeItem(CONSTANT.STORE_AUTH_SOURCE_NAME)
      }
    },
    syncFromStorage(this: SessionStoreState) {
      const token = localStorage.getItem(CONSTANT.STORE_TOKEN_NAME) || ''
      this.token = token
      this.isLogin = token.trim() !== ''
      this.authSource = (localStorage.getItem(CONSTANT.STORE_AUTH_SOURCE_NAME) || '') as AuthSource
    },
    clear(this: SessionStoreState) {
      this.token = ''
      this.isLogin = false
      this.authSource = ''
      this.username = ''
      localStorage.removeItem(CONSTANT.STORE_TOKEN_NAME)
      localStorage.removeItem(CONSTANT.STORE_AUTH_SOURCE_NAME)
    },
    async logout(this: SessionStoreState & { clear: () => void }) {
      this.isLoggingOut = true
      try {
        await logout()
      } finally {
        this.clear()
        this.isLoggingOut = false
      }
    }
  }
})
