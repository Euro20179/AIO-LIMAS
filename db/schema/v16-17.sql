/*
add 2 (F_MOD_UNOWNED) to modifiers
also set format to 13 (F_OTHER)
*/
UPDATE entryInfo SET
    format_modifiers = format_modifiers | 2,
    format = 13
WHERE format = 16;
