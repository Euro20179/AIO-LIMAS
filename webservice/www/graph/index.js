function getCtx(id) {
    const canv = /**@type {HTMLCanvasElement}*/(document.getElementById(id))
    return canv.getContext("2d")
}

const typeColors = {
    "Manga": "lightyellow",
    "Show": "pink",
    "Movie": "lightblue",
    "MovieShort": "lightskyblue",
    "Song": "lightgreen",
    "Game": "#fed8b1",
    "Book": "gainsboro",
    "Collection": "violet"
}

const ctx = getCtx("by-year")
const typePieCtx = getCtx("type-pie")
const rbyCtx = getCtx("rating-by-year")

/**
* @param {Record<any, any>} obj
* @param {string} label
*/
function fillGap(obj, label) {
    obj[label] = []
    if (!((Number(label) + 1) in obj)) {
        fillGap(obj, String(Number(label) + 1))
    }
}

/**@type {any}*/
let countByFormatChart = null
/**
 * @param {InfoEntry[]} entries
 */
function countByFormat(entries) {
    let data = Object.groupBy(entries, i => formatToName(i.Format))

    let labels = Object.keys(data)
        .sort((a, b) => data[b].length - data[a].length)
    let counts = Array.from(labels, (v, k) => data[v].length)

    if(countByFormatChart) {
        countByFormatChart.destroy()
    }
    countByFormatChart = new Chart(getCtx("count-by-format"), {
        type: 'pie',
        data: {
            labels: labels,
            datasets: [{
                label: "Types",
                data: counts,
                borderWidth: 1,
            }]
        },
        options: {
            plugins: {
                title: {
                    display: true,
                    text: "Format count",
                }
            },
            responsive: true
        }
    });
}
async function treeFilterForm() {
    let form = /**@type {HTMLFormElement}*/(document.getElementById("sidebar-form"))
    let data = new FormData(form)
    let sortBy = data.get("sort-by")
    let status = /**@type {string[]}*/(data.getAll("status"))
    let type = /**@type {string[]}*/(data.getAll("type"))
    let format = /**@type {string[]}*/(data.getAll('format')).filter(n => n !== "")

    let search = /**@type {string}*/(data.get("search-query"))

    let tags = /**@type {string[]}*/(data.getAll("tags"))

    let pgt = /**@type {string}*/(data.get("price-gt"))
    let plt = /**@type {string}*/(data.get("price-lt"))

    let rgt = /**@type {string}*/(data.get("rating-gt"))
    let rlt = /**@type {string}*/(data.get("rating-lt"))

    let formatN = undefined
    if (format.length) {
        formatN = format.map(Number)
    }

    //TODO:
    //Add hasTags, notHasTags, and maybeHasTags
    //allow the user to type #tag #!tag and #?tag in the search bar
    /**@type {DBQuery}*/
    let queryData = {
        status: status.join(","),
        type: type.join(","),
        format: formatN,
        tags: tags.join(","),
        purchasePriceGt: Number(pgt),
        purchasePriceLt: Number(plt),
        userRatingGt: Number(rgt),
        userRatingLt: Number(rlt),
    }

    let shortcuts = {
        "userRatingGt": "r>",
        "userRatingLt": "r<",
        "purchasePriceGt": "p>",
        "purchasePriceLt": "p<",
    }
    for (let word of search.split(" ")) {
        for (let property in queryData) {
            //@ts-ignore
            let shortcut = shortcuts[property]
            let value
            if (word.startsWith(shortcut)) {
                value = word.slice(shortcut.length)
            } else if (word.startsWith(`${property}:`)) {
                value = word.slice(property.length + 1)
            } else {
                continue
            }
            search = search.replace(word, "").trim()
            //@ts-ignore
            queryData[property] = value
        }
    }

    queryData.title = search

    let entries = await loadQueriedEntries(queryData)

    if (sortBy != "") {
        if (sortBy == "rating") {
            entries = entries.sort((a, b) => {
                let aUInfo = findUserEntryById(a.ItemId)
                let bUInfo = findUserEntryById(b.ItemId)
                if (!aUInfo || !bUInfo) return 0
                return bUInfo?.UserRating - aUInfo?.UserRating
            })
        } else if (sortBy == "cost") {
            entries = entries.sort((a, b) => {
                return b.PurchasePrice - a.PurchasePrice
            })
        } else if (sortBy == "general-rating") {
            entries = entries.sort((a, b) => {
                let am = findMetadataById(a.ItemId)
                let bm = findMetadataById(b.ItemId)
                if (!bm || !am) return 0
                return normalizeRating(bm.Rating, bm.RatingMax || 100) - normalizeRating(am.Rating, am.RatingMax || 100)
            })
        } else if (sortBy == "rating-disparity") {
            entries = entries.sort((a, b) => {
                let am = findMetadataById(a.ItemId)
                let au = findUserEntryById(a.ItemId)
                let bm = findMetadataById(b.ItemId)
                let bu = findUserEntryById(b.ItemId)
                if (!bm || !am) return 0
                let bGeneral = normalizeRating(bm.Rating, bm.RatingMax || 100)
                let aGeneral = normalizeRating(am.Rating, am.RatingMax || 100)

                let aUser = Number(au?.UserRating)
                let bUser = Number(bu?.UserRating)


                return (aGeneral - aUser) - (bGeneral - bUser)
            })
        } else if (sortBy == "release-year") {
            entries = entries.sort((a, b) => {
                let am = findMetadataById(a.ItemId)
                let bm = findMetadataById(b.ItemId)
                return (bm?.ReleaseYear || 0) - (am?.ReleaseYear || 0)
            })
        }
    }

    makeGraphs(entries)
}

/**@type {any}*/
let costChart = null
/**
 * @param {InfoEntry[]} entries
 */
function costByFormat(entries) {
    entries = entries.filter(v => v.PurchasePrice > 0)

    let data = Object.groupBy(entries, i => formatToName(i.Format))

    let totals = Object.fromEntries(
        Object.entries(data)
            .map(([name, data]) => [name, data.reduce((p, c) => p + c.PurchasePrice, 0)])
            .sort((a, b) => b[1] - a[1])
    )

    let labels = Object.keys(data)
    totals = Object.values(totals)

    if(costChart) {
        costChart.destroy()
    }
    costChart = new Chart(getCtx("cost-by-format"), {
        type: 'pie',
        data: {
            labels: labels,
            datasets: [{
                label: "Types",
                data: totals,
                borderWidth: 1,
            }]
        },
        options: {
            plugins: {
                title: {
                    display: true,
                    text: "Cost by format"
                }
            },
            responsive: true
        }
    });
}


/**@type {any}*/
let costByTypeChart = null
/**
 * @param {InfoEntry[]} entries
 */
function costByTypePie(entries) {
    entries = entries.filter(v => v.PurchasePrice > 0)
    let data = Object.groupBy(entries, i => i.Type)

    let totals = Object.fromEntries(
        Object.entries(data)
            .map(([name, data]) => [name, data.reduce((p, c) => p + c.PurchasePrice, 0)])
            .sort((a, b) => b[1] - a[1])
    )
    let labels = Object.keys(data)
    let totalList = Object.values(totals)

    let colors = labels.map(v => typeColors[v])

    if(costByTypeChart) {
        costByTypeChart.destroy()
    }

    costByTypeChart = new Chart(getCtx("cost-by-type"), {
        type: 'pie',
        data: {
            labels: labels,
            datasets: [{
                label: "Cost",
                data: totalList,
                borderWidth: 1,
                backgroundColor: colors
            }]
        },
        options: {
            plugins: {
                title: {
                    display: true,
                    text: "Cost by type"
                }
            },
            responsive: true
        }
    });
}


/**@type {any}*/
let typechart = null
/**
 * @param {InfoEntry[]} entries
 */
function typePieChart(entries) {
    let data = Object.groupBy(entries, i => i.Type)

    let labels = Object.keys(data)
        .sort((a, b) => data[b].length - data[a].length)
    let counts = Array.from(labels, (v, k) => data[v].length)

    let colors = labels.map(v => typeColors[v])

    if(typechart) {
        typechart.destroy()
    }
    typechart = new Chart(typePieCtx, {
        type: 'pie',
        data: {
            labels: labels,
            datasets: [{
                label: "Types",
                data: counts,
                borderWidth: 1,
                backgroundColor: colors
            }]
        },
        options: {
            plugins: {
                title: {
                    display: true,
                    text: "Type count",
                }
            },
            responsive: true
        }
    });
}

/**@type {any}*/
let rbyChart = null
/**
 * @param {InfoEntry[]} entries
 */
async function ratingByYear(entries) {
    let user = await loadList("/engagement/list-entries")
    let met = await loadList("/metadata/list-entries")

    let data = Object.groupBy(entries, i => {
        let meta = findEntryById(i.ItemId, met)
        return meta.ReleaseYear
    })

    let highestYear = Object.keys(data).sort((a, b) => +b - +a)[0]
    for (let year in data) {
        let yearInt = Number(year)
        if (highestYear == yearInt) break
        if (yearInt < 1970) continue
        if (!((yearInt + 1) in data)) {
            fillGap(data, yearInt + 1)
        }
    }

    delete data["0"]
    const years = Object.keys(data)
    const ratings = Object.values(data)
        .map(v => v
            .map(i => {
                let thisUser = findEntryById(i.ItemId, user)
                return thisUser.UserRating
            })
            .reduce((p, c, i) => (p * i + c) / (i + 1), 0)
        )

    if (rbyChart) {
        rbyChart.destroy()
    }
    rbyChart = new Chart(rbyCtx, {
        type: 'bar',
        data: {
            labels: years,
            datasets: [{
                label: 'ratings',
                data: ratings,
                borderWidth: 1
            }]
        },
        options: {
            responsive: true,
            scales: {
                y: {
                    beginAtZero: true
                }
            }
        }
    });
}

let bycChart = null
/**
 * @param {InfoEntry[]} entries
 */
async function byc(entries) {
    let met = await loadList("metadata/list-entries")
    let data = Object.groupBy(entries, i => {
        let meta = findEntryById(i.ItemId, met)
        return meta.ReleaseYear
    })
    let highestYear = Object.keys(data).sort((a, b) => +b - +a)[0]
    for (let year in data) {
        let yearInt = Number(year)
        if (+highestYear == yearInt) break
        if (yearInt < 1970) continue
        if (!((yearInt + 1) in data)) {
            fillGap(data, String(yearInt + 1))
        }
    }

    delete data["0"]

    const years = Object.keys(data)
    const counts = Object.values(data).map(v => v.length)
    let total = counts.reduce((p, c) => p + c, 0)
    console.log(`total items by year: ${total}`)

    if (bycChart) {
        bycChart.destroy()
    }
    bycChart = new Chart(ctx, {
        type: 'bar',
        data: {
            labels: years,
            datasets: [{
                label: '#items',
                data: counts,
                borderWidth: 1
            }]
        },
        options: {
            responsive: true,
            scales: {
                y: {
                    beginAtZero: true
                },
            },
        }
    });
}

function treeToEntriesList(tree) {
    let entries = []
    for (let id in tree) {
        tree[id].EntryInfo.ItemId = BigInt(id)
        entries.push(tree[id].EntryInfo)
    }
    return entries
}

/**
 * @param {bigint} id
 * @param {Record<string, any>} entryTable
 * @returns {any}
 */
function findEntryById(id, entryTable) {
    for (let item in entryTable) {
        let entry = entryTable[item]
        if (entry.ItemId === id) {
            return entry
        }
    }
    return null
}

/**
* @param {object} json
*/
function makeGraphs(entries) {
    byc(entries)
    typePieChart(entries)
    ratingByYear(entries)
    costByTypePie(entries)
    costByFormat(entries)
    countByFormat(entries)
}

function makeGraphsWithTree(tree) {
    const entries = treeToEntriesList(tree)
    makeGraphs(entries)
}

let searchQueryElem = /**@type {HTMLInputElement}*/(document.getElementById("search-query"))
searchQueryElem.value = "status:Finished"
treeFilterForm()
