import { ref } from 'vue'

// Module-level so App.vue can watch it for the title indicator
const count = ref(0)

// The most recent project-chat message that raised the counter — surfaced in
// the tab-title blink (App.vue) so it can say which project it came from
// instead of a bare, unattributed "New message!". Cleared together with the
// counter since it only describes the latest contributor to it, not a queue.
const source = ref(null) // { projectName, projectSlug, senderName } | null

export function useProjectChatUnread() {
  return {
    projectChatUnread: count,
    projectChatUnreadSource: source,
    addProjectChatUnread: (info) => {
      count.value++
      if (info) source.value = info
    },
    clearProjectChatUnread: () => {
      count.value = 0
      source.value = null
    },
  }
}
