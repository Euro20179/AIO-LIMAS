/**
 * @typedef UserEntry
 * @type {object}
 * @property {bigint} ItemId
 * @property {string} Status
 * @property {number} ViewCount
 * @property {string} StartDate
 * @property {string} EndDate
 * @property {number} UserRating
 * @property {string} Notes
 */

/**
 * @typedef InfoEntry
 * @type {object}
 * @property {bigint} ItemId
 * @property {string} Collection
 * @property {number} Format
 * @property {boolean} IsAnime
 * @property {string} Location
 * @property {string} Native_Title
 * @property {bigint} Parent
 * @property {number} PurchasePrice
 * @property {string} Type
 * @property {string} En_Title
 */

/**
 * @typedef MetadataEntry
 * @type {object}
 * @property {bigint} ItemId
 * @property {number} Rating
 * @property {string} Description
 * @property {number} ReleaseYear
 * @property {string} Thumbnail
 * @property {string} MediaDependant
 * @property {string} Datapoints
 */

/**@type { {formats: Record<number, string>, userEntries: UserEntry[], metadataEntries: MetadataEntry[]} }*/
const globals = { formats: {}, userEntries: [], metadataEntries: [] }

/**
 * @param {bigint} id
 * @returns {UserEntry?}
 */
function getUserEntry(id) {
    for (let entry of globals.userEntries) {
        if (entry.ItemId === id) {
            return entry
        }
    }
    return null
}

/**
 * @param {bigint} id
 * @returns {MetadataEntry?}
 */
function getMetadataEntry(id) {
    for (let entry of globals.metadataEntries) {
        if (entry.ItemId === id) {
            return entry
        }
    }
    return null
}

/**
 * @param {bigint} id
 * @param {MetadataEntry?} newMeta
 */
function setMetadataEntry(id, newMeta) {
    for(let i = 0; i < globals.metadataEntries.length; i++) {
        let entry = globals.metadataEntries[i]
        if(entry.ItemId === id) {
            //@ts-ignore
            globals.metadataEntries[i] = newMeta
            return
        }
    }
}

/**@param {string} jsonl*/
function mkStrItemId(jsonl) {
    return jsonl
        .replace(/"ItemId":\s*(\d+),/, "\"ItemId\": \"$1\",")
        .replace(/"Parent":\s*(\d+),/, "\"Parent\": \"$1\",")
}

/**@param {string} jsonl*/
function parseJsonL(jsonl) {
    const bigIntProperties = ["ItemId", "Parent"]
    return JSON.parse(jsonl, (key, v) => bigIntProperties.includes(key) ? BigInt(v) : v)
}

async function loadFormats() {
    const res = await fetch(`${apiPath}/type/format`)
    const json = await res.json()
    let fmtJson = Object.fromEntries(
        Object.entries(json).map(([key, val]) => [Number(key), val])
    )
    globals.formats = fmtJson
}

/**
 * @param {string} text
 * @param {string} ty
 */
function basicElement(text, ty = "span") {
    const el = document.createElement(ty)
    el.append(text)
    return el
}

/**
 * @param {HTMLElement} container
 * @param {InfoEntry} item
 */
function fillBasicInfoSummary(container, item) {
    /**@type {HTMLDetailsElement}*/
    const basicInfoEl = /**@type {HTMLDetailsElement}*/(container.querySelector(".basic-info ul"))
    basicInfoEl.append(basicElement(`Item id: ${item.ItemId}`, "li"))
    basicInfoEl.append(basicElement(`Title: ${item.En_Title}`, "li"))
    basicInfoEl.append(basicElement(`Native title: ${item.Native_Title}`, "li"))
    let typeText = `Type: ${item.Type}`
    if (item.IsAnime) {
        typeText += ` +anime`
    }
    basicInfoEl.append(basicElement(typeText, "li"))

    const MOD_DIGITAL = Number(Object.entries(globals.formats).filter(([_, val]) => val == "MOD_DIGITAL")[0][0])

    let digitized = false
    let fmtNo = item.Format
    if ((item.Format & MOD_DIGITAL) === MOD_DIGITAL) {
        fmtNo -= MOD_DIGITAL
        digitized = true
    }
    let fmtText = `Format: ${globals.formats[fmtNo]}`
    if (digitized) {
        fmtText += ` +digitized`
    }
    basicInfoEl.append(basicElement(fmtText, "li"))
}

/**
 * @param {HTMLElement} container
 * @param {UserEntry} item
 */
function fillUserInfo(container, item) {

    /**@type {HTMLDetailsElement}*/
    const userInfoEl = /**@type {HTMLDetailsElement}*/(container.querySelector(".user-info ul"))

    userInfoEl.append(basicElement(`View count: ${item.ViewCount}`, "li"));
    userInfoEl.append(basicElement(`Status: ${item.Status}`, "li"));

    const viewTableBody = /**@type {HTMLTableElement}*/(container.querySelector(".view-table tbody"));

    const startDates = JSON.parse(item.StartDate)
    const endDates = JSON.parse(item.EndDate)
    for (let i = 0; i < startDates.length; i++) {
        let start = startDates[i]
        let end = endDates[i]

        let sd = new Date(start)
        let ed = new Date(end)
        let sText = `${sd.getMonth() + 1}/${sd.getDate()}/${sd.getFullYear()}`
        let eText = `${ed.getMonth() + 1}/${ed.getDate()}/${ed.getFullYear()}`

        viewTableBody.innerHTML += `
<tr>
    <td>${start === 0 ? "unknown" : sText}</td>
    <td>${end === 0 ? "unknown" : eText}</td>
</tr>
`
    }
}

/**
 * @param {InfoEntry} item
 */
function createItemEntry(item) {
    const out = /**@type {HTMLElement}*/(document.getElementById("all-entries"))
    const itemTemplate = /**@type {HTMLTemplateElement}*/ (document.getElementById("item-entry"))
    /**@type {HTMLElement}*/
    const clone = /**@type {HTMLElement}*/(itemTemplate.content.cloneNode(true));

    const root = /**@type {HTMLElement}*/(clone.querySelector(".entry"));
    root.setAttribute("data-type", item.Type);


    /**@type {HTMLElement}*/(clone.querySelector(".name")).innerHTML = item.En_Title

    const userEntry = getUserEntry(item.ItemId);

    /**@type {HTMLElement}*/(clone.querySelector(".rating")).innerHTML = String(userEntry?.UserRating) || "#N/A";

    if(item.PurchasePrice) {
        /**@type {HTMLElement}*/(clone.querySelector(".cost")).innerHTML = String(item.PurchasePrice)
    }

    /**@type {HTMLElement}*/(clone.querySelector(".notes")).innerHTML = String(userEntry?.Notes || "");

    if (item.Location) {
        const locationA = /**@type {HTMLAnchorElement}*/(clone.querySelector(".location"));
        locationA.innerText = item.Location
        locationA.href = item.Location
    }

    if (item.Collection) {
        /**@type {HTMLElement}*/(clone.querySelector(".collection")).innerHTML = `Collection: ${item.Collection}`
    }

    if (userEntry?.UserRating) {
        if (userEntry.UserRating >= 80) {
            root.classList.add("good")
        } else {
            root.classList.add("bad")
        }
    }

    const img = /**@type {HTMLImageElement}*/(clone.querySelector(".img"));

    const meta = getMetadataEntry(item.ItemId)

    if (meta?.Thumbnail) {
        img.src = meta.Thumbnail
    }

    fillBasicInfoSummary(clone, item)
    fillUserInfo(clone, /**@type {UserEntry}*/(userEntry))


    const metaRefresher = /**@type {HTMLButtonElement}*/(clone.querySelector(".meta-fetcher"));
    metaRefresher.onclick = async function(e) {
        let res = await fetch(`${apiPath}/metadata/fetch?id=${item.ItemId}`).catch(console.error)
        if(res?.status != 200) {
            console.error(res)
            return
        }

        res = await fetch(`${apiPath}/metadata/retrieve?id=${item.ItemId}`).catch(console.error)
        if(res?.status != 200) {
            console.error(res)
            return
        }

        const json = /**@type {MetadataEntry}*/(await res.json())

        setMetadataEntry(item.ItemId, json)
        img.src = json.Thumbnail
    }
    out.appendChild(clone)
}

async function loadCollections() {
    const res = await fetch(`${apiPath}/list-collections`).catch(console.error)
    if (!res) {
        alert("Could not load collections")
        return
    }
    /**@type {string}*/
    const text = /**@type {string}*/(await res.text().catch(console.error))
    const lines = text.split("\n").filter(Boolean)
    return lines
}


/**@param {string[] | undefined} collections*/
function addCollections(collections) {
    if (!collections) {
        return
    }
    const collectionsSection = /**@type {HTMLElement}*/ (document.getElementById("collections"))
    for (let collection of collections) {
        const elem = /**@type {HTMLElement}*/ (document.createElement("entry-collection"));
        /**@type {HTMLElement}*/ (elem.querySelector(".name")).innerText = collection
        collectionsSection.appendChild(elem)
    }
}

async function loadAllEntries() {
    const res = await fetch(`${apiPath}/list-entries`)
        .catch(console.error)
    if (!res) {
        alert("Could not load entries")
    } else {
        let itemsText = await res.text()
        /**@type {string[]}*/
        let jsonL = itemsText.split("\n").filter(Boolean)
        return jsonL
            .map(mkStrItemId)
            .map(parseJsonL)
    }
}

/**
* @param {InfoEntry[] | undefined} items 
*/
async function addAllEntries(items) {
    console.log(items)
    if (!items) {
        return
    }
    items = items.sort((a, b) => {
        const aUE = getUserEntry(a.ItemId)
        const bUE = getUserEntry(b.ItemId)
        return (bUE?.UserRating || 0) - (aUE?.UserRating || 0)
    })
    for (const item of items) {
        if(item.Parent) {
            //TODO: put a link to children on each entry
            //when the link is clicked, all entries will be removed in favor of that item's children
            //also render the item itself
            continue
        }
        createItemEntry(item)
    }
}

/**@returns {Promise<UserEntry[]>}*/
async function loadUserEntries() {
    const res = await fetch(`${apiPath}/engagement/list-entries`)
    if (!res) {
        return []
    }

    const text = await res.text()
    if (!text) {
        return []
    }

    const lines = text.split("\n").filter(Boolean)
    globals.userEntries = lines
        .map(mkStrItemId)
        .map(parseJsonL)

    return globals.userEntries
}

/**@returns {Promise<MetadataEntry[]>}*/
async function loadMetadata() {
    const res = await fetch(`${apiPath}/metadata/list-entries`)
    if (!res) {
        return []
    }

    const text = await res.text()
    if (!text) {
        return []
    }

    const lines = text.split("\n").filter(Boolean)
    globals.metadataEntries = lines
        .map(mkStrItemId)
        .map(parseJsonL)

    return globals.metadataEntries
}

function main() {
    loadFormats()
        // .then(loadCollections)
        // .then(addCollections)
        .then(loadUserEntries)
        .then(loadMetadata)
        .then(loadAllEntries)
        .then(addAllEntries)
        .catch(console.error)
}
main()
