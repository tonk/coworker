export function parseLineItems(json) {
  try { return JSON.parse(json || '[]') } catch { return [] }
}
