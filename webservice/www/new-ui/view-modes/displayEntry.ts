let displayQueue: InfoEntry[] = []

function mkGenericTbl(root: HTMLElement, data: Record<any, any>) {
    let html = `
                <thead>
                    <tr>
                        <th>Name</th>
                        <th>Value</th>
                    </tr>
                </thead>
                <tbody>
            `

    for (let key in data) {
        html += `<tr><td>${key}</td><td contenteditable>${data[key]}</td></tr>`
    }
    html += "</tbody>"
    root.innerHTML = html
}

const displayEntryIntersected: Set<string> = new Set()

function onIntersection(entries: IntersectionObserverEntry[]) {
    for (let entry of entries) {
        const entryId = entry.target.getAttribute("data-item-id") || "NA"
        if (entry.isIntersecting && displayQueue.length && !displayEntryIntersected.has(entryId)) {
            displayEntryIntersected.add(entryId)

            let newItem = displayQueue.shift()
            if (!newItem) continue
            modeDisplayEntry.add(newItem, false)
        }
    }
}

function deleteEvent(el: HTMLElement, ts: number, after: number) {
    if (!confirm("Are you sure you would like to delete this event")) {
        return
    }
    const itemId = getIdFromDisplayElement(el)
    apiDeleteEvent(itemId, ts, after)
        .then(res => res.text())
        .then(() =>
            refreshInfo().then(() =>
                refreshDisplayItem(globalsNewUi.entries[String(itemId)])
            ).catch(alert)
        )
        .catch(alert)

}

function newEvent(form: HTMLFormElement) {
    const data = new FormData(form)
    const name = data.get("name")
    if (name == null) {
        alert("Name required")
        return
    }
    const tsStr = data.get("timestamp")
    const aftertsStr = data.get("after")
    //@ts-ignore
    let ts = new Date(tsStr).getTime()
    if (isNaN(ts)) {
        ts = 0
    }
    //@ts-ignore
    let afterts = new Date(aftertsStr).getTime()
    if (isNaN(afterts)) {
        afterts = 0
    }
    const itemId = getIdFromDisplayElement(form)
    apiRegisterEvent(itemId, name.toString(), ts, afterts)
        .then(res => res.text())
        .then(() => refreshInfo().then(() => {
            form.parentElement?.hidePopover()
            refreshDisplayItem(globalsNewUi.entries[String(itemId)])
        }
        ))
        .catch(alert)
    //TODO: should be a modal thing for date picking
}

const observer = new IntersectionObserver(onIntersection, {
    root: document.querySelector("#entry-output"),
    rootMargin: "0px",
    threshold: 0.1
})

const modeDisplayEntry: DisplayMode = {
    add(entry, updateStats = true) {
        updateStats && changeResultStatsWithItem(entry)
        renderDisplayItem(entry)
    },

    sub(entry, updateStats = true) {
        updateStats && changeResultStatsWithItem(entry, -1)
        removeDisplayItem(entry)
    },

    addList(entry, updateStats = true) {
        displayEntryIntersected.clear()

        updateStats && changeResultStatsWithItemList(entry, 1)
        for (let i = 0; i < entry.length; i++) {
            if (i > 5) {
                displayQueue.push(entry[i])
            } else {
                renderDisplayItem(entry[i])
            }
        }
    },

    subList(entry, updateStats = true) {
        updateStats && changeResultStatsWithItemList(entry, -1)

        const itemIdsToRemove = entry.map(v => v.ItemId)
        displayQueue = displayQueue.filter(i => !itemIdsToRemove.includes(i.ItemId))

        for (let item of entry) {
            removeDisplayItem(item)
        }
    },

    putSelectedInCollection() {
        const selected = globalsNewUi.selectedEntries
        const collectionName = prompt("Id of collection")
        if (!collectionName) return

        let waiting = []
        for (let item of selected) {
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
        for (let item of globalsNewUi.selectedEntries) {
            addEntryTags(item.ItemId, tagsList)
        }
        //FIXME: tags do not update immediately
    }
}

function hookActionButtons(shadowRoot: ShadowRoot, item: InfoEntry) {
    for (let btn of shadowRoot.querySelectorAll("[data-action]") || []) {
        let action = btn.getAttribute("data-action")
        btn.addEventListener("click", _ => {
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

            const tz = Intl.DateTimeFormat().resolvedOptions().timeZone
            fetch(`${apiPath}/engagement/${action?.toLowerCase()}-media${queryParams}&timezone=${encodeURIComponent(tz)}`)
                .then(res => res.text())
                .then(text => {
                    alert(text)
                    return refreshInfo()
                })
                .then(() => refreshDisplayItem(item))
        })
    }


    let imgEl = shadowRoot.querySelector(".thumbnail") as HTMLImageElement
    const fileUpload = (shadowRoot.getElementById("thumbnail-file-upload")) as HTMLInputElement

    fileUpload.onchange = async function(_) {
        const reader = new FileReader()
        const blob = fileUpload.files?.[0]
        if (!blob) return
        reader.readAsDataURL(blob)
        reader.onload = () => {
            if (!reader.result) return
            updateThumbnail(item.ItemId, reader.result.toString())
                .then(() => {
                    refreshInfo().then(() => {
                        refreshDisplayItem(item)
                    })
                })
        }
    }
    imgEl.onclick = function(_) {
        if (!fileUpload) return

        fileUpload.click()
        console.log(fileUpload.value)
    }
}

function updateDisplayEntryContents(item: InfoEntry, user: UserEntry, meta: MetadataEntry, events: UserEvent[], el: ShadowRoot) {
    const displayEntryTitle = el.querySelector(".title") as HTMLHeadingElement
    const displayEntryNativeTitle = el.querySelector(".official-native-title") as HTMLHeadingElement
    const imgEl = el.querySelector(".thumbnail") as HTMLImageElement
    const costEl = el.querySelector(".cost") as HTMLSpanElement
    const descEl = el.querySelector(".description") as HTMLParagraphElement
    const notesEl = el.querySelector(".notes") as HTMLParagraphElement
    const ratingEl = el.querySelector(".rating") as HTMLSpanElement
    const audienceRatingEl = el.querySelector(".audience-rating") as HTMLElement
    const infoRawTbl = el.querySelector(".info-raw") as HTMLTableElement
    const metaRawtbl = el.querySelector(".meta-info-raw") as HTMLTableElement
    const viewCountEl = el.querySelector(".entry-progress .view-count") as HTMLSpanElement
    const progressEl = el.querySelector(".entry-progress progress") as HTMLProgressElement
    const captionEl = el.querySelector(".entry-progress figcaption") as HTMLElement
    const mediaInfoTbl = el.querySelector("figure .media-info") as HTMLTableElement
    const eventsTbl = el.querySelector(".user-actions") as HTMLTableElement

    //tags
    const tagsRoot = el.querySelector(".tags") as HTMLDivElement
    tagsRoot.innerHTML = ""
    for (let tag of item.Tags || []) {
        tag = tag.trim()
        if (!tag) continue
        const outer = document.createElement("div")

        const del = document.createElement("button")
        del.innerText = "🗑"
        del.classList.add("delete")

        del.onclick = function() {
            deleteEntryTags(item.ItemId, [tag])
                .then(res => {
                    if (res.status !== 200) return ""
                    res.text().then(() => {
                        item.Tags = item.Tags.filter((t: string) => t != tag)
                        changeDisplayItemData(item, user, meta, events, el.host as HTMLElement)
                    })
                })
                .catch(console.error)
        }

        outer.append(del)

        const span = document.createElement("span")
        span.classList.add("tag")
        span.innerText = tag

        outer.append(span)

        tagsRoot?.append(outer)
    }

    //type icon
    let typeIcon = typeToSymbol(item.Type)
    displayEntryTitle?.setAttribute("data-type-icon", typeIcon)


    //format
    formatToName(item.Format).then(name => {
        displayEntryTitle?.setAttribute("data-format-name", name)
    })


    //Title
    displayEntryTitle.innerText = meta.Title || item.En_Title
    //only set the title of the heading to the user's title if the metadata title exists
    //otherwise it looks dumb
    if (meta.Title && item.En_Title) {
        displayEntryTitle.title = item.En_Title
    }

    console.log(meta)

    //Native title
    displayEntryNativeTitle.innerText = meta.Native_Title || item.Native_Title
    //Same as with the regular title
    if (meta.Native_Title && item.Native_Title) {
        displayEntryNativeTitle.title = item.Native_Title
    }

    //Thumbnail
    imgEl.alt = meta.Title || item.En_Title
    imgEl.src = meta.Thumbnail

    //Cost
    costEl.innerText = `$${item.PurchasePrice}`

    //Description
    descEl.innerHTML = meta.Description

    //Notes
    notesEl.innerHTML = user.Notes

    //Rating
    if (user.UserRating) {
        applyUserRating(user.UserRating, ratingEl)
        ratingEl.innerHTML = String(user.UserRating)
    } else {
        ratingEl.innerText = "Unrated"
    }

    //Audience Rating
    let max = meta.RatingMax
    if (meta.Rating) {
        let rating = meta.Rating
        let normalizedRating = rating
        if (max !== 0) {
            normalizedRating = rating / max * 100
        }
        applyUserRating(normalizedRating, audienceRatingEl)
        audienceRatingEl.innerHTML = String(rating)
    } else if (audienceRatingEl) {
        audienceRatingEl.innerText = "Unrated"
    }

    //Info table raw
    mkGenericTbl(infoRawTbl, item)

    //Meta table raw
    let data = meta
    mkGenericTbl(metaRawtbl, data)

    //View count
    let viewCount = user.ViewCount
    if (viewCount) {
        let mediaDependant
        try {
            mediaDependant = JSON.parse(data["MediaDependant"])
        } catch (err) {
            console.error("Could not parse media dependant meta info json")
            return
        }
        viewCountEl.setAttribute("data-time-spent", String(Number(viewCount) * Number(mediaDependant["Show-length"] || mediaDependant["Movie-length"] || 0) / 60 || "unknown"))
        viewCountEl.innerText = String(viewCount)
    }


    //Media dependant
    let type = item.Type
    type = String(type)
    let mediaDeptData
    try {
        mediaDeptData = JSON.parse(meta.MediaDependant)
    }
    catch (err) {
        console.error("Could not parse json", meta.MediaDependant)
        return
    }
    //remove the <Media>- part from the key looks ugly
    let modifiedKeys: { [k: string]: string } = {}
    for (let key in mediaDeptData) {
        const val = mediaDeptData[key]
        key = key.split("-").slice(1).join(" ")
        modifiedKeys[key] = val
    }
    mkGenericTbl(mediaInfoTbl, modifiedKeys)


    el.host.setAttribute("data-user-status", user.Status)
    if (mediaDeptData[`${type}-episodes`] && user.Status === "Viewing") {
        progressEl.max = mediaDeptData[`${type}-episodes`]

        let pos = Number(user.CurrentPosition)
        progressEl.value = pos

        captionEl.innerText = `${pos}/${progressEl.max}`
        captionEl.title = `${Math.round(pos / progressEl.max * 1000) / 10}%`
    }

    //Current position
    progressEl.title = user.CurrentPosition
    if (progressEl.max) {
        progressEl.title = `${user.CurrentPosition}/${progressEl.max}`
    }

    //Events
    if (events.length) {
        let html = `
            <thead>
                <tr>
                    <!-- this nonsense is so that the title lines up with the events -->
                    <th>
                        <div class="grid column">
                            <button popovertarget="new-event-form">➕︎</button><span style="text-align: center">Event</span>
                        </div>
                    </th>
                    <th>Time</th>
                </tr>
            </thead>
            <tbody>
        `
        for (let event of events) {
            const ts = event.Timestamp
            const afterts = event.After
            const timeZone = event.TimeZone || "UTC"
            const name = event.Event

            let date = new Date(event.Timestamp)
            let afterDate = new Date(event.After)
            let timeTd = ""
            if (ts !== 0) {
                let time = date.toLocaleTimeString("en", { timeZone })
                let dd = date.toLocaleDateString("en", { timeZone })
                timeTd = `<td title="${time} (${timeZone})">${dd}</td>`
            } else if (afterts !== 0) {
                let time = afterDate.toLocaleTimeString("en", { timeZone })
                let dd = afterDate.toLocaleDateString("en", { timeZone })
                timeTd = `<td title="${time} (${timeZone})">after: ${dd}</td>`
            } else {
                timeTd = `<td title="unknown">unknown</td>`
            }
            html += `<tr>
                        <td>
                            <div class="grid column">
                                <button class="delete" onclick="deleteEvent(this, ${ts}, ${afterts})">🗑</button>
                                ${name}
                            </div>
                        </td>
                            ${timeTd}
                        </tr>`
        }
        html += "</tbody>"
        eventsTbl.innerHTML = html
    } else {
        //there are no events
        eventsTbl.innerHTML = ""
    }
}

function renderDisplayItem(item: InfoEntry, parent: HTMLElement | DocumentFragment = displayItems) {
    let el = document.createElement("display-entry")

    observer.observe(el)

    let meta = findMetadataById(item.ItemId)
    let user = findUserEntryById(item.ItemId)
    let events = findUserEventsById(item.ItemId)
    if (!user || !meta || !events) return


    parent.append(el)

    let root = el.shadowRoot
    if (!root) return

    function createRelationButtons(elementParent: HTMLElement, relationGenerator: Generator<InfoEntry>) {
        let relationships = relationGenerator.toArray()
        let titles = relationships.map(i => i.En_Title)
        relationships = relationships.sort((a, b) => {
            return (sequenceNumberGrabber(a.En_Title, titles) || 0) - (sequenceNumberGrabber(b.En_Title, titles) || 0)
        })
        for (let child of relationships) {
            let meta = findMetadataById(child.ItemId)
            let el: HTMLElement
            if (meta?.Thumbnail) {
                el = document.createElement("img")
                formatToName(child.Format).then(name => {
                    el.title = `${child.En_Title} (${typeToSymbol(child.Type)} on ${name})`
                })
                //@ts-ignore
                el.src = meta.Thumbnail
            } else {
                el = document.createElement("button")
                el.innerText = child.En_Title
            }
            elementParent.append(el)
            el.onclick = () => toggleItem(child)
        }
    }

    let childEl = root.querySelector(".descendants div") as HTMLElement
    createRelationButtons(childEl, findDescendants(item.ItemId))

    let copyEl = root.querySelector(".copies div") as HTMLElement
    createRelationButtons(copyEl, findCopies(item.ItemId))

    hookActionButtons(root, item)

    const newTag = (root.getElementById("create-tag")) as HTMLButtonElement
    newTag.onclick = function() {
        const name = prompt("Tag name (, seperated)")
        if (!name) return
        let names = name.split(",")
        item.Tags = item.Tags.concat(names)
        addEntryTags(item.ItemId, name.split(","))
            .then(res => {
                if (res.status !== 200) return ""
                res.text().then(() => changeDisplayItemData(item, user, meta, events, el))
            })
            .catch(console.error)
    }

    el.addEventListener("data-changed", function(_e) {
        let e = _e as CustomEvent
        const event = e
        const item = /**@type {InfoEntry}*/(event.detail.item)
        const user = /**@type {UserEntry}*/(event.detail.user)
        const meta = /**@type {MetadataEntry}*/(event.detail.meta)
        const events = /**@type {UserEvent[]}*/(event.detail.events)
        updateDisplayEntryContents(item, user, meta, events, el.shadowRoot as ShadowRoot)
    })

    changeDisplayItemData(item, user, meta, events, el)

    for (let el of root.querySelectorAll("[contenteditable]")) {
        (el as HTMLElement).addEventListener("keydown", handleRichText)
    }
}

function removeDisplayItem(item: InfoEntry) {
    displayEntryIntersected.delete(String(item.ItemId))
    const el = /**@type {HTMLElement}*/(displayItems.querySelector(`[data-item-id="${item.ItemId}"]`))
    if (!el) return
    el.remove()
    observer.unobserve(el)
}

function refreshDisplayItem(item: InfoEntry) {
    let el = document.querySelector(`display-entry[data-item-id="${item.ItemId}"]`) as HTMLElement
    if (el) {
        let user = findUserEntryById(item.ItemId)
        let events = findUserEventsById(item.ItemId)
        let meta = findMetadataById(item.ItemId)
        if (!user || !events || !meta) return
        changeDisplayItemData(item, user, meta, events, el)
    } else {
        renderDisplayItem(item)
    }
}

function getIdFromDisplayElement(element: HTMLElement) {
    let rootNode = element.getRootNode() as ShadowRoot
    let host = rootNode.host
    if (!host) {
        return 0n
    }
    return BigInt(String(host.getAttribute("data-item-id")))
}

function displayEntryAction(func: (item: InfoEntry, root: ShadowRoot) => any) {
    return function(elem: HTMLElement) {
        let id = getIdFromDisplayElement(elem)
        let item;
        (item = findInfoEntryById(id)) && func(item, elem.getRootNode() as ShadowRoot)
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
        .then(() => {
            refreshInfo().then(() => {
                refreshDisplayItem(item)
            })
        })
        .catch(console.error)
})

const displayEntryProgress = displayEntryAction(async (item, root) => {
    let progress = root.querySelector(".entry-progress progress") as HTMLProgressElement

    let newEp = promptNumber("Current position:", "Not a number, current position")
    if (!newEp) return

    await setPos(item.ItemId, String(newEp))
    root.host.setAttribute("data-user-current-position", String(newEp))
    progress.value = Number(newEp)
})

const displayEntryRating = displayEntryAction(item => {
    let user = findUserEntryById(item.ItemId)
    if (!user) {
        alert("Failed to get current rating")
        return
    }
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
    apiRegisterEvent(item.ItemId, `rating-change - ${user?.UserRating} -> ${newRating}`, Date.now(), 0).catch(console.error)
})
