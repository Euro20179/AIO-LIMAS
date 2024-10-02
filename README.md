# The **A**ll **I**n **O**ne **Li**brary **Ma**gement **S**ystem

### AIO LIMAS For short

![entries](./readme-assets/entries.png)
![graph](./readme-assets/graph.png)

### The Point

I made this program because i had a system for managing which shows/movies i had watched/planned,

And I had another system for keeping track of how much I've spent on Manga, DVDS, and the like.

I realized that I could make a program that combines both of these problems into one massive
inventory management thingy

For some extra challenge, I also want to support as many media types as possible on as many
formats as possible

example formats:
xbox 360
digital
blu ray
dvd

example media types:
Movie
Show
Manga
Book
Game
BoardGame
Song

### Running


> [!IMPORTANT]
Be sure to export the ACCOUNT_NUMBER env var
This is used as the login password
(plans to disable this by default)


> [!TIP]
To use the omdb provider, get an omdb key and export the OMDB_KEY variable


> [!NOTE]
Only tested on linux

```bash
git clone https://github.com/euro20179/aio-limas

cd aio-limas

go run .
```

A server and web ui will then be running on `localhost:8080`


### TODO

- [x] enable/disable children/copies
- [ ] enable/disable the "collection" type's stats being the sum of it's children's stats
- [ ] steam account linking
- [x] image media type
    - [ ] when fetching metadata, use ai to generate a description of the image
- [ ] search by description
- [x] disable ACCOUNT_NUMBER by default
- [ ] documentation
    - [ ] ui
- [ ] edit info/user/meta tables from ui
    - [x] info
    - [ ] meta
    - [ ] user
    - do this by letting the user click on a button that opens an editable table basically
