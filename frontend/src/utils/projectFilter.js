/**
 * Cascading project list for a selected customer, shared by the weekly-sheet "add row" form,
 * the macro editor, and the calendar view's entry modal.
 *
 * - No customer selected: every project (regular + time-tracking-only).
 * - A time-tracking-only customer: only time-tracking-only projects.
 * - A regular customer: that customer's projects plus the time-tracking-only ones.
 */
export function filterProjectsForCustomer(customerId, { allProjects, ttCustomers, ttProjects, projects }) {
  if (!customerId) return allProjects
  if (ttCustomers.some(c => c.id === customerId)) return ttProjects
  return [...projects.filter(p => p.customer_id === customerId), ...ttProjects]
}
