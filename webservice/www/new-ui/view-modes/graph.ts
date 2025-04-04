function getCtx2(id: string): CanvasRenderingContext2D {
    const canv = document.getElementById(id) as HTMLCanvasElement
    return canv.getContext("2d") as CanvasRenderingContext2D
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

const groupBySelect = document.getElementById("group-by") as HTMLSelectElement
const typeSelection = document.getElementById("chart-type") as HTMLSelectElement

const groupByInput = document.getElementById("group-by-expr") as HTMLInputElement

function ChartManager(mkChart: (entries: InfoEntry[]) => Promise<any>) {
    let chrt: any = null
    return async function(entries: InfoEntry[]) {
        if (chrt) {
            chrt.destroy()
        }
        chrt = await mkChart(entries)
    }
}

function mkPieChart(ctx: CanvasRenderingContext2D, labels: string[], data: number[], labelText: string, colors: string[] = []) {
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

function mkBarChart(ctx: CanvasRenderingContext2D, x: any[], y: any[], labelText: string) {
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

function mkXTypeChart(ctx: CanvasRenderingContext2D, x: any[], y: any[], labelText: string) {
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
        let dict: Record<any, any> = {}
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

function getWatchTime(watchCount: number, meta: MetadataEntry): number {
    if (!meta.MediaDependant) {
        return 0
    }
    let data
    try{
        data = JSON.parse(meta.MediaDependant)
    }
    catch(err){
        console.error("Could not parse json", meta.MediaDependant)
        return 0
    }
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

function fillGap(obj: Record<any, any>, label: string) {
    obj[label] = []
    if (!((Number(label) + 1) in obj)) {
        fillGap(obj, String(Number(label) + 1))
    }
}

function organizeDataByExpr(entries: InfoEntry[]): Record<string, InfoEntry[]> {
    let expr = groupByInput.value

    let group: Record<string, InfoEntry[]> = Object.groupBy(entries, item => {
        let meta = findMetadataById(item.ItemId)
        let user = findUserEntryById(item.ItemId)
        let symbols = makeSymbolsTableFromObj({ ...item, ...meta, ...user })

        return parseExpression(expr, symbols).toStr().jsValue
    }) as Record<string, InfoEntry[]>
    //Sometimes there's just empty groups i'm not sure why
    for (let key in group) {
        if (group[key].length === 0) {
            delete group[key]
        }
    }
    return group
}

async function organizeData(entries: InfoEntry[]): Promise<[string[], InfoEntry[][]]> {
    let groupBy = groupBySelect.value

    const groupings: Record<string, (i: InfoEntry) => any> = {
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
        "Format": i => formats[i.Format] || "N/A",
        "Status": i => globalsNewUi.userEntries[String(i.ItemId)].Status,
        "View-count": i => globalsNewUi.userEntries[String(i.ItemId)].ViewCount,
        "Is-anime": i => (i.ArtStyle & 1) == 1,
        "Item-name": i => i.En_Title
    }

    let data: Record<string, InfoEntry[]>
    if (groupBy === "Tags") {
        data = {}
        for (let item of entries) {
            for (let tag of item.Tags) {
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
        data = Object.groupBy(entries, (groupings[/**@type {keyof typeof groupings}*/(groupBy)])) as Record<string, InfoEntry[]>
    }


    let sortBy = document.getElementsByName("sort-by")[0] as HTMLInputElement

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

function sortXY(x: string[], y: any[]) {
    let associated: Record<string, any> = {}
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
            // let min = Math.min(...ratings)

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

function makeGraphs(entries: InfoEntry[]) {
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

const modeGraphView: DisplayMode = {
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
