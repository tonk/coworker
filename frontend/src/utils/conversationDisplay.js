// Returns the other participant in a 1-on-1 conversation
export function getOtherMember(conv, currentUserId) {
  if (!conv?.members) return null
  const m = conv.members.find(m => m.user_id !== currentUserId)
  return m?.user || null
}

// Human-readable conversation name
export function getConversationDisplayName(conv, currentUserId) {
  if (!conv) return ''
  if (conv.name) return conv.name
  if (conv.is_group) {
    return conv.members
      ?.filter(m => m.user_id !== currentUserId)
      .map(m => m.user?.display_name || m.user?.username)
      .join(', ') || 'Group'
  }
  const other = getOtherMember(conv, currentUserId)
  return other?.display_name || other?.username || 'Unknown'
}
