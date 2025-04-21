# Clients

This should document some things about implementing a client

## Assumptions

There are a couple of assumptions the client needs to make.

1. the Collection row in entryInfo is for tags, separated by `\x1F` (ASCII Unit)
    Example (`|` represents ASCII unit):
    ```
    |tag1||tag2||tag3|
    ```
    each tag has exactly one ASCII Unit on either side of it.
2. The Notes section part of userViewingInfo can be in any format, the web client supports HTML and phpbb like syntax.
    A client MAY choose to display notes, and if notes are displayed:
    the client MUST implement phpbb like syntax for the following tags:
    `[b]bold[/b]`, `[i]italic[/i]`, `[spoiler]spoiler[/spoiler]`
3. The UserRating can be anything the user desires, it could be a scale of 1-10, it could be a non-linear scale, it's up to the user.
    However the client may enforce a certain rating scale,
    The web client enforces my rating scale, of a non-linear scale where 80 is average

Besides these assumptions, the client can do whatever it wants.
