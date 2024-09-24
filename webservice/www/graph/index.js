/**
 * @param {string} id
 */
function getCtx(id) {
    const canv = /**@type {HTMLCanvasElement}*/(document.getElementById(id))
    return /**@type {CanvasRenderingContext2D}*/(canv.getContext("2d"))
}

const typeColors = {
    "Manga": "#e5c890",
    "Show": "#f7c8e0",
    "Movie": "#95bdff",
    "MovieShort": "#b4e4ff",
    "Song": "#dfffd8",
    "Game": "#fed8b1",
    "Book": "gainsboro",
    "Collection": "#b4befe"
}

const ctx = getCtx("by-year")
const typePieCtx = getCtx("type-pie")
const rbyCtx = getCtx("rating-by-year")

const groupBySelect = /**@type {HTMLSelectElement}*/(document.getElementById("group-by"))

/**
 * @param {CanvasRenderingContext2D} ctx
 * @param {string[]} labels
 * @param {number[]} data
 * @param {string} labelText
 * @param {string[]} [colors=[]]
 */
function mkPieChart(ctx, labels, data, labelText, colors=[]) {
    let obj = {
        type: 'pie',
        data: {
            labels: labels,
            datasets: [{
                label:labelText,
                data: data,
                borderWidth: 1,
            }]
        },
        options: {
            plugins: {
                legend: {
                    labels: {
                        color: "white"
                    }
                },
                title: {
                    color: "white",
                    display: true,
                    text: "Format count",
                }
            },
            responsive: true
        }
    }
    if(colors.length) {
        //@ts-ignore
        obj.data.datasets[0].backgroundColor = colors
    }
    //@ts-ignore
    return new Chart(ctx, obj)
}

/**
 * @param {CanvasRenderingContext2D} ctx
 * @param {any[]} x
 * @param {any[]} y
 * @param {string} labelText
 */
function mkBarChart(ctx, x, y, labelText) {
    //@ts-ignore
    return new Chart(ctx, {
        type: 'bar',
        data: {
            labels: x,
            datasets: [{
                label: labelText,
                data: y,
                borderWidth: 1,
                backgroundColor: "#95bdff"
            }]
        },
        options: {
            plugins: {
                legend: {
                    labels: {
                        color: "white"
                    }
                }
            },
            responsive: true,
            scales: {
                y: {
                    ticks: {
                        color: "white"
                    },
                    beginAtZero: true
                },
                x: {
                    ticks: {
                        color: "white"
                    }
                }
            }
        }
    })
}

/**
 * @param {number} watchCount
 * @param {MetadataEntry} meta
 * @returns {number}
 */
function getWatchTime(watchCount, meta) {
    if (!meta.MediaDependant) {
        return 0
    }
    let data = JSON.parse(meta.MediaDependant)
    let length = 0
    for (let type of ["Show", "Movie"]) {
        if (!(`${type}-length` in data)) {
            continue
        }
        length = Number(data[`${type}-length`])
        break
    }
    if (isNaN(length)) {
        return 0
    }
    return length * watchCount
}

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

/**
 * @param {InfoEntry[]} entries
 * @returns {Promise<Record<string, InfoEntry[]>>}
 */
async function organizeData(entries) {
    let met = await loadList("/metadata/list-entries")
    let user = await loadList("/engagement/list-entries")

    let groupBy = groupBySelect.value

    /**@type {Record<string, (i: InfoEntry) => any>}*/
    const groupings = {
        "Year": i => findEntryById(i.ItemId, met).ReleaseYear,
        "Type": i => i.Type,
        "Status": i => {
            let u = findEntryById(i.ItemId, user)
            return u.Status
        },
        "View-count": i => {
            let u = findEntryById(i.ItemId, user)
            return u.ViewCount
        },
        "Is-anime": i => {
            return i.IsAnime
        }
    }

    let data = Object.groupBy(entries, (groupings[/**@type {keyof typeof groupings}*/(groupBy)]))

    if (groupBy === "Year") {
        delete data['0']
    }

    if (groupBy === "Year") {
        let highestYear = +Object.keys(data).sort((a, b) => +b - +a)[0]
        for (let year in data) {
            let yearInt = Number(year)
            if (highestYear == yearInt) break
            if (yearInt < 1970) continue
            if (!((yearInt + 1) in data)) {
                fillGap(data, String(yearInt + 1))
            }
        }
    }
    return /**@type {Record<string, InfoEntry[]>}*/(data)
}

/**@type {any}*/
let wtbyChart = null
/**
 * @param {InfoEntry[]} entries
 */
async function watchTimeByYear(entries) {
    let user = await loadList("/engagement/list-entries")
    let met = await loadList("/metadata/list-entries")

    let data = await organizeData(entries)

    const years = Object.keys(data)

    const watchTimes = Object.values(data)
        .map(v => {
            return v.map(i => {
                let watchCount = findEntryById(i.ItemId, user).ViewCount
                let thisMeta = findEntryById(i.ItemId, met)
                let watchTime = getWatchTime(watchCount, thisMeta)
                return watchTime / 60
            }).reduce((p, c) => p + c, 0)
        })
    console.log(watchTimes)

    if (wtbyChart) {
        wtbyChart.destroy()
    }
    wtbyChart = mkBarChart(getCtx("watch-time-by-year"), years, watchTimes, "Watch time")
}

/**@type {any}*/
let adjRatingChart = null
/**
 * @param {InfoEntry[]} entries
 */
async function adjRatingByYear(entries) {
    let user = await loadList("/engagement/list-entries")
    let data = await organizeData(entries)

    const years = Object.keys(data)

    let items = Object.values(data)
    let totalItems = 0
    for (let item of items) {
        totalItems += item.length
    }
    let avgItems = totalItems / items.length
    const ratings = Object.values(data)
        .map(v => {
            let ratings = v.map(i => {
                let thisUser = findEntryById(i.ItemId, user)
                return thisUser.UserRating
            })
            let totalRating = ratings
                .reduce((p, c) => (p + c), 0)

            let avgRating = totalRating / v.length
            let min = Math.min(...ratings)
            return (avgRating + v.length / (Math.log10(avgItems) / avgItems)) + min
            // return (avgRating + v.length / (Math.log10(avgItems) / avgItems)) + min - max
            //avg + (ROOT(count/avgItems, (count/(<digit-count> * 10))))
        })

    if (adjRatingChart) {
        adjRatingChart.destroy()
    }
    adjRatingChart = mkBarChart(getCtx("adj-rating-by-year"), years, ratings, 'adj ratings')
}

/**@type {any}*/
let countByFormatChart = null
/**
 * @param {InfoEntry[]} entries
 */
function countByFormat(entries) {
    let data = /**@type {Record<string, InfoEntry[]>}*/(Object.groupBy(entries, i => formatToName(i.Format)))

    let labels = Object.keys(data)
        .sort((a, b) => data[b].length - data[a].length)
    let counts = Array.from(labels, (v, k) => data[v].length)

    if (countByFormatChart) {
        countByFormatChart.destroy()
    }
    countByFormatChart = mkPieChart(getCtx("count-by-format"),labels,counts, "Types")
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
        let [property, value] = word.split(":")
        if (!value) continue
        //@ts-ignore
        let shortcut = shortcuts[property]
        if (word.startsWith(shortcut)) {
            value = word.slice(shortcut.length)
        }
        search = search.replace(word, "").trim()
        //@ts-ignore
        queryData[property] = value
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

    let data = /**@type {Record<string, InfoEntry[]>}*/(Object.groupBy(entries, i => formatToName(i.Format)))

    let totals = Object.fromEntries(
        Object.entries(data)
            .map(([name, data]) => [name, data.reduce((p, c) => p + c.PurchasePrice, 0)])
            .sort((a, b) => +b[1] - +a[1])
    )

    let labels = Object.keys(data)
    totals = Object.values(totals)

    if (costChart) {
        costChart.destroy()
    }
    costChart = mkPieChart(getCtx("cost-by-format"),labels,totals, "Types")
}


/**@type {any}*/
let costByTypeChart = null
/**
 * @param {InfoEntry[]} entries
 */
function costByTypePie(entries) {
    entries = entries.filter(v => v.PurchasePrice > 0)
    let data = /**@type {Record<string, InfoEntry[]>}*/(Object.groupBy(entries, i => i.Type))

    let totals = Object.fromEntries(
        Object.entries(data)
            .map(([name, data]) => [name, data.reduce((p, c) => p + c.PurchasePrice, 0)])
            .sort((a, b) => +b[1] - +a[1])
    )
    let labels = Object.keys(data)
    let totalList = Object.values(totals)

    let colors = labels.map(v => typeColors[/**@type {keyof typeof typeColors}*/(v)])

    if (costByTypeChart) {
        costByTypeChart.destroy()
    }

    costByTypeChart = mkPieChart(getCtx("cost-by-type"), labels, totalList, "Cost", colors)
}


/**@type {any}*/
let typechart = null
/**
 * @param {InfoEntry[]} entries
 */
function typePieChart(entries) {
    let data = /**@type {Record<string, InfoEntry[]>}*/(Object.groupBy(entries, i => i.Type))

    let labels = Object.keys(data)
        .sort((a, b) => data[b].length - data[a].length)
    let counts = Array.from(labels, (v, k) => data[v].length)

    let colors = labels.map(v => typeColors[/**@type {keyof typeof typeColors}*/(v)])

    if (typechart) {
        typechart.destroy()
    }
    typechart = mkPieChart(typePieCtx, labels, counts, "Types", colors)
}

/**@type {any}*/
let rbyChart = null
/**
 * @param {InfoEntry[]} entries
 */
async function ratingByYear(entries) {
    let user = await loadList("/engagement/list-entries")

    let data = await organizeData(entries)

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
    rbyChart = mkBarChart(rbyCtx, years, ratings, 'ratings')
}

/**@type {any}*/
let bycChart = null
/**
 * @param {InfoEntry[]} entries
 */
async function byc(entries) {
    let data = await organizeData(entries)

    const years = Object.keys(data)
    const counts = Object.values(data).map(v => v.length)

    if (bycChart) {
        bycChart.destroy()
    }
    bycChart = mkBarChart(ctx, years, counts, '#items')
}

// function treeToEntriesList(tree) {
//     let entries = []
//     for (let id in tree) {
//         tree[id].EntryInfo.ItemId = BigInt(id)
//         entries.push(tree[id].EntryInfo)
//     }
//     return entries
// }

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
* @param {InfoEntry[]} entries
*/
function makeGraphs(entries) {
    byc(entries)
    typePieChart(entries)
    ratingByYear(entries)
    adjRatingByYear(entries)
    costByTypePie(entries)
    costByFormat(entries)
    countByFormat(entries)
    watchTimeByYear(entries)
}

groupBySelect.onchange = function() {
    treeFilterForm()
}

let searchQueryElem = /**@type {HTMLInputElement}*/(document.getElementById("search-query"))
searchQueryElem.value = "status:Finished"
treeFilterForm()
