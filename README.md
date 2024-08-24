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
- [ ] Library features
    - [ ] copy of (id)
        - that way if the user has say, a digital movie and blu ray of something
        - they can have 2 entries, and the ui can know to display the same user/metadata for both
    - [x] media dependant metadata
    - [ ] proper collections
        - [ ] create collection
        - [ ] delete collection
        - [ ] cannot put entry in a collection unless collection exists
    - [ ] automatic metadata
        - [x] anlist
            - [x] Anime
            - [ ] Manga
        - [ ] tmdb
    - [ ] ~nfo files~
        - impossibly non-standard format
    - [x] allow user to change metadata
    - [ ] allow user to identify a media, and pull related metadata (like how jellyfin does)
    - [x] scanning folders
    - [ ] search
        - [ ] search based on metadata entries
        - [] search based on user info entries
        - [x] search based on info entries
        - [ ] search based on a combination of various entries
        - [x] search filters
    - [x] Store purchace price
    - [ ] monitor folders, automatically add items in them
        - [ ] allow user to specify type of media, metadata source, and other stuff for everything in the folder
- [x] Ability to act as a proxy for the given {location} of an entry, and stream it
  - [ ] reencoding?

