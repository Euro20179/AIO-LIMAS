/**
 * @typedef GlobalsNewUi
 * @type {object}
 * @property {UserEntry[]} userEntries
 * @property {MetadataEntry[]} metadataEntries
 * @property {EntryTree} tree
 * @property {UserEvent[]} events
 */
/**@type {GlobalsNewUi}*/

let globalsNewUi = {
    userEntries: [],
    metadataEntries: [],
    tree: {},
    events: [],
}
const sidebarItems = /**@type {HTMLElement}*/(document.querySelector(".sidebar--items"))
const displayItems = /**@type {HTMLElement}*/(document.getElementById("entry-output"))

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

/**
 * @param {bigint} id
 * @returns {MetadataEntry?}
 */
function findMetadataById(id) {
    for (let item in globalsNewUi.metadataEntries) {
        let entry = globalsNewUi.metadataEntries[item]
        if (entry.ItemId === id) {
            return entry
        }
    }
    return null
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
    for (let item in globalsNewUi.userEntries) {
        let entry = globalsNewUi.userEntries[item]
        if (entry.ItemId === id) {
            userEntryCache[String(id)] = entry
            return entry
        }
    }
    return null
}

/**
 * @param {bigint} id
 * @returns {InfoEntry?}
 */
function findInfoEntryById(id) {
    for (let item in globalsNewUi.tree) {
        let entry = globalsNewUi.tree[item]
        if (entry.EntryInfo.ItemId === id) {
            return entry.EntryInfo
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
}

/**
 * @param {InfoEntry} item
 */
function renderDisplayItem(item) {
    let el = document.createElement("display-entry")
    el.setAttribute("data-title", item.En_Title)
    el.setAttribute("data-item-id", String(item.ItemId))

    let meta = findMetadataById(item.ItemId)
    let user = findUserEntryById(item.ItemId)

    if (meta?.Thumbnail) {
        el.setAttribute("data-thumbnail-src", meta.Thumbnail)
    }
    displayItems.append(el)
}

/**
 * @param {InfoEntry} item
 */
function removeDisplayItem(item) {
    let el = /**@type {HTMLElement}*/(displayItems.querySelector(`[data-item-id="${item.ItemId}"]`))
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

    let viewCkBox = /**@type {HTMLInputElement?}*/(elem.shadowRoot?.querySelector("[name='view']"))
    if (viewCkBox) {
        viewCkBox.addEventListener("change", e => {
            if (viewCkBox?.checked && !document.querySelector(`[data-item-id="${item.ItemId}"]`)) {
                renderDisplayItem(item)
            } else {
                removeDisplayItem(item)
            }
        })
    }
}

document.getElementById("view-all")?.addEventListener("change", e => {
    clearMainDisplay()
    if (e.target?.checked) {
        for (let elem of document.querySelectorAll("sidebar-entry")) {
            let id = elem.getAttribute("data-entry-id")
            if (!id) continue
            let item = findInfoEntryById(BigInt(id))
            if (!item) continue
            renderDisplayItem(item)
        }
    }
})

/**
* @param {EntryTree} tree
*/
function renderSidebar(tree) {
    for (let id in tree) {
        let item = tree[id]
        renderSidebarItem(item.EntryInfo)
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

    let formatN = undefined
    if (format.length) {
        formatN = format.map(Number)
    }

    let entries = await loadQueriedEntries({
        status: status.join(","),
        type: type.join(","),
        format: formatN
    })

    if (sortBy != "") {
        if (sortBy == "rating") {
            entries = entries.sort((a, b) => {
                let aUInfo = findUserEntryById(a.ItemId)
                let bUInfo = findUserEntryById(b.ItemId)
                return bUInfo?.UserRating - aUInfo?.UserRating
            })
        }
    }

    for (let item of entries) {
        renderSidebarItem(item)
    }
}



async function main() {
    let tree = await loadEntryTree()
    globalsNewUi.tree = tree
    await loadMetadata()
    await loadUserEntries()

    renderSidebar(sortTree(tree, ([_, aInfo], [__, bInfo]) => {
        let aUInfo = findUserEntryById(aInfo.EntryInfo.ItemId)
        let bUInfo = findUserEntryById(bInfo.EntryInfo.ItemId)
        return bUInfo?.UserRating - aUInfo?.UserRating
    }))
}
main()
