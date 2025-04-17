
type DisplayMode = {
    add: (entry: InfoEntry, updateStats?: boolean) => any
    sub: (entry: InfoEntry, updateStats?: boolean) => any
    addList: (entry: InfoEntry[], updateStats?: boolean) => any
    subList: (entry: InfoEntry[], updateStats?: boolean) => any
    refresh?: (id: bigint) => any
    putSelectedInCollection?: () => any
    addTagsToSelected?: () => any
    [key: string]: any
}

const modes = [modeDisplayEntry, modeGraphView, modeCalc, modeGallery]
const modeOutputIds = ["entry-output", "graph-output", "calc-output", "gallery-output"]

let idx = modeOutputIds.indexOf(location.hash.slice(1))

let mode = modes[idx]

function selectItem(item: InfoEntry, mode: DisplayMode, updateStats: boolean = true) {
    globalsNewUi.selectedEntries.push(item)
    mode.add(item, updateStats)
}

function deselectItem(item: InfoEntry, updateStats: boolean = true) {
    globalsNewUi.selectedEntries = globalsNewUi.selectedEntries.filter(a => a.ItemId !== item.ItemId)
    mode.sub(item, updateStats)
}

function selectItemList(itemList: InfoEntry[], mode: DisplayMode, updateStats: boolean = true) {
    globalsNewUi.selectedEntries = globalsNewUi.selectedEntries.concat(itemList)
    mode.addList(itemList, updateStats)
}

function toggleItem(item: InfoEntry, updateStats: boolean = true) {
    if (globalsNewUi.selectedEntries.find(a => a.ItemId === item.ItemId)) {
        deselectItem(item, updateStats)
    } else {
        selectItem(item, mode, updateStats)
    }
}

function clearItems() {
    mode.subList(globalsNewUi.selectedEntries)
    globalsNewUi.selectedEntries = []
}

function putSelectedToCollection() {
    if (mode.putSelectedInCollection) {
        mode.putSelectedInCollection()
    } else {
        alert("This mode does not support putting items into a collection")
    }
}

function addTagsToSelected() {
    if (mode.addTagsToSelected) {
        mode.addTagsToSelected()
    } else {
        alert("This mode does not support adding tags to selected items")
    }
}

document.getElementById("view-toggle")?.addEventListener("change", e => {
    mode.subList(globalsNewUi.selectedEntries)

    let name = (e.target as HTMLSelectElement).value

    let curModeIdx = modeOutputIds.indexOf(name)

    mode = modes[curModeIdx]
    location.hash = name

    mode.addList(globalsNewUi.selectedEntries)
})

const viewAllElem = document.getElementById("view-all") as HTMLInputElement
viewAllElem.addEventListener("change", e => {
    clearItems()
    if ((e.target as HTMLInputElement)?.checked) {
        selectItemList(globalsNewUi.results, mode)
    } else {
        resultStatsProxy.reset()
    }
})
