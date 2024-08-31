/**
 * @typedef TemplateFiller
 * @type {Record<string, string | ((el: HTMLElement) => any)>}
 * @description key is the selector, value is the text to put in the found element
 */

/**
    * @param {string | HTMLTemplateElement} templateId
    * @param {TemplateFiller} fills
    * @returns {HTMLElement} The root element
*/
function fillTemplate(templateId, fills) {
    if(typeof templateId === 'string') {
        templateId = /**@type {HTMLTemplateElement}*/(document.getElementById(templateId))
    }

    const clone = /**@type {HTMLElement}*/(templateId.content.cloneNode(true))

    const root = /**@type {HTMLElement}*/(clone.querySelector(":first-child"))
    fillRoot(root, fills)

    return root
}

/**
    * @param {HTMLElement} root
    * @param {TemplateFiller} fills
*/
function fillRoot(root, fills) {
    for(let key in fills) {
        let value = fills[key]
        let el = /**@type {HTMLElement}*/(root.querySelector(key))
        if (typeof value === 'function') {
            value(el)
        } else {
            el.append(value)
        }
    }


}
