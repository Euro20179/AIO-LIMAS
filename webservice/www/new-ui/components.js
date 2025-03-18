/**
 * @param {HTMLElement | ShadowRoot} root
 * @param {string} selector
 * @param {string} text
 * @param {"append" | "innerhtml" | `attribute=${string}`} [fillmode="append"] 
 */
function fillElement(root, selector, text, fillmode = "append") {
    let elem = /**@type {HTMLElement}*/(root.querySelector(selector))
    if (!elem) {
        return
    }
    if (fillmode === "append") {
        elem.innerText = text
    } else if (fillmode.match(/attribute=.+/)) {
        let attribute = fillmode.split("=")[1]
        elem.setAttribute(attribute, text)
    } else {
        elem.innerHTML = text
    }
}

/**
 * @param {number} rating
 * @param {HTMLElement} root
 */
function applyUserRating(rating, root) {
    if (rating > 100) {
        root.classList.add("splus-tier")
    } else if (rating > 96) {
        root.classList.add("s-tier")
    } else if (rating > 87) {
        root.classList.add("a-tier")
    } else if (rating > 78) {
        root.classList.add("b-tier")
    } else if (rating > 70) {
        root.classList.add("c-tier")
    } else if (rating > 65) {
        root.classList.add("d-tier")
    } else if (rating > 0) {
        root.classList.add('f-tier')
    } else {
        root.classList.add("z-tier")
    }
}

customElements.define("display-entry", class extends HTMLElement {
    static observedAttributes = [
        "data-type",
        "data-format",
        "data-title",
        "data-thumbnail-src",
        "data-cost",
        "data-description",
        "data-user-notes",
        "data-native-title",
        "data-user-rating",
        "data-audience-rating",
        "data-audience-rating-max",
        "data-info-raw",
        "data-meta-info-raw",
        "data-user-status",
        "data-view-count",
        "data-media-dependant",
        "data-user-events",
        "data-user-current-position"
    ]

    constructor() {
        super()
        let template = /**@type {HTMLTemplateElement}*/(document.getElementById("display-entry"))
        let content = /**@type {HTMLElement}*/(template.content.cloneNode(true))
        let root = this.attachShadow({ mode: "open" })
        root.appendChild(content)
        this.root = root
    }
    /**
     * @param {HTMLElement} root
     * @param {Record<any, any>} data
     */
    mkGenericTbl(root, data) {
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

    /**
     * @param {string} name
     * @param {string} ov
     * @param {string} nv
     */
    attributeChangedCallback(name, ov, nv) {
        let root = this.shadowRoot
        if (!root) return

        if (name in this) {
            //@ts-ignore
            this[name](nv)
        }
    }

    /**
     * @param {string} val
     */
    ["data-type"](val) {
        let typeIcon = typeToSymbol(val)
        let title = this.root.querySelector(".title")
        title?.setAttribute("data-type-icon", typeIcon)
        this.setAttribute("data-type-icon", typeIcon)
    }

    /**
     * @param {string} val
     */
    ["data-format"](val) {
        let formatName = formatToName(Number(val))
        let title = /**@type {HTMLElement}*/(this.root.querySelector(".title"))
        title.setAttribute("data-format-name", formatName)
    }

    /**
     * @param {string} val
     */
    ["data-title"](val) {
        fillElement(this.root, ".title", val)

        let imgEl = /**@type {HTMLImageElement}*/(this.root.querySelector(".thumbnail"))
        imgEl.alt = val
    }

    /**
     * @param {string} val
     */
    ["data-thumbnail-src"](val) {
        let imgEl = /**@type {HTMLImageElement}*/(this.root.querySelector(".thumbnail"))
        imgEl.src = val
    }

    /**
     * @param {string} val
     */
    ["data-cost"](val) {
        fillElement(this.root, ".cost", `$${val}`)
    }

    /**
     * @param {string} val
     */
    ["data-description"](val) {
        fillElement(this.root, ".description", val, "innerhtml")
    }

    /**
     * @param {string} val
     */
    ["data-user-notes"](val) {
        fillElement(this.root, ".notes", val, "innerhtml")
    }

    /**
     * @param {string} val
     */
    ["data-native-title"](val) {
        let el = /**@type {HTMLElement}*/(this.root.querySelector(".title"))
        el.title = val
    }

    /**
     * @param {string} val
     */
    ["data-user-rating"](val) {
        let ratingE = /**@type {HTMLElement?}*/(this.root.querySelector(".rating"))
        if (val && ratingE) {
            let rating = Number(val)
            applyUserRating(rating, ratingE)
            ratingE.innerHTML = val
        } else if (ratingE) {
            ratingE.innerText = "Unrated"
        }
    }

    /**
     * @param {string} val
     */
    ["data-audience-rating"](val) {
        let ratingE = /**@type {HTMLElement?}*/(this.root.querySelector(".audience-rating"))
        let max = Number(this.getAttribute("data-audience-rating-max"))
        if (val && ratingE) {
            let rating = Number(val)
            let normalizedRating = rating
            if (max !== 0) {
                normalizedRating = rating / max * 100
            }
            applyUserRating(normalizedRating, ratingE)
            ratingE.innerHTML = val
        } else if (ratingE) {
            ratingE.innerText = "Unrated"
        }
    }


    /**
     * @param {string} val
     */
    ["data-info-raw"](val) {
        let infoRawTbl = /**@type {HTMLTableElement}*/(this.root.querySelector(".info-raw"))
        try {
            this.mkGenericTbl(infoRawTbl, JSON.parse(val))
        }
        catch (err) {
            console.error("Could not parse json")
        }
    }

    /**
     * @param {string} val
     */
    ["data-meta-info-raw"](val) {
        let infoRawTbl = /**@type {HTMLTableElement}*/(this.root.querySelector(".meta-info-raw"))
        let data
        try {
            data = JSON.parse(val)
        }
        catch (err) {
            console.error("Could not parse meta info json")
            return
        }
        this.mkGenericTbl(infoRawTbl, data)
        let viewCount = this.getAttribute("data-view-count")
        if (viewCount) {
            let mediaDependant
            try {
                mediaDependant = JSON.parse(data["MediaDependant"])
            } catch (err) {
                console.error("Could not parse media dependant meta info json")
                return
            }
            fillElement(this.root, ".entry-progress .view-count", String(Number(viewCount) * Number(mediaDependant["Show-length"] || mediaDependant["Movie-length"] || 0) / 60 || "unknown"), "attribute=data-time-spent")
        }
    }

    /**
     * @param {string} val
     */
    ["data-user-status"](val) {
        fillElement(this.root, ".entry-progress .status", val, "innerhtml")
    }

    /**
     * @param {string} val
     */
    ["data-user-current-position"](val) {
        let progress = /**@type {HTMLProgressElement}*/(this.root.querySelector(".entry-progress progress"))
        let caption = /**@type {HTMLElement}*/(this.root.querySelector(".entry-progress figcaption"))

        progress.title = `${val}`

        if (progress.max) {
            progress.title = `${val}/${progress.max}`
        }

        caption.innerHTML = `${val}/${progress.max}`
    }

    /**
     * @param {string} val
     */
    ["data-view-count"](val) {
        fillElement(this.root, ".entry-progress .view-count", val, "innerhtml")
        let infoRawTbl = this.getAttribute("data-meta-info-raw")
        if (infoRawTbl) {
            let meta
            try {
                meta = JSON.parse(infoRawTbl)
            }
            catch (err) {
                return
            }
            fillElement(this.root, ".entry-progress .view-count", String(Number(val) * Number(meta["Show-length"] || meta["Movie-length"])), "attribute=data-time-spent")
        }
    }

    /**
     * @param {string} val
     */
    ["data-media-dependant"](val) {
        let type = this.getAttribute("data-type")
        type = String(type)

        let caption = /**@type {HTMLElement}*/(this.root.querySelector(".entry-progress figcaption"))

        let data
        try {
            data = JSON.parse(val)
        }
        catch (err) {
            console.error("Could not parse json", val)
            return
        }

        let mediaInfoTbl = /**@type {HTMLTableElement}*/(this.root.querySelector("figure .media-info"))
        this.mkGenericTbl(mediaInfoTbl, data)

        if (data[`${type}-episodes`] && this.getAttribute("data-user-status") === "Viewing") {
            let progress = /**@type {HTMLProgressElement}*/(this.root.querySelector("progress.entry-progress"))
            progress.max = data[`${type}-episodes`]

            let pos = Number(this.getAttribute("data-user-current-position"))
            progress.value = pos

            progress.title = `${pos}/${progress.max}`

            caption.innerText = `${pos}/${progress.max}`
            caption.title = `${Math.round(pos / progress.max * 1000) / 10}%`
        }
    }

    /**
     * @param {string} val
     */
    ["data-user-events"](val) {
        let eventsTbl = /**@type {HTMLTableElement}*/(this.root.querySelector(".user-actions"))
        if (val) {
            let html = `
                <thead>
                    <tr>
                        <!-- this nonsense is so that the title lines up with the events -->
                        <th class="grid column"><button popovertarget="new-event-form">➕︎</button><span style="text-align: center">Event</span></th>
                        <th>Time</th>
                    </tr>
                </thead>
                <tbody>
            `
            for (let event of val.split(",")) {
                let [name, ts, afterts, timeZone] = event.split(":")
                timeZone ||= "UTC"
                let date = new Date(Number(ts))
                let afterDate = new Date(Number(afterts))
                let timeTd = ""
                if (ts !== "0") {
                    let time = date.toLocaleTimeString("en", {timeZone})
                    let dd = date.toLocaleDateString("en", {timeZone})
                    timeTd = `<td title="${time}">${dd}</td>`
                } else if(afterts !== "0") {
                    let time = afterDate.toLocaleTimeString("en", {timeZone})
                    let dd = afterDate.toLocaleDateString("en", {timeZone})
                    timeTd = `<td title="${time}">after: ${dd}</td>`
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
        }
    }
})

customElements.define("sidebar-entry", class extends HTMLElement {
    /**@type {string[]}*/
    static observedAttributes = [
        "data-title",
        "data-thumbnail-src",
        "data-cost",
        "data-user-rating",
        "data-type",
        "data-release-year"
    ]
    constructor() {
        super()
        let template = /**@type {HTMLTemplateElement}*/(document.getElementById("sidebar-entry"))
        let content = /**@type {HTMLElement}*/(template.content.cloneNode(true))
        let root = this.attachShadow({ mode: "open" })
        root.appendChild(content)
        this.root = root

    }

    /**
     * @param {string} val
     */
    ["data-title"](val) {
        fillElement(this.root, ".title", val)
        let imgEl = /**@type {HTMLImageElement}*/(this.root.querySelector(".thumbnail"))
        imgEl.alt = `${val} thumbnail`
    }

    /**
     * @param {string} val
     */
    ["data-thumbnail-src"](val) {
        let imgEl = /**@type {HTMLImageElement}*/(this.root.querySelector(".thumbnail"))
        if (imgEl.src === val) return
        imgEl.src = String(val)
    }

    /**
     * @param {string} val
     */
    ["data-cost"](val) {
        fillElement(this.root, ".cost", `$${val}`)
    }

    /**
     * @param {string} val
     */
    ["data-user-rating"](val) {
        let ratingA = val
        if (ratingA) {
            let rating = Number(ratingA)
            let ratingE = /**@type {HTMLElement?}*/(this.root.querySelector(".rating"))
            if (ratingE) {
                applyUserRating(rating, ratingE)
                ratingE.innerHTML = ratingA
            }
        }
    }


    /**
     * @param {string} val
     */
    ["data-release-year"](val) {
        let title = /**@type {HTMLElement}*/(this.root.querySelector(".title"))
        title.setAttribute("data-release-year", val)
    }

    /**
     * @param {string} val
     */
    ["data-type"](val) {
        let typeIcon = typeToSymbol(String(val))
        let title = /**@type {HTMLElement}*/(this.root.querySelector(".title"))
        title.setAttribute("data-type-icon", typeIcon)
    }

    /**
     * @param {string} name
     * @param {string} ov
     * @param {string} nv
     */
    attributeChangedCallback(name, ov, nv) {
        let root = this.shadowRoot
        if (!root) return

        if (name in this) {
            //@ts-ignore
            this[name](nv)
        }
    }
})

customElements.define("entries-statistic", class extends HTMLElement {
    static observedAttributes = ["data-value"]

    /**
    * @param {string} name
    * @param {string} ov
    * @param {string} nv
    */
    attributeChangedCallback(name, ov, nv) {
        if (name != "data-value") return
        this.innerText = String(Math.round(Number(nv) * 100) / 100)
    }
})

customElements.define("calc-entry", class extends HTMLElement {
    static observedAttributes = ["data-expression-output"]

    constructor() {
        super()
        let template = /**@type {HTMLTemplateElement}*/(document.getElementById("calc-entry"))
        let content = /**@type {HTMLElement}*/(template.content.cloneNode(true))
        let root = this.attachShadow({ mode: "open" })
        root.appendChild(content)
        this.root = root

    }

    /**
     * @param {string} name
     * @param {string} ov
     * @param {string} nv
     */
    attributeChangedCallback(name, ov, nv) {
        if (name !== "data-expression-output") return

        let el = this.root.querySelector(".expression-output")
        if (!el) return
        el.innerHTML = nv
    }
})
