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

    await refreshInfo()
}

function resetResultStats() {
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
    el.innerText = String(resultStats[key])
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
        if (item.PurchasePrice) {
            console.log(item.En_Title, item.PurchasePrice)
        }
        changeResultStatsWithItem(item, multiplier)
    }
}

/**
    * @returns {Promise<EntryTree>}
*/
async function loadEntryTree() {
    const res = await fetch(`${apiPath}/list-tree`)
        .catch(console.error)
    if (!res) {
        alert("Could not load entries")
        return {}
    } else {
        let itemsText = await res.text()
        itemsText = itemsText
            .replaceAll(/"ItemId":\s*(\d+),/g, "\"ItemId\": \"$1\",")
            .replaceAll(/"Parent":\s*(\d+),/g, "\"Parent\": \"$1\",")
            .replaceAll(/"CopyOf":\s*(\d+),/g, "\"CopyOf\": \"$1\"")
        const bigIntProperties = ["ItemId", "Parent", "CopyOf"]
        let json = JSON.parse(itemsText, (key, v) => bigIntProperties.includes(key) ? BigInt(v) : v)
        return json
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
                        .then(() => refreshDisplayItem(item))
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
        renderDisplayItem(item, el, false)
    } else {
        renderDisplayItem(item, null, false)
    }
}

/**
 * @param {InfoEntry} item
 * @param {HTMLElement?} [el=null] 
 * @param {boolean} [updateStats=true]
 */
function renderDisplayItem(item, el = null, updateStats = true) {
    let doEventHooking = false
    if (!el) {
        el = document.createElement("display-entry")
        doEventHooking = true
    }

    el.setAttribute("data-title", item.En_Title)
    el.setAttribute("data-item-id", String(item.ItemId))

    let meta = findMetadataById(item.ItemId)
    let user = findUserEntryById(item.ItemId)
    let events = findUserEventsById(item.ItemId)

    if (updateStats) {
        changeResultStatsWithItem(item)
    }

    if (meta?.Thumbnail) {
        el.setAttribute("data-thumbnail-src", meta.Thumbnail)
    }

    if (user?.UserRating) {
        el.setAttribute("data-user-rating", String(user.UserRating))
    }

    if (user?.Notes) {
        el.setAttribute('data-user-notes', user.Notes)
    }

    if (item.PurchasePrice) {
        el.setAttribute("data-cost", String(item.PurchasePrice))
    }

    if (meta?.Description) {
        el.setAttribute("data-description", meta.Description)
    }

    if (events.length) {
        let eventsStr = events.map(e => `${e.Event}:${e.Timestamp}`).join(",")
        el.setAttribute("data-user-events", eventsStr)
    }

    displayItems.append(el)

    let root = el.shadowRoot
    if (!root) return

    if (doEventHooking) {
        hookActionButtons(root, item)

        let closeButton = root.querySelector(".close")
        closeButton?.addEventListener("click", e => {
            removeDisplayItem(item)
        })

        let deleteBtn = root.querySelector(".delete")
        deleteBtn?.addEventListener("click", e => {
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
        })
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
 */
function renderSidebarItem(item) {
    let elem = document.createElement("sidebar-entry")
    let meta = findMetadataById(item.ItemId)
    let user = findUserEntryById(item.ItemId)

    elem.setAttribute("data-entry-id", String(item.ItemId))

    elem.setAttribute("data-title", item.En_Title)
    if (meta?.Thumbnail) {
        elem.setAttribute("data-thumbnail-src", meta.Thumbnail)
    }

    if (user?.Status) {
        elem.setAttribute("data-user-status", user.Status)
    }

    if (user?.UserRating) {
        elem.setAttribute("data-user-rating", String(user.UserRating))
    }

    if (item.PurchasePrice) {
        elem.setAttribute("data-cost", String(Math.round(item.PurchasePrice * 100) / 100))
    }


    sidebarItems.append(elem)

    let img = elem.shadowRoot?.querySelector("img")
    if (img) {
        img.addEventListener("click", e => {
            if (!isItemDisplayed(item.ItemId)) {
                renderDisplayItem(item)
            } else {
                removeDisplayItem(item)
            }
        })
    }
}

document.getElementById("view-all")?.addEventListener("change", e => {
    clearMainDisplay()
    if (/**@type {HTMLInputElement}*/(e.target)?.checked) {
        changeResultStatsWithItemList(globalsNewUi.results)
        for (let item of globalsNewUi.results) {
            renderDisplayItem(item, null, false)
        }
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

/**
 * @param {EntryTree} tree
 * @param {(a: [string, TreeNode], b: [string, TreeNode]) => number} sorter
 */
function sortTree(tree, sorter) {
    return Object.fromEntries(
        Object.entries(tree)
            .sort(sorter)
    )
}

/**
 * @param {EntryTree} tree
 * @param {(item: [string, TreeNode]) => boolean} filter
 */
function filterTree(tree, filter) {
    return Object.fromEntries(
        Object.entries(tree).filter(filter)
    )
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

    let formatN = undefined
    if (format.length) {
        formatN = format.map(Number)
    }

    let entries = await loadQueriedEntries({
        status: status.join(","),
        type: type.join(","),
        format: formatN,
        title: search
    })

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
        }
    }

    clearMainDisplay()
    if (entries.length === 0) {
        alert("No results")
        return
    }
    renderSidebar(entries)
}

async function refreshInfo() {
    return Promise.all([
        loadInfoEntries(),
        loadMetadata(),
        loadUserEntries(),
        loadUserEvents()
    ])
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
