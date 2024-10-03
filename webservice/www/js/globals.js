const apiPath = "/api/v1"

/**
    * @param {string} text
    * @param {string} textFail
    * @param {NumberConstructor | BigIntConstructor} [numberConverter=Number]
    */
function promptNumber(text, textFail, numberConverter = Number) {
    let n = prompt(text)
    while (n !== null && n !== "" && isNaN(Number(n))) {
        n = prompt(textFail)
    }
    if (n === null || n === "") return null
    return numberConverter(n)
}
