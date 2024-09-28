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
            html += `<tr><td>${key}</td><td>${data[key]}</td></tr>`
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
        if (val && ratingE){
            let rating = Number(val)
            applyUserRating(rating, ratingE)
            ratingE.innerHTML = val
        } else if(ratingE){
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
        } else if(ratingE){
            ratingE.innerText = "Unrated"
        }
    }


    /**
     * @param {string} val
     */
    ["data-info-raw"](val) {
        let infoRawTbl = /**@type {HTMLTableElement}*/(this.root.querySelector(".info-raw"))
        this.mkGenericTbl(infoRawTbl, JSON.parse(val))
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

        progress.title = `${val}`

        if(progress.max) {
            progress.title = `${val}/${progress.max}`
        }
    }

    /**
     * @param {string} val
     */
    ["data-view-count"](val) {
        fillElement(this.root, ".entry-progress .view-count", val, "innerhtml")
    }

    /**
     * @param {string} val
     */
    ["data-media-dependant"](val) {
        let type = this.getAttribute("data-type")
        type = String(type)

        let caption = /**@type {HTMLElement}*/(this.root.querySelector(".entry-progress figcaption"))

        let data = JSON.parse(val)

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
                        <th>Event</th>
                        <th>Time</th>
                    </tr>
                </thead>
                <tbody>
            `
            for (let event of val.split(",")) {
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
        console.log(imgEl.src, val)
        if(imgEl.src === val) return
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
            if(ratingE) {
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
        this.innerText = nv
    }
})
