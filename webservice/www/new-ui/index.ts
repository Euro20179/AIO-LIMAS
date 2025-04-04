//TODO:
//Add 3 checkboxes for what the stats of a collection should be
//* itself
//* children
//* copies
//eg: if itself + children is selected
//the stats of the collection will be the stats of itself + stats of children
//eg: the total cost

type DisplayMode = {
    add: (entry: InfoEntry, updateStats?: boolean) => any
    sub: (entry: InfoEntry, updateStats?: boolean) => any
    addList: (entry: InfoEntry[], updateStats?: boolean) => any
    subList: (entry: InfoEntry[], updateStats?: boolean) => any
    putSelectedInCollection?: () => any
    addTagsToSelected?: () => any
    [key: string]: any
}


type GlobalsNewUi = {
    userEntries: Record<string, UserEntry>
    metadataEntries: Record<string, MetadataEntry>
    entries: Record<string, InfoEntry>
    events: UserEvent[]
    results: InfoEntry[]
    selectedEntries: InfoEntry[]
}

let globalsNewUi: GlobalsNewUi = {
    userEntries: {},
    metadataEntries: {},
    entries: {},
    results: [],
    events: [],
    selectedEntries: []
}

const viewAllElem = document.getElementById("view-all") as HTMLInputElement

const displayItems = document.getElementById("entry-output") as HTMLElement

const statsOutput = document.querySelector(".result-stats") as HTMLElement

const modes = [modeDisplayEntry, modeGraphView, modeCalc, modeGallery]
const modeOutputIds = ["entry-output", "graph-output", "calc-output", "gallery-output"]

let idx = modeOutputIds.indexOf(location.hash.slice(1))

let mode = modes[idx]

function* findDescendants(itemId: bigint) {
    let entries = Object.values(globalsNewUi.entries)
    yield* entries.values()
        .filter(v => v.ParentId === itemId)
}

function* findCopies(itemId: bigint) {
    let entries = Object.values(globalsNewUi.entries)
    yield* entries.values()
        .filter(v => v.CopyOf === itemId)
}

function sumCollectionStats(collectionEntry: InfoEntry, itself: boolean = true, children: boolean = true, copies: boolean = false) {
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


function selectItem(item: InfoEntry, mode: DisplayMode, updateStats: boolean = true) {
    globalsNewUi.selectedEntries.push(item)
    mode.add(item, updateStats)
}

function deselectItem(item: InfoEntry, updateStats: boolean = true) {
    globalsNewUi.selectedEntries = globalsNewUi.selectedEntries.filter(a => a.ItemId !== item.ItemId)
    mode.sub(item, updateStats)
}

function selectItemList(itemList: InfoEntry[], mode: DisplayMode, updateStats: boolean = true) {
    globalsNewUi.selectedEntries = globalsNewUi.selectedEntries.concat(itemList)
    mode.addList(itemList, updateStats)
}

function toggleItem(item: InfoEntry, updateStats: boolean = true) {
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

function putSelectedToCollection() {
    if(mode.putSelectedInCollection) {
        mode.putSelectedInCollection()
    } else {
        alert("This mode does not support putting items into a collection")
    }
}

function addTagsToSelected() {
    if(mode.addTagsToSelected) {
        mode.addTagsToSelected()
    } else {
        alert("This mode does not support adding tags to selected items")
    }
}


document.querySelector(".view-toggle")?.addEventListener("change", e => {
    mode.subList(globalsNewUi.selectedEntries)

    let name = (e.target as HTMLSelectElement).value

    let curModeIdx = modeOutputIds.indexOf(name)

    mode = modes[curModeIdx]
    location.hash = name

    mode.addList(globalsNewUi.selectedEntries)

})


async function newEntry() {
    const form = document.getElementById("new-item-form") as HTMLFormElement
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

    let validEntries: Record<string, FormDataEntryValue> = {}
    for (let [name, value] of data.entries()) {
        if (value == "") continue
        validEntries[name] = value
    }
    const queryString = "?" + Object.entries(validEntries).map(v => `${v[0]}=${encodeURIComponent(String(v[1]))}`).join("&") + `&art-style=${artStyle}`

    const tz = Intl.DateTimeFormat().resolvedOptions().timeZone

    let res = await fetch(`${apiPath}/add-entry${queryString}&timezone=${encodeURIComponent(tz)}`)
    let text = await res.text()
    if (res.status !== 200) {
        alert(text)
        return
    }
    await refreshInfo()

    clearItems()

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
    results: 0,
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
        let el = statsOutput.querySelector(`[data-stat-name="${String(prop)}"]`) as HTMLElement
        el.setAttribute("data-value", String(value))
        return true
    }
})

type ResultStats = {
    totalCost: number
    count: number
    results: number
}

function setResultStat(key: keyof ResultStats, value: number) {
    resultStatsProxy[key] = value
}

function changeResultStats(key: keyof ResultStats, value: number) {
    resultStatsProxy[key] += value
}

function changeResultStatsWithItem(item: InfoEntry, multiplier: number = 1) {
    changeResultStats("totalCost", item.PurchasePrice * multiplier)
    changeResultStats("count", 1 * multiplier)
}

function changeResultStatsWithItemList(items: InfoEntry[], multiplier: number = 1) {
    for (let item of items) {
        changeResultStatsWithItem(item, multiplier)
    }
}

async function loadInfoEntries() {
    const res = await fetch(`${apiPath}/list-entries?uid=${uid}`).catch(console.error)

    if (!res) {
        alert("Could not load all entries")
        return
    }

    const text = await res.text()
    let jsonL = text.split("\n").filter(Boolean)
    let obj: Record<string, InfoEntry> = {}
    for (let item of jsonL
        .map(mkStrItemId)
        .map(parseJsonL)
    ) {
        obj[item.ItemId] = item
    }

    setResultStat("results", jsonL.length)

    return globalsNewUi.entries = obj
}

function findMetadataById(id: bigint): MetadataEntry | null {
    return globalsNewUi.metadataEntries[String(id)]
}

function findUserEntryById(id: bigint): UserEntry | null {
    return globalsNewUi.userEntries[String(id)]
}

function findUserEventsById(id: bigint): UserEvent[] {
    return globalsNewUi.events.filter(v => v.ItemId === id)
}

function findInfoEntryById(id: bigint): InfoEntry | null {
    return globalsNewUi.entries[String(id)]
}

async function loadUserEntries(): Promise<Record<string, UserEntry>>  {
    let items = await loadList<UserEntry>("engagement/list-entries")
    let obj: Record<string, UserEntry> = {}
    for (let item of items) {
        obj[String(item.ItemId)] = item
    }
    return globalsNewUi.userEntries = obj
}

async function loadUserEvents() {
    return globalsNewUi.events = await loadList("engagement/list-events")
}

async function loadMetadata(): Promise<Record<string, MetadataEntry>> {
    let items = await loadList<MetadataEntry>("metadata/list-entries")
    let obj: Record<string, MetadataEntry> = {}
    for (let item of items) {
        obj[String(item.ItemId)] = item
    }
    return globalsNewUi.metadataEntries = obj
}

function saveItemChanges(root: ShadowRoot, item: InfoEntry) {
    if (!confirm("Are you sure you want to save changes?")) {
        return
    }

    const userEn_title = root.querySelector(".title") as HTMLHeadingElement

    if (userEn_title) {
        item.En_Title = userEn_title.innerText
    }

    let userEntry = findUserEntryById(item.ItemId)
    if (!userEntry) return

    let notes = (root?.querySelector(".notes"))?.innerHTML
    if (notes === "<br>") {
        notes = ""
    }
    userEntry.Notes = notes || ""

    let infoTable = root.querySelector("table.info-raw")
    let metaTable = root.querySelector("table.meta-info-raw")
    if (!infoTable || !metaTable) return

    const updateWithTable: (table: Element, item: InfoEntry | MetadataEntry) => void = (table, item) => {
        for (let row of table?.querySelectorAll("tr") || []) {
            let nameChild = row.firstElementChild as HTMLElement
            let valueChild = row.firstElementChild?.nextElementSibling as HTMLElement
            let name = nameChild.innerText.trim()
            let value = valueChild.innerText.trim()
            if (!(name in item)) {
                console.log(`${name} NOT IN ITEM`)
                continue
            } else if (name === "ItemId") {
                console.log("Skipping ItemId")
                continue
            }
            let ty = item[name as keyof typeof item].constructor
            //@ts-ignore
            item[name] = ty(value)
        }
    }

    updateWithTable(infoTable, item)
    let meta = findMetadataById(item.ItemId)
    if (!meta) return
    updateWithTable(metaTable, meta)


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

function deleteEntry(item: InfoEntry) {
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


function overwriteEntryMetadata(_root: ShadowRoot, item: InfoEntry) {
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

function changeDisplayItemData(item: InfoEntry, user: UserEntry, meta: MetadataEntry, events: UserEvent[], el: HTMLElement) {
    const e = new CustomEvent("data-changed", {
        detail: {
            item,
            user,
            meta,
            events,
        }
    })
    el.dispatchEvent(e)
    el.setAttribute("data-item-id", String(item.ItemId))
}

async function titleIdentification(provider: string, search: string, selectionElemOutput: HTMLElement): Promise<string> {
    let res = await identify(search, provider)
    let text = await res.text()
    let [_, rest] = text.split("\x02")

    let items: any[]
    try {
        items = rest.split("\n").filter(Boolean).map(v => JSON.parse(v))
    }
    catch (err) {
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

async function itemIdentification(form: HTMLFormElement) {
    form.parentElement?.hidePopover()
    let data = new FormData(form)

    let provider = data.get("provider") as string

    let queryType = data.get("query-type") as "by-title" | "by-id"

    let search = data.get("search") as string

    let shadowRoot = form.getRootNode() as ShadowRoot

    let itemId = shadowRoot.host.getAttribute("data-item-id")

    if (!itemId) {
        alert("Could not get item id")
        return
    }

    let finalItemId = ""

    switch (queryType) {
        case "by-title":
            let titleSearchContainer = shadowRoot.getElementById("identify-items") as HTMLDialogElement
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
    if ((e.target as HTMLInputElement)?.checked) {
        selectItemList(globalsNewUi.results, mode)
    } else {
        resultStatsProxy.reset()
    }
})

type ClientSearchFilters = {
    filterRules: string[]
    newSearch: string
    sortBy: string
    children: boolean
    copies: boolean
}

/**
 * @description pulls out relevant filters for the client
 * has the side effect of modifying search-query, removing any parts deemed filters for the client
 * eg: \[start:end\]
 */
function parseClientsideSearchFiltering(searchForm: FormData): ClientSearchFilters {
    // let start = 0
    // let end = -1

    let search = searchForm.get("search-query") as string
    let filters;
    [search, ...filters] = search.split("->")
    filters = filters.map(v => v.trim())

    let sortBy = searchForm.get("sort-by") as string

    let children = searchForm.get("children") as string
    let copies = searchForm.get("copies") as string

    return {
        filterRules: filters,
        newSearch: search,
        sortBy,
        children: children === "on",
        copies: copies === "on"
    }
}

function applyClientsideSearchFiltering(entries: InfoEntry[], filters: ClientSearchFilters) {
    if (!filters.children) {
        entries = entries.filter(v => v.ParentId === 0n)
    }
    if (!filters.copies) {
        entries = entries.filter(v => v.CopyOf === 0n)
    }

    if (filters.sortBy !== "") {
        entries = sortEntries(entries, filters.sortBy)
    }

    for (let filter of filters.filterRules) {
        filter = filter.trim()
        if (filter.startsWith("is")) {
            let ty = filter.split("is")[1]?.trim()
            if (!ty) continue
            ty = ty.replace(/(\w)(\S*)/g, (_, $1, $2) => {
                return $1.toUpperCase() + $2
            })
            entries = entries.filter(v => v.Type == ty)
        }

        let slicematch = filter.match(/\[(\d*):(-?\d*)\]/)
        if (slicematch) {
            let start = +slicematch[1]
            let end = +slicematch[2]
            entries = entries.slice(start, end)
        }

        if (filter.startsWith("/") && filter.endsWith("/")) {
            let re = new RegExp(filter.slice(1, filter.length - 1))
            entries = entries.filter(v => v.En_Title.match(re))
        } else if (filter == "shuffle" || filter == "shuf") {
            entries = entries.sort(() => Math.random() - Math.random())
        } else if (filter.startsWith("head")) {
            const n = filter.slice("head".length).trim() || 1
            entries = entries.slice(0, Number(n))
        } else if (filter.startsWith("tail")) {
            const n = filter.slice("tail".length).trim() || 1
            entries = entries.slice(entries.length - Number(n))
        } else if (filter.startsWith("sort")) {
            let type = filter.slice("sort".length).trim() || "a"
            const reversed = type.startsWith("-") ? -1 : 1
            if (reversed == -1) type = type.slice(1)
            switch (type[0]) {
                case "a":
                    entries.sort((a, b) => (a.En_Title > b.En_Title ? 1 : -1) * reversed)
                    break;
                case "e": {
                    const fn = type.slice(1)
                    entries.sort((a, b) => eval(fn))
                    break
                }
            }
        } else if (filter.startsWith("filter")) {
            let expr = filter.slice("filter".length).trim()
            entries = entries.filter((a) => eval(expr))
        }
    }

    // if (filters.end < 0) {
    //     filters.end += entries.length + 1
    // }
    return entries
    //
    // return entries.slice(filters.start, filters.end)
}

async function loadSearch() {
    let form = document.getElementById("sidebar-form") as HTMLFormElement

    let formData = new FormData(form)

    let filters = parseClientsideSearchFiltering(formData)

    let entries = await doQuery3(String(filters.newSearch))

    entries = applyClientsideSearchFiltering(entries, filters)

    setResultStat("results", entries.length)

    globalsNewUi.results = entries

    clearItems()
    if (entries.length === 0) {
        alert("No results")
        return
    }
    renderSidebar(entries)
}

function normalizeRating(rating: number, maxRating: number) {
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

function handleRichText(e: KeyboardEvent) {
    const actions: Record<string, () => ([string, string[]] | [string, string[]][])> = {
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
    if (initialSearch) {
        let entries = await doQuery3(initialSearch)

        setResultStat("results", entries.length)

        globalsNewUi.results = entries

        await refreshInfo()

        if (entries.length === 0) {
            alert("No results")
            return
        }
        renderSidebar(entries)
    } else {
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

        //FIXME: this should work, but for some reason just doesn't
        if (thumbnail.startsWith("data:")) continue
        // if (thumbnail.startsWith(`${location.origin}${apiPath}/resource/thumbnail`)) {
        //     updateThumbnail(metadata.ItemId, `${apiPath}/resource/thumbnail?id=${metadata.ItemId}`)
        //     continue
        // }

        console.log(`${userTitle || userNativeTitle || metadata.Title || metadata.Native_Title} Has a remote image url, downloading`)

        fetch(`${apiPath}/resource/download-thumbnail?id=${metadata.ItemId}&uid=${uid}`).then(res => {
            if (res.status !== 200) return ""
            return res.text()
        }).then(hash => {
            if (!hash) return
            console.log(`THUMBNAIL HASH: ${hash}`)
            updateThumbnail(metadata.ItemId, `${apiPath}/resource/get-thumbnail?hash=${hash}`).then(res => res.text()).then(console.log)
        })

        await new Promise(res => setTimeout(res, 200))

        if (!servicing) break
    }

    console.log("ALL IMAGES HAVE BEEN INLINED")

    servicing = false
}
