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

function countByFormat(json) {
    let values = Object.values(json)
    let data = Object.groupBy(values, i => formatToName(i.EntryInfo.Format))

    let labels = Object.keys(data)
        .sort((a, b) => data[b].length - data[a].length)
    let counts = Array.from(labels, (v, k) => data[v].length)

    new Chart(getCtx("count-by-format"), {
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

function costByFormat(json) {
    let values = Object.values(json)
        .filter(v => v.EntryInfo.PurchasePrice > 0)
    let data = Object.groupBy(values, i => formatToName(i.EntryInfo.Format))

    let totals = Object.fromEntries(
        Object.entries(data)
            .map(([name, data]) => [name, data.reduce((p, c) => p + c.EntryInfo.PurchasePrice, 0)])
            .sort((a, b) => b[1] - a[1])
    )

    let labels = Object.keys(data)
    totals = Object.values(totals)

    new Chart(getCtx("cost-by-format"), {
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

function costByTypePie(json) {
    let values = Object.values(json)
        .filter(v => v.EntryInfo.PurchasePrice > 0)
    let data = Object.groupBy(values, i => i.EntryInfo.Type)

    let totals = Object.fromEntries(
        Object.entries(data)
            .map(([name, data]) => [name, data.reduce((p, c) => p + c.EntryInfo.PurchasePrice, 0)])
            .sort((a, b) => b[1] - a[1])
    )
    let labels = Object.keys(data)
    let totalList = Object.values(totals)

    let colors = labels.map(v => typeColors[v])

    new Chart(getCtx("cost-by-type"), {
        type: 'pie',
        data: {
            labels: labels,
            datasets: [{
                label: "Types",
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

function typePieChart(json) {
    let values = Object.values(json)
    let data = Object.groupBy(values, i => i.EntryInfo.Type)

    let labels = Object.keys(data)
        .sort((a, b) => data[b].length - data[a].length)
    let counts = Array.from(labels, (v, k) => data[v].length)

    let colors = labels.map(v => typeColors[v])
    console.log(colors)

    new Chart(typePieCtx, {
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

function ratingByYear(json) {
    let finishedValues = Object.values(json)
        .filter(v => v.UserInfo.Status == "Finished" && v.MetaInfo.ReleaseYear != 0 && v.EntryInfo.CopyOf == 0)
    let data = Object.groupBy(finishedValues, i => i.MetaInfo.ReleaseYear)

    let highestYear = Object.keys(data).sort((a, b) => b - a)[0]
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
            .map(i => i.UserInfo.UserRating)
            .reduce((p, c, i) => (p * i + c) / (i + 1), 0)
        )

    let chart = new Chart(rbyCtx, {
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

function byYearChart(json) {
    let finishedValues = Object.values(json)
        .filter(v => v.UserInfo.Status == "Finished" && v.MetaInfo.ReleaseYear != 0 && v.EntryInfo.CopyOf == 0)

    let data = Object.groupBy(finishedValues, i => i.MetaInfo.ReleaseYear)
    let highestYear = Object.keys(data).sort((a, b) => b - a)[0]
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
    const counts = Object.values(data).map(v => v.length)
    let total = counts.reduce((p, c) => p + c, 0)
    console.log(`total items by year: ${total}`)

    new Chart(ctx, {
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

/**
    * @param {object} json
    */
function makeGraphs(json) {
    //TODO:
    //have some way to let the user filter information in the json
    //then display the filtered json in the charts
    byYearChart(json)
    typePieChart(json)
    ratingByYear(json)
    costByTypePie(json)
    costByFormat(json)
    countByFormat(json)
}

fetch("http://localhost:8080/api/v1/list-tree")
    .then(res => res.json())
    .then(makeGraphs)
