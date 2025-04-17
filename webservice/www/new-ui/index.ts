type ClientSearchFilters = {
    filterRules: string[]
    newSearch: string
    sortBy: string
}

type GlobalsNewUi = {
    userEntries: Record<string, UserEntry>
    metadataEntries: Record<string, MetadataEntry>
    entries: Record<string, InfoEntry>
    events: UserEvent[]
    results: InfoEntry[]
    selectedEntries: InfoEntry[]
    libraries: Record<string, InfoEntry>
    viewingLibrary: bigint
}

let globalsNewUi: GlobalsNewUi = {
    userEntries: {},
    metadataEntries: {},
    entries: {},
    results: [],
    events: [],
    selectedEntries: [],
    libraries: {},
    viewingLibrary: 0n,
}

const displayItems = document.getElementById("entry-output") as HTMLElement

const statsOutput = document.getElementById("result-stats") as HTMLElement

const librarySelector = document.getElementById("library-selector") as HTMLSelectElement
const newEntryLibrarySelector = document.querySelector("[name=\"libraryId\"]") as HTMLSelectElement

librarySelector.onchange = function() {
    let val = librarySelector.value
    globalsNewUi.viewingLibrary = BigInt(val)

    loadSearch()
}

function getUserExtra(user: UserEntry, prop: string) {
    return JSON.parse(user.Extra).AIOWeb?.[prop] || null
}

function setUserExtra(user: UserEntry, prop: string, value: string) {
    let extra = JSON.parse(user.Extra)
    let AIOWeb = extra.AIOWeb || {}
    AIOWeb[prop] = value
    extra.AIOWeb = AIOWeb
    user.Extra = JSON.stringify(extra)
}

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
    set(_obj, prop, value) {
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

function updateLibraryDropdown() {
    for (let i in globalsNewUi.libraries) {
        let item = globalsNewUi.libraries[i]
        const opt = document.createElement("option")
        opt.value = String(item["ItemId"])
        opt.innerText = item["En_Title"]
        librarySelector.append(opt)
    }
    newEntryLibrarySelector.innerHTML = librarySelector.innerHTML
}

async function loadLibraries() {
    const items = await doQuery3("type = 'Library'")

    librarySelector.innerHTML = '<option value="0">Library</option>'
    for (let item of items) {
        globalsNewUi.libraries[String(item["ItemId"])] = item
    }
    updateLibraryDropdown()
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

async function loadUserEntries(): Promise<Record<string, UserEntry>> {
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

    return {
        filterRules: filters,
        newSearch: search,
        sortBy,
    }
}

function applyClientsideSearchFiltering(entries: InfoEntry[], filters: ClientSearchFilters) {

    entries = entries.filter(v => v.Library === globalsNewUi.viewingLibrary)

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
                    entries.sort(() => eval(fn))
                    break
                }
            }
        } else if (filter.startsWith("filter")) {
            let expr = filter.slice("filter".length).trim()
            entries = entries.filter(() => eval(expr))
        } else if (filter === "!child") {
            entries = entries.filter(v => v.ParentId === 0n)
        } else if (filter === "!copy") {
            entries = entries.filter(v => v.CopyOf === 0n)
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

    let entries = await doQuery3(String(filters.newSearch) || "#")

    entries = applyClientsideSearchFiltering(entries, filters)

    setResultStat("results", entries.length)

    globalsNewUi.results = entries

    clearItems()
    if (entries.length === 0) {
        setError("No results")
        clearSidebar()
        return
    }
    renderSidebar(entries)
}

function normalizeRating(rating: number, maxRating: number) {
    return rating / maxRating * 100
}

function genericMetadata(itemId: bigint): MetadataEntry {
    return {
        ItemId: itemId,
        Rating: 0,
        RatingMax: 0,
        Description: "",
        ReleaseYear: 0,
        Thumbnail: "",
        MediaDependant: "",
        Datapoints: "{}",
        Title: "",
        Native_Title: "",
    }
}

function genericUserEntry(itemId: bigint): UserEntry {
    return {
        ItemId: itemId,
        Status: "",
        ViewCount: 0,
        UserRating: 0,
        Notes: "",
        CurrentPosition: "",
        Extra: "{}",
    }
}

/**
 * @description Updates information in the global information table, also refreshes the mode's dispaly info, and the sidebar info
 */
function updateInfo({
    entries, metadataEntries, events, userEntries
}: Partial<GlobalsNewUi>, del = false) {

    if(del && events) {
        let specific = events.map(v => `${v.ItemId}:${v.Timestamp || v.After}:${v.Event}`)
        globalsNewUi.events = globalsNewUi.events.filter(v => !specific.includes(`${v.ItemId}:${v.Timestamp || v.After}:${v.Event}`))
    } else if(events) {
        globalsNewUi.events = globalsNewUi.events.concat(events)
    }

    let updatedLibraries = false

    for (let entry in entries) {
        let e = entries[entry]
        if (e.Type === "Library") {
            if (!del) {
                globalsNewUi.libraries[entry] = e
            } else {
                delete globalsNewUi.libraries[entry]
            }
            updatedLibraries = true
        }


        if (!del) {
            let metadata = metadataEntries?.[entry] || findMetadataById(e.ItemId) || genericMetadata(e.ItemId)
            let user = userEntries?.[entry] || findUserEntryById(e.ItemId) || genericUserEntry(e.ItemId)

            globalsNewUi.entries[entry] = e
            globalsNewUi.metadataEntries[entry] = metadata
            globalsNewUi.userEntries[entry] = user

            refreshSidebarItem(e)
            if(mode.refresh)
                mode.refresh(e.ItemId)
        } else {
            delete globalsNewUi.entries[entry]
            delete globalsNewUi.metadataEntries[entry]
            delete globalsNewUi.userEntries[entry]
        }
    }

    if (updatedLibraries) {
        updateLibraryDropdown()
    }
}

async function refreshInfo() {
    return Promise.all([
        loadLibraries(),
        loadInfoEntries(),
        loadMetadata(),
        loadUserEntries(),
        loadUserEvents()
    ])
}

async function main() {
    await refreshInfo()

    if (initialSearch) {
        let entries = await doQuery3(initialSearch)

        setResultStat("results", entries.length)

        globalsNewUi.results = entries

        if (entries.length === 0) {
            alert("No results")
            return
        }
        renderSidebar(entries)
    } else {
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
        if (thumbnail.startsWith(`${apiPath}/resource/thumbnail`) || thumbnail.startsWith(`${apiPath}/resource/get-thumbnail`)) continue
        console.log(thumbnail)

        //FIXME: this should work, but for some reason just doesn't
        if (thumbnail.startsWith("data:")) continue
        // if (thumbnail.startsWith(`${location.origin}${apiPath}/resource/thumbnail`)) {
        //     updateThumbnail(metadata.ItemId, `${apiPath}/resource/thumbnail?id=${metadata.ItemId}`)
        //     continue
        // }

        console.log(`${userTitle || userNativeTitle || metadata.Title || metadata.Native_Title} Has a remote image url, downloading`)

        fetch(`${apiPath}/resource/download-thumbnail?id=${metadata.ItemId}`).then(res => {
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
