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
 * @property {boolean} IsAnime
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
 * @typedef DBQuery
 * @type {object}
 * @property {string} [title]
 * @property {string} [type]
 * @property {number[]} [format]
 * @property {boolean} [children]
 * @property {boolean} [copies]
 * @property {string} [tags]
 * @property {string} [status]
 * @property {number} [userRatingLt]
 * @property {number} [userRatingGt]
 * @property {number} [purchasePriceGt]
 * @property {number} [purchasePriceLt]
 * @property {number} [releasedGe]
 * @property {number} [releasedLe]
 * @property {boolean} [isAnime]
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
        .replace(/"CopyOf":\s*(\d+),/, "\"CopyOf\": \"$1\"")
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
 * @param {DBQuery} search
 * @returns {Promise<InfoEntry[]>}
 */
async function loadQueriedEntries(search) {
    let queryString = "?_"
    if (search.title) {
        queryString += `&title=${encodeURIComponent(search.title)}`
    }
    if (search.format && search.format?.[0] != -1) {
        queryString += `&formats=${encodeURIComponent(search.format.join(","))}`
    }
    if (search.type) {
        queryString += `&types=${encodeURIComponent(search.type)}`
    }
    if (search.tags) {
        queryString += `&tags=${encodeURIComponent(search.tags)}`
    }
    if (search.status) {
        queryString += `&user-status=${encodeURIComponent(search.status)}`
    }
    if (search.userRatingGt) {
        queryString += `&user-rating-gt=${encodeURIComponent(String(search.userRatingGt))}`
    }
    if (search.userRatingLt) {
        queryString += `&user-rating-lt=${encodeURIComponent(String(search.userRatingLt))}`
    }
    if (search.isAnime) {
        queryString += `&is-anime=${Number(search.isAnime) + 1}`
    }
    if (search.purchasePriceGt) {
        queryString += `&purchase-gt=${encodeURIComponent(String(search.purchasePriceGt))}`
    }
    if (search.purchasePriceLt) {
        queryString += `&purchase-lt=${encodeURIComponent(String(search.purchasePriceLt))}`
    }
    if (search.releasedGe) {
        queryString += `&released-ge=${encodeURIComponent(String(search.releasedGe))}`
    }
    if (search.releasedLe) {
        queryString += `&released-le=${encodeURIComponent(String(search.releasedLe))}`
    }
    const res = await fetch(`${apiPath}/query${queryString}`)
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
* @param {HTMLFormElement} form
*/
async function doQuery(form) {

    let data = new FormData(form)
    let status = /**@type {string[]}*/(data.getAll("status"))
    let type = /**@type {string[]}*/(data.getAll("type"))
    let format = /**@type {string[]}*/(data.getAll('format')).filter(n => n !== "")

    let search = /**@type {string}*/(data.get("search-query"))

    let tags = /**@type {string[]}*/(data.getAll("tags"))

    let pgt = /**@type {string}*/(data.get("price-gt"))
    let plt = /**@type {string}*/(data.get("price-lt"))

    let rgt = /**@type {string}*/(data.get("rating-gt"))
    let rlt = /**@type {string}*/(data.get("rating-lt"))

    let formatN = undefined
    if (format.length) {
        formatN = format.map(Number)
    }

    //TODO:
    //Add hasTags, notHasTags, and maybeHasTags
    //allow the user to type #tag #!tag and #?tag in the search bar
    /**@type {DBQuery}*/
    let queryData = {
        status: status.join(","),
        type: type.join(","),
        format: formatN,
        tags: tags.join(","),
        purchasePriceGt: Number(pgt),
        purchasePriceLt: Number(plt),
        userRatingGt: Number(rgt),
        userRatingLt: Number(rlt),
    }

    /**
     * @type {Record<string, string | ((value: string) => Record<string, string>)>}
     */
    let shortcuts = {
        "r>": "userRatingGt",
        "r<": "userRatingLt",
        "p>": "purchasePriceGt",
        "p<": "purchasePriceLt",
        "y>=": "releasedGe",
        "y<=": "releasedLe",
        "y=": (value) => { return { "releasedGe": value, "releasedLe": value } },
    }

    for (let word of search.split(" ")) {
        /**@type {string | ((value: string) => Record<string, string>)}*/
        let property
        let value
        [property, value] = word.split(":")
        for (let shortcut in shortcuts) {
            if (word.startsWith(shortcut)) {
                value = word.slice(shortcut.length)
                //@ts-ignore
                property = shortcuts[/**@type {keyof typeof shortcuts}*/(shortcut)]
                break
            }
        }
        if (!value) continue
        if (typeof property === 'function') {
            let res = property(value)
            for (let key in res) {
                //@ts-ignore
                queryData[key] = res[key]
            }
        }
        search = search.replace(word, "").trim()
        //@ts-ignore
        queryData[property] = value
    }

    queryData.title = search

    return await loadQueriedEntries(queryData)
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
