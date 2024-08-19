# Format

0 F_VHS
1 F_CD
2 F_DVD
3 F_BLURAY
4 F_4KBLURAY
5 F_MANGA
6 F_BOOK
7 F_DIGITAL
8 F_BOARDGAME
9 F_STEAM
10 F_NIN_SWITCH_DIGITAL
11 F_XBOXONE_DIGITAL
12 F_XBOX360_DISC
13 F_NIN_SWITCH_DISC

# Status

VIEWING
FINISHED
DROPPED
PLANNED
REVIEWING
PAUSED

# Media Dependant Json
(all items in all jsons are optional)

Show + Movie
```json
{
    "length": number (milliseconds),
    "actors": string[],
    "directors": string[]
}
```

Book + Manga
```json
{
    "volumes": number,
    "chapters": number,
    "author": string,
}
```

Songs
```json
{
    "singer": string,
    "record-label": string
}
```
