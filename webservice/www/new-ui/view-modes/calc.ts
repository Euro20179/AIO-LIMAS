const calcItems = document.getElementById("calc-items") as HTMLDivElement
const expressionInput = document.getElementById("calc-expression") as HTMLTextAreaElement

const modeCalc: DisplayMode = {
    add(entry, updateStats = true) {
        updateStats && changeResultStatsWithItem(entry)
        renderCalcItem(entry)
    },

    sub(entry, updateStats = true) {
        updateStats && changeResultStatsWithItem(entry, -1)
        removecCalcItem(entry)
    },

    addList(entry, updateStats = true) {
        updateStats && changeResultStatsWithItemList(entry, 1)
        for (let item of entry) {
            renderCalcItem(item)
        }
    },

    subList(entry, updateStats = true) {
        updateStats && changeResultStatsWithItemList(entry, -1)
        for (let item of entry) {
            removecCalcItem(item)
        }
    },

    *_getValidEntries() {
        const selected = Object.groupBy(globalsNewUi.selectedEntries, item => String(item.ItemId))
        let elems = calcItems.querySelectorAll(`[data-item-id]`)

        for (let elem of elems) {
            let val = Number(elem.getAttribute("data-expression-output"))
            let id = elem.getAttribute("data-item-id")

            if (!id || val <= 0 || !selected[id]) continue

            yield selected[id][0]
        }
    },

    putSelectedInCollection() {
        const collectionName = prompt("Id of collection")
        if (!collectionName) return

        let waiting = []
        for (let item of this._getValidEntries()) {
            waiting.push(setParent(item.ItemId, BigInt(collectionName)))
        }
        Promise.all(waiting).then(res => {
            for (let r of res) {
                console.log(r.status)
            }
        })
    },

    addTagsToSelected() {
        const tags = prompt("tags (, seperated)")
        if (!tags) return

        const tagsList = tags.split(",")
        for (let item of this._getValidEntries()) {
            addEntryTags(item.ItemId, tagsList)
        }
    }
}

expressionInput.onchange = function() {
    for (let entry of globalsNewUi.selectedEntries) {
        let val = updateExpressionOutput(entry)
        let el = calcItems.querySelector(`[data-item-id="${entry.ItemId}"]`)
        el?.setAttribute("data-expression-output", val.jsStr())
    }
}

function updateExpressionOutput(item: InfoEntry) {
    let expr = expressionInput.value

    let meta = findMetadataById(item.ItemId)
    let user = findUserEntryById(item.ItemId)

    let all = { ...item, ...meta, ...user }
    let symbols = makeSymbolsTableFromObj(all)

    let val = new Str(`${meta?.Description || ""}<br>${meta?.Rating || 0}`)
    if (expr) {
        val = parseExpression(expr, symbols)
    }
    return val
}

function removecCalcItem(item: InfoEntry) {
    let el = calcItems.querySelector(`[data-item-id="${item.ItemId}"]`)
    el?.remove()
}


function renderCalcItem(item: InfoEntry, parent: HTMLElement | DocumentFragment = calcItems) {
    let el = document.createElement("calc-entry")

    let root = el.shadowRoot
    if (!root) return


    let meta = findMetadataById(item.ItemId)

    let val = updateExpressionOutput(item)

    let name = root.getElementById('name') as HTMLElement
    name.innerText = item.En_Title

    let img = root.getElementById("calc-thumbnail") as HTMLImageElement
    if (meta?.Thumbnail) {
        img.src = meta?.Thumbnail
    }
    parent.append(el)
    el.setAttribute("data-expression-output", String(val.jsStr()))
    el.setAttribute("data-item-id", String(item.ItemId))
}

function sortCalcDisplay() {
    let elements = [...calcItems.querySelectorAll(`[data-item-id]`)]
    elements.sort((a, b) => {
        let exprA = /**@type {string}*/(a.getAttribute("data-expression-output"))
        let exprB = /**@type {string}*/(b.getAttribute("data-expression-output"))
        return Number(exprB) - Number(exprA)
    })
    calcItems.innerHTML = ""
    for (let elem of elements) {
        calcItems.append(elem)
    }
}
