const apiPath = "/api/v1"

const urlParams = new URLSearchParams(document.location.search)
const uid = urlParams.get("uid")

if(!uid) {
    (document.getElementById("no-uid") as HTMLDivElement).style.display = "block"
}

const initialSearch = urlParams.get("q")

const searchInput = document.querySelector("[name=\"search-query\"]") as HTMLInputElement
if(searchInput && initialSearch) {
    searchInput.value = decodeURIComponent(initialSearch)
}

function promptNumber(text: string, textFail: string, numberConverter: NumberConstructor | BigIntConstructor = Number) {
    let n = prompt(text)
    while (n !== null && n !== "" && isNaN(Number(n))) {
        n = prompt(textFail)
    }
    if (n === null || n === "") return null
    return numberConverter(n)
}

/**
* @description Grabs the sequence number from a string, given a list of all items in the sequence
*/
function sequenceNumberGrabber(text: string, allItems: string[]): number | null {
    //match sequence indicator (non-word character, vol/ova/e/s)
    //followed by sequence number (possibly a float)
    //followed by non-word character
    //eg: S01
    //eg: E6.5
    const regex = /(?:[\W_\-\. EesS]|[Oo][Vv][Aa]|[Vv](?:[Oo][Ll])?\.?)?(\d+(?:\.\d+)?)[\W_\-\. ]?/g

    const matches = text.matchAll(regex).toArray()
    if (matches[0] == null) {
        return null
    }
    return Number(matches.filter(match => {
        for (let item of allItems) {
            if (item === text) continue

            if (item.includes(match[0]))
                return false
            return true
        }
    })[0][1])
}
