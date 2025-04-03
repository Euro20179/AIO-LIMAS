const calcItems = /**@type {HTMLDivElement}*/(document.getElementById("calc-items"))
const expressionInput = /**@type {HTMLTextAreaElement}*/(document.getElementById("calc-expression"))

/**
 * @type {DisplayMode}
 */
const modeCalc = {
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
    }
}

expressionInput.onchange = function() {
    for (let entry of globalsNewUi.selectedEntries) {
        let val = updateExpressionOutput(entry)
        let el = calcItems.querySelector(`[data-item-id="${entry.ItemId}"]`)
        el?.setAttribute("data-expression-output", val.jsStr())
    }
}

/**
 * @param {InfoEntry} item
 */
function updateExpressionOutput(item) {
    let expr = expressionInput.value

    let meta = findMetadataById(item.ItemId)
    let user = findUserEntryById(item.ItemId)

    let all = {...item, ...meta, ...user}
    let symbols = makeSymbolsTableFromObj(all)

    let val = new Str(`${meta?.Description || ""}<br>${meta?.Rating || 0}`)
    if (expr) {
        val = parseExpression(expr, symbols)
    }
    return val
}

/**
 * @param {InfoEntry} item
 */
function removecCalcItem(item) {
    let el = calcItems.querySelector(`[data-item-id="${item.ItemId}"]`)
    el?.remove()
}


/**
    * @param {InfoEntry} item
    * @param {HTMLElement | DocumentFragment} [parent=calcItems]
    */
function renderCalcItem(item, parent = calcItems) {
    let el = document.createElement("calc-entry")

    let root = el.shadowRoot
    if(!root) return


    let meta = findMetadataById(item.ItemId)

    let val = updateExpressionOutput(item)

    let name = /**@type {HTMLElement}*/(root.querySelector('.name'))
    name.innerText = item.En_Title

    let img = /**@type {HTMLImageElement}*/(root.querySelector(".thumbnail"))
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
    for(let elem of elements) {
        calcItems.append(elem)
    }
}
