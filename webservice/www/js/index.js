// /**@type { {collections: string[]} }*/
// const globals = {}

function basicSpan(text) {
    const el = document.createElement("span")
    el.append(text)
    return el
}

/**
 * @param {string} name
 * @returns {HTMLDivElement}
 */
function createItemEntry(item) {
    const out = document.getElementById("plain-entries")
    /**@type {HTMLTemplateElement}*/
    const itemTemplate = document.getElementById("item-entry")
    /**@type {HTMLElement}*/
    const clone = itemTemplate.content.cloneNode(true)

    clone.querySelector("#name").innerHTML = item.En_Title

    /**@type {HTMLDetailsElement}*/
    const basicInfoEl = clone.querySelector("#basic-info")
    basicInfoEl.append(basicSpan(`Item id: ${item.ItemId}`))
    basicInfoEl.append(basicSpan(`Title: ${item.En_Title}`))
    basicInfoEl.append(basicSpan(`Native title: ${item.Native_Title}`))
    let typeText = `Type: ${item.Type}`
    if(item.IsAnime) {
        typeText += ` (anime)`
    }
    basicInfoEl.append(basicSpan(typeText))
    basicInfoEl.append(basicSpan(`Format: ${item.Format}`))

    out.appendChild(clone)
}

async function loadCollections() {
    const res = await fetch(`${apiPath}/list-collections`).catch(console.error)
    /**@type {string}*/
    const text = await res.text().catch(console.error)
    const lines = text.split("\n").filter(Boolean)
    return lines
}


function addCollections(collections) {
    const collectionsSection = document.getElementById("collections")
    for (let collection of collections) {
        const elem = document.createElement("entry-collection")
        elem.querySelector("#name").innerText = collection
        collectionsSection.appendChild(elem)
    }
}

async function loadPlainEntries() {
    const res = await fetch(`${apiPath}/query?collections=`)
        .catch(console.error)
    let itemsText = await res.text()
    return itemsText.split("\n").filter(Boolean)
        .map(JSON.parse)
}

async function addPlainEntries(items) {
    for(const item of items) {
        createItemEntry(item)
    }
}

function main() {
    loadCollections().then(addCollections)
    loadPlainEntries().then(addPlainEntries)
}
main()
