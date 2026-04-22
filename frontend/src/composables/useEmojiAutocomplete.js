import { ref, computed } from 'vue';
import emojis from '@/utils/emoji-data.json'; // Assuming a standard emoji list

/**
 * Composable to handle ':' triggered emoji autocomplete in textareas.
 */
export function useEmojiAutocomplete(textareaRef) {
  const showPopup = ref(false);
  const query = ref('');
  const selectedIndex = ref(0);
  const cursorPosition = ref(0);

  const filteredEmojis = computed(() => {
    if (!query.value) return emojis.slice(0, 10);
    return emojis
      .filter(e => e.name.toLowerCase().includes(query.value.toLowerCase()))
      .slice(0, 10);
  });

  const handleInput = (e) => {
    const text = e.target.value;
    const pos = e.target.selectionStart;
    cursorPosition.value = pos;

    // Look back from cursor to find a colon
    const lastColon = text.lastIndexOf(':', pos - 1);
    if (lastColon !== -1 && !text.slice(lastColon, pos).includes(' ')) {
      query.value = text.slice(lastColon + 1, pos);
      showPopup.value = true;
      selectedIndex.value = 0;
    } else {
      showPopup.value = false;
    }
  };

  const selectEmoji = (emoji) => {
    const text = textareaRef.value.value;
    const pos = cursorPosition.value;
    const lastColon = text.lastIndexOf(':', pos - 1);

    const newText = text.slice(0, lastColon) + emoji.char + ' ' + text.slice(pos);
    textareaRef.value.value = newText;

    showPopup.value = false;
    // Trigger input event for v-model compatibility
    textareaRef.value.dispatchEvent(new Event('input'));

    // Reset focus and cursor
    textareaRef.value.focus();
    const newPos = lastColon + emoji.char.length + 1;
    setTimeout(() => {
      textareaRef.value.setSelectionRange(newPos, newPos);
    }, 0);
  };

  return {
    showPopup,
    query,
    selectedIndex,
    filteredEmojis,
    handleInput,
    selectEmoji
  };
}