import { ref } from 'vue'

const STORAGE_KEY = 'chat_layout'
const LAYOUTS = ['bubble', 'comfortable', 'compact', 'cozy', 'grouped']

// Module-level ref so all chat components share one setting
const layout = ref(LAYOUTS.includes(localStorage.getItem(STORAGE_KEY)) ? localStorage.getItem(STORAGE_KEY) : 'bubble')

function setLayout(l) {
  if (!LAYOUTS.includes(l)) return
  layout.value = l
  localStorage.setItem(STORAGE_KEY, l)
}

export function useChatLayout() {
  return { layout, setLayout, LAYOUTS }
}
