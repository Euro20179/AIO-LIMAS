/**
 * @param {HTMLElement | ShadowRoot} root
 * @param {string} selector
 * @param {string} text
 * @param {"append" | "innerhtml"} [fillmode="append"] 
 */
function fillElement(root, selector, text, fillmode="append") {
    let elem = /**@type {HTMLElement}*/(root.querySelector(selector))
    if(!elem) {
        return
    }
    if(fillmode === "append") {
        elem.append(text)
    } else {
        elem.innerHTML = text
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
        let title = /**@type {HTMLElement}*/(this.root.querySelector(".title"))
        let titleText = this.getAttribute("data-title")
        title.append(String(titleText))

        let imgEl = /**@type {HTMLImageElement}*/(this.root.querySelector(".thumbnail"))
        let thA = this.getAttribute("data-thumbnail-src")
        if (thA) {
            imgEl.src = thA
            imgEl.alt = String(titleText)
        }

        let costA = this.getAttribute("data-cost")
        if(costA) {
            fillElement(this.root, ".cost", `$${costA}`)
        }

        let descA = this.getAttribute("data-description")
        if(descA) {
            fillElement(this.root, ".description", descA, "innerhtml")
        }

        let notes = this.getAttribute("data-user-notes")
        if(notes) {
            fillElement(this.root, ".notes", notes, "innerhtml")
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
                if(ts !== "0") {
                    time = date.toLocaleTimeString("en", {timeZone: "America/Los_Angeles"})
                    dd = date.toLocaleDateString("en", {timeZone: "America/Los_Angeles"})
                }
                html += `<tr><td>${name}</td><td title="${time}">${dd}</td></tr>`
            }
            html += "</tbody>"
            eventsTbl.innerHTML = html
        }
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
    connectedCallback() {
        let title = /**@type {HTMLElement}*/(this.root.querySelector(".title"))
        let titleText = this.getAttribute("data-title")
        title.append(String(titleText))

        let imgEl = /**@type {HTMLImageElement}*/(this.root.querySelector(".thumbnail"))
        imgEl.src = String(this.getAttribute("data-thumbnail-src"))
        imgEl.alt = `${titleText} thumbnail`

        let costE = /**@type {HTMLElement}*/(this.root.querySelector(".cost"))
        let cost = this.getAttribute("data-cost")
        if (cost) {
            costE.innerText = `$${cost}`
        }

        let ratingA = this.getAttribute("data-user-rating")
        if (ratingA) {
            let rating = Number(ratingA)
            let ratingE = /**@type {HTMLElement}*/(this.root.querySelector(".rating"))
            if (rating > 100) {
                ratingE.classList.add("splus-tier")
            } else if (rating > 96) {
                ratingE.classList.add("s-tier")
            } else if (rating > 87) {
                ratingE.classList.add("a-tier")
            } else if (rating > 78) {
                ratingE.classList.add("b-tier")
            } else if (rating > 70) {
                ratingE.classList.add("c-tier")
            } else if (rating > 65) {
                ratingE.classList.add("d-tier")
            } else {
                ratingE.classList.add('f-tier')
            }
            ratingE.append(ratingA)
        }
    }
})
