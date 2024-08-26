/**
 * @typedef UserEntry
 * @type {object}
 * @property {bigint} ItemId
 * @property {string} Status
 * @property {number} ViewCount
 * @property {string} StartDate
 * @property {string} EndDate
 * @property {number} UserRating
 * @property {string} Notes
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
 * @property {string} title
 * @property {string} type
 * @property {number[]} format
 * @property {boolean} children
 * @property {boolean} copies
 */

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
async function findTotalCostDeep(entry) {
    if (String(entry.ItemId) in costCache) {
        return costCache[String(entry.ItemId)]
    }
    let res = await fetch(`${apiPath}/total-cost?id=${entry.ItemId}`)
    let text = await res.text()
    costCache[String(entry.ItemId)] = Number(text)
    return Number(text)
}

class EntryTree {
    /**
    * @param {InfoEntry} entry
    * @param {InfoEntry?} [parent=null] 
    */
    constructor(entry, parent = null) {
        this.entry = entry
        /**@type {EntryTree[]}*/
        this.children = []
        /**@type {EntryTree[]}*/
        this.copies = []
        /**@type {InfoEntry?}*/
        this.parent = parent
    }

    /**
    * @param {InfoEntry} entry
    */
    addChild(entry) {
        let tree = new EntryTree(entry, this.entry)
        this.children.push(tree)
    }

    /**
    * @param {InfoEntry} entry
    */
    addCopy(entry) {
        this.copies.push(new EntryTree(entry))
    }
}
