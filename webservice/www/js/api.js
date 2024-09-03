/**
 * @typedef UserEntry
 * @type {object}
 * @property {bigint} ItemId
 * @property {string} Status
 * @property {number} ViewCount
 * @property {number} UserRating
 * @property {string} Notes
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
 * @property {bigint} Parent
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
 * @property {string} Description
 * @property {number} ReleaseYear
 * @property {string} Thumbnail
 * @property {string} MediaDependant
 * @property {string} Datapoints
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
        .replace(/"Parent":\s*(\d+),/, "\"Parent\": \"$1\",")
        .replace(/"CopyOf":\s*(\d+),/, "\"CopyOf\": \"$1\"")
}

/**@param {string} jsonl*/
function parseJsonL(jsonl) {
    const bigIntProperties = ["ItemId", "Parent", "CopyOf"]
    return JSON.parse(jsonl, (key, v) => bigIntProperties.includes(key) ? BigInt(v) : v)
}
/**
 * @param {bigint} id 
 * @returns {Promise<InfoEntry[]>}
 */
async function getChildren(id) {
    let res = await fetch(`${apiPath}/query?parent-ids=${id}`)
    let text = await res.text()
    return /**@type {InfoEntry[]}*/ text.split("\n")
        .filter(Boolean)
        .map(mkStrItemId)
        .map(parseJsonL)
}

/**
 * @param {bigint} id 
 * @returns {Promise<InfoEntry[]>}
 */
async function getDescendants(id) {
    let res = await fetch(`${apiPath}/list-descendants?id=${id}`)
    let text = await res.text()
    return /**@type {InfoEntry[]}*/ text.split("\n")
        .filter(Boolean)
        .map(mkStrItemId)
        .map(parseJsonL)
}

/**
 * @param {bigint} id 
 * @returns {Promise<InfoEntry[]>}
 */
async function getCopies(id) {
    let res = await fetch(`${apiPath}/query?copy-ids=${id}`)
    let text = await res.text()
    return /**@type {InfoEntry[]}*/ text.split("\n")
        .filter(Boolean).map(mkStrItemId).map(parseJsonL)
}

/**
 * @param {InfoEntry} entry
 * @returns {Promise<number>} cost
 */
async function getTotalCostDeep(entry) {
    if (String(entry.ItemId) in costCache) {
        return costCache[String(entry.ItemId)]
    }
    let res = await fetch(`${apiPath}/total-cost?id=${entry.ItemId}`)
    let text = await res.text()
    costCache[String(entry.ItemId)] = Number(text)
    return Number(text)
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
   return  lines
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
        queryString += `&title=${encodeURI(search.title)}`
    }
    if (search.format && search.format?.[0] != -1) {
        queryString += `&formats=${encodeURI(search.format.join(","))}`
    }
    if (search.type) {
        queryString += `&types=${encodeURI(search.type)}`
    }
    if (search.tags) {
        queryString += `&tags=${encodeURI(search.tags)}`
    }
    if (search.status) {
        queryString += `&user-status=${encodeURI(search.status)}`
    }
    if (search.userRatingGt) {
        queryString += `&user-rating-gt=${encodeURI(String(search.userRatingGt))}`
    }
    if (search.userRatingLt) {
        queryString += `&user-rating-lt=${encodeURI(String(search.userRatingLt))}`
    }
    if (search.isAnime) {
        queryString += `&is-anime=${search.isAnime}`
    }
    if (search.purchasePriceGt) {
        queryString += `&purchase-gt=${encodeURI(String(search.purchasePriceGt))}`
    }
    if (search.purchasePriceLt) {
        queryString += `&purchase-lt=${encodeURI(String(search.purchasePriceLt))}`
    }
    const res = await fetch(`${apiPath}/query${queryString}`)
        .catch(console.error)
    if (!res) {
        alert("Could not query entries")
        return
    }
    let itemsText = await res.text()
    let jsonL = itemsText.split("\n")
        .filter(Boolean)
        .map(mkStrItemId)
        .map(parseJsonL)
    return jsonL
}
