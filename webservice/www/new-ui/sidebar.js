const sidebarItems = /**@type {HTMLElement}*/(document.querySelector(".sidebar--items"))

/**@type {InfoEntry[]}*/
const sidebarQueue = []

/**@type {Set<string>} */
const sidebarIntersected = new Set()
const sidebarObserver = new IntersectionObserver((entries) => {
    for (let entry of entries) {
        //keep track of which items have already triggered a new item to be added
        //otherwise each time the user scrolls, more items are added which is ugly lol
        const entryId = entry.target.getAttribute("data-entry-id") || "NA"
        if (entry.isIntersecting && sidebarQueue.length && !sidebarIntersected.has(entryId)) {
            sidebarIntersected.add(entryId)
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

    sidebarIntersected.clear()

    sidebarQueue.length = 0
    while (sidebarItems.children.length) {
        if (sidebarItems.firstChild?.nodeType === 1) {
            sidebarObserver.unobserve(sidebarItems.firstChild)
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
        changeSidebarItemData(item, user, meta, el.shadowRoot)
    } else {
        sidebarObserver.unobserve(el)
        renderSidebarItem(item)
    }
}


/**
 * @param {InfoEntry} item
 */
function removeSidebarItem(item) {
    sidebarIntersected.delete(String(item.ItemId))

    sidebarItems.querySelector(`[data-entry-id="${item.ItemId}"]`)?.remove()
}
/**
 * @param {InfoEntry} item
 * @param {UserEntry} user
 * @param {MetadataEntry} meta
 * @param {ShadowRoot} el
 */
function updateSidebarEntryContents(item, user, meta, el) {
    const titleEl = /**@type {HTMLDivElement}*/(el.querySelector(".title"))
    const imgEl = /**@type {HTMLImageElement}*/(el.querySelector(".thumbnail"))

    //Title
    titleEl.innerText = item.En_Title || item.Native_Title
    titleEl.title = meta.Title

    imgEl.alt = "thumbnail"

    //Type
    let typeIcon = typeToSymbol(String(item.Type))
    titleEl.setAttribute("data-type-icon", typeIcon)

    //Thumbnail
    if (imgEl.src !== meta.Thumbnail) {
        imgEl.src = meta.Thumbnail
    }

    //Release year
    if (meta.ReleaseYear)
        titleEl.setAttribute("data-release-year", String(meta.ReleaseYear))
    else
        titleEl.setAttribute("data-release-year", "unknown")
}

/**
 * @param {InfoEntry} item
 * @param {UserEntry} user
 * @param {MetadataEntry} meta
 * @param {HTMLElement} el
 */
function changeSidebarItemData(item, user, meta, el) {
    const e = new CustomEvent("data-changed", {
        detail: {
            item,
            user,
            meta,
        }
    })
    el.dispatchEvent(e)
    el.setAttribute("data-entry-id", String(item.ItemId))
}

/**
 * @param {InfoEntry} item
 * @param {DisplayMode} mode
 */
function dblclickSideBarEntry(item, mode) {
    clearItems()
    selectItem(item, mode)
}

/**
 * @param {InfoEntry} item
 */
function clickSideBarEntry(item) {
                toggleItem(item)
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

    sidebarParent.append(elem)

    let img = elem.shadowRoot?.querySelector("img")
    if (img) {
        img.addEventListener("click", _e => {
            (innerWidth / innerHeight >= 1)
            ? clickSideBarEntry(item)
            : dblclickSideBarEntry(item, mode)
        })
        img.addEventListener("dblclick", _e => {
            (innerWidth / innerHeight >= 1)
            ? dblclickSideBarEntry(item, mode)
            : clickSideBarEntry(item)
        })
    }

    elem.addEventListener("data-changed", function(e) {
        const event = /**@type {CustomEvent}*/(e)
        const item = /**@type {InfoEntry}*/(event.detail.item)
        const user = /**@type {UserEntry}*/(event.detail.user)
        const meta = /**@type {MetadataEntry}*/(event.detail.meta)
        const events = /**@type {UserEvent[]}*/(event.detail.events)
        updateSidebarEntryContents(item, user, meta, elem.shadowRoot)
    })

    changeSidebarItemData(item, user, meta, elem)
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

