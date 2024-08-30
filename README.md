# The **A**ll **I**n **O**ne **Li**brary **Ma**gement **S**ystem

### AIO LIMAS For short

---

# Goals

The goal of this project is to be a management system for your physical, and digital content library

The user should be able to serach for something (with certain criteria like dvd, blu ray, digital, etc) and it will show matching results

In addition, it will contain a myanimelist/letterboxd like experience for all media managed by the LIMAS

Physical content will need to be entered in manually, however metadata should be fetched automatically (similar to how manual entries in jellyfin work)

- [x] MAL/letterboxd system
  - [x] user rating system
  - [x] notes
  - [x] watch dates table
  - [x] ability to plan, start, stop shows
    - [x] status (planning, viewing, etc)
    - [x] start
    - [x] finish
    - [x] plan
    - [x] drop
    - [x] pause
    - [x] continue
  - [x] Replace the start/end lists with event lists
    - eg: `[["Planned", <unixtime>], ["Viewing", <unixtime>], ["Finished", <unixtime>]]`
    - The events should be the same as statuses
    - unixtime is when that status was set
    - this actually might be really good as it's own table that looks something like this
    `itemId | timestamp | event`
    - Even if the items are entered into the table out of order, and the ids aren't in order, we can always use `SELECT * FROM userEvents WHERE itemId = ? ORDER BY timestamp`
- [ ] Library features
  - [x] copy of (id)
    - that way if the user has say, a digital movie and blu ray of something
    - they can have 2 entries, and the ui can know to display the same user/metadata for both
  - [x] media dependant metadata
  - [x] proper collections
    - [x] create collection
    - [x] delete collection
    - [x] cannot put entry in a collection unless collection exists
    - the way this is accomplished is simply by having a Collection TY.
  - [ ] automatic metadata
    - [x] anlist
      - [x] Anime
      - [ ] Manga
    - [ ] Steam games
    - [ ] Books
    - [ ] ~tmdb~
      - requires billing address
    - [x] omdb
  - [ ] ~nfo files~
    - impossibly non-standard format
  - [x] allow user to change metadata
  - [ ] allow user to identify a media, and pull related metadata (like how jellyfin does)
  - [x] scanning folders
  - [ ] search
    - [ ] search based on metadata entries
    - [ ] search based on user info entries
    - [x] search based on info entries
    - [ ] search based on a combination of various entries
    - [x] search filters
  - [x] Store purchace price
  - [ ] monitor folders, automatically add items in them
    - [ ] allow user to specify type of media, metadata source, and other stuff for everything in the folder
- [x] Ability to act as a proxy for the given {location} of an entry, and stream it
  - [ ] reencoding?

- [ ] ui
  - [ ] sorting
    - [ ] by price
    - [ ] by rating
    - [ ] alpha
  - [ ] search
    - [ ] by price
    - [ ] by collection
    - [ ] by rating
    - [x] by type
      - [x] multiple
    - [x] by format
      - [ ] multiple
    - [x] by title
  - [x] add item
  - [ ] edit item
  - [x] delete item
  - [x] start/stop, etc
  - [x] fetch metadata
  - [ ] display metadata
  - [x] display user info
  - [x] display general info
  - [x] display cost
  - [ ] display total stats
    - [ ] total stats
    - [x] total items
  - [x] thumbnails
  - [ ] if an item is marked as a copy, display the copy user viewing entry
