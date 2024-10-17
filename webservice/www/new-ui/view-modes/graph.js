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
const rbyCtx = getCtx2("rating-by-year")

const groupBySelect = /**@type {HTMLSelectElement}*/(document.getElementById("group-by"))
const typeSelection = /**@type {HTMLSelectElement}*/(document.getElementById("chart-type"))

const groupByInput = /**@type {HTMLInputElement}*/(document.getElementById("group-by-expr"))

/**
 * @param {(entries: InfoEntry[]) => Promise<any>} mkChart
 */
function ChartManager(mkChart) {
    /**@type {any}*/
    let chrt = null
    /**
     * @param {InfoEntry[]} entries
     */
    return async function(entries) {
        if (chrt) {
            chrt.destroy()
        }
        chrt = await mkChart(entries)
    }
}

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
 * @param {CanvasRenderingContext2D} ctx
 * @param {any[]} x
 * @param {any[]} y
 * @param {string} labelText
 */
function mkXTypeChart(ctx, x, y, labelText) {
    const ty = typeSelection.value
    if (ty === "bar") {
        return mkBarChart(ctx, x, y, labelText)
    } else {
        let totalY = ty === "pie-percentage"
            ? y.reduce((p, c) => p + c, 0)
            : 0

        //sort y values, keep x in same order as y
        //this makes the pie chart look better

        //put x, y into a dict so that a given x can be assigned to a given y easily
        /**@type {Record<any, any>}*/
        let dict = {}
        for (let i = 0; i < y.length; i++) {
            dict[x[i]] = ty === 'pie-percentage' ? y[i] / totalY * 100 : y[i]
        }

        //sort entries based on y value
        const sorted = Object.entries(dict).sort(([_, a], [__, b]) => b - a)

        //y is the 2nd item in the entry list
        y = sorted.map(i => i[1])
        //x is the 1st item in the entry list
        x = sorted.map(i => i[0])

        return mkPieChart(ctx, x, y, labelText)
    }
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
 * @returns {Record<string, InfoEntry[]>}
 */
function organizeDataByExpr(entries) {
    let expr = groupByInput.value

    let group = /**@type {Record<string, InfoEntry[]>}*/(Object.groupBy(entries, item => {
        let meta = findMetadataById(item.ItemId)
        let user = findUserEntryById(item.ItemId)
        let symbols = makeSymbolsTableFromObj({ ...item, ...meta, ...user })

        return parseExpression(expr, symbols).toStr().jsValue
    }))
    //Sometimes there's just empty groups i'm not sure why
    for (let key in group) {
        if (group[key].length === 0) {
            delete group[key]
        }
    }
    return group
}

/**
 * @param {InfoEntry[]} entries
 * @returns {Promise<[string[], InfoEntry[][]]>}
 */
async function organizeData(entries) {
    let groupBy = groupBySelect.value

    /**@type {Record<string, (i: InfoEntry) => any>}*/
    const groupings = {
        "Year": i => globalsNewUi.metadataEntries[String(i.ItemId)].ReleaseYear,
        "Decade": i => {
            const year = String(globalsNewUi.metadataEntries[String(i.ItemId)].ReleaseYear)
            if (year == "0") {
                return "0"
            }
            let century = year.slice(0, 2)
            let decade = year.slice(2)[0]
            return `${century}${decade}0s`
        },
        "Century": i => {
            const year = String(globalsNewUi.metadataEntries[String(i.ItemId)].ReleaseYear)
            if (year == "0") {
                return "0"
            }
            let century = year.slice(0, 2)
            return `${century}00s`
        },
        "Type": i => i.Type,
        "Format": i => formatToName(i.Format),
        "Status": i => globalsNewUi.userEntries[String(i.ItemId)].Status,
        "View-count": i => globalsNewUi.userEntries[String(i.ItemId)].ViewCount,
        "Is-anime": i => (i.ArtStyle & 1) == 1,
        "Item-name": i => i.En_Title
    }

    /**@type {Record<string, InfoEntry[]>}*/
    let data
    if (groupBy === "Tags") {
        data = {}
        for (let item of entries) {
            for (let tag of item.Collection.split(",")) {
                if (data[tag]) {
                    data[tag].push(item)
                } else {
                    data[tag] = [item]
                }
            }
        }
    }
    else if (groupByInput.value) {
        data = organizeDataByExpr(entries)
    }
    else {
        data = /**@type {Record<string, InfoEntry[]>}*/(Object.groupBy(entries, (groupings[/**@type {keyof typeof groupings}*/(groupBy)])))
    }


    let sortBy = /**@type {HTMLInputElement}*/(document.getElementsByName("sort-by")[0])

    //filling in years messes up the sorting, idk why
    if (sortBy.value == "") {
        //this is the cutoff year because this is when jaws came out and changed how movies were produced
        const cutoffYear = 1975
        if (groupBy === "Year") {
            let highestYear = +Object.keys(data).sort((a, b) => +b - +a)[0]
            for (let year in data) {
                let yearInt = +year
                if (highestYear === yearInt) break
                if (yearInt < cutoffYear) continue
                if (!((yearInt + 1) in data)) {
                    fillGap(data, String(yearInt + 1))
                }
            }
        }
    }

    let x = Object.keys(data)
    let y = Object.values(data)
    if (sortBy.value == "rating") {
        let sorted = Object.entries(data).sort((a, b) => {
            let aRating = a[1].reduce((p, c) => {
                let user = findUserEntryById(c.ItemId)
                return p + (user?.UserRating || 0)
            }, 0)
            let bRating = b[1].reduce((p, c) => {
                let user = findUserEntryById(c.ItemId)
                return p + (user?.UserRating || 0)
            }, 0)

            return bRating / b[1].length - aRating / a[1].length
        })
        x = sorted.map(v => v[0])
        y = sorted.map(v => v[1])
    } else if (sortBy.value === "cost") {
        let sorted = Object.entries(data).sort((a, b) => {
            let aCost = a[1].reduce((p, c) => {
                return p + (c?.PurchasePrice || 0)
            }, 0)
            let bCost = b[1].reduce((p, c) => {
                return p + (c?.PurchasePrice || 0)
            }, 0)

            return bCost - aCost
        })
        x = sorted.map(v => v[0])
        y = sorted.map(v => v[1])
    }
    return [x, y]
}

/**
 * @param {string[]} x
 * @param {any[]} y
 */
function sortXY(x, y) {
    /**@type {Record<string, any>}*/
    let associated = {}
    for (let i = 0; i < x.length; i++) {
        associated[x[i]] = y[i]
    }

    let associatedList = Object.entries(associated).sort(([_, a], [__, b]) => b - a)
    x = associatedList.map(x => x[0])
    y = associatedList.map(x => x[1])
    return [x, y]
}

const watchTimeByYear = ChartManager(async (entries) => {
    let [years, data] = await organizeData(entries)

    let watchTimes = data
        .map(v => {
            return v.map(i => {
                let watchCount = globalsNewUi.userEntries[String(i.ItemId)].ViewCount
                let thisMeta = globalsNewUi.metadataEntries[String(i.ItemId)]
                let watchTime = getWatchTime(watchCount, thisMeta)
                return watchTime / 60
            }).reduce((p, c) => p + c, 0)
        });

    return mkXTypeChart(getCtx2("watch-time-by-year"), years, watchTimes, "Watch time")
})

const adjRatingByYear = ChartManager(async (entries) => {
    let [years, data] = await organizeData(entries)

    let items = data
    let totalItems = 0
    let totalRating = 0
    for (let item of items) {
        totalItems += item.length
        totalRating += item.reduce((p, c) => p + globalsNewUi.userEntries[String(c.ItemId)].UserRating, 0)
    }
    let avgItems = totalItems / items.length
    let generalAvgRating = totalRating / totalItems
    const ratings = data
        .map(v => {
            let ratings = v.map(i => {
                let thisUser = globalsNewUi.userEntries[String(i.ItemId)]
                return thisUser.UserRating
            })
            let totalRating = ratings
                .reduce((p, c) => (p + c), 0)

            let avgRating = totalRating / v.length
            let min = Math.min(...ratings)

            return (avgRating - generalAvgRating) + (v.length - avgItems)

            // return (avgRating + v.length / (Math.log10(avgItems) / avgItems)) + min
        })

    return mkXTypeChart(getCtx2("adj-rating-by-year"), years, ratings, 'adj ratings')
})

const costByFormat = ChartManager(async (entries) => {
    entries = entries.filter(v => v.PurchasePrice > 0)
    let [labels, data] = await organizeData(entries)
    let totals = data.map(v => v.reduce((p, c) => p + c.PurchasePrice, 0))

    return mkXTypeChart(getCtx2("cost-by-format"), labels, totals, "Cost by")
})

const ratingByYear = ChartManager(async (entries) => {
    let [years, data] = await organizeData(entries)
    const ratings = data
        .map(v => v
            .map(i => {
                let thisUser = globalsNewUi.userEntries[String(i.ItemId)]
                return thisUser.UserRating
            })
            .reduce((p, c, i) => (p * i + c) / (i + 1), 0)
        )

    return mkXTypeChart(rbyCtx, years, ratings, 'ratings')
})

const generalRating = ChartManager(async (entries) => {
    let [years, data] = await organizeData(entries)
    const ratings = data.map(v => {
        return v.map(i => {
            let meta = findMetadataById(i.ItemId)
            let rating = meta?.Rating
            let max = meta?.RatingMax
            if (rating && max) {
                return (rating / max) * 100
            }
            return 0
        }).reduce((p, c, i) => (p * i + c) / (i + 1), 0)
    })
    return mkXTypeChart(getCtx2("general-rating-by-year"), years, ratings, "general ratings")
})

const ratingDisparityGraph = ChartManager(async (entries) => {
    let [years, data] = await organizeData(entries)
    const disparity = data.map(v => {
        return v.map(i => {
            let meta = findMetadataById(i.ItemId)
            let user = findUserEntryById(i.ItemId)
            let rating = meta?.Rating
            let max = meta?.RatingMax
            if (rating && max) {
                let general = (rating / max) * 100
                return (user?.UserRating || 0) - general
            }
            return user?.UserRating || 0
        }).reduce((p, c) => p + c, 0)
    })
    return mkXTypeChart(getCtx2("rating-disparity-graph"), years, disparity, "Rating disparity")
})

const byc = ChartManager(async (entries) => {
    let [years, data] = await organizeData(entries)
    const counts = data.map(v => v.length)

    return mkXTypeChart(ctx, years, counts, '#items')
})

/**
* @param {InfoEntry[]} entries
*/
function makeGraphs(entries) {
    byc(entries)
    ratingByYear(entries)
    adjRatingByYear(entries)
    costByFormat(entries)
    watchTimeByYear(entries)
    generalRating(entries)
    ratingDisparityGraph(entries)
}

groupByInput.onchange = function() {
    makeGraphs(globalsNewUi.selectedEntries)
}

groupBySelect.onchange = typeSelection.onchange = function() {
    makeGraphs(globalsNewUi.selectedEntries)
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
