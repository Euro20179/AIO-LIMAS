local function comp(left, right)
    return left .. " == " .. right
end

local function parseDateParams(paramString, startOrEnd)
    local month = 1
    if startOrEnd == "start" then
        month = 1
    else
        month = 12
    end

    --default required fields, set to 2025, 1, 1
    local time = {
        year = 2025,
        month = month,
        day = 1,
        hour = 0,
        minute = 0,
        second = 0,
    }

    local curKey = ""
    local curVal = ""
    ---@type "key" | "val"
    local parsing = "key"
    for i = 1, #paramString do
        local ch = paramString:sub(i, i)
        if ch == ":" then
            parsing = "val"
            goto continue
        elseif ch == "/" then
            parsing = "key"
            local full = ({
                y = "year",
                m = "month",
                d = "day",
                H = "hour",
                M = "minute",
                S = "second"
            })[curKey] or ((time[curKey] ~= nil) and curKey) or "year"
            time[full] = tonumber(curVal)
            curKey = ""
            curVal = ""
            goto continue
        end

        if parsing == "key" then
            curKey = curKey .. ch
        else
            curVal = curVal .. ch
        end
        ::continue::
    end

    local ok, t = pcall(os.time, time)
    if not ok then
        return "false"
    end
    return tostring(t * 1000)
end

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

local asName2I = {}
for k, v in pairs(aio.artStyles) do
    asName2I[v] = k
end

local prefixMacros = {
    ["s:"] = function(macro)
        local text = aio.title(string.sub(macro, 3))
        return comp("status", '"' .. text .. '"'), ""
    end,

    ["t:"] = function(macro)
        return comp("type", '"' .. aio.title(string.sub(macro, 3)) .. '"'), ""
    end,

    ["a:"] = function(macro)
        local itemList = string.sub(macro, 3)
        local items = aio.split(itemList, "+")
        local query = ""
        for _, item in ipairs(items) do
            local titledArg = aio.title(item)
            local as_int = asName2I[titledArg]
            if as_int == nil then
                return "", "Invalid art style: " .. titledArg
            end
            if query ~= "" then
                query = query .. string.format("and (artStyle & %d == %d)", as_int, as_int)
            else
                --extra ( because i want to encase the whole thing with ()
                query = string.format("( (artStyle & %d == %d)", as_int, as_int)
            end
        end
        return query .. ")", ""
    end,

    ["f:"] = function(macro)
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
    end,

    ["tag:"] = function(macro)
        local tag = string.sub(macro, 5)
        return "Collection LIKE ('%' || char(31) || '" .. tag .. "' || char(31) || '%')", ""
    end,

    ["md:"] = function(macro)
        local name = string.sub(macro, 4)
        return string.format("mediaDependant != '' and json_extract(mediaDependant, '$.%s')", name), ""
    end,

    ["mdi:"] = function(macro)
        local name = string.sub(macro, 5)
        return string.format("mediaDependant != '' and CAST(json_extract(mediaDependant, '$.%s') as decimal)", name), ""
    end,
}

---@param macro string
function Expand_macro(macro)
    local mediaTypes = {
        "Show",
        "Episode",
        "Documentary",
        "Movie",
        "MovieShort",
        "Game",
        "BoardGame",
        "Song",
        "Book",
        "Manga",
        "Collection",
        "Picture",
        "Meme",
        "Library",
        "Video"
    }

    local statuses = {
        "Viewing",
        "Finished",
        "Dropped",
        "Paused",
        "Planned",
        "ReViewing",
        "Waiting"
    }

    local basicMacros = {
        isAnime = string.format("(artStyle & %d == %d)", asName2I['Anime'], asName2I['Anime']),
        r = "userRating",
        R = "rating",
        t = "en_title",
        T = "title",
        d = "description",
        ts = "timestamp",
        ["s:v"] = comp("status", "\"Viewing\"") .. " or " .. comp("status", "\"ReViewing\""),
        ep = "CAST(json_extract(mediadependant, format('$.%s-episodes', type)) as DECIMAL)",
        len = "CAST(json_extract(mediadependant, format('$.%s-length', type)) as DECIMAL)",
        epd = "CAST(json_extract(mediadependant, format('$.%s-episode-duration', type)) as DECIMAL)"
    }

    for _, item in ipairs(mediaTypes) do
        basicMacros[string.lower(item)] = '(type = \'' .. item .. '\')'
    end

    for _, item in ipairs(statuses) do
        basicMacros[string.lower(item)] = '(status = \'' .. item .. '\')'
    end

    local _, e, _ = string.find(macro, ":")
    local prefix = string.sub(macro, 0, e)

    if basicMacros[macro] ~= nil then
        return basicMacros[macro], ""
    elseif prefixMacros[prefix] ~= nil then
        return prefixMacros[prefix](macro)
    elseif string.sub(macro, 0, 4) == "date" then
        local beginOrEnd = "start"
        if string.sub(macro, 5, 5) == ">" then
            beginOrEnd = "end"
        end


        local time = string.sub(macro, 6)
        if time == "" then
            return "false", ""
        end

        return parseDateParams(time, beginOrEnd), ""
    else
        return string.format("(en_title LIKE '%%%s%%')", string.sub(macro, 1)), ""
    end
end

aio.listen("MacroExpand", Expand_macro)
