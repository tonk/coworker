<template>
  <div class="emoji-panel" ref="panelEl">
    <!-- Category tabs -->
    <div class="emoji-cats">
      <button
        v-for="cat in CATEGORIES"
        :key="cat.name"
        class="emoji-cat-btn"
        :class="{ active: activeCat === cat.name }"
        :title="cat.name"
        :aria-label="cat.name"
        @click="activeCat = cat.name"
      >{{ cat.icon }}</button>
    </div>
    <!-- Search -->
    <div class="emoji-search-wrap">
      <input
        class="emoji-search"
        v-model="search"
        placeholder="Search…"
        aria-label="Search emoji"
        ref="searchEl"
        @keydown.enter.prevent="onEnterSearch"
        @keydown.escape.prevent.stop="onEscapeSearch"
      />
    </div>
    <!-- Grid -->
    <div class="emoji-grid">
      <button
        v-for="e in visibleEmojis"
        :key="e"
        class="emoji-btn"
        @mousedown.prevent="pick(e)"
        :title="e"
      >{{ e }}</button>
      <span v-if="!visibleEmojis.length" class="emoji-empty">No results</span>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount, nextTick } from 'vue'
import { gemoji } from 'gemoji'

const props = defineProps({ initialSearch: { type: String, default: '' } })
const emit = defineEmits(['pick', 'close', 'escape'])

const panelEl = ref(null)
const searchEl = ref(null)
const search = ref(props.initialSearch || '')
const activeCat = ref('Smileys')

const CATEGORIES = [
  { icon: '😀', name: 'Smileys', emojis: ['😀','😁','😂','🤣','😃','😄','😅','😆','😉','😊','😋','😎','😍','🥰','😘','🙂','🤗','🤔','😐','🙄','😏','😒','😔','😟','😢','😭','😤','😠','😡','🤬','😱','😳','🥺','😴','🤒','🤧','🥳','🤩','🥸','🫠','😵','🥴','🤪','😜','😛','🤭','🫣','🫢','🤫','🫡','🤐','🫥'] },
  { icon: '👋', name: 'People',  emojis: ['👋','🤚','🖐','✋','🤙','👌','🤌','✌️','🤞','🤟','🤘','👈','👉','👆','👇','👍','👎','✊','👊','🤛','🤜','👏','🙌','👐','🙏','🤲','💪','🦾','💅','🤳','💑','👫','👨‍👩‍👧','❤️','🧡','💛','💚','💙','💜','🖤','🤍','🤎','💔','❣️','💕','💞','💓','💗','💖','💘','💝','💯'] },
  { icon: '🐶', name: 'Animals', emojis: ['🐶','🐱','🐭','🐹','🐰','🦊','🐻','🐼','🐨','🐯','🦁','🐮','🐷','🐸','🐵','🐔','🐧','🐦','🦆','🦉','🦇','🐝','🦋','🐌','🐞','🐢','🐍','🦎','🐙','🦑','🦈','🐬','🐳','🦒','🦓','🐘','🦄','🌵','🌲','🌳','🌴','🌺','🌸','🌻','🌹','⭐','🌟','🌈','❄️','🔥','💧','🌊','⚡','🌙','☀️','🌤'] },
  { icon: '🍕', name: 'Food',    emojis: ['🍎','🍊','🍋','🍌','🍉','🍇','🍓','🫐','🍒','🍑','🥭','🍍','🥥','🥝','🍅','🍆','🥑','🥦','🥕','🌽','🌶','🍄','🥐','🍞','🧀','🥚','🍳','🥞','🧇','🥓','🥩','🍔','🍟','🍕','🌮','🌯','🥗','🍜','🍣','🍱','🍩','🍪','🎂','🍰','🍦','🍫','🍬','🍭','☕','🍵','🥤','🧋','🍺','🍷','🥂','🫖'] },
  { icon: '⚽', name: 'Activity',emojis: ['⚽','🏀','🏈','⚾','🥎','🏐','🏉','🎾','🏓','🏸','🥊','🥋','🎯','⛳','🎣','🏊','🚵','🏋️','🤸','⛷','🏂','🏄','🎽','🎮','🕹','🎲','🧩','🎨','🎸','🎷','🎺','🎻','🥁','🎵','🎶','🎤','🎧','🎬','🎭','🎉','🎊','🎁','🏆','🥇','🎈','🪄','🎇','🎆','🎃','🎄'] },
  { icon: '🚗', name: 'Travel',  emojis: ['🚗','🚕','🚙','🚌','🏎','🚑','🚒','🚜','🏍','🛵','🚲','🛴','⛽','🚨','🚦','⚓','🛶','⛵','🚤','🚢','✈','🛩','🚀','🛸','🚁','🏠','🏡','🏢','🏥','🏦','🏨','🏫','🏭','🏰','⛺','🌁','🌃','🏙','🌅','🌆','🌇','🌉','🌌','🗺','🧭','🗼','🗽','⛪','🕌','🏛'] },
  { icon: '💡', name: 'Objects', emojis: ['💡','🔦','🕯','💰','💳','🪙','📝','📋','📁','📱','💻','🖥','⌨','🖱','💾','💿','🔭','🔬','🔍','💊','💉','🩺','🩹','🔑','🗝','🔒','🔓','🔨','⚙','🔧','🔩','🪛','🧰','🔗','🧲','📡','🎁','🎀','🧸','🪆','🪅','🎏','🧪','🧫','🧬','🪤','🧲','🔋','💡','📚','📖','📰'] },
  { icon: '✅', name: 'Symbols', emojis: ['✅','❌','❓','❗','‼️','⚠️','🚫','⛔','🔞','♻️','💯','✨','💫','💥','🎵','🎶','💬','💭','🗯','➕','➖','✖️','➗','♾️','🔄','▶️','⏩','⏪','⏸','⏹','🔔','🔕','📢','📣','🔅','🔆','🆒','🆕','🆓','🔝','🆙','🆗','💲','🔱','⚜️','🔰','🅰️','🅱️','🆎','🅾️','🆑'] },
]

const ALL_EMOJIS = Array.from(new Set(CATEGORIES.flatMap(c => c.emojis)))

const EMOJI_META = new Map()
for (const entry of gemoji) {
  if (!EMOJI_META.has(entry.emoji)) {
    EMOJI_META.set(entry.emoji, { names: [], tags: [], descriptions: [] })
  }
  const meta = EMOJI_META.get(entry.emoji)
  meta.names.push(...(entry.names || []))
  meta.tags.push(...(entry.tags || []))
  if (entry.description) meta.descriptions.push(entry.description)
}

function normalizeQuery(q) {
  return String(q || '')
    .trim()
    .toLowerCase()
    .replace(/^:/, '')
    .replace(/:$/, '')
}

function matchesSearch(emoji, query) {
  if (!query) return true
  if (emoji.includes(query)) return true

  const meta = EMOJI_META.get(emoji)
  if (!meta) return false

  const names = [...meta.names, ...meta.tags]
  for (const name of names) {
    const n = String(name || '').toLowerCase()
    if (!n) continue
    if (n.includes(query)) return true
    if (n.replace(/_/g, '').includes(query.replace(/_/g, ''))) return true
  }

  return meta.descriptions.some(d => String(d || '').toLowerCase().includes(query))
}

const visibleEmojis = computed(() => {
  if (search.value.trim()) {
    const q = normalizeQuery(search.value)
    return ALL_EMOJIS.filter(e => matchesSearch(e, q)).slice(0, 120)
  }
  return CATEGORIES.find(c => c.name === activeCat.value)?.emojis || []
})

function pick(emoji) {
  emit('pick', emoji)
}

function onEnterSearch() {
  if (visibleEmojis.value.length === 1) {
    pick(visibleEmojis.value[0])
  }
}

function onEscapeSearch() {
  emit('escape')
  emit('close')
}

function onClickOutside(e) {
  if (panelEl.value && !panelEl.value.contains(e.target)) {
    emit('close')
  }
}

onMounted(() => {
  document.addEventListener('mousedown', onClickOutside)
  nextTick(() => searchEl.value?.focus())
})
onBeforeUnmount(() => document.removeEventListener('mousedown', onClickOutside))
</script>

<style scoped>
.emoji-panel {
  position: absolute;
  bottom: calc(100% + 6px);
  left: 0;
  width: 300px;
  background: var(--color-surface);
  border: 1px solid var(--color-border);
  border-radius: 10px;
  box-shadow: 0 8px 24px rgba(0,0,0,.15);
  z-index: 400;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.emoji-cats {
  display: flex;
  border-bottom: 1px solid var(--color-border);
  padding: 2px 4px;
  gap: 2px;
  overflow-x: auto;
  scrollbar-width: none;
}
.emoji-cats::-webkit-scrollbar { display: none; }

.emoji-cat-btn {
  flex-shrink: 0;
  background: none;
  border: none;
  cursor: pointer;
  font-size: 16px;
  padding: 4px 5px;
  border-radius: 6px;
  line-height: 1;
  opacity: .6;
  transition: opacity .1s, background .1s;
}
.emoji-cat-btn:hover { opacity: 1; background: var(--color-bg); }
.emoji-cat-btn.active { opacity: 1; background: var(--color-bg); }

.emoji-search-wrap {
  padding: 6px 8px 4px;
}
.emoji-search {
  width: 100%;
  padding: 4px 8px;
  border: 1px solid var(--color-border);
  border-radius: 6px;
  background: var(--color-bg);
  color: var(--color-text);
  font-size: 12px;
  outline: none;
  box-sizing: border-box;
}
.emoji-search:focus { border-color: var(--color-primary); }

.emoji-grid {
  display: grid;
  grid-template-columns: repeat(8, 1fr);
  padding: 4px 6px 8px;
  max-height: 200px;
  overflow-y: auto;
  gap: 1px;
}

.emoji-btn {
  background: none;
  border: none;
  cursor: pointer;
  font-size: 20px;
  padding: 3px;
  border-radius: 6px;
  line-height: 1;
  text-align: center;
  transition: background .1s, transform .08s;
  aspect-ratio: 1;
}
.emoji-btn:hover { background: var(--color-bg); transform: scale(1.25); }

.emoji-empty {
  grid-column: 1 / -1;
  text-align: center;
  padding: 16px;
  font-size: 12px;
  color: var(--color-text-muted);
}
</style>
