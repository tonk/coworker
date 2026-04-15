- ~~Bugfix:
  On Fedora 43 this setting must be done to not show the blank screen:
  LD_PRELOAD=/usr/lib64/libwayland-client.so
  Client was built on Ubuntu 24.04. Could be that this library is
  LD_PRELOAD=/usr/lib/libwayland-client.so on Ubuntu
  Fix this, once and for all~~ **Done in v0.5.3**

- ~~In the client, an issue created in Forgejo is displayed on the card, but
  when clicked, nothing happens. The link should be opened in a browser.
  This works in the browser version.~~ **Done in v0.5.2**

- ~~When a Forgejo webhook is configured and triggered, the card shows the
  Gitea logo. This is not really an error, but can be better.
  Show the Forgejo logo when the event comes from Forgejo.~~ **Done in v0.5.2**

- ~~In the project overview, allow projects to be moved, change the order
  Drag-to-reorder on dashboard (admin only)~~ **Done in v0.5.2**

- ~~After zooming in or out, save the zoom setting for the next session~~ **Already implemented (localStorage)**

- ~~Currently the theme is blue, also allow, per user, a red, green,
  orange, black and white theme (Save buttons, highlight lines in the
  sidebar, etc). This, of course, needs to be a user setting~~ **Done in v0.6.3 (blue, red, green, orange — black and white skipped)**

- ~~In the sidebar, allow for re-ordering the starred customers and
  projects~~ **Done in v0.6.1**

- ~~In the sidebar show all customers~~ **Done in v0.6.1**

- ~~In the cards, the date-pickers for the "Start Date" and "Due Date"
  don't work~~ **Done in v0.6.1**

- ~~In the contract editor the date fields should follow the configured
  date format~~ **Done in v0.6.1**

- ~~In the client, when a new version is available and the "Release notes"
  are clicked, nothing happens.
  This works in the browser version.~~ **Done in v0.5.2**

- ~~In the chat show an indication when someone else in the chat is typing
  and who it is~~ **Done in v0.5.2**

- ~~Add a `/api/v1/metrics`, non public, API endpoint for Prometheus
  monitoring, and create a user type that is only allowed to read these
  metrics. User type should be "metrics"~~ **Done in v0.5.2**

- ~~Re-add '@..' mentions in cards, the description and comments, for users
  and for teams, this worked before. This does work in the chat~~ **Done in v0.5.2**

- ~~Add cross reference between cards~~ **Done in v0.6.0**

- ~~Pressing "Esc" in an open card, should close the card, if no changes
  where made~~ **Done in v0.5.2**

- ~~Pressing "Cancel" with changes in the card should show a pop-up with
  "Save" and "No save"~~ **Done in v0.5.2**

- ~~Investigate:
  When starting the Windows client it takes rather long before
  it responds, it does show the login screen, but typing lags behind and
  is very slow, for userid/password entry. After logging in, performance
  is OK~~ **Fix applied (pending Windows verification): passive keydown listener for zoom + `::-ms-reveal { display: none }` to suppress WebView2 password-reveal IPC round-trip**

- ~~Change projects to add an extra layer
  so that we can have "Customer / Projects" and between the "Customer / Projects"
  there should be the option for "Contract". So:
  * "Customer / Projects"
  * "Customer / Contract / Projects"~~ **Done in v0.5.3**

- ~~Also add 'Customers' and "Favorite customers" to the sidebar
  with, of course, the possibility to "Star" and "Un-star"  customer.~~ **Done in v0.5.3**

- ~~Add indicator to card, to show that sub-cards are present~~ **Already implemented (BoardCard.vue)**

- ~~"Edit customer" shows empty fields~~ **Fixed: openEdit() was never called**

- ~~Switching customers in the sidebar doesn't update the overview on the
  right hand side~~ **Fixed: added watch(custId) to CustomerDetailView**

- ~~Add customers and contracts, demo data, to `warmdesk-seed`~~ **Done**

- ~~Implement sub-cards in the cards~~ **Done in v0.5.3**

- ~~Implement a Gantt chart per project~~ **Done in v0.6.0**

- Allow external repo servers (Gitea, Forgejo, Github, Gitlab) to follow
  projects, so that the comments (with the Gitea issue tag) are also
  linked in the issue
