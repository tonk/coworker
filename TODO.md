- Bugfix:
  On Fedora 43 this setting must be done to not show the blank screen:
  LD_PRELOAD=/usr/lib64/libwayland-client.so
  Client was built on Ubuntu 24.04. Could be that this library is
  LD_PRELOAD=/usr/lib/libwayland-client.so on Ubuntu
  Fix this, once and for all

- In the client, an issue created in Forgejo is displayed on the card, but
  when clicked, nothing happens. The link should be opened in a browser.
  This works in the browser version.

- When a Forgejo webhook is configured and triggered, the card shows the
  Gitea logo. This is not really an error, but can be better.
  Show the Forgejo logo when the event comes from Forgejo. Forgejo sends
  this:
  * `X-Forgejo-Delivery:`
  * `X-Forgejo-Event: issues`
  * `X-Forgejo-Event-Type: issues`
  * `X-Forgejo-Signature:`

- In the project overview, allow projects to be moved, change the order

- Allow external repo servers (Gitea, Forgejo, Github, Gitlab) to follow
  projects, so that the comments (with the Gitea issue tag) are also
  linked in the issue

- After zooming in or out, save the zoom setting for the nest session

- Currently the theme is blue, also allow, per user, a red, green,
  orange, black and white theme (Save buttons, highlight lines in the
  sidebar, etc). This, of course, needs to be a user setting

- In the client, when a new version is available and the "Release notes"
  are clicked, nothing happens.
  This works in the browser version.

- In the chat show an indication when someone else in the chat is typing
  and who it is

- Add a `/api/v1/metrics`, non public, API endpoint for Prometheus
  monitoring, and create a user type that is only allowed to read these
  metrics. User type should be "metrics"
  Metrics should show:
  * Number of projects
  * Project names
  * Number of columns per projects
  * Column names
  * Number of open and closed cards per column

- Re-add '@..' mentions in cards, the description and comments, for users
  and for teams, this worked before. This does work in the chat

- Add cross reference between cards

- Pressing "Esc" in an open card, should close the card, if no changes
  where made

- Pressing "Cancel" with changes in the card should show a pop-up with
  "Save" and "No save"

- Investigate:
  When starting the Windows client it takes rather long before
  it responds… for userid/password entry

- Change projects to add an extra layer
  so that we can have "Customer / Projects" and between the "Customer / Projects"
  there should be the option for "Contract". So:
  * "Customer / Projects"
  * "Customer / Contract / Projects"

- Also add 'Customers' and "Favorite customers" to the sidebar
  with, of course, the possibility to "Star" and "Un-star"  customer.

- Implement sub-cards in the cards

- Implement a Gantt chart per project
