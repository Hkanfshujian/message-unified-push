// store/index.js
import { createPinia } from 'pinia';

const pinia = createPinia();

export default pinia;
export { useAppStore } from './app'
export { useRbacStore } from './rbac'
export { useSessionStore } from './session'
export { useMessageCenterStore } from './message-center'
