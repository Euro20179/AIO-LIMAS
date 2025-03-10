const sidebarItems = /**@type {HTMLElement}*/(document.querySelector(".sidebar--items"))

/**@type {InfoEntry[]}*/
const sidebarQueue = []

const sidebarObserver = new IntersectionObserver((entries) => {
    for (let entry of entries) {
        if (entry.isIntersecting && sidebarQueue.length) {
            let newItem = sidebarQueue.shift()
            if (!newItem) continue
            renderSidebarItem(newItem)
        }
    }
}, {
    root: document.querySelector("nav.sidebar"),
    rootMargin: "0px",
    threshold: 0.1
})

function clearSidebar() {
    sidebarQueue.length = 0
    while (sidebarItems.children.length) {
        try {
            sidebarObserver.unobserve(sidebarItems.firstChild)
        } catch (err) {
            console.error(err)
        }
        sidebarItems.firstChild?.remove()
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
        sidebarObserver.unobserve(el)
        renderSidebarItem(item)
    }
}


/**
 * @param {InfoEntry} item
 */
function removeSidebarItem(item) {
    sidebarItems.querySelector(`[data-entry-id="${item.ItemId}"]`)?.remove()
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

    meta?.Thumbnail && el.setAttribute("data-thumbnail-src", meta.Thumbnail)
    meta.ReleaseYear && el.setAttribute("data-release-year", String(meta.ReleaseYear))
    user?.Status && el.setAttribute("data-user-status", user.Status)
    user.ViewCount > 0 && el.setAttribute("data-user-rating", String(user.UserRating))
    item.PurchasePrice && el.setAttribute("data-cost", String(Math.round(item.PurchasePrice * 100) / 100))
}

/**
 * @param {InfoEntry} item
 * @param {HTMLElement | DocumentFragment} [sidebarParent=sidebarItems]
 */
function renderSidebarItem(item, sidebarParent = sidebarItems) {


    let elem = document.createElement("sidebar-entry")

    sidebarObserver.observe(elem)

    let meta = findMetadataById(item.ItemId)
    let user = findUserEntryById(item.ItemId)
    if (!user || !meta) return

    applySidebarAttrs(item, user, meta, elem)

    sidebarParent.append(elem)

    let img = elem.shadowRoot?.querySelector("img")
    if (img) {
        img.addEventListener("click", _e => {
            toggleItem(item)
        })
        img.addEventListener("dblclick", _e => {
            clearItems()
            selectItem(item, mode)
        })
    }
}

/**
* @param {InfoEntry[]} entries
*/
function renderSidebar(entries) {
    if (viewAllElem.checked) {
        selectItemList(entries, mode)
    } else {
        selectItem(entries[0], mode)
    }
    clearSidebar()
    for (let i = 0; i < entries.length; i++) {
        if (i > 15) {
            sidebarQueue.push(entries[i])
        } else {
            renderSidebarItem(entries[i], sidebarItems)
        }
    }
    // for (let item of entries) {
    //     renderSidebarItem(item, frag)
    // }
}

