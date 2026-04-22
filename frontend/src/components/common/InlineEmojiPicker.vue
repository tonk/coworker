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
        @click="activeCat = cat.name"
      >{{ cat.icon }}</button>
    </div>
    <!-- Search -->
    <div class="emoji-search-wrap">
      <input class="emoji-search" v-model="search" placeholder="Search…" ref="searchEl" />
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

const props = defineProps({ initialSearch: { type: String, default: '' } })
const emit = defineEmits(['pick', 'close'])

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

const visibleEmojis = computed(() => {
  if (search.value.trim()) {
    const q = search.value.toLowerCase()
    const all = CATEGORIES.flatMap(c => c.emojis)
    return all.filter(e => e.includes(q)).slice(0, 60)
  }
  return CATEGORIES.find(c => c.name === activeCat.value)?.emojis || []
})

function pick(emoji) {
  emit('pick', emoji)
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
