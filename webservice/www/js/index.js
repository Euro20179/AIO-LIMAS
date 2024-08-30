/**
 * @typedef EntryTree
 * @type {Record<string, TreeNode>}
 *
 * @typedef TreeNode
 * @type {object}
 * @property {InfoEntry} EntryInfo
 * @property {bigint[]} Children
 * @property {bigint[]} Copies
 */

/**@type { {formats: Record<number, string>, userEntries: UserEntry[], metadataEntries: MetadataEntry[], entries: InfoEntry[], tree: EntryTree }}*/
const globals = { formats: {}, userEntries: [], metadataEntries: [], entries: [], tree: {} }

function resubmitSearch() {
    /**@type {HTMLInputElement?}*/
    let queryButton = document.querySelector("#query [type=submit][value*=ocate]")
    if (!queryButton) {
        return
    }
    queryButton.click()
}

/**
 * @param {bigint} id
 * @returns {InfoEntry[]}
 */
function findChildren(id) {
    let childrenIds = globals.tree[String(id)]?.Children || []
    let entries = []
    for (let childId of childrenIds) {
        entries.push(globals.tree[String(childId)].EntryInfo)
    }
    return entries
}

/**
 * @param {bigint} id
 * @returns {InfoEntry[]}
 */
function findCopies(id) {
    let copyIds = globals.tree[String(id)]?.Copies || []
    let entries = []
    for (let copyId of copyIds) {
        entries.push(globals.tree[String(copyId)].EntryInfo)
    }
    return entries
}

/**
 * @param {bigint} id
 * @returns {number}
 */
function findTotalCostDeep(id, recurse = 0, maxRecursion = 10) {
    if (recurse > maxRecursion) {
        return 0
    }
    let item = globals.tree[String(id)]
    let cost = item.EntryInfo.PurchasePrice

    for (let child of item.Children || []) {
        cost += findTotalCostDeep(child, recurse + 1, maxRecursion)
    }

    return cost
}

/**
 * @param {Record<string, any>} stats
 */
function setGlobalStats(stats) {
    let out = /**@type {HTMLElement}*/(document.getElementById("total-stats"))
    while (out.children.length) {
        out.firstChild?.remove()
    }
    for (let name in stats) {
        let e = basicElement(`${name}: ${stats[name]}`, "li")
        out.append(e)
    }
}

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

/**@type {Record<string, number>}*/
const costCache = {}

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
    const bigIntProperties = ["ItemId", "Parent", "CopyOf"]
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
    if (item.CopyOf) {
        basicInfoEl.append(basicElement(`Copy of: ${item.CopyOf}`, "li"))
    }
    if (item.PurchasePrice) {
        basicInfoEl.append(basicElement(`Purchase: $${item.PurchasePrice}`))
    }
}

/**
 * @param {HTMLElement} container
 *@param {MetadataEntry} item
 */
function fillMetaInfo(container, item) {
    const metaEl = /**@type {HTMLDetailsElement}*/(container.querySelector(".metadata-info ul"))

    const descBox = /**@type {HTMLElement}*/(container.querySelector(".description"))
    descBox.innerText = item.Description

    metaEl.append(basicElement(`Release year: ${item.ReleaseYear}`, "li"))
    metaEl.append(basicElement(`General rating: ${item.Rating}`, "li"))
    try{ 
        const mediaDependant = JSON.parse(item.MediaDependant)
        for(let name in mediaDependant) {
            metaEl.append(basicElement(`${name}: ${mediaDependant[name]}`, "li"))
        }
    } catch(err) {
        console.warn(err)
    }
    try{
        const datapoints = JSON.parse(item.Datapoints)
        for(let name in datapoints) {
            metaEl.append(basicElement(`${name}: ${datapoints[name]}`, "li"))
        }
    } catch(err){
        console.warn(err)
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

    const viewTableBody = /**@type {HTMLTableElement}*/(container.querySelector(".event-table tbody"));
    fetch(`${apiPath}/engagement/get-events?id=${item.ItemId}`)
        .then(res => res.text())
        .then(text => {
            const json = text.split("\n")
                .filter(Boolean)
                .map(mkStrItemId)
                .map(parseJsonL)
            for (let item of json) {
                let tText = "unknown"
                if (item.Timestamp !== 0) {
                    let time = new Date(item.Timestamp)
                    tText = `${time.getMonth() + 1}/${time.getDate()}/${time.getFullYear()}`
                } else if (item.After !== 0) {
                    let time = new Date(item.After)
                    tText = `After: ${time.getMonth() + 1}/${time.getDate()}/${time.getFullYear()}`
                }

                let eventTd = basicElement(item.Event, "td")
                let timeTd = basicElement(tText, "td")
                let tr = document.createElement("tr")
                tr.append(eventTd)
                tr.append(timeTd)
                viewTableBody.append(tr)
            }
        })

    //     const startDates = JSON.parse(item.StartDate)

    //     const endDates = JSON.parse(item.EndDate)
    //     for (let i = 0; i < startDates.length; i++) {
    //         let start = startDates[i]
    //         let end = endDates[i]
    //
    //         let sd = new Date(start)
    //         let ed = new Date(end)
    //         let sText = `${sd.getMonth() + 1}/${sd.getDate()}/${sd.getFullYear()}`
    //         let eText = `${ed.getMonth() + 1}/${ed.getDate()}/${ed.getFullYear()}`
    //
    //         viewTableBody.innerHTML += `
    // <tr>
    //     <td>${start === 0 ? "unknown" : sText}</td>
    //     <td>${end === 0 ? "unknown" : eText}</td>
    // </tr>
    // `
    //     }
}

/**
 * @param {string} type
 * @returns {string}
 */
function typeToEmoji(type) {
    return {
        "Movie": "🎬︎",
        "Manga": "本",
        "Book": "📚︎",
        "Show": "📺︎",
        "Collection": "🗄",
        "Game": "🎮︎",
    }[type] || type
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

    /**@type {TemplateFiller}*/
    const fills = {};


    if (userEntry?.UserRating) {
        fills[".rating"] = String(userEntry.UserRating || "#N/A")
    }
    fills[".notes"] = e => e.innerHTML = userEntry.Notes || "";

    if (item.Location) {
        fills[".location"] = e => {
            let el = /**@type {HTMLAnchorElement}*/(e)
            el.href = item.Location
            el.append(`${item.En_Title} (${typeToEmoji(item.Type)} ${formatToStr(item.Format).toLowerCase()})`)
            if (item.Native_Title) {
                el.title = `Native: ${item.Native_Title}`
            }
        }
    } else {
        fills[".location"] = e => {
            e.append(`${item.En_Title} (${typeToEmoji(item.Type)} ${formatToStr(item.Format).toLowerCase()})`)
            if (item.Native_Title) {
                e.title = `Native: ${item.Native_Title}`
            }
        }
    }

    if (item.Collection) {
        fills[".tags"] = e => {
            e.append("tags: ")
            for (let tag of item.Collection.split(",")) {
                let tSpan = basicElement(tag, "span")
                tSpan.classList.add("tag")
                e.append(tSpan)
            }
        }
    }

    if (item.Parent) {
        fills["a.parent"] = e => {
            let el = /**@type {HTMLAnchorElement}*/(e)
            el.href = `javascript:displayEntry([${item.Parent}n])`
            el.hidden = false
        }
    }

    fills[".img"] = e => {
        const el = /**@type {HTMLImageElement}*/(e)
        el.src = meta.Thumbnail
    }

    let totalCost = findTotalCostDeep(item.ItemId)
    let children = findChildren(item.ItemId)

    if (totalCost !== 0) {
        fills['.cost'] = `$${totalCost}`
    }

    fills[".children"] = e => {
        if (!children.length) return
        //@ts-ignore
        e.parentNode.hidden = false

        let allA = /**@type {HTMLAnchorElement}*/(basicElement("all children", "a"))
        allA.href = `javascript:displayEntry([${children.map(i => i.ItemId).join("n, ")}n])`
        e.append(allA)

        for (let child of children) {
            let childCost = findTotalCostDeep(child.ItemId)
            let a = /**@type {HTMLAnchorElement}*/ (basicElement(
                basicElement(
                    `${child.En_Title} (${formatToStr(child.Format).toLowerCase()}) - $${childCost}`,
                    "li"
                ),
                "a"
            ))
            a.href = `javascript:displayEntry([${child.ItemId.toString()}n])`
            e.append(a)
        }
    }

    fills[".copies"] = e => {
        let copies = findCopies(item.ItemId)
        if (!copies.length) return
        //@ts-ignore
        e.parentNode.hidden = false

        let allA = /**@type {HTMLAnchorElement}*/(basicElement("all copies", "a"))
        allA.href = `javascript:displayEntry([${copies.map(i => i.ItemId).join("n, ")}n])`
        e.append(allA)

        for (let child of copies) {
            let a = /**@type {HTMLAnchorElement}*/ (basicElement(
                basicElement(
                    `${child.En_Title} - $${child.PurchasePrice}`,
                    "li"
                ),
                "a"
            ))
            a.href = `javascript:displayEntry([${child.ItemId.toString()}n])`
            e.append(a)
        }
    }

    if (item.PurchasePrice > 0) {
        fills[".cost"] = `$${item.PurchasePrice}`
    }

    let root = fillTemplate("item-entry", fills)

    root.setAttribute("data-entry-id", String(item.ItemId))

    if (userEntry?.UserRating) {
        if (userEntry.UserRating >= 80) {
            root.classList.add("good")
        } else {
            root.classList.add("bad")
        }
    }

    let metadata = getMetadataEntry(item.ItemId)

    fillBasicInfoSummary(root, item)
    fillUserInfo(root, /**@type {UserEntry}*/(userEntry))
    fillMetaInfo(root, /**@type {MetadataEntry}*/(metadata))

    const metaRefresher = /**@type {HTMLButtonElement}*/(root.querySelector(".meta-fetcher"));
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
        const img = /**@type {HTMLImageElement}*/(root.querySelector(".img"))
        img.src = json.Thumbnail
    }

    const deleter = /**@type {HTMLButtonElement}*/ (root.querySelector(".deleter"))
    deleter.onclick = async function() {
        if (!confirm("Are you sure you want to delete this item")) {
            return
        }
        let res = await fetch(`${apiPath}/delete-entry?id=${item.ItemId}`)
        if (res?.status != 200) {
            console.error(res)
            alert("Failed to delete item")
            return
        }
        alert(`Deleted: ${item.En_Title} (${item.Native_Title} : ${item.ItemId})`)
        resubmitSearch()
    }

    out.appendChild(root)
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

async function loadEntryTree() {
    const res = await fetch(`${apiPath}/list-tree`)
        .catch(console.error)
    if (!res) {
        alert("Could not load entries")
    } else {
        let itemsText = await res.text()
        itemsText = itemsText
            .replaceAll(/"ItemId":\s*(\d+),/g, "\"ItemId\": \"$1\",")
            .replaceAll(/"Parent":\s*(\d+),/g, "\"Parent\": \"$1\",")
            .replaceAll(/"CopyOf":\s*(\d+),/g, "\"CopyOf\": \"$1\"")
        const bigIntProperties = ["ItemId", "Parent", "CopyOf"]
        let json = JSON.parse(itemsText, (key, v) => bigIntProperties.includes(key) ? BigInt(v) : v)
        globals.tree = json
        return globals.tree
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
    if (search.tags) {
        queryString += `&tags=${encodeURI(search.tags)}`
    }
    if (search.status) {
        queryString += `&user-status=${encodeURI(search.status)}`
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
 * @param {EntryTree | undefined} items
 * @param {boolean} [ignoreChildren=true] 
 * @param {boolean} [ignoreCopies=true]
 */
async function renderEntryTree(items, ignoreChildren = true, ignoreCopies = true) {
    if (!items) {
        return
    }
    items = Object.fromEntries(
        Object.entries(items)
            .sort(([id, _], [idB, __]) => {
                const aUE = getUserEntry(BigInt(id))
                const bUE = getUserEntry(BigInt(idB))
                return (bUE?.UserRating || 0) - (aUE?.UserRating || 0)
            })
    )

    let costFinders = []
    let count = 0
    let totalCount = 0
    for (let id in items) {
        totalCount++
        let item = items[id]

        if (item.EntryInfo.Parent && ignoreChildren) continue
        if (item.EntryInfo.CopyOf && ignoreCopies) continue

        let user = getUserEntry(item.EntryInfo.ItemId)
        costFinders.push(getTotalCostDeep(item.EntryInfo))
        let meta = item.EntryInfo.Parent ?
            getMetadataEntry(item.EntryInfo.Parent) :
            getMetadataEntry(item.EntryInfo.ItemId)
        createItemEntry(item.EntryInfo, user, meta)
        count++
    }
    let hiddenItems = totalCount - count

    let totalCost = (await Promise.all(costFinders)).reduce((p, c) => p + c, 0)
    setGlobalStats({
        "Results": count,
        "Hidden": hiddenItems,
        "Cost": totalCost
    })
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
    let count = 0
    let costFinders = []
    for (const item of items) {
        if (item.Parent && ignoreChildren) {
            continue
        }
        if (item.CopyOf && ignoreCopies) {
            continue
        }
        let user = getUserEntry(item.ItemId)
        costFinders.push(getTotalCostDeep(item))
        let meta = item.Parent ?
            getMetadataEntry(item.Parent) :
            getMetadataEntry(item.ItemId)
        createItemEntry(item, user, meta)
        count++
    }
    let hiddenItems = items.length - count

    let totalCost = (await Promise.all(costFinders)).reduce((p, c) => p + c, 0)
    setGlobalStats({
        "Results": count,
        "Hidden": hiddenItems,
        "Cost": totalCost
    })
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

    let tags = /**@type {string}*/(data.get("tags"))

    let formats = []
    formats.push(data.getAll("format").map(Number))

    let status = /**@type {string}*/(data.get("status"))

    /**@type {DBQuery}*/
    let query = {
        title: enTitle,
        type: ty,
        //@ts-ignore
        format: formats.length ? formats : [-1],
        tags: tags,
        status
        // format: Number(format)
    }

    loadQueriedEntries(query).then(entries => {
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

async function newEntry() {
    const form = /**@type {HTMLFormElement}*/(document.getElementById("new-item-form"))
    const data = new FormData(form)
    /**@type {Record<string, FormDataEntryValue>}*/
    let validEntries = {}
    for (let [name, value] of data.entries()) {
        if (value == "") continue
        validEntries[name] = value
    }
    const queryString = "?" + Object.entries(validEntries).map(v => `${v[0]}=${encodeURI(String(v[1]))}`).join("&")

    let res = await fetch(`${apiPath}/add-entry${queryString}`)
    let text = await res.text()
    alert(text)

    await Promise.all([
        loadUserEntries(),
        loadMetadata(),
        loadAllEntries(),
        loadEntryTree()
    ])
}

function main() {
    loadFormats()
        .then(loadUserEntries)
        .then(loadMetadata)
        .then(loadAllEntries)
        .then(loadEntryTree)
        .then(renderEntryTree)
        .catch(console.error)
}
main()
