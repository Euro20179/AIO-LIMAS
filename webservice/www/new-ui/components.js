/**
 * @param {HTMLElement | ShadowRoot} root
 * @param {string} selector
 * @param {string} text
 * @param {"append" | "innerhtml"} [fillmode="append"] 
 */
function fillElement(root, selector, text, fillmode = "append") {
    let elem = /**@type {HTMLElement}*/(root.querySelector(selector))
    if (!elem) {
        return
    }
    if (fillmode === "append") {
        elem.innerText = text
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
    constructor() {
        super()
        let template = /**@type {HTMLTemplateElement}*/(document.getElementById("display-entry"))
        let content = /**@type {HTMLElement}*/(template.content.cloneNode(true))
        let root = this.attachShadow({ mode: "open" })
        root.appendChild(content)
        this.root = root
    }

    connectedCallback() {
        let type = this.getAttribute("data-type")
        type = String(type)
        let typeIcon = typeToSymbol(type)
        let format = this.getAttribute("data-format")
        let formatName = formatToName(Number(format))

        let title = /**@type {HTMLElement}*/(this.root.querySelector(".title"))
        let titleText = this.getAttribute("data-title")
        title.innerText = String(titleText)
        title.setAttribute("data-type-icon", typeIcon)
        title.setAttribute("data-format-name", formatName)

        let imgEl = /**@type {HTMLImageElement}*/(this.root.querySelector(".thumbnail"))
        let thA = this.getAttribute("data-thumbnail-src")
        if (thA) {
            imgEl.src = thA
            imgEl.alt = String(titleText)
        }

        let costA = this.getAttribute("data-cost")
        if (costA) {
            fillElement(this.root, ".cost", `$${costA}`)
        }

        let descA = this.getAttribute("data-description")
        if (descA) {
            fillElement(this.root, ".description", descA, "innerhtml")
        }

        let notes = this.getAttribute("data-user-notes")
        if (notes) {
            fillElement(this.root, ".notes", notes, "innerhtml")
        }

        let nativeTitle = this.getAttribute("data-native-title")
        if (nativeTitle) {
            let el = /**@type {HTMLElement}*/(this.root.querySelector(".title"))
            el.title = nativeTitle
        }

        let ratingA = this.getAttribute("data-user-rating")
        let ratingE = /**@type {HTMLElement}*/(this.root.querySelector(".rating"))
        if (ratingA) {
            let rating = Number(ratingA)
            applyUserRating(rating, ratingE)
            ratingE.innerHTML = ratingA
        } else {
            ratingE.innerText = "Unrated"
        }


        /**
         * @param {HTMLElement} root
         * @param {Record<any, any>} data
         */
        function mkGenericTbl(root, data) {
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
                html += `<tr><td>${key}</td><td>${data[key]}</td></tr>`
            }
            html += "</tbody>"
            root.innerHTML = html
        }

        let infoRawTbl = /**@type {HTMLTableElement}*/(this.root.querySelector(".info-raw"))
        let infoRaw = this.getAttribute("data-info-raw")
        if (infoRaw) {
            mkGenericTbl(infoRawTbl, JSON.parse(infoRaw))
        }


        let figure = /**@type {HTMLElement}*/(this.root.querySelector("figure.entry-progress"))

        let userStatus = this.getAttribute("data-user-status")
        if (userStatus) {
            fillElement(this.root, ".entry-progress .status", userStatus, "innerhtml")
        }

        let caption = /**@type {HTMLElement}*/(this.root.querySelector(".entry-progress figcaption"))

        let viewCount = this.getAttribute("data-view-count")
        if (viewCount) {
            fillElement(this.root, ".entry-progress .view-count", viewCount, "innerhtml")
        }

        let mediaInfoTbl = /**@type {HTMLTableElement}*/(this.root.querySelector("figure .media-info"))
        let mediaInfoRaw = this.getAttribute("data-media-dependant")
        if (mediaInfoRaw) {
            let data = JSON.parse(mediaInfoRaw)
            mkGenericTbl(mediaInfoTbl, data)
            if (data[`${type}-episodes`] && this.getAttribute("data-user-status") === "Viewing") {
                let progress = /**@type {HTMLProgressElement}*/(this.root.querySelector("progress.entry-progress"))
                progress.max = data[`${type}-episodes`]
                let pos = Number(this.getAttribute("data-user-current-position"))
                progress.value = pos
                caption.innerText = `${pos}/${progress.max}`
                caption.title = `${Math.round(pos / progress.max * 1000) / 10}%`
            }
        }

        let eventsTbl = /**@type {HTMLTableElement}*/(this.root.querySelector(".user-actions"))
        let eventsA = this.getAttribute("data-user-events")
        if (eventsA) {
            let html = `
                <thead>
                    <tr>
                        <th>Event</th>
                        <th>Time</th>
                    </tr>
                </thead>
                <tbody>
            `
            for (let event of eventsA.split(",")) {
                let [name, ts] = event.split(":")
                let date = new Date(Number(ts))
                let time = "unknown"
                let dd = "unknown"
                if (ts !== "0") {
                    time = date.toLocaleTimeString("en", { timeZone: "America/Los_Angeles" })
                    dd = date.toLocaleDateString("en", { timeZone: "America/Los_Angeles" })
                }
                html += `<tr><td>${name}</td><td title="${time}">${dd}</td></tr>`
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
        "data-type"
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
            let ratingE = /**@type {HTMLElement}*/(this.root.querySelector(".rating"))
            applyUserRating(rating, ratingE)
            ratingE.innerHTML = ratingA
        }
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
