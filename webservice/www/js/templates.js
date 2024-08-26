/**
 * @typedef TemplateFiller
 * @type {Record<string, string | ((el: HTMLElement) => any)>}
 * @description key is the selector, value is the text to put in the found element
 */

/**
    * @param {string} templateId
    * @param {TemplateFiller} fills
    * @returns {HTMLElement} The root element
*/
function fillTemplate(templateId, fills) {
    const template = /**@type {HTMLTemplateElement}*/(document.getElementById(templateId))

    const clone = /**@type {HTMLElement}*/(template.content.cloneNode(true))

    const root = /**@type {HTMLElement}*/(clone.querySelector(":first-child"))

    for(let key in fills) {
        let value = fills[key]
        let el = /**@type {HTMLElement}*/(root.querySelector(key))
        if (typeof value === 'function') {
            value(el)
        } else {
            el.append(value)
        }
    }

    return root
}
