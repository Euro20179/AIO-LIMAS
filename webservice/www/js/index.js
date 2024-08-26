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

/**@type { {formats: Record<number, string>, userEntries: UserEntry[], metadataEntries: MetadataEntry[], entries: InfoEntry[] }}*/
const globals = { formats: {}, userEntries: [], metadataEntries: [], entries: [] }

/**
 * @param {bigint} id
 * @returns {UserEntry?}
 */
function getUserEntry(id) {
    for (let entry of globals.userEntries) {
        if (entry.ItemId === id) {
            return entry
        }
    }
    return null
}

/**
 * @param {bigint} id
 * @returns {MetadataEntry?}
 */
function getMetadataEntry(id) {
    for (let entry of globals.metadataEntries) {
        if (entry.ItemId === id) {
            return entry
        }
    }
    return null
}

/**
 * @param {bigint} id
 * @returns {InfoEntry?}
 */
function getInfoEntry(id) {
    for (let entry of globals.entries) {
        if (entry.ItemId === id) {
            return entry
        }
    }
    return null
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
async function getCopies(id) {
    let res = await fetch(`${apiPath}/query?copy-ids=${id}`)
    let text = await res.text()
    return /**@type {InfoEntry[]}*/ text.split("\n")
        .filter(Boolean).map(mkStrItemId).map(parseJsonL)
}

/**
 * @param {InfoEntry} entry
 * @param {InfoEntry[]} children
 * @returns number
 */
function findTotalCost(entry, children) {
    let cost = entry.PurchasePrice || 0
    for (let child of children) {
        cost += child.PurchasePrice || 0
    }
    return cost
}

/**
 * @param {bigint} id
 * @param {MetadataEntry?} newMeta
 */
function setMetadataEntry(id, newMeta) {
    for (let i = 0; i < globals.metadataEntries.length; i++) {
        let entry = globals.metadataEntries[i]
        if (entry.ItemId === id) {
            //@ts-ignore
            globals.metadataEntries[i] = newMeta
            return
        }
    }
}

/**@param {string} jsonl*/
function mkStrItemId(jsonl) {
    return jsonl
        .replace(/"ItemId":\s*(\d+),/, "\"ItemId\": \"$1\",")
        .replace(/"Parent":\s*(\d+),/, "\"Parent\": \"$1\",")
        .replace(/"CopyOf":\s*(\d+),/, "\"CopyOf\": \"$1\"")
}

/**@param {string} jsonl*/
function parseJsonL(jsonl) {
    const bigIntProperties = ["ItemId", "Parent"]
    return JSON.parse(jsonl, (key, v) => bigIntProperties.includes(key) ? BigInt(v) : v)
}

async function loadFormats() {
    const res = await fetch(`${apiPath}/type/format`)
    const json = await res.json()
    let fmtJson = Object.fromEntries(
        Object.entries(json).map(([key, val]) => [Number(key), val])
    )
    globals.formats = fmtJson
}

/**
 * @param {string | HTMLElement} text
 * @param {string} ty
 */
function basicElement(text, ty = "span") {
    const el = document.createElement(ty)
    el.append(text)
    return el
}

/**
 * @param {HTMLElement} container
 * @param {InfoEntry} item
 */
function fillBasicInfoSummary(container, item) {
    /**@type {HTMLDetailsElement}*/
    const basicInfoEl = /**@type {HTMLDetailsElement}*/(container.querySelector(".basic-info ul"))
    basicInfoEl.append(basicElement(`Item id: ${item.ItemId}`, "li"))
    basicInfoEl.append(basicElement(`Title: ${item.En_Title}`, "li"))
    basicInfoEl.append(basicElement(`Native title: ${item.Native_Title}`, "li"))
    let typeText = `Type: ${item.Type}`
    if (item.IsAnime) {
        typeText += ` +anime`
    }
    basicInfoEl.append(basicElement(typeText, "li"))

    basicInfoEl.append(basicElement(`Format: ${formatToStr(item.Format)}`, "li"))
    if(item.CopyOf) {
        basicInfoEl.append(basicElement(`Copy of: ${item.CopyOf}`, "li"))
    }
    if (item.PurchasePrice) {
        basicInfoEl.append(basicElement(`$Purchase: $${item.PurchasePrice}`))
    }
}

/**
 * @param {HTMLElement} container
 * @param {UserEntry} item
 */
function fillUserInfo(container, item) {

    /**@type {HTMLDetailsElement}*/
    const userInfoEl = /**@type {HTMLDetailsElement}*/(container.querySelector(".user-info ul"))

    userInfoEl.append(basicElement(`View count: ${item.ViewCount}`, "li"));
    userInfoEl.append(basicElement(`Status: ${item.Status}`, "li"));

    const viewTableBody = /**@type {HTMLTableElement}*/(container.querySelector(".view-table tbody"));

    const startDates = JSON.parse(item.StartDate)
    const endDates = JSON.parse(item.EndDate)
    for (let i = 0; i < startDates.length; i++) {
        let start = startDates[i]
        let end = endDates[i]

        let sd = new Date(start)
        let ed = new Date(end)
        let sText = `${sd.getMonth() + 1}/${sd.getDate()}/${sd.getFullYear()}`
        let eText = `${ed.getMonth() + 1}/${ed.getDate()}/${ed.getFullYear()}`

        viewTableBody.innerHTML += `
<tr>
    <td>${start === 0 ? "unknown" : sText}</td>
    <td>${end === 0 ? "unknown" : eText}</td>
</tr>
`
    }
}

/**
 * @param {number} format
 * @returns {string}
 */
function formatToStr(format) {
    const MOD_DIGITAL = Number(Object.entries(globals.formats).filter(([_, val]) => val == "MOD_DIGITAL")[0][0])
    let fmtNo = format
    let digitized = false
    if ((format & MOD_DIGITAL) === MOD_DIGITAL) {
        fmtNo -= MOD_DIGITAL
        digitized = true
    }
    let fmtText = `${globals.formats[fmtNo]}`
    if (digitized) {
        fmtText += ` +digitized`
    }
    return fmtText
}

/**
 * @param {InfoEntry} item
 * @param {UserEntry} userEntry
 * @param {MetadataEntry} meta
 */
function createItemEntry(item, userEntry, meta) {
    const out = /**@type {HTMLElement}*/(document.getElementById("all-entries"))
    const itemTemplate = /**@type {HTMLTemplateElement}*/ (document.getElementById("item-entry"))
    /**@type {HTMLElement}*/
    const clone = /**@type {HTMLElement}*/(itemTemplate.content.cloneNode(true));

    const root = /**@type {HTMLElement}*/(clone.querySelector(".entry"));
    root.setAttribute("data-type", item.Type);
    root.setAttribute("data-entry-id", String(item.ItemId));

    /**@type {HTMLElement}*/(clone.querySelector(".location")).append(`${item.En_Title} (${formatToStr(item.Format).toLowerCase()})`)

    if (userEntry?.UserRating) {
        /**@type {HTMLElement}*/(clone.querySelector(".rating")).innerHTML = String(userEntry?.UserRating) || "#N/A";
    }


    /**@type {HTMLElement}*/(clone.querySelector(".notes")).innerHTML = String(userEntry?.Notes || "");

    if (item.Location) {
        const locationA = /**@type {HTMLAnchorElement}*/(clone.querySelector(".location"));
        locationA.href = item.Location
    }

    if (item.Collection) {
        /**@type {HTMLElement}*/(clone.querySelector(".collection")).innerHTML = `Collection: ${item.Collection}`
    }

    if (item.Parent) {
        const parentA = /**@type {HTMLAnchorElement}*/(root.querySelector("a.parent"))
        parentA.href = `javascript:displayEntry([${item.Parent}n])`
        parentA.hidden = false
    }

    if (userEntry?.UserRating) {
        if (userEntry.UserRating >= 80) {
            root.classList.add("good")
        } else {
            root.classList.add("bad")
        }
    }

    const img = /**@type {HTMLImageElement}*/(clone.querySelector(".img"));

    if (meta?.Thumbnail) {
        img.src = meta.Thumbnail
    }

    getChildren(item.ItemId).then(children => {
        if (!children.length) return
        const list = /**@type {HTMLElement}*/ (root.querySelector(".children"))
        //@ts-ignore
        list.parentNode.hidden = false

        let allA = /**@type {HTMLAnchorElement}*/(basicElement("all children", "a"))
        allA.href = `javascript:displayEntry([${children.map(i => i.ItemId).join("n, ")}n])`
        list.append(allA)

        for (let child of children) {
            let a = /**@type {HTMLAnchorElement}*/ (basicElement(
                basicElement(
                    `${child.En_Title} - $${child.PurchasePrice}`,
                    "li"
                ),
                "a"
            ))
            a.href = `javascript:displayEntry([${child.ItemId.toString()}n])`
            list.append(a)
        }

        let totalCost = findTotalCost(item, children);
        /**@type {HTMLElement}*/(root.querySelector(".cost")).innerHTML = `$${totalCost}`
    })

    getCopies(item.ItemId).then(copies => {
        if (!copies.length) return
        const list = /**@type {HTMLElement}*/ (root.querySelector(".copies"))
        //@ts-ignore
        list.parentNode.hidden = false

        let allA = /**@type {HTMLAnchorElement}*/(basicElement("all copies", "a"))
        allA.href = `javascript:displayEntry([${copies.map(i => i.ItemId).join("n, ")}n])`
        list.append(allA)

        for (let child of copies) {
            let a = /**@type {HTMLAnchorElement}*/ (basicElement(
                basicElement(
                    `${child.En_Title} - $${child.PurchasePrice}`,
                    "li"
                ),
                "a"
            ))
            a.href = `javascript:displayEntry([${child.ItemId.toString()}n])`
            list.append(a)
        }
    })


    if (item.PurchasePrice) {
        /**@type {HTMLElement}*/(root.querySelector(".cost")).innerHTML = `$${item.PurchasePrice}`
    }


    fillBasicInfoSummary(clone, item)
    fillUserInfo(clone, /**@type {UserEntry}*/(userEntry))


    const metaRefresher = /**@type {HTMLButtonElement}*/(clone.querySelector(".meta-fetcher"));
    metaRefresher.onclick = async function(e) {
        let res = await fetch(`${apiPath}/metadata/fetch?id=${item.ItemId}`).catch(console.error)
        if (res?.status != 200) {
            console.error(res)
            return
        }

        res = await fetch(`${apiPath}/metadata/retrieve?id=${item.ItemId}`).catch(console.error)
        if (res?.status != 200) {
            console.error(res)
            return
        }

        const json = /**@type {MetadataEntry}*/(await res.json())

        setMetadataEntry(item.ItemId, json)
        img.src = json.Thumbnail
    }
    out.appendChild(clone)
}

async function loadCollections() {
    const res = await fetch(`${apiPath}/list-collections`).catch(console.error)
    if (!res) {
        alert("Could not load collections")
        return
    }
    /**@type {string}*/
    const text = /**@type {string}*/(await res.text().catch(console.error))
    const lines = text.split("\n").filter(Boolean)
    return lines
}


/**@param {string[] | undefined} collections*/
function addCollections(collections) {
    if (!collections) {
        return
    }
    const collectionsSection = /**@type {HTMLElement}*/ (document.getElementById("collections"))
    for (let collection of collections) {
        const elem = /**@type {HTMLElement}*/ (document.createElement("entry-collection"));
        /**@type {HTMLElement}*/ (elem.querySelector(".name")).innerText = collection
        collectionsSection.appendChild(elem)
    }
}

async function loadAllEntries() {
    const res = await fetch(`${apiPath}/list-entries`)
        .catch(console.error)
    if (!res) {
        alert("Could not load entries")
    } else {
        let itemsText = await res.text()
        /**@type {string[]}*/
        let jsonL = itemsText.split("\n").filter(Boolean)
        globals.entries = jsonL
            .map(mkStrItemId)
            .map(parseJsonL)
        return globals.entries
    }
}

/**
 * @param {DBQuery} search
 */
async function loadQueriedEntries(search) {
    let queryString = "?_"
    if (search.title) {
        queryString += `&title=${encodeURI(search.title)}`
    }
    if (search.format[0] != -1) {
        queryString += `&formats=${encodeURI(search.format.join(","))}`
    }
    if (search.type) {
        queryString += `&types=${encodeURI(search.type)}`
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

/**
* @param {InfoEntry[] | undefined} items 
*/
async function addEntries(items, ignoreChildren = true, ignoreCopies = true) {
    if (!items) {
        return
    }
    items = items.sort((a, b) => {
        const aUE = getUserEntry(a.ItemId)
        const bUE = getUserEntry(b.ItemId)
        return (bUE?.UserRating || 0) - (aUE?.UserRating || 0)
    })
    for (const item of items) {
        console.log(item, ignoreChildren)
        if (item.Parent && ignoreChildren) {
            //TODO: put a link to children on each entry
            //when the link is clicked, all entries will be removed in favor of that item's children
            //also render the item itself
            continue
        }
        if(item.CopyOf && ignoreCopies) {
            continue
        }
        let user = getUserEntry(item.ItemId)
        let meta = item.Parent ?
            getMetadataEntry(item.Parent) :
            getMetadataEntry(item.ItemId)
        createItemEntry(item, user, meta)
    }
}

/**@returns {Promise<UserEntry[]>}*/
async function loadUserEntries() {
    const res = await fetch(`${apiPath}/engagement/list-entries`)
    if (!res) {
        return []
    }

    const text = await res.text()
    if (!text) {
        return []
    }

    const lines = text.split("\n").filter(Boolean)
    globals.userEntries = lines
        .map(mkStrItemId)
        .map(parseJsonL)

    return globals.userEntries
}

/**@returns {Promise<MetadataEntry[]>}*/
async function loadMetadata() {
    const res = await fetch(`${apiPath}/metadata/list-entries`)
    if (!res) {
        return []
    }

    const text = await res.text()
    if (!text) {
        return []
    }

    const lines = text.split("\n").filter(Boolean)
    globals.metadataEntries = lines
        .map(mkStrItemId)
        .map(parseJsonL)

    return globals.metadataEntries
}

function removeEntries() {
    const entryEl = document.getElementById("all-entries")
    while (entryEl?.children.length) {
        entryEl.firstChild?.remove()
    }
}

function query() {
    removeEntries()
    const form = /**@type {HTMLFormElement}*/(document.getElementById("query"))
    let data = new FormData(form)
    let enTitle = /**@type {string}*/(data.get("query"))
    // let format = /**@type {string}*/(data.get("format"))
    let ty = /**@type {string}*/(data.get("type"))

    let displayChildren = /**@type {string}*/(data.get("children"))
    let displayCopies = /**@type {string}*/(data.get("copies"))

    let formats = []
    formats.push(data.getAll("format").map(Number))

    /**@type {DBQuery}*/
    let query = {
        title: enTitle,
        type: ty,
        //@ts-ignore
        format: formats.length ? formats : [-1]
        // format: Number(format)
    }

    loadQueriedEntries(query).then(entries => {
        console.log(displayChildren, displayCopies)
        addEntries(entries, displayChildren !== "on", displayCopies !== "on")
    })
}

/**
* @param {bigint[]} ids
*/
function displayEntry(ids) {
    if (!Array.isArray(ids)) {
        console.error("displayEntry: ids is not an array")
        return
    }
    removeEntries()

    addEntries(
        ids.map(getInfoEntry),
        false,
        false
    )
}

function main() {
    loadFormats()
        // .then(loadCollections)
        // .then(addCollections)
        .then(loadUserEntries)
        .then(loadMetadata)
        .then(loadAllEntries)
        .then(addEntries)
        .catch(console.error)
}
main()
