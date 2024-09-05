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
  - [ ] Track current position as string
    - Then whichever ui the user uses to play content, will send a request to the server saying that the user is now at position X
  - [x] Replace the start/end lists with event lists
    - eg: `[["Planned", <unixtime>], ["Viewing", <unixtime>], ["Finished", <unixtime>]]`
    - The events should be the same as statuses
    - unixtime is when that status was set
    - this actually might be really good as it's own table that looks something like this
      `itemId | timestamp | event`
    - Even if the items are entered into the table out of order, and the ids aren't in order, we can always use `SELECT * FROM userEvents WHERE itemId = ? ORDER BY timestamp`
- [ ] Library features
  - [ ] internet search
    - i can do something like `/api/v1/internet-search` with similar params to addentry, except
    - instead of adding an entry, it uses the (yet to be implemented) identify feature to gather search results
    - [ ] search
      - as described above
    - [ ] lookup
      - will lookup a specific entry using a specific provider
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
  - [x] allow user to identify a media, and pull related metadata (like how jellyfin does)
  - [x] scanning folders
  - [ ] search
    - [ ] search based on metadata entries
    - [ ] search based on user info entries
    - [x] search based on info entries
    - [x] search based on a combination of various entries
    - [x] search filters
  - [x] Store purchace price
  - [ ] store true title + native title in metadata instead of info entry
    - the info entry title and native title can remain, but they will be user determined
  - [ ] monitor folders, automatically add items in them
    - [ ] allow user to specify type of media, metadata source, and other stuff for everything in the folder
- [x] Ability to act as a proxy for the given {location} of an entry, and stream it

  - [ ] reencoding?

<del>
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
</del>

- [ ] new ui
  - [ ] sorting
    - [x] by price
    - [x] by rating
    - [ ] alpha
  - [ ] search
    - [ ] by price
    - [ ] by collection
    - [ ] by rating
    - [ ] tags
    - [x] by type
      - [x] multiple
    - [x] by format
      - [x] multiple
    - [ ] by title
  - [x] add item
  - [ ] edit item
    - [x] rating
    - [x] notes
    - [ ] thumbnail
    - [ ] title
  - [x] delete item
  - [x] start/stop, etc
  - [x] fetch metadata
  - [x] display metadata
  - [ ] identify item
  - [x] display user info
  - [x] display general info
  - [x] display cost
  - [ ] ability to view pie chart of which tags cost the most money
  - [x] display total stats
    - [x] total cost of inspected items
    - [x] total items
  - [x] thumbnails
  - [ ] if an item is marked as a copy, display the copy user viewing entry
  - [ ] if an item has children, display the children within the item (if the item is in the inspection area)

- [ ] terminal ui
  - simply a list of all items
  - [ ] allow user to filter by status
  - [ ] start/stop, etc
