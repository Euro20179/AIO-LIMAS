/*
 * This file is for functions that should be called by the ui to handle api operations
 * with client-side inputs
*/

function deleteEntryUI(item: InfoEntry) {
    if (!confirm("Are you sure you want to delete this item")) {
        return
    }

    deleteEntry(item.ItemId).then(res => {
        if (res?.status != 200) {
            console.error(res)
            alert("Failed to delete item")
            return
        }
        alert(`Deleted: ${item.En_Title} (${item.Native_Title} : ${item.ItemId})`)
        updateInfo({ entries: { [String(item.ItemId)]: item } }, true)
        deselectItem(item)
        removeSidebarItem(item)
    })
}

function overwriteEntryMetadataUI(_root: ShadowRoot, item: InfoEntry) {
    if (!confirm("Are you sure you want to overwrite the metadata with a refresh")) {
        return
    }

    overwriteMetadataEntry(item.ItemId).then(res => {
        if (res.status !== 200) {
            console.error(res)
            alert("Failed to get metadata")
            return
        }
        loadMetadata()
            .then(() => {
                updateInfo({
                    entries: {
                        [String(item.ItemId)]: item
                    }
                })
            })
    })
}

async function newEntryUI() {
    const form = document.getElementById("new-item-form") as HTMLFormElement
    document.getElementById("new-entry")?.hidePopover()
    const data = new FormData(form)

    let artStyle = 0

    const styles = ['is-anime', 'is-cartoon', 'is-handrawn', 'is-digital', 'is-cgi', 'is-live-action']
    for (let i = 0; i < styles.length; i++) {
        let style = styles[i]
        if (data.get(style)) {
            artStyle |= 2 ** i
            data.delete(style)
        }
    }

    let validEntries: Record<string, FormDataEntryValue> = {}
    for (let [name, value] of data.entries()) {
        if (value == "") continue
        validEntries[name] = value
    }

    if(validEntries["libraryId"] && validEntries["libraryId"] == "0") {
        delete validEntries["libraryId"]
    }

    const queryString = "?" + Object.entries(validEntries).map(v => `${v[0]}=${encodeURIComponent(String(v[1]))}`).join("&") + `&art-style=${artStyle}`

    const tz = Intl.DateTimeFormat().resolvedOptions().timeZone

    let res = await fetch(`${apiPath}/add-entry${queryString}&timezone=${encodeURIComponent(tz)}`)
    let text = await res.text()
    if (res.status !== 200) {
        alert(text)
        return
    }

    let json = parseJsonL(mkStrItemId(text))
    loadMetadata().then(() => {
        updateInfo({
            entries: { [json.ItemId]: json },
        })
    })

    clearItems()

    selectItem(json, mode, true)
    renderSidebarItem(json)
}


