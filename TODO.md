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

- In the client, when a new version is available and the "Release notes"
  are clicked, nothing happens

- In the chat show an indication when someone else in the chat is typing
  and who it is

- Add a `/api/v1/metrics`, non public, API endpoint for Prometheus
  monitoring, and create a user type that is only allowed to read these
  metrics

- Investigate:
  When starting the Windows client it takes rather long before
  it responds… for userid/password entry

- Change projects to add an extra layer
  so that we can have "Customer / Projects" and between the "Customer / Projects"
  the should be the option for "Contract". So:
  * "Customer / Projects"
  * "Customer / Contract / Projects"

- Implement sub-cards in the cards

- Implement a Gantt chart per project
