/**
 * @typedef UserEntry
 * @type {object}
 * @property {bigint} ItemId
 * @property {string} Status
 * @property {number} ViewCount
 * @property {number} UserRating
 * @property {string} Notes
 * @property {string} CurrentPosition
 */

/**
 * @typedef UserEvent
 * @type {object}
 * @property {bigint} ItemId
 * @property {string} Event
 * @property {number} Timestamp
 * @property {number} After
 * @property {string} TimeZone
 */

/**
 * @typedef InfoEntry
 * @type {object}
 * @property {bigint} ItemId
 * @property {string} Collection
 * @property {number} Format
 * @property {number} ArtStyle
 * @property {string} Location
 * @property {string} Native_Title
 * @property {bigint} ParentId
 * @property {number} PurchasePrice
 * @property {string} Type
 * @property {string} En_Title
 * @property {bigint} CopyOf
 */

/**
 * @typedef MetadataEntry
 * @type {object}
 * @property {bigint} ItemId
 * @property {number} Rating
 * @property {number} RatingMax
 * @property {string} Description
 * @property {number} ReleaseYear
 * @property {string} Thumbnail
 * @property {string} MediaDependant
 * @property {string} Datapoints
 * @property {string} Title
 * @property {string} Native_Title
 */

/**
 * @typedef IdentifyResult
 * @property {number} id
 * @property {string} Description
 * @property {string} Thumbnail
 */

/**@param {string} jsonl*/
function mkStrItemId(jsonl) {
    return jsonl
        .replace(/"ItemId":\s*(\d+),/, "\"ItemId\": \"$1\",")
        .replace(/"ParentId":\s*(\d+),/, "\"ParentId\": \"$1\",")
        .replace(/"CopyOf":\s*(\d+)(,)?/, "\"CopyOf\": \"$1\"$2")
}

/**@param {string} jsonl*/
function mkIntItemId(jsonl) {
    return jsonl
        .replace(/"ItemId":"(\d+)",/, "\"ItemId\": $1,")
        .replace(/"ParentId":"(\d+)",/, "\"ParentId\": $1,")
        .replace(/"CopyOf":"(\d+)"(,)?/, "\"CopyOf\": $1$2")
}

/**@param {string} jsonl*/
function parseJsonL(jsonl) {
    const bigIntProperties = ["ItemId", "ParentId", "CopyOf"]
    try {
        return JSON.parse(jsonl, (key, v) => bigIntProperties.includes(key) ? BigInt(v) : v)
    }
    catch (err) {
        console.error("Could not parse json", err)
    }
}

/**
 * @function
 * @template T
 * @param {string} endpoint 
 * @returns {Promise<T[]>}
*/
async function loadList(endpoint) {
    const res = await fetch(`${apiPath}/${endpoint}?uid=${uid}`)
    if (!res) {
        return []
    }

    const text = await res.text()
    if (!text) {
        return []
    }

    const lines = text.split("\n").filter(Boolean)
    return lines
        .map(mkStrItemId)
        .map(parseJsonL)
}

/**
* @param {bigint} oldid
* @param {bigint} newid
*/
async function copyUserInfo(oldid, newid) {
    return await fetch(`${apiPath}/engagement/copy?src-id=${oldid}&dest-id=${newid}`).catch(console.error)
}

/**
    * @param {string} type
    */
function typeToSymbol(type) {
    const conversion = {
        "Show": "📺︎",
        "Movie": "📽",
        "MovieShort": "⏯",
        "Book": "📚︎",
        "Manga": "本",
        "Game": "🎮︎",
        "Song": "♫",
        "Collection": "🗄",
        "BoardGame": "🎲︎",
        "Picture": "🖼",
        "Meme": "🃏",
        "Unowned": "X"
    }
    if (type in conversion) {
        //@ts-ignore
        return conversion[type]
    }
    return type
}

/**
 * @param {string} name
 * @returns {number}
 */
function nameToFormat(name) {
    const DIGI_MOD = 0x1000
    let val = 0
    name = name.toLowerCase()
    if (name.includes("+digital")) {
        name = name.replace("+digital", "")
        val |= DIGI_MOD
    }
    const formats = {
        "vhs": 0,
        "cd": 1,
        "dvd": 2,
        "bluray": 3,
        "4kBLURAY": 4,
        "manga": 5,
        "book": 6,
        "digital": 7,
        "boardgame": 8,
        "steam": 9,
        "ninSWITCH": 10,
        "xboxONE": 11,
        "xbox360": 12,
        "other": 13,
        "vinyl": 14,
        "image": 15
    }
    val |= formats[/**@type {keyof typeof formats}*/(name)]
    return val

}

/**
    * @param {number} format
    */
function formatToName(format) {
    const DIGI_MOD = 0x1000
    let out = ""
    if ((format & DIGI_MOD) === DIGI_MOD) {
        format -= DIGI_MOD
        out = " +digital"
    }
    const formats = [
        "VHS",
        "CD",
        "DVD",
        "BLURAY",
        "4K BLURAY",
        "MANGA",
        "BOOK",
        "DIGITAL",
        "BOARDGAME",
        "STEAM",
        "NIN SWITCH",
        "XBOX ONE",
        "XBOX 360",
        "OTHER",
        "VINYL",
        "IMAGE",
        "UNOWNED"
    ]
    if (format >= formats.length) {
        return `unknown${out}`
    }
    return `${formats[format]}${out}`
}

/**
    * @param {string} title
    * @param {string} provider
    */
async function identify(title, provider) {
    return await fetch(`${apiPath}/metadata/identify?title=${encodeURIComponent(title)}&provider=${provider}`)
}

/**
    * @param {string} identifiedId
    * @param {string} provider
    * @param {bigint} applyTo
    */
async function finalizeIdentify(identifiedId, provider, applyTo) {
    identifiedId = encodeURIComponent(identifiedId)
    provider = encodeURIComponent(provider)
    return await fetch(`${apiPath}/metadata/finalize-identify?identified-id=${identifiedId}&provider=${provider}&apply-to=${applyTo}`)
}

/**
 * @param {bigint} id
 * @param {string} thumbnail
 */
async function updateThumbnail(id, thumbnail) {
    return await fetch(`${apiPath}/metadata/mod-entry?id=${id}&thumbnail=${encodeURIComponent(thumbnail)}`)
}

/**
 * @param {string} searchString
 */
async function doQuery3(searchString) {
    const res = await fetch(`${apiPath}/query-v3?search=${encodeURIComponent(searchString)}&uid=${uid}`).catch(console.error)
    if (!res) return []

    let itemsText = await res.text()
    if (res.status !== 200) {
        alert(itemsText)
        return []
    }

    try {
        let jsonL = itemsText.split("\n")
            .filter(Boolean)
            .map(mkStrItemId)
            .map(parseJsonL)
        return jsonL
    } catch (err) {
        console.error(err)
    }

    return []
}

/**
 * @param {bigint} id
 * @param {string} pos
 */
async function setPos(id, pos) {
    return fetch(`${apiPath}/engagement/mod-entry?id=${id}&current-position=${pos}`)
}

/**
* @param {InfoEntry[]} entries
* @param {string} sortBy
*/
function sortEntries(entries, sortBy) {
    if (sortBy != "") {
        if (sortBy == "rating") {
            entries = entries.sort((a, b) => {
                let aUInfo = findUserEntryById(a.ItemId)
                let bUInfo = findUserEntryById(b.ItemId)
                if (!aUInfo || !bUInfo) return 0
                return bUInfo?.UserRating - aUInfo?.UserRating
            })
        } else if (sortBy == "cost") {
            entries = entries.sort((a, b) => {
                return b.PurchasePrice - a.PurchasePrice
            })
        } else if (sortBy == "general-rating") {
            entries = entries.sort((a, b) => {
                let am = findMetadataById(a.ItemId)
                let bm = findMetadataById(b.ItemId)
                if (!bm || !am) return 0
                return normalizeRating(bm.Rating, bm.RatingMax || 100) - normalizeRating(am.Rating, am.RatingMax || 100)
            })
        } else if (sortBy == "rating-disparity") {
            entries = entries.sort((a, b) => {
                let am = findMetadataById(a.ItemId)
                let au = findUserEntryById(a.ItemId)
                let bm = findMetadataById(b.ItemId)
                let bu = findUserEntryById(b.ItemId)
                if (!bm || !am) return 0
                let bGeneral = normalizeRating(bm.Rating, bm.RatingMax || 100)
                let aGeneral = normalizeRating(am.Rating, am.RatingMax || 100)

                let aUser = Number(au?.UserRating)
                let bUser = Number(bu?.UserRating)


                return (aGeneral - aUser) - (bGeneral - bUser)
            })
        } else if (sortBy == "release-year") {
            entries = entries.sort((a, b) => {
                let am = findMetadataById(a.ItemId)
                let bm = findMetadataById(b.ItemId)
                return (bm?.ReleaseYear || 0) - (am?.ReleaseYear || 0)
            })
        }
    }
    return entries
}

/**
 * @param {BigInt} itemId
 * @param {number} ts
 * @param {number} after
*/
async function apiDeleteEvent(itemId, ts, after) {
    return await fetch(`${apiPath}/engagement/delete-event?id=${itemId}&after=${after}&timestamp=${ts}`)
}

/**
 * @param {BigInt} itemId
 * @param {string} name
 * @param {number} ts
 * @param {number} after
*/
async function apiRegisterEvent(itemId, name, ts, after) {
    const tz = Intl.DateTimeFormat().resolvedOptions().timeZone
    return await fetch(`${apiPath}/engagement/register-event?name=${encodeURIComponent(name)}&id=${itemId}&after=${after}&timestamp=${ts}&timezone=${encodeURIComponent(tz)}`)
}
