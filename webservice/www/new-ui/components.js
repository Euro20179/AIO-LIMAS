customElements.define("sidebar-entry", class extends HTMLElement {
    constructor() {
        super()
        let template = /**@type {HTMLTemplateElement}*/(document.getElementById("sidebar-entry"))
        let content = /**@type {HTMLElement}*/(template.content.cloneNode(true))
        let root = this.attachShadow({mode: "open"})
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
        if(cost) {
            costE.innerText = `$${cost}`
        }

        let ratingA = this.getAttribute("data-user-rating")
        if(ratingA) {
            let rating = Number(ratingA)
            let ratingE = /**@type {HTMLElement}*/(this.root.querySelector(".rating"))
            if(rating > 100) {
                ratingE.classList.add("splus-tier")
            } else if(rating > 96) {
                ratingE.classList.add("s-tier")
            } else if(rating > 87) {
                ratingE.classList.add("a-tier")
            } else if(rating > 78) {
                ratingE.classList.add("b-tier")
            } else if(rating > 70) {
                ratingE.classList.add("c-tier")
            } else if(rating > 65) {
                ratingE.classList.add("d-tier")
            } else {
                ratingE.classList.add('f-tier')
            }
            ratingE.append(ratingA)
        }
    }
})
