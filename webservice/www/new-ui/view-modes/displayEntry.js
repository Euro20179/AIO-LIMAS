/**@type {HTMLElement[]}*/
const displaying = []
/**@type {InfoEntry[]}*/
const displayQueue = []

/**
 * @param {IntersectionObserverEntry[]} entries
 */
function onIntersection(entries) {
    for(let entry of entries) {
        if(entry.isIntersecting && displayQueue.length) {
            let newItem = displayQueue.shift()
            if(!newItem) continue
            //bypass the mode.add mechanism, because
            //the add function needs to put in the back of the queue otherwise
            //something might get rendered ahead of somethign in the queue
            renderDisplayItem(newItem)
        }
    }
}

const observer = new IntersectionObserver(onIntersection, {
    root: document.querySelector("#entry-output"),
    rootMargin: "0px",
    threshold: 0.1
})

/**
 * @type {DisplayMode}
 */
const modeDisplayEntry = {
    add(entry, updateStats = true) {
        updateStats && changeResultStatsWithItem(entry)
        //if there is an item in the queue, this item should not render ahead of it
        if (displayQueue.length) 
            displayQueue.push(entry)
        else
            renderDisplayItem(entry)
    },

    sub(entry, updateStats = true) {
        updateStats && changeResultStatsWithItem(entry, -1)
        removeDisplayItem(entry)
    },

    addList(entry, updateStats = true) {
        updateStats && changeResultStatsWithItemList(entry, 1)

        //only render the first five items
        entry.slice(0, 5).forEach(v => renderDisplayItem(v))
        //the rest are queued
        entry.slice(5).forEach(v => displayQueue.push(v))
    },

    subList(entry, updateStats = true) {
        updateStats && changeResultStatsWithItemList(entry, -1)
        for (let item of entry) {
            removeDisplayItem(item)
        }
    }
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

            let queryParams = `?id=${item.ItemId}`
            if (action === "Finish") {
                let rating = promptNumber("Rating", "Not a number\nRating")
                if (rating !== null) {
                    queryParams += `&rating=${rating}`
                }
            }

            fetch(`${apiPath}/engagement/${action?.toLowerCase()}-media${queryParams}`)
                .then(res => res.text())
                .then(text => {
                    alert(text)
                    return refreshInfo()
                })
                .then(() => refreshDisplayItem(item))
        })
    }
}

/**
 * @param {InfoEntry} item
 * @param {HTMLElement | DocumentFragment} [parent=displayItems]
 */
function renderDisplayItem(item, parent = displayItems) {
    let el = document.createElement("display-entry")

    displaying.push(el)

    observer.observe(el)

    let meta = findMetadataById(item.ItemId)
    let user = findUserEntryById(item.ItemId)
    let events = findUserEventsById(item.ItemId)
    if (!user || !meta || !events) return

    applyDisplayAttrs(item, user, meta, events, el)


    parent.append(el)

    let root = el.shadowRoot
    if (!root) return

    /**
     * @param {HTMLElement} elementParent
     * @param {Generator<InfoEntry>} relationGenerator
     */
    function createRelationButtons(elementParent, relationGenerator) {
        for (let child of relationGenerator) {
            let meta = findMetadataById(child.ItemId)
            let el
            if (meta?.Thumbnail) {
                el = document.createElement("img")
                el.title = `${child.En_Title} (${typeToSymbol(child.Type)} on ${formatToName(child.Format)})`
                el.src = meta.Thumbnail
            } else {
                el = document.createElement("button")
                el.innerText = child.En_Title
            }
            elementParent.append(el)
            el.onclick = () => toggleItem(child)
        }
    }

    let childEl = /**@type {HTMLElement}*/(root.querySelector(".descendants div"))
    createRelationButtons(childEl, findDescendants(item.ItemId))

    let copyEl = /**@type {HTMLElement}*/(root.querySelector(".copies div"))
    createRelationButtons(copyEl, findCopies(item.ItemId))

    hookActionButtons(root, item)

    for (let el of root.querySelectorAll("[contenteditable]")) {
        /**@type {HTMLElement}*/(el).addEventListener("keydown", handleRichText)
    }
}

/**
 * @param {InfoEntry} item
 */
function removeDisplayItem(item) {
    const el = /**@type {HTMLElement}*/(displayItems.querySelector(`[data-item-id="${item.ItemId}"]`))
    if(!el) return
    el.remove()
    observer.unobserve(el)
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
        renderDisplayItem(item)
    }
}

/**
* @param {HTMLElement} element
*/
function getIdFromDisplayElement(element) {
    let rootNode = /**@type {ShadowRoot}*/(element.getRootNode())
    let host = rootNode.host
    if (!host) {
        return 0n
    }
    return BigInt(String(host.getAttribute("data-item-id")))
}

/**
 * @param {(item: InfoEntry, root: ShadowRoot) => any} func
 */
function displayEntryAction(func) {
    /**@param {HTMLElement} elem*/
    return function(elem) {
        let id = getIdFromDisplayElement(elem)
        let item;
        (item = findInfoEntryById(id)) && func(item, /**@type {ShadowRoot}*/(elem.getRootNode()))
    }
}

const displayEntryDelete = displayEntryAction(item => deleteEntry(item))
const displayEntryRefresh = displayEntryAction((item, root) => overwriteEntryMetadata(root, item))
const displayEntrySave = displayEntryAction((item, root) => saveItemChanges(root, item))
const displayEntryClose = displayEntryAction(item => deselectItem(item))

const displayEntryCopyTo = displayEntryAction(item => {
    let id = promptNumber("Copy user info to (item id)", "Not a number, mmust be item id number", BigInt)
    if (id === null) return
    let idInt = BigInt(id)

    copyUserInfo(item.ItemId, idInt)
        .then(res => res?.text())
        .then(console.log)
})

const displayEntryViewCount = displayEntryAction(item => {
    let count = promptNumber("New view count", 'Not a number, view count')
    if (count === null) return

    fetch(`${apiPath}/engagement/mod-entry?id=${item.ItemId}&view-count=${count}`)
        .then(res => res.text())
        .then(alert)
        .catch(console.error)
})

const displayEntryProgress = displayEntryAction(async (item, root) => {
    let progress = /**@type {HTMLProgressElement}*/(root.querySelector(".entry-progress progress"))

    let newEp = promptNumber("Current position:", "Not a number, current position")
    if (!newEp) return

    await setPos(item.ItemId, String(newEp))
    root.host.setAttribute("data-user-current-position", String(newEp))
    progress.value = Number(newEp)
})

const displayEntryRating = displayEntryAction(item => {
    let newRating = prompt("New rating")
    if (!newRating || isNaN(Number(newRating))) {
        return
    }

    fetch(`${apiPath}/engagement/mod-entry?id=${item.ItemId}&rating=${newRating}`)
        .then(refreshInfo)
        .then(() => {
            let newItem = globalsNewUi.entries[String(item.ItemId)]
            refreshDisplayItem(newItem)
        })
        .catch(console.error)
})
