//TODO:
//display statistics eg:
//  total watch time
//  total items
//  total spent
//
// TODO: add back group by

/**
 * @param {string} id
 */
function getCtx2(id) {
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

const ctx = getCtx2("by-year")
const typePieCtx = getCtx2("type-pie")
const rbyCtx = getCtx2("rating-by-year")

const groupBySelect = /**@type {HTMLSelectElement}*/(document.getElementById("group-by"))

/**
 * @param {CanvasRenderingContext2D} ctx
 * @param {string[]} labels
 * @param {number[]} data
 * @param {string} labelText
 * @param {string[]} [colors=[]]
 */
function mkPieChart(ctx, labels, data, labelText, colors = []) {
    let obj = {
        type: 'pie',
        data: {
            labels: labels,
            datasets: [{
                label: labelText,
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
                    text: labelText,
                }
            },
            responsive: true
        }
    }
    if (colors.length) {
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
                    grid: {
                        color: "grey"
                    },
                    ticks: {
                        color: "white"
                    },
                    beginAtZero: true
                },
                x: {
                    grid: {
                        color: "grey"
                    },
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

    if (wtbyChart) {
        wtbyChart.destroy()
    }
    wtbyChart = mkBarChart(getCtx2("watch-time-by-year"), years, watchTimes, "Watch time")
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
    adjRatingChart = mkBarChart(getCtx2("adj-rating-by-year"), years, ratings, 'adj ratings')
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
    countByFormatChart = mkPieChart(getCtx2("count-by-format"), labels, counts, "Formats")
}
async function treeFilterForm() {
    let form = /**@type {HTMLFormElement}*/(document.getElementById("sidebar-form"))

    let entries = await doQuery(form)

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
    costChart = mkPieChart(getCtx2("cost-by-format"), labels, totals, "Types")
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

    costByTypeChart = mkPieChart(getCtx2("cost-by-type"), labels, totalList, "Cost", colors)
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

/**
 * @type {DisplayMode}
 */
const modeGraphView = {
    add(entry, updateStats = true) {
        updateStats && changeResultStatsWithItem(entry)
        makeGraphs(globalsNewUi.selectedEntries)
    },

    sub(entry, updateStats = true) {
        updateStats && changeResultStatsWithItem(entry, -1)
        makeGraphs(globalsNewUi.selectedEntries)
    },

    addList(entries, updateStats = true) {
        updateStats && changeResultStatsWithItemList(entries)

        makeGraphs(globalsNewUi.selectedEntries)
    },

    subList(entries, updateStats = true) {
        updateStats && changeResultStatsWithItemList(entries, -1)
    }
}
