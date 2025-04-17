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
    constructor() {
        super()
        let template = /**@type {HTMLTemplateElement}*/(document.getElementById("sidebar-entry"))
        let content = /**@type {HTMLElement}*/(template.content.cloneNode(true))
        let root = this.attachShadow({ mode: "open" })
        root.appendChild(content)
        this.root = root

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

        let el = this.root.getElementById("expression-output")
        if (!el) return
        el.innerHTML = nv
    }
})
