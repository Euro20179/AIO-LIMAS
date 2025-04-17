const sidebarItems = document.getElementById("sidebar-items") as HTMLElement

const sidebarIntersected: Set<string> = new Set()
const sidebarObserver = new IntersectionObserver((entries) => {
    for (let entry of entries) {
        if (entry.isIntersecting) {
            entry.target.dispatchEvent(new Event("on-screen-appear"))
        }
    }
}, {
    root: document.getElementById("sidebar"),
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
        changeSidebarItemData(item, user, meta, el)
    } else {
        renderSidebarItem(item)
    }

    let meta = findMetadataById(item.ItemId)
    if(meta)
        updateSidebarThumbnail(item.ItemId, meta?.Thumbnail)
}


function removeSidebarItem(item: InfoEntry) {
    sidebarIntersected.delete(String(item.ItemId))

    sidebarItems.querySelector(`[data-entry-id="${item.ItemId}"]`)?.remove()
}

function updateSidebarEntryContents(item: InfoEntry, user: UserEntry, meta: MetadataEntry, el: ShadowRoot) {
    const titleEl = el.getElementById("sidebar-title") as HTMLInputElement
    const imgEl = el.getElementById("sidebar-thumbnail") as HTMLImageElement

    //Title
    titleEl.value = item.En_Title
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

function changeSidebarItemData(item: InfoEntry, user: UserEntry, meta: MetadataEntry, el: HTMLElement) {
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

function sidebarEntryOpenMultiple(item: InfoEntry, mode: DisplayMode) {
    clearItems()
    selectItem(item, mode)
}

function sidebarEntryOpenOne(item: InfoEntry) {
    toggleItem(item)
}

function updateSidebarThumbnail(id: bigint, src: string) {
    const elem = sidebarItems.querySelector(`[data-entry-id="${id}"]`)
    if(!elem) return
    let img = elem.shadowRoot?.querySelector("img") as HTMLImageElement
    img.src = src
}

function renderSidebarItem(item: InfoEntry, sidebarParent: HTMLElement | DocumentFragment = sidebarItems) {
    let elem = document.createElement("sidebar-entry")

    sidebarObserver.observe(elem)

    let meta = findMetadataById(item.ItemId)
    let user = findUserEntryById(item.ItemId)
    if (!user || !meta) return elem

    sidebarParent.append(elem)

    let img = elem.shadowRoot?.querySelector("img") as HTMLImageElement
    if (img) {
        img.addEventListener("click", e => {
            if (e.ctrlKey) {
                sidebarEntryOpenOne(item)
            } else {

                sidebarEntryOpenMultiple(item, mode)
            }
        })
    }

    let title = elem.shadowRoot?.getElementById("sidebar-title") as HTMLInputElement
    if(title) {
        title.onchange = function() {
            if(title.value)
                updateInfoTitle(item.ItemId, title.value)
        }
    }

    elem.addEventListener("on-screen-appear", function(e) {
        if (img.src !== meta.Thumbnail) {
            img.src = meta.Thumbnail
        }
    })

    elem.addEventListener("data-changed", function(e) {
        const event = e as CustomEvent
        const item = event.detail.item as InfoEntry
        const user = event.detail.user as UserEntry
        const meta = event.detail.meta as MetadataEntry
        const events = event.detail.events as UserEvent[]
        updateSidebarEntryContents(item, user, meta, elem.shadowRoot as ShadowRoot)
    })

    changeSidebarItemData(item, user, meta, elem)

    return elem
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

