//TODO:
//Add 3 checkboxes for what the stats of a collection should be
//* itself
//* children
//* copies
//eg: if itself + children is selected
//the stats of the collection will be the stats of itself + stats of children
//eg: the total cost


/**
 * @typedef DisplayMode
 * @type {object}
 * @property {(entry: InfoEntry, updateStats?: boolean) => any} add
 * @property {(entry: InfoEntry, updateStats?: boolean) => any} sub
 * @property {(entry: InfoEntry[], updateStats?: boolean) => any} addList
 * @property {(entry: InfoEntry[], updateStats?: boolean) => any} subList
 */


/**
 * @typedef GlobalsNewUi
 * @type {object}
 * @property {Record<string, UserEntry>} userEntries
 * @property {Record<string, MetadataEntry>} metadataEntries
 * @property {Record<string, InfoEntry>} entries
 * @property {UserEvent[]} events
 * @property {InfoEntry[]} results
 * @property {InfoEntry[]} selectedEntries
 */
/**@type {GlobalsNewUi}*/

let globalsNewUi = {
    userEntries: {},
    metadataEntries: {},
    entries: {},
    results: [],
    events: [],
    selectedEntries: []
}

/**
 * @param {bigint} itemId
 */
function* findDescendants(itemId) {
    let entries = Object.values(globalsNewUi.entries)
    yield* entries.values()
        .filter(v => v.ParentId === itemId)
}

/**
 * @param {bigint} itemId
 */
function* findCopies(itemId) {
    let entries = Object.values(globalsNewUi.entries)
    yield* entries.values()
        .filter(v => v.CopyOf === itemId)
}

/**
 * @param {InfoEntry} collectionEntry
 * @param {boolean} itself
 * @param {boolean} children
 * @param {boolean} copies
 */
function sumCollectionStats(collectionEntry, itself = true, children = true, copies = false) {
    const stats = {
        totalItems: 0,
        cost: 0
    }
    if (itself) {
        stats.totalItems++
        stats.cost += collectionEntry.PurchasePrice
    }
    if (children) {
        for (let child of findDescendants(collectionEntry.ItemId)) {
            stats.totalItems++
            stats.cost += child.PurchasePrice
        }
    }
    if (copies) {
        for (let copy of findCopies(collectionEntry.ItemId)) {
            stats.totalItems++
            stats.cost += copy.PurchasePrice
        }
    }
    return stats
}

const viewAllElem = /**@type {HTMLInputElement}*/(document.getElementById("view-all"))

const displayItems = /**@type {HTMLElement}*/(document.getElementById("entry-output"))

const statsOutput = /**@type {HTMLElement}*/(document.querySelector(".result-stats"))

const modes = [modeDisplayEntry, modeGraphView, modeCalc, modeGallery]
const modeOutputIds = ["entry-output", "graph-output", "calc-output", "gallery-output"]

let idx = modeOutputIds.indexOf(location.hash.slice(1))

let mode = modes[idx]

/**
 * @param {InfoEntry} item
 * @param {DisplayMode} mode
 * @param {boolean} [updateStats=true]
 */
function selectItem(item, mode, updateStats = true) {
    globalsNewUi.selectedEntries.push(item)
    mode.add(item, updateStats)
}

/**
 * @param {InfoEntry} item
 * @param {boolean} [updateStats=true]
 */
function deselectItem(item, updateStats = true) {
    globalsNewUi.selectedEntries = globalsNewUi.selectedEntries.filter(a => a.ItemId !== item.ItemId)
    mode.sub(item, updateStats)
}

/**
 * @param {InfoEntry[]} itemList
 * @param {DisplayMode} mode
 * @param {boolean} [updateStats=true]
 */
function selectItemList(itemList, mode, updateStats = true) {
    globalsNewUi.selectedEntries = globalsNewUi.selectedEntries.concat(itemList)
    mode.addList(itemList, updateStats)
}

/**
 * @param {InfoEntry} item
 * @param {boolean} [updateStats=true]
 */
function toggleItem(item, updateStats = true) {
    if (globalsNewUi.selectedEntries.find(a => a.ItemId === item.ItemId)) {
        deselectItem(item, updateStats)
    } else {
        selectItem(item, mode, updateStats)
    }
}

function clearItems() {
    mode.subList(globalsNewUi.selectedEntries)
    globalsNewUi.selectedEntries = []
}

document.querySelector(".view-toggle")?.addEventListener("change", e => {
    mode.subList(globalsNewUi.selectedEntries)

    let name = /**@type {HTMLSelectElement}*/(e.target).value

    let curModeIdx = modeOutputIds.indexOf(name)

    mode = modes[curModeIdx]
    location.hash = name

    mode.addList(globalsNewUi.selectedEntries)

})


async function newEntry() {
    const form = /**@type {HTMLFormElement}*/(document.getElementById("new-item-form"))
    document.getElementById("new-entry")?.hidePopover()
    const data = new FormData(form)

    let artStyle = 0

    const styles = ['is-anime', 'is-cartoon', 'is-handrawn', 'is-digital', 'is-cgi', 'is-live-action']
    for (let i = 0; i < styles.length; i++) {
        let style = styles[i]
        if (data.get(style)) {
            artStyle |= 2 ** i
            data.delete(style)
        }
    }

    /**@type {Record<string, FormDataEntryValue>}*/
    let validEntries = {}
    for (let [name, value] of data.entries()) {
        if (value == "") continue
        validEntries[name] = value
    }
    const queryString = "?" + Object.entries(validEntries).map(v => `${v[0]}=${encodeURIComponent(String(v[1]))}`).join("&") + `&art-style=${artStyle}`

    let res = await fetch(`${apiPath}/add-entry${queryString}`)
    let text = await res.text()
    if (res.status !== 200) {
        alert(text)
        return
    }
    await refreshInfo()

    clearItems()
    clearSidebar()

    let json = parseJsonL(mkStrItemId(text))
    selectItem(json, mode, true)
    renderSidebarItem(json)
}

function resetResultStats() {
    for (let node of statsOutput.querySelectorAll(".stat") || []) {
        node.setAttribute("data-value", "0")
    }
    return {
        totalCost: 0,
        count: 0
    }
}

let resultStatsProxy = new Proxy({
    count: 0,
    totalCost: 0,
    reset() {
        this.count = 0
        this.totalCost = 0
    }
}, {
    set(obj, prop, value) {
        //@ts-ignore
        if (!Reflect.set(...arguments)) {
            return false
        }
        let el = /**@type {HTMLElement}*/(statsOutput.querySelector(`[data-stat-name="${String(prop)}"]`))
        el.setAttribute("data-value", String(value))
        return true
    }
})

/**
 * @typedef ResultStats
 * @type {object}
 * @property {number} totalCost
 * @property {number} count
 */

/**
 * @param {keyof ResultStats} key
 * @param {number} value
 */
function changeResultStats(key, value) {
    resultStatsProxy[key] += value
}

/**
 * @param {InfoEntry} item
 * @param {number} [multiplier=1]
 */
function changeResultStatsWithItem(item, multiplier = 1) {
    changeResultStats("totalCost", item.PurchasePrice * multiplier)
    changeResultStats("count", 1 * multiplier)
}

/**
 * @param {InfoEntry[]} items
 * @param {number} [multiplier=1]
 */
function changeResultStatsWithItemList(items, multiplier = 1) {
    for (let item of items) {
        changeResultStatsWithItem(item, multiplier)
    }
}

async function loadInfoEntries() {
    const res = await fetch(`${apiPath}/query-v3?search=${encodeURIComponent(`status=="Planned" | status=="Viewing"`)}`)
        .catch(console.error)
    if (!res) {
        alert("Could not load entries")
        return
    }
    let itemsText = await res.text()
    /**@type {string[]}*/
    let jsonL = itemsText.split("\n").filter(Boolean)
    /**@type {Record<string, InfoEntry>}*/
    let obj = {}
    for (let item of jsonL
        .map(mkStrItemId)
        .map(parseJsonL)
    ) {
        obj[item.ItemId] = item
    }

    //globalsNewUi.entries should contain ALL entries
    {
        const res = await fetch(`${apiPath}/list-entries`).catch(console.error)

        if (!res) {
            alert("Could not load all entries")
            return
        }

        const text = await res.text()
        let jsonL = text.split("\n").filter(Boolean)
        /**@type {Record<string, InfoEntry>}*/
        let obj = {}
        for (let item of jsonL
            .map(mkStrItemId)
            .map(parseJsonL)
        ) {
            obj[item.ItemId] = item
        }
        return globalsNewUi.entries = obj
    }
}

/**
 * @param {bigint} id
 * @returns {MetadataEntry?}
 */
function findMetadataById(id) {
    return globalsNewUi.metadataEntries[String(id)]
}

/**
 * @param {bigint} id
 * @returns {UserEntry?}
 */
function findUserEntryById(id) {
    return globalsNewUi.userEntries[String(id)]
}

/**
 * @param {bigint} id
 * @returns {UserEvent[]}
 */
function findUserEventsById(id) {
    return globalsNewUi.events.filter(v => v.ItemId === id)
}

/**
 * @param {bigint} id
 * @returns {InfoEntry?}
 */
function findInfoEntryById(id) {
    return globalsNewUi.entries[String(id)]
}

/**@returns {Promise<Record<string, UserEntry>>}*/
async function loadUserEntries() {
    let items = await loadList("engagement/list-entries")
    /**@type {Record<string, UserEntry>}*/
    let obj = {}
    for (let item of items) {
        obj[item.ItemId] = item
    }
    return globalsNewUi.userEntries = obj
}

async function loadUserEvents() {
    return globalsNewUi.events = await loadList("engagement/list-events")
}

/**@returns {Promise<Record<string, MetadataEntry>>}*/
async function loadMetadata() {
    let items = await loadList("metadata/list-entries")
    /**@type {Record<string, MetadataEntry>}*/
    let obj = {}
    for (let item of items) {
        obj[item.ItemId] = item
    }
    return globalsNewUi.metadataEntries = obj
}

/**
 * @param {ShadowRoot} root
 * @param {InfoEntry} item
 */
function saveItemChanges(root, item) {
    if (!confirm("Are you sure you want to save changes?")) {
        return
    }

    let userEntry = findUserEntryById(item.ItemId)
    if (!userEntry) return

    let notes = /**@type {HTMLElement}*/(root?.querySelector(".notes")).innerHTML
    if (notes === "<br>") {
        notes = ""
    }
    userEntry.Notes = notes

    let infoTable = root.querySelector("table.info-raw")
    let metaTable = root.querySelector("table.meta-info-raw")
    if (!infoTable || !metaTable) return

    /**@type {(table: Element, item: InfoEntry | MetadataEntry) => void}*/
    const updateWithTable = (table, item) => {
        for (let row of table?.querySelectorAll("tr") || []) {
            let nameChild = /**@type {HTMLElement}*/(row.firstElementChild)
            let valueChild = /**@type {HTMLElement}*/(row.firstElementChild?.nextElementSibling)
            let name = nameChild.innerText.trim()
            let value = valueChild.innerText.trim()
            if (!(name in item)) {
                console.log(`${name} NOT IN ITEM`)
                continue
            } else if (name === "ItemId") {
                console.log("Skipping ItemId")
                continue
            }
            let ty = item[/**@type {keyof typeof item}*/(name)].constructor
            //@ts-ignore
            item[name] = ty(value)
        }
    }

    updateWithTable(infoTable, item)
    let meta = findMetadataById(item.ItemId)
    if (!meta) return
    updateWithTable(metaTable, meta)

    //TODO: also save meta table


    const infoStringified = mkIntItemId(
        JSON.stringify(
            item,
            (_, v) => typeof v === 'bigint' ? String(v) : v
        )
    )

    const metaStringified = mkIntItemId(
        JSON.stringify(
            meta, (_, v) => typeof v === 'bigint' ? String(v) : v
        )
    )

    const userStringified = mkIntItemId(
        JSON.stringify(
            userEntry,
            (_, v) => typeof v === "bigint" ? String(v) : v
        )
    )

    let promises = []

    let engagementSet = fetch(`${apiPath}/engagement/set-entry`, {
        body: userStringified,
        method: "POST"
    })
        .then(res => res.text())
        .then(console.log)
        .catch(console.error)

    promises.push(engagementSet)

    let entrySet = fetch(`${apiPath}/set-entry`, {
        body: infoStringified,
        method: "POST"
    })
        .then(res => res.text())
        .then(console.log)
        .catch(console.error)

    promises.push(entrySet)

    let metaSet = fetch(`${apiPath}/metadata/set-entry`, {
        body: metaStringified,
        method: "POST"
    }).then(res => res.text())
        .then(console.log)
        .catch(console.error)

    promises.push(metaSet)

    Promise.all(promises).then(() => {
        refreshInfo().then(() => {
            refreshDisplayItem(item)
        })
    })

}

/**
 * @param {InfoEntry} item
 */
function deleteEntry(item) {
    if (!confirm("Are you sure you want to delete this item")) {
        return
    }
    fetch(`${apiPath}/delete-entry?id=${item.ItemId}`).then(res => {
        if (res?.status != 200) {
            console.error(res)
            alert("Failed to delete item")
            return
        }
        alert(`Deleted: ${item.En_Title} (${item.Native_Title} : ${item.ItemId})`)
        refreshInfo()
            .then(() => {
                deselectItem(item)
                removeSidebarItem(item)
            })
    })
}


/**
 * @param {ShadowRoot} _root
 * @param {InfoEntry} item
 */
function overwriteEntryMetadata(_root, item) {
    if (!confirm("Are you sure you want to overwrite the metadata with a refresh")) {
        return
    }

    fetch(`${apiPath}/metadata/fetch?id=${item.ItemId}`).then(res => {
        if (res.status !== 200) {
            console.error(res)
            alert("Failed to get metadata")
            return
        }
        refreshInfo()
            .then(() => {
                refreshDisplayItem(item)
                refreshSidebarItem(item)
            })
    })
}

/**
 * @param {InfoEntry} item
 * @param {UserEntry} user
 * @param {MetadataEntry} meta
 * @param {UserEvent[]} events
 * @param {HTMLElement} el
 */
function applyDisplayAttrs(item, user, meta, events, el) {
    el.setAttribute("data-title", item.En_Title)
    el.setAttribute("data-item-id", String(item.ItemId))
    el.setAttribute("data-format", String(item.Format))
    el.setAttribute("data-view-count", String(user.ViewCount))
    el.setAttribute("data-type", item.Type)

    user.CurrentPosition && el.setAttribute('data-user-current-position', user.CurrentPosition)
    meta.Thumbnail && el.setAttribute("data-thumbnail-src", meta.Thumbnail)
    meta.Title && el.setAttribute("data-true-title", meta.Title)
    meta.Native_Title && el.setAttribute("data-native-title", meta.Native_Title)
    user.ViewCount > 0 && el.setAttribute("data-user-rating", String(user.UserRating))
    user.Notes && el.setAttribute('data-user-notes', user.Notes)
    user.Status && el.setAttribute("data-user-status", user.Status)
    item.PurchasePrice && el.setAttribute("data-cost", String(item.PurchasePrice))
    meta.Description && el.setAttribute("data-description", meta.Description)
    meta?.MediaDependant && el.setAttribute("data-media-dependant", meta.MediaDependant)

    if (events.length) {
        let eventsStr = events.map(e => `${e.Event}:${e.Timestamp}`).join(",")
        el.setAttribute("data-user-events", eventsStr)
    }

    if (meta.Rating) {
        el.setAttribute("data-audience-rating-max", String(meta.RatingMax))
        el.setAttribute("data-audience-rating", String(meta.Rating))
    }

    el.setAttribute("data-info-raw", JSON.stringify(item, (_, v) => typeof v === 'bigint' ? v.toString() : v))
    el.setAttribute("data-meta-info-raw", JSON.stringify(meta, (_, v) => typeof v === 'bigint' ? v.toString() : v))
}

/**
 * @param {string} provider
 * @param {string} search
 * @param {HTMLElement} selectionElemOutput
 * @returns {Promise<string>}
 */
async function titleIdentification(provider, search, selectionElemOutput) {
    let res = await identify(search, provider)
    let text = await res.text()
    let [_, rest] = text.split("\x02")

    /**@type {any[]}*/
    let items
    try{
        items = rest.split("\n").filter(Boolean).map(v => JSON.parse(v))
    }
    catch(err){
        console.error("Could not parse json", rest.split('\n'))
        return ""
    }

    while (selectionElemOutput.children.length) {
        selectionElemOutput.firstChild?.remove()
    }

    selectionElemOutput.showPopover()

    return await new Promise(RETURN => {
        for (let result of items) {
            let fig = document.createElement("figure")

            let img = document.createElement("img")
            img.src = result.Thumbnail
            img.style.cursor = "pointer"
            img.width = 100

            img.addEventListener("click", _e => {
                selectionElemOutput.hidePopover()
                RETURN(result.ItemId)
            })

            let title = document.createElement("h3")
            title.innerText = result.Title || result.Native_Title
            title.title = result.Native_Title || result.Title

            fig.append(title)
            fig.append(img)
            selectionElemOutput.append(fig)
        }
    })
}

/**
 * @param {HTMLFormElement} form
 */
async function itemIdentification(form) {
    form.parentElement?.hidePopover()
    let data = new FormData(form)

    let provider = /**@type {string}*/(data.get("provider"))

    let queryType = /**@type {"by-title" | "by-id"}*/(data.get("query-type"))

    let search = /**@type {string}*/(data.get("search"))

    let shadowRoot = /**@type {ShadowRoot}*/(form.getRootNode())

    let itemId = shadowRoot.host.getAttribute("data-item-id")

    if (!itemId) {
        alert("Could not get item id")
        return
    }

    let finalItemId = ""

    switch (queryType) {
        case "by-title":
            let titleSearchContainer = /**@type {HTMLDialogElement}*/(shadowRoot.getElementById("identify-items"))
            finalItemId = await titleIdentification(provider, search, titleSearchContainer)
            break
        case "by-id":
            finalItemId = search
            break
    }
    finalizeIdentify(finalItemId, provider, BigInt(itemId))
        .then(refreshInfo)
        .then(() => {
            let newItem = globalsNewUi.entries[itemId]
            refreshDisplayItem(newItem)
            refreshSidebarItem(newItem)
        })
}

viewAllElem.addEventListener("change", e => {
    clearItems()
    if (/**@type {HTMLInputElement}*/(e.target)?.checked) {
        selectItemList(globalsNewUi.results, mode)
    } else {
        resultStatsProxy.reset()
    }
})

/**
 * @typedef ClientSearchFilters
 * @type {object}
 * @property {number} start
 * @property {number} end
 * @property {string} newSearch
 * @property {string} sortBy
 * @property {boolean} children
 * @property {boolean} copies
 */

/**
 * @param {FormData} searchForm
 * @returns {ClientSearchFilters}
 *
 * @description pulls out relevant filters for the client
 * has the side effect of modifying search-query, removing any parts deemed filters for the client
 * eg: \[start:end\]
 */
function parseClientsideSearchFiltering(searchForm) {
    let start = 0
    let end = -1

    let search = /**@type {string}*/(searchForm.get("search-query"))

    let slicematch = search.match(/\[(\d*):(-?\d*)\]/)
    if (slicematch) {
        start = +slicematch[1]
        end = +slicematch[2]
        search = search.replace(slicematch[0], "")

        searchForm.set("search-query", search)
    }

    let sortBy = /**@type {string}*/(searchForm.get("sort-by"))

    let children = /**@type {string}*/(searchForm.get("children"))
    let copies = /**@type {string}*/(searchForm.get("copies"))

    return {
        start,
        end,
        newSearch: search,
        sortBy,
        children: children === "on",
        copies: copies === "on"
    }
}

/**
 * @param {InfoEntry[]} entries
 * @param {ClientSearchFilters} filters
 */
function applyClientsideSearchFiltering(entries, filters) {
    if (!filters.children) {
        entries = entries.filter(v => v.ParentId === 0n)
    }
    if (!filters.copies) {
        entries = entries.filter(v => v.CopyOf === 0n)
    }

    if (filters.sortBy !== "") {
        entries = sortEntries(entries, filters.sortBy)
    }

    if (filters.end < 0) {
        filters.end += entries.length + 1
    }

    return entries.slice(filters.start, filters.end)
}

async function loadSearch() {
    let form = /**@type {HTMLFormElement}*/(document.getElementById("sidebar-form"))

    let formData = new FormData(form)

    let filters = parseClientsideSearchFiltering(formData)

    let entries = await doQuery3(String(formData.get("search-query")))

    entries = applyClientsideSearchFiltering(entries, filters)

    globalsNewUi.results = entries

    clearItems()
    if (entries.length === 0) {
        alert("No results")
        return
    }
    renderSidebar(entries)
}

/**
 * @param {number} rating
 * @param {number} maxRating
 */
function normalizeRating(rating, maxRating) {
    return rating / maxRating * 100
}

async function refreshInfo() {
    return Promise.all([
        loadInfoEntries(),
        loadMetadata(),
        loadUserEntries(),
        loadUserEvents()
    ])
}

/**@param {KeyboardEvent} e*/
function handleRichText(e) {
    /**@type {Record<string, () => ([string, string[]] | [string, string[]][])>}*/
    const actions = {
        "b": () => ["bold", []],
        "i": () => ["italic", []],
        "h": () => ["hiliteColor", [prompt("highlight color (yellow)") || "yellow"]],
        "f": () => ["foreColor", [prompt("Foreground color (black)") || "black"]],
        "t": () => ["fontName", [prompt("Font name (sans-serif)") || "sans-serif"]],
        "T": () => ["fontSize", [prompt("Font size (12pt)") || "12pt"]],
        "I": () => ["insertImage", [prompt("Image url (/favicon.ico)") || "/favicon.ico"]],
        "e": () => [
            ["enableObjectResizing", []],
            ["enableAbsolutePositionEditor", []],
            ["enableInlineTableEditing", []],
        ],
        "s": () => ["strikeThrough", []],
        "u": () => ['underline', []],
        "m": () => ["insertHTML", [getSelection()?.toString() || prompt("html (html)") || "html"]],
        "f12": () => ["removeFormat", []]
    }
    if (!e.ctrlKey) return
    let key = e.key
    if (key in actions) {
        let res = actions[key]()
        if (typeof res[0] === "string") {
            let [name, args] = res
            //@ts-ignore
            document.execCommand(name, false, ...args)
        } else {
            for (let [name, args] of res) {
                //@ts-ignore
                document.execCommand(name, false, ...args)
            }
        }
        e.preventDefault()
    }

}

async function main() {
    await refreshInfo()

    let tree = Object.values(globalsNewUi.entries).sort((a, b) => {
        let aUInfo = findUserEntryById(a.ItemId)
        let bUInfo = findUserEntryById(b.ItemId)
        if (!aUInfo || !bUInfo) return 0
        return bUInfo?.UserRating - aUInfo?.UserRating
    })

    globalsNewUi.results = tree
    renderSidebar(tree)
}

main()

let servicing = false
async function remote2LocalThumbService() {
    if (servicing) return

    servicing = true
    for (let item in globalsNewUi.metadataEntries) {
        let metadata = globalsNewUi.metadataEntries[item]
        let thumbnail = metadata.Thumbnail

        let userTitle = globalsNewUi.entries[item].En_Title
        let userNativeTitle = globalsNewUi.entries[item].Native_Title

        if (!thumbnail) continue
        if (thumbnail.startsWith(`${apiPath}/resource/thumbnail`)) continue
        if (thumbnail.startsWith(`${location.origin}${apiPath}/resource/thumbnail`)) {
            updateThumbnail(metadata.ItemId, `${apiPath}/resource/thumbnail?id=${metadata.ItemId}`)
            continue
        }

        console.log(`${userTitle || userNativeTitle || metadata.Title || metadata.Native_Title} Has a remote image url, downloading`)

        fetch(`${apiPath}/resource/download-thumbnail?id=${metadata.ItemId}`).then(res => res.text()).then(console.log)

        updateThumbnail(metadata.ItemId, `${apiPath}/resource/thumbnail?id=${metadata.ItemId}`).then(res => res.text()).then(console.log)

        await new Promise(res => setTimeout(res, 200))

        if (!servicing) break
    }

    console.log("ALL IMAGES HAVE BEEN INLINED")

    servicing = false
}
