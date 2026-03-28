Create an application that has all these features and requirements

- Do not use Docker or Podman
- Written in Golang, minumum version 1.25
- Webbased
- Kanban board
- Multi user
- Multi project
- Multi language support
  * English
  * Dutch
  * German
  * French
  * Spanish
- Database support for
  * SQLlite
  * MySQL
  * PostgreSQL
- Database configuration through config file
- Allow another config file as an option on the command line
- Role based access control
- Internal chat function
- Group chat between members
- Allow chat to be scaled horizontally
- All text editors must support Github markdown
- Allow for dark, light and follow-system interface
- Add settings per user
  * First name
  * Last name
  * Email address
  * Avatar, with Gravatar support
  * Change password
  * Date / time format
    ** Use IЅO timeformat as the default
  * Timezone
    ** Use UTC as the default
  * Interface font and font size
- Allow Admin settings for
  * For all users
    ** Change all user settings
    ** Change all project settings
  * Create new users
  * Disable users
  * Delete users
  * Switch the register on the login page on/off
    ** Remove "Don't have an account? Register" from the login page if
       the setting is off
  * Global system settings, can be overruled by the user
    ** Date / time format
       ** Use IЅO timeformat as the default
    * * Timezone
      ** Use UTC as the default
    ** Interface color, dark, light or follow-system interface
    ** Interface font and font size
    ** Default language setting, English by default
  * Add, change and remove projects
- Keep track of last login
- Keep track of last user settings change
- Enable a sidebar
  * Collapsible list of starred projects
  * Collapsible list of all projects, starred project marked and at the top
  * Collapsible list of all users, in chat users marked and at the top
  * Make the users clickable and when clicked, open an direct chat box to them
- Allow sidebar to be moved to the left or right
- In the projects
  * Allow the column names to be edited
  * Allow the columns to be moved
  * Allow project to be starred for the sidebar
  * In the "Invite members" box, give a pull down with all active users
- When a new project is created, automatically create a default column
  called "Inbox"
- In the ticket overview, place the users avatar at the top right of the
  box
- In the tickets
  * Keep history of column changes
  * In the "Edit Card" box
    ** Allow the box to be resized horizontal and vertical
    ** Add "Save" button
    ** Start the date picker with the current date
    ** "Due date" field must follow user date / time settings
    ** "Due date" must be saved with the card
- Enable an API to control all tickets
  * Add
  * Comment
  * Move to other column
- Add a footer to the UI
  * Application name and version number
    ** Align left
  * Logged users full name
    ** Align right

- Add README.md with explanation of Coworker
- Add systemd example service files
- Add Nginx example configuration file
- Add Apache example configuration file
- Add a decent .gitignore file
- Build production version
- Create installation manual
- Create distribution package containing everything needed
- Create a Github actions file to build a new release when a new tag is
  pushed in the main branch
