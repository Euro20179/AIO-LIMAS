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
    return JSON.parse(jsonl, (key, v) => bigIntProperties.includes(key) ? BigInt(v) : v)
}

/**
 * @function
 * @template T
 * @param {string} endpoint 
 * @returns {Promise<T[]>}
*/
async function loadList(endpoint) {
    const res = await fetch(`${apiPath}/${endpoint}`)
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
 * @param {[string[], string[], string[], number[]]} search
 */
async function loadQueriedEntries2(search) {
    let names = encodeURIComponent(search[0].join(","))
    let values = encodeURIComponent(search[1].join(","))
    let checkers = encodeURIComponent(search[2].join(","))
    let gates = encodeURIComponent(search[3].join(","))
    const res = await fetch(`${apiPath}/query-v2?names=${names}&values=${values}&checkers=${checkers}&gates=${gates}`)
        .catch(console.error)
    if (!res) {
        alert("Could not query entries")
        return []
    }
    let itemsText = await res.text()
    let jsonL = itemsText.split("\n")
        .filter(Boolean)
        .map(mkStrItemId)
        .map(parseJsonL)
    return jsonL

}

/**
* @param {bigint} oldid
* @param {bigint} newid
*/
async function copyUserInfo(oldid, newid) {
    console.log(oldid, newid)
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
        out = "+digital"
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
        "IMAGE"
    ]
    if (format >= formats.length) {
        return `unknown ${out}`
    }
    return `${formats[format]} ${out}`
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
 * @param {FormData} data
 */
async function doQuery2(data) {
    //TODO:
    //make a proper parser
    //i want to handle stuff like
    //"exact search"
    //contains these words
    //r > 200
    //-(r > 200)
    //-"cannot contain this exact search"
    //-(cannot contain these words)
    //
    //quoted strings can use the "like" operator, and the user can add % if they want
    //non-quoted strings should essentially add a % sign to the start and end
    //use the like opeartor, and be all ored together
    let search = /**@type {string}*/(data.get("search-query"))

    let operatorPairs = {
        ">": "<=",
        "<": ">=",
        "=": "!=",
        "~": "!~",
        "^": "!^"
    }

    let operator2Name = {
        ">": "GT",
        "<": "LT",
        "<=": "LE",
        ">=": "GE",
        "=": "EQ",
        "!=": "NE",
        "~": "LIKE",
        "!~": "NOTLIKE",
        "^": "IN",
        "!^": "NOTIN"
    }
    //sort by length because the longer ops need to be tested first
    //they need to be tested first because of a scenario such as:
    //userRating!=20, if = is tested first,
    //it'll succeed but with the wrong property of userRating!
    let operatorList = [...Object.keys(operator2Name).sort((a, b) => b.length - a.length)]

    let shortcuts = {
        "r": "userRating",
        "y": "releaseYear",
        "p": "purchasePrice",
        "f": "format",
        "t": "type",
        "s": "status",
    }

    let words = search.split(" ")

    let names = []
    let values = []
    let operators = []
    let gates = []

    const gateNames = {
        "and": 0,
        "or": 1
    }
    let nextGate = gateNames["and"]

    for (let word of words) {
        if (word == "||") {
            nextGate = gateNames["or"]
            search = search.replace(word, "").trim()
            continue
        } else if(word === "&&") {
            nextGate = gateNames["and"]
            search = search.replace(word, "").trim()
            continue
        }
        for (let op of operatorList) {
            if (!word.includes(op)) continue
            let [property, value] = word.split(op)

            if (property.startsWith("|")) {
                property = property.slice(1)

            }
            if (property.startsWith("-")) {
                property = property.slice(1)
                op = operatorPairs[/**@type {keyof typeof operatorPairs}*/(op)]
            }

            let opName = operator2Name[/**@type {keyof typeof operator2Name}*/(op)]

            for (let shortcut in shortcuts) {
                if (property === shortcut) {
                    property = shortcuts[/**@type {keyof typeof shortcuts}*/(shortcut)]
                    break
                }
            }

            if (property === "format") {
                let formats = []
                for (let format of value.split(":")) {
                    if (isNaN(Number(format))) {
                        formats.push(nameToFormat(format))
                    }
                    else {
                        formats.push(format)
                    }
                }
                value = formats.join(":")
            }

            names.push(property)
            values.push(value)
            operators.push(opName)
            gates.push(nextGate)

            nextGate = gateNames['and']

            search = search.replace(word, "").trim()

            break
        }
    }

    search = search.trim()

    if (search) {
        names.push("entryInfo.en_title")
        values.push(search)
        operators.push("LIKE")
        gates.push(nextGate)
    }

    return await loadQueriedEntries2([names, values, operators, gates])
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
