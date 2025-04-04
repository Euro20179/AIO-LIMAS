const galleryItems = document.getElementById("gallery-items") as HTMLDivElement

function renderGalleryItem(item: InfoEntry, parent = galleryItems){
    let el = new Image()
    let meta = findMetadataById(item.ItemId)
    if(meta?.Thumbnail) {
        el.src = meta.Thumbnail
    }
    el.title = item.En_Title
    el.setAttribute("data-item-id", String(item.ItemId))

    parent.appendChild(el)
}

function removeGalleryItem(entry: InfoEntry) {
    let el = galleryItems.querySelector(`[data-item-id="${entry.ItemId}"]`)
    el?.remove()
}

const modeGallery: DisplayMode = {
    add(entry, updateStats = true) {
        updateStats && changeResultStatsWithItem(entry)
        renderGalleryItem(entry)
    },

    sub(entry, updateStats = true) {
        updateStats && changeResultStatsWithItem(entry, -1)
        removeGalleryItem(entry)
    },

    addList(entry, updateStats = true) {
        updateStats && changeResultStatsWithItemList(entry, 1)
        for (let item of entry) {
            renderGalleryItem(item)
        }
    },

    subList(entry, updateStats = true) {
        updateStats && changeResultStatsWithItemList(entry, -1)
        for (let item of entry) {
            removeGalleryItem(item)
        }
    }
}
