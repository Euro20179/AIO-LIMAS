/**
 * @typedef GlobalsNewUi
 * @type {object}
 * @property {UserEntry[]} userEntries
 * @property {MetadataEntry[]} metadataEntries
 * @property {InfoEntry[]} entries
 * @property {UserEvent[]} events
 * @property {InfoEntry[]} results
 */
/**@type {GlobalsNewUi}*/

let globalsNewUi = {
    userEntries: [],
    metadataEntries: [],
    entries: [],
    results: [],
    events: [],
}

const sidebarItems = /**@type {HTMLElement}*/(document.querySelector(".sidebar--items"))
const displayItems = /**@type {HTMLElement}*/(document.getElementById("entry-output"))

const statsOutput = /**@type {HTMLElement}*/(document.querySelector(".result-stats"))

async function newEntry() {
    const form = /**@type {HTMLFormElement}*/(document.getElementById("new-item-form"))
    document.getElementById("new-entry")?.hidePopover()
    const data = new FormData(form)
    /**@type {Record<string, FormDataEntryValue>}*/
    let validEntries = {}
    for (let [name, value] of data.entries()) {
        if (value == "") continue
        validEntries[name] = value
    }
    const queryString = "?" + Object.entries(validEntries).map(v => `${v[0]}=${encodeURIComponent(String(v[1]))}`).join("&")

    let res = await fetch(`${apiPath}/add-entry${queryString}`)
    let text = await res.text()
    if (res.status !== 200) {
        alert(text)
        return
    }
    await refreshInfo()

    clearMainDisplay()
    clearSidebar()

    let json = parseJsonL(mkStrItemId(text))
    renderDisplayItem(json)
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

/**
 * @typedef ResultStats
 * @type {object}
 * @property {number} totalCost
 * @property {number} count
 */
let resultStats = resetResultStats()

/**
 * @param {keyof ResultStats} key
 * @param {number} value
 */
function changeResultStats(key, value) {
    let el = /**@type {HTMLElement}*/(statsOutput.querySelector(`[data-name="${key}"]`))
    resultStats[key] += value
    el.setAttribute("data-value", String(resultStats[key]))
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
    const res = await fetch(`${apiPath}/list-entries`)
        .catch(console.error)
    if (!res) {
        alert("Could not load entries")
    } else {
        let itemsText = await res.text()
        /**@type {string[]}*/
        let jsonL = itemsText.split("\n").filter(Boolean)
        globalsNewUi.entries = jsonL
            .map(mkStrItemId)
            .map(parseJsonL)
        return globalsNewUi.entries
    }
}

/**
 * @param {bigint} id
 * @param {Record<string, any>} entryTable
 * @returns {any}
 */
function findEntryById(id, entryTable) {
    for (let item in entryTable) {
        let entry = entryTable[item]
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
function findMetadataById(id) {
    return findEntryById(id, globalsNewUi.metadataEntries)
}

/**@type {Record<string, UserEntry>}*/
let userEntryCache = {}
/**
 * @param {bigint} id
 * @returns {UserEntry?}
 */
function findUserEntryById(id) {
    if (userEntryCache[`${id}`]) {
        return userEntryCache[`${id}`]
    }
    return findEntryById(id, globalsNewUi.userEntries)
}

/**
 * @param {bigint} id
 * @returns {UserEvent[]}
 */
function findUserEventsById(id) {
    let events = []
    for (let item in globalsNewUi.events) {
        let entry = globalsNewUi.events[item]
        if (entry.ItemId === id) {
            events.push(entry)
        }
    }
    return events
}

/**
 * @param {bigint} id
 * @returns {InfoEntry?}
 */
function findInfoEntryById(id) {
    for (let item of globalsNewUi.entries) {
        if (item.ItemId === id) {
            return item
        }
    }
    return null
}

/**@returns {Promise<UserEntry[]>}*/
async function loadUserEntries() {
    userEntryCache = {}
    return globalsNewUi.userEntries = await loadList("engagement/list-entries")
}

async function loadUserEvents() {
    return globalsNewUi.events = await loadList("engagement/list-events")
}

/**@returns {Promise<MetadataEntry[]>}*/
async function loadMetadata() {
    return globalsNewUi.metadataEntries = await loadList("metadata/list-entries")
}

function clearSidebar() {
    while (sidebarItems.children.length) {
        sidebarItems.firstChild?.remove()
    }
}

function clearMainDisplay() {
    while (displayItems.children.length) {
        displayItems.firstChild?.remove()
    }
    resultStats = resetResultStats()
}

/**
 * @param {ShadowRoot} shadowRoot
 * @param {InfoEntry} item 
 */
function hookActionButtons(shadowRoot, item) {
    for (let btn of shadowRoot.querySelectorAll("[data-action]") || []) {
        let action = btn.getAttribute("data-action")
        btn.addEventListener("click", e => {
            if (!confirm(`Are you sure you want to ${action} this entry`)) {
                return
            }

            let queryParams = ""
            if (action === "Finish") {
                let rating = prompt("Rating")
                while (isNaN(Number(rating))) {
                    rating = prompt("Not a number\nrating")
                }
                queryParams += `&rating=${rating}`
            }

            fetch(`${apiPath}/engagement/${action?.toLowerCase()}-media?id=${item.ItemId}${queryParams}`)
                .then(res => res.text())
                .then(text => {
                    alert(text)
                    refreshInfo()
                        .then(() => {
                            refreshDisplayItem(item)
                        })
                })
        })
    }

}

/**
 * @param {InfoEntry} item
 */
function refreshDisplayItem(item) {
    let el = /**@type {HTMLElement}*/(document.querySelector(`display-entry[data-item-id="${item.ItemId}"]`))
    if (el) {
        let user = findUserEntryById(item.ItemId)
        let events = findUserEventsById(item.ItemId)
        let meta = findMetadataById(item.ItemId)
        if (!user || !events || !meta) return
        applyDisplayAttrs(item, user, meta, events, el)
    } else {
        renderDisplayItem(item, null, false)
    }
}

/**
 * @param {InfoEntry} item
 */
function refreshSidebarItem(item) {
    let el = /**@type {HTMLElement}*/(document.querySelector(`sidebar-entry[data-entry-id="${item.ItemId}"]`))
    if (el) {
        let user = findUserEntryById(item.ItemId)
        let meta = findMetadataById(item.ItemId)
        if (!user || !meta) return
        applySidebarAttrs(item, user, meta, el)
    } else {
        renderSidebarItem(item)
    }
}

/**
 * @param {ShadowRoot} root
 * @param {InfoEntry} item
 */
function saveItemChanges(root, item) {
    if (!confirm("Are you sure you want to save changes?")) {
        return
    }

    let queryParams = ""

    let notes = /**@type {HTMLElement}*/(root?.querySelector(".notes")).innerHTML
    if (notes === "<br>") {
        notes = ""
    }
    queryParams += `&notes=${encodeURIComponent(notes)}`

    fetch(`${apiPath}/engagement/mod-entry?id=${item.ItemId}${queryParams}`)
        .then(res => {
            return res.text()
        })
        .then(console.log)
        .catch(console.error)

    queryParams = ""
    let title = /**@type {HTMLElement}*/(root.querySelector(".title")).innerText
    queryParams += `&en-title=${encodeURIComponent(title)}`
    fetch(`${apiPath}/mod-entry?id=${item.ItemId}${queryParams}`)
        .then(res => res.text())
        .then(console.log)
        .catch(console.error)
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
                removeDisplayItem(item)
                removeSidebarItem(item)
            })
    })
}


/**
 * @param {ShadowRoot} root
 * @param {InfoEntry} item
 */
function overwriteEntryMetadata(root, item) {
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

    if (user.CurrentPosition) {
        el.setAttribute('data-user-current-position', user.CurrentPosition)
    }

    if (meta.Thumbnail) {
        el.setAttribute("data-thumbnail-src", meta.Thumbnail)
    }

    if (meta.Title) {
        el.setAttribute("data-true-title", meta.Title)
    }
    if (meta.Native_Title) {
        el.setAttribute("data-native-title", meta.Native_Title)
    }

    if (user.ViewCount > 0) {
        el.setAttribute("data-user-rating", String(user.UserRating))
    }

    if (meta.Rating) {
        el.setAttribute("data-audience-rating-max", String(meta.RatingMax))
        el.setAttribute("data-audience-rating", String(meta.Rating))
    }

    if (user.Notes) {
        el.setAttribute('data-user-notes', user.Notes)
    }
    if (user.Status) {
        el.setAttribute("data-user-status", user.Status)
    }

    if (item.PurchasePrice) {
        el.setAttribute("data-cost", String(item.PurchasePrice))
    }

    if (meta.Description) {
        el.setAttribute("data-description", meta.Description)
    }

    if (events.length) {
        let eventsStr = events.map(e => `${e.Event}:${e.Timestamp}`).join(",")
        el.setAttribute("data-user-events", eventsStr)
    }

    if (meta?.MediaDependant) {
        el.setAttribute("data-media-dependant", meta.MediaDependant)
    }

    el.setAttribute("data-info-raw", JSON.stringify(item, (_, v) => typeof v === 'bigint' ? v.toString() : v))
}

/**
 * @param {InfoEntry} item
 * @param {HTMLElement?} [el=null] 
 * @param {boolean} [updateStats=true]
 */
function renderDisplayItem(item, el = null, updateStats = true) {
    //TODO:
    //when the user clicks on the info table on the side
    //open a dialog box with the information pre filled in with the information already there
    //and allow the user to edit each item
    //when the user clicks submit, it'll send a mod-entry request
    let doEventHooking = false
    if (!el) {
        el = document.createElement("display-entry")
        doEventHooking = true
    }


    let meta = findMetadataById(item.ItemId)
    let user = findUserEntryById(item.ItemId)
    let events = findUserEventsById(item.ItemId)
    if (!user || !meta || !events) return

    if (updateStats) {
        changeResultStatsWithItem(item)
    }

    applyDisplayAttrs(item, user, meta, events, el)

    displayItems.append(el)

    let root = el.shadowRoot
    if (!root) return

    if (doEventHooking) {
        hookActionButtons(root, item)

        let closeButton = root.querySelector(".close")
        closeButton?.addEventListener("click", _ => {
            removeDisplayItem(item)
        })

        let copyToBtn = root.querySelector(".copy-to")
        copyToBtn?.addEventListener("click", _ => {
            let id, idInt
            do {
                id = prompt("Copy user info to (item id)")
                idInt = BigInt(String(id))
            }
            while (isNaN(Number(idInt)))
            copyUserInfo(item.ItemId, idInt)
                .then(res => res?.text())
                .then(console.log)
        })

        let viewCount = root.querySelector(".view-count")
        if (viewCount) {
            viewCount.addEventListener("click", e => {
                let count
                do {
                    count = prompt("New view count")
                    if (count == null) {
                        return
                    }
                } while (isNaN(Number(count)))
                fetch(`${apiPath}/engagement/mod-entry?id=${item.ItemId}&view-count=${count}`)
                    .then(res => res.text())
                    .then(alert)
                    .catch(console.error)
            })
        }

        let deleteBtn = root.querySelector(".delete")
        deleteBtn?.addEventListener("click", _ => {
            deleteEntry(item)
        })

        let refreshBtn = root.querySelector(".refresh")
        refreshBtn?.addEventListener("click", _ => {
            overwriteEntryMetadata(/**@type {ShadowRoot}*/(root), item)
        })

        let saveBtn = root.querySelector(".save")
        saveBtn?.addEventListener("click", _ => {
            saveItemChanges(/**@type {ShadowRoot}*/(root), item)
        })

        let identifyBtn = /**@type {HTMLButtonElement}*/(root.querySelector(".identify"))
        identifyBtn?.addEventListener("click", e => {
            let provider = prompt("provider: anilist, omdb")
            let title = prompt("title search: ")
            identify(String(title), provider || "anilist")
                .then(res => res.text())
                .then(jsonL => {
                    let [provider, rest] = jsonL.split("\x02")
                    jsonL = rest
                    const data = jsonL
                        .split("\n")
                        .filter(Boolean)
                        .map((v) => JSON.parse(v))
                    let container = /**@type {HTMLDialogElement}*/(root?.getElementById("identify-items"))
                    container.innerHTML = ""

                    for (let result of data) {
                        let fig = document.createElement("figure")

                        let img = document.createElement("img")
                        img.src = result.Thumbnail
                        img.style.cursor = "pointer"

                        img.addEventListener("click", e => {
                            finalizeIdentify(result.ItemId, provider, item.ItemId)
                                .then(refreshInfo)
                                .then(() => {
                                    let newItem = findEntryById(item.ItemId, globalsNewUi.entries)
                                    refreshDisplayItem(newItem)
                                    refreshSidebarItem(newItem)
                                })
                                .catch(console.error)
                        })

                        let title = document.createElement("h3")
                        title.innerText = result.Title
                        title.title = result.Native_Title

                        fig.append(title)
                        fig.append(img)
                        container.append(fig)
                    }

                })

        })

        let ratingSpan = root.querySelector(".rating")
        ratingSpan?.addEventListener("click", _ => {
            let newRating = prompt("New rating")
            if (isNaN(Number(newRating)) || newRating === null || newRating === "") {
                return
            }

            fetch(`${apiPath}/engagement/mod-entry?id=${item.ItemId}&rating=${newRating}`)
                .then(refreshInfo)
                .then(() => {
                    let newItem = findEntryById(item.ItemId, globalsNewUi.entries)
                    refreshDisplayItem(newItem)
                    refreshSidebarItem(newItem)
                })
                .catch(console.error)
        })

        for (let el of root.querySelectorAll("[contenteditable]")) {
            /**@type {HTMLElement}*/(el).addEventListener("keydown", handleRichText)
        }
    }
}

/**
 * @param {bigint} id
 * @returns {boolean}
 */
function isItemDisplayed(id) {
    let elem = document.querySelector(`[data-item-id="${id}"]`)
    if (elem) {
        return true
    }
    return false
}

/**
 * @param {InfoEntry} item
 */
function removeDisplayItem(item) {
    let el = /**@type {HTMLElement}*/(displayItems.querySelector(`[data-item-id="${item.ItemId}"]`))
    if (el) {
        el.remove()
        changeResultStatsWithItem(item, -1)
    }
}

/**
 * @param {InfoEntry} item
 */
function removeSidebarItem(item) {
    let el = /**@type {HTMLElement}*/(sidebarItems.querySelector(`[data-entry-id="${item.ItemId}"]`))
    if (el) {
        el.remove()
    }
}

/**
 * @param {InfoEntry} item
 * @param {UserEntry} user
 * @param {MetadataEntry} meta
 * @param {HTMLElement} el
 */
function applySidebarAttrs(item, user, meta, el) {
    el.setAttribute("data-entry-id", String(item.ItemId))
    el.setAttribute("data-title", item.En_Title)

    el.setAttribute("data-type", item.Type)
    if (meta?.Thumbnail) {
        el.setAttribute("data-thumbnail-src", meta.Thumbnail)
    }
    if (meta.ReleaseYear) {
        el.setAttribute("data-release-year", String(meta.ReleaseYear))
    }

    if (user?.Status) {
        el.setAttribute("data-user-status", user.Status)
    }

    if (user.ViewCount > 0) {
        el.setAttribute("data-user-rating", String(user.UserRating))
    }

    if (item.PurchasePrice) {
        el.setAttribute("data-cost", String(Math.round(item.PurchasePrice * 100) / 100))
    }
}

/**
 * @param {InfoEntry} item
 * @param {HTMLElement?} [elem=null] 
 */
function renderSidebarItem(item, elem = null) {
    let doEventHooking = false
    if (!elem) {
        doEventHooking = true
        elem = document.createElement("sidebar-entry")
    }
    let meta = findMetadataById(item.ItemId)
    let user = findUserEntryById(item.ItemId)
    if (!user || !meta) return

    applySidebarAttrs(item, user, meta, elem)

    sidebarItems.append(elem)

    if (doEventHooking) {
        let img = elem.shadowRoot?.querySelector("img")
        if (img) {
            img.addEventListener("click", e => {
                toggleDisplayItem(item)
            })
            img.addEventListener("dblclick", e => {
                clearMainDisplay()
                renderDisplayItem(item)
            })
        }
    }
}

/**
 * @param {InfoEntry} item
 */
function toggleDisplayItem(item) {
    if (!isItemDisplayed(item.ItemId)) {
        renderDisplayItem(item)
    } else {
        removeDisplayItem(item)
    }

}

document.getElementById("view-all")?.addEventListener("change", e => {
    clearMainDisplay()
    if (/**@type {HTMLInputElement}*/(e.target)?.checked) {
        changeResultStatsWithItemList(globalsNewUi.results)
        for (let item of globalsNewUi.results) {
            renderDisplayItem(item, null, false)
        }
    } else {
        resultStats = resetResultStats()
    }
})

/**
* @param {InfoEntry[]} entries
*/
function renderSidebar(entries) {
    renderDisplayItem(entries[0])
    for (let item of entries) {
        renderSidebarItem(item)
    }
}

async function treeFilterForm() {
    clearSidebar()
    let form = /**@type {HTMLFormElement}*/(document.getElementById("sidebar-form"))
    let data = new FormData(form)
    let sortBy = data.get("sort-by")
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

    let shortcuts = {
        "userRatingGt": "r>",
        "userRatingLt": "r<",
        "purchasePriceGt": "p>",
        "purchasePriceLt": "p<",
    }

    for (let word of search.split(" ")) {
        let [property, value] = word.split(":")
        if (!value) continue
        //@ts-ignore
        let shortcut = shortcuts[property]
        if (word.startsWith(shortcut)) {
            value = word.slice(shortcut.length)
        }
        search = search.replace(word, "").trim()
        //@ts-ignore
        queryData[property] = value
    }

    queryData.title = search

    let entries = await loadQueriedEntries(queryData)

    globalsNewUi.results = entries

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

    clearMainDisplay()
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

    let tree = globalsNewUi.entries.sort((a, b) => {

        let aUInfo = findUserEntryById(a.ItemId)
        let bUInfo = findUserEntryById(b.ItemId)
        if (!aUInfo || !bUInfo) return 0
        return bUInfo?.UserRating - aUInfo?.UserRating
    })

    globalsNewUi.results = tree
    renderSidebar(tree)
}

main()
