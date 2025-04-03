const sidebarItems = document.querySelector(".sidebar--items") as HTMLElement

const sidebarQueue: InfoEntry[] = []

const sidebarIntersected: Set<string> = new Set()
const sidebarObserver = new IntersectionObserver((entries) => {
    for (let entry of entries) {
        if (entry.isIntersecting) {
            entry.target.dispatchEvent(new Event("on-screen-appear"))
        }
    }
}, {
    root: document.querySelector("nav.sidebar"),
    rootMargin: "0px",
    threshold: 0.1
})

function clearSidebar() {

    sidebarIntersected.clear()

    while (sidebarItems.firstElementChild) {
        sidebarObserver.unobserve(sidebarItems.firstElementChild)
        sidebarItems.firstElementChild.remove()
    }
}

function refreshSidebarItem(item: InfoEntry) {
    let el = document.querySelector(`sidebar-entry[data-entry-id="${item.ItemId}"]`) as HTMLElement
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

    //thumbnail source is updated in `on-screen-appear` event as to make sure it doesn't request 300 images at once
    imgEl.alt = "thumbnail"

    //Type
    let typeIcon = typeToSymbol(String(item.Type))
    titleEl.setAttribute("data-type-icon", typeIcon)

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

function renderSidebarItem(item: InfoEntry, sidebarParent: HTMLElement | DocumentFragment = sidebarItems) {
    let elem = document.createElement("sidebar-entry")

    sidebarObserver.observe(elem)

    let meta = findMetadataById(item.ItemId)
    let user = findUserEntryById(item.ItemId)
    if (!user || !meta) return

    sidebarParent.append(elem)

    let img = elem.shadowRoot?.querySelector("img") as HTMLImageElement
    if (img) {
        img.addEventListener("click", e => {
            if (e.ctrlKey) {
                clickSideBarEntry(item)
            } else {

                dblclickSideBarEntry(item, mode)
            }
        })
    }

    elem.addEventListener("on-screen-appear", function(e) {
        if (img.src !== meta.Thumbnail) {
            img.src = meta.Thumbnail
        }
    })

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

function renderSidebar(entries: InfoEntry[]) {
    if (viewAllElem.checked) {
        selectItemList(entries, mode)
    } else {
        selectItem(entries[0], mode)
    }
    clearSidebar()
    for (let i = 0; i < entries.length; i++) {
            renderSidebarItem(entries[i], sidebarItems)
    }
}

