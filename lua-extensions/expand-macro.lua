local function comp(left, right)
    return left .. " == " .. right
end

function Expand_macro(macro)
    local asName2I = {}
    for k, v in pairs(aio.artStyles) do
        asName2I[v] = k
    end

    local basicMacros = {
        isAnime = string.format("(artStyle & %d == %d)", asName2I['Anime'], asName2I['Anime']),
        r = "userRating",
        R = "rating",
        t = "en_title",
        T = "title",
        d = "description"
    }

    if basicMacros[macro] ~= nil then
        return basicMacros[macro], ""

    elseif aio.hasprefix(macro, "s:") then
        return comp("status", '"' .. aio.title(string.sub(macro, 3)) .. '"'), ""

    elseif aio.hasprefix(macro, "t:") then
        return comp("type", '"' .. aio.title(string.sub(macro, 3)) .. '"'), ""

    elseif aio.hasprefix(macro, "a:") then
        local titledArg = aio.title(string.sub(macro, 3))
        local as_int = asName2I[titledArg]
        if as_int == nil then
            return  "", "Invalid art style: " .. titledArg
        end
        return string.format("(artStyle & %d == %d)", as_int, as_int), ""
    end

    return macro, ""
end

return Expand_macro
