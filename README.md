# The **A**ll **I**n **O**ne **Li**brary **Ma**gement **S**ystem

### AIO LIMAS For short

---

# Goals

The goal of this project is to be a management system for your physical, and digital content library

The user should be able to serach for something (with certain criteria like dvd, blu ray, digital, etc) and it will show matching results

In addition, it will contain a myanimelist/letterboxd like experience for all media managed by the LIMAS

Physical content will need to be entered in manually, however metadata should be fetched automatically (similar to how manual entries in jellyfin work)

- [ ] MAL/letterboxd system
    - [ ] user rating system
    - [ ] notes
    - [x] watch dates table
    - [ ] ability to plan, start, stop shows
        - [x] status (planning, viewing, etc)
        - [x] start
        - [x] finish
        - [ ] plan
        - [ ] drop
        - [ ] pause
        - [ ] continue
- [ ] Library features
    - [ ] automatic metadata
    - [ ] scanning folders
    - [ ] search
        - [ ] search filters
    - [x] Store purchace price
    
- [ ] Ability to act as a proxy for the given {location} of an entry, and stream it
  - [ ] reencoding?

---

# Project layout

All data will be stored in one big database

Maybe containing the following tables

Below is an example table for the library
Each title will have a randomly generated id

The user should be able to put items into collections
And search by that collection

# User General Info table

| ID  | Title        | Format  | Location | Purchase price | Collection   |
| --- | ------------ | ------- | -------- | -------------- | ------------ |
| Fx  | Friends s01  | DVD     | Library  | $xx.xx         | Friends      |
| xx  | Erased Vol 1 | Manga   | Library  | $xx.xx         | Erased:Anime |
| Yx  | Your Name    | Digital | {link}   | $xx.xx         | Anime        |

# Generated Metadata table

| ID  | Rating | Description | Length (seconds) |
| --- | ------ | ----------- | ---------------- |
| Fx  | 80     | ...         | ...              |

# User Viewing Info

<!--the user will be displayed the title, instead of the id-->

| ID  | Status   | ViewCount | Start date | End date | User Rating |
| --- | -------- | --------- | ---------- | -------- | ----------- |
| Yx  | Finished | 3         | unixtime   | unixtime | 94          |

---

