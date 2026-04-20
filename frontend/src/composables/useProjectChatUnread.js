import { ref } from 'vue'

// Module-level so App.vue can watch it for the title indicator
const count = ref(0)

export function useProjectChatUnread() {
  return {
    projectChatUnread: count,
    addProjectChatUnread: () => { count.value++ },
    clearProjectChatUnread: () => { count.value = 0 },
  }
}
