const apiPath = "/api/v1"

/**
    * @param {string} text
    * @param {string} textFail
    */
function promptNumber(text, textFail) {
    let n = prompt(text)
    while (n !== null && n !== "" && isNaN(Number(n))) {
        n = prompt(textFail)
    }
    if (n === null || n === "") return null
    return Number(n)
}
