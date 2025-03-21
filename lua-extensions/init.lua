local function comp(left, right)
    return left .. " == " .. right
end

---@param macro string
function Expand_macro(macro)
    local asName2I = {}
    for k, v in pairs(aio.artStyles) do
        asName2I[v] = k
    end

    local mediaTypes = {
        "Show",
        "Movie",
        "MovieShort",
        "Game",
        "BoardGame",
        "Song",
        "Book",
        "Manga",
        "Collection",
        "Picture",
        "Meme"
    }

    local formats = {
        VHS          = 0,
        CD           = 1,
        DVD          = 2,
        BLURAY       = 3,
        ['4KBLURAY'] = 4,
        MANGA        = 5,
        BOOK         = 6,
        DIGITAL      = 7,
        BOARDGAME    = 8,
        STEAM        = 9,
        NIN_SWITCH   = 10,
        XBOXONE      = 11,
        XBOX360      = 12,
        OTHER        = 13,
        VINYL        = 14,
        IMAGE        = 15,
        UNOWNED      = 16,
    }
    local F_MOD_DIGITAL = 0x1000

    local statuses = {
        "Viewing",
        "Finished",
        "Dropped",
        "Paused",
        "Planned",
        "ReViewing"
    }

    local basicMacros = {
        isAnime = string.format("(artStyle & %d == %d)", asName2I['Anime'], asName2I['Anime']),
        r = "userRating",
        R = "rating",
        t = "en_title",
        T = "title",
        d = "description",
    }

    for _, item in ipairs(mediaTypes) do
        basicMacros[string.lower(item)] = '(type == "' .. item .. '")'
    end

    for _, item in ipairs(statuses) do
        basicMacros[string.lower(item)] = '(status == "' .. item .. '")'
    end

    if basicMacros[macro] ~= nil then
        return basicMacros[macro], ""
    elseif aio.hasprefix(macro, "f:") then
        local reqFmt = string.upper(string.sub(macro, 3))
        if macro:match("%+d") then
            reqFmt = string.sub(reqFmt, 0, #reqFmt - 2)
            --add in the digital modifier
            reqFmt = tostring(aio.bor(formats[reqFmt], F_MOD_DIGITAL))
            --the user explicitely asked for +digital modifier, only return those matching it
            return comp("Format", reqFmt), ""
        end

        --the user wants ONLY the non-digital version
        if macro:match("-d") then
            reqFmt = string.sub(reqFmt, 0, #reqFmt - 2)
            return comp("Format", formats[reqFmt]), ""
        end

        --return any matching format OR the digitally modified version
        return comp("Format", formats[reqFmt]) .. " or " .. comp("Format", aio.bor(formats[reqFmt], F_MOD_DIGITAL)), ""
    elseif macro == "s:v" then
        return comp("status", "\"Viewing\"") .. " or " .. comp("status", "\"ReViewing\""), ""
    elseif aio.hasprefix(macro, "s:") then
        return comp("status", '"' .. aio.title(string.sub(macro, 3)) .. '"'), ""
    elseif aio.hasprefix(macro, "t:") then
        return comp("type", '"' .. aio.title(string.sub(macro, 3)) .. '"'), ""
    elseif aio.hasprefix(macro, "a:") then
        local titledArg = aio.title(string.sub(macro, 3))
        local as_int = asName2I[titledArg]
        if as_int == nil then
            return "", "Invalid art style: " .. titledArg
        end
        return string.format("(artStyle & %d == %d)", as_int, as_int), ""
    else
        return string.format("(en_title LIKE \"%%%s%%\")", string.sub(macro, 1)), ""
    end
end

aio.listen("MacroExpand", Expand_macro)
