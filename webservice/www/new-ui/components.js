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
    constructor() {
        super()
        let template = /**@type {HTMLTemplateElement}*/(document.getElementById("display-entry"))
        let content = /**@type {HTMLElement}*/(template.content.cloneNode(true))
        let root = this.attachShadow({ mode: "open" })
        root.appendChild(content)
        this.root = root
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
