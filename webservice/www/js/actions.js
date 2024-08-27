/**
* @param {HTMLButtonElement} elem 
*/
function findRoot(elem) {
    let entryElem = elem.parentElement
    while(entryElem && !entryElem?.getAttribute("data-entry-id")) {
        entryElem = entryElem.parentElement
    }
    return entryElem
}

/**
* @param {HTMLButtonElement} elem 
* @param {string} action 
* @param {string} queryParams 
*/
function mediaAction(elem, action, queryParams = "") {
    if(!confirm(`Are you sure you want to ${action} this entry`)) {
        return
    }
    let entryElem = findRoot(elem)

    let entryId = BigInt(entryElem?.getAttribute("data-entry-id") || 0)
    if(entryId == 0n) {
        alert(`Could not ${action} entry`)
        return
    }

    fetch(`${apiPath}/engagement/${action}-media?id=${entryId}${queryParams}`)
        .then(res => res.text())
        .then(alert)
        .catch(console.error)
}

/**
* @param {HTMLButtonElement} elem 
*/
function beginMedia(elem) {
    mediaAction(elem, "begin", "")
}

/**
* @param {HTMLButtonElement} elem 
*/
function endMedia(elem) {
    let rating = prompt("Rating")
    while(isNaN(Number(rating))) {
        rating = prompt("Not a number\nRating")
    }
    mediaAction(elem, "finish", `&rating=${rating}`)
}
