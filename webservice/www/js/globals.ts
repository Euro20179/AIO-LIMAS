const apiPath = "/api/v1"

const urlParams = new URLSearchParams(document.location.search)
const uid = urlParams.get("uid")
const initialSearch = urlParams.get("q")

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

/**
* @param {string} text
* @param {string[]} allItems
* @returns {number | null}
* @description Grabs the sequence number from a string, given a list of all items in the sequence
* eg: `text: "friends s01"`, `allItems: ["friends s01", "friends s02", ...]`
* This would return `1` because `1` is the sequence number in `text`
*/
function sequenceNumberGrabber(text, allItems) {
    //match sequence indicator (non-word character, vol/ova/e/s)
    //followed by sequence number (possibly a float)
    //followed by non-word character
    //eg: S01
    //eg: E6.5
    const regex = /(?:[\W_\-\. EesS]|[Oo][Vv][Aa]|[Vv](?:[Oo][Ll])?\.?)?(\d+(?:\.\d+)?)[\W_\-\. ]?/g

    const matches = text.matchAll(regex).toArray()
    if(matches[0] == null) {
        return null
    }
    return Number(matches.filter(match => {
        for(let item of allItems) {
            if(item === text) continue

            if(item.includes(match[0]))
                return false
            return true
        }
    })[0][1])
}
