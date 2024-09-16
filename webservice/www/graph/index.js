function getCtx(id) {
    const canv = /**@type {HTMLCanvasElement}*/(document.getElementById(id))
    return canv.getContext("2d")
}

const ctx = getCtx("by-year")
const typePieCtx = getCtx("type-pie")
const rbyCtx = getCtx("rating-by-year")

const totalizer = {
    id: 'totalizer',

    beforeUpdate: function(chart) {
        chart.$totalizer = {
            totals: {},
        };
    },

    afterDatasetUpdate: (chart, { index: datasetIndex }) => {
        const dataset = chart.data.datasets[datasetIndex];
        const meta = chart.getDatasetMeta(datasetIndex);
        const total = dataset.data.reduce((total, value, index) => {
            return total + (meta.data[index].hidden ? 0 : value);
        }, 0);

        chart.$totalizer.totals[datasetIndex] = total;
    }
};

function fillGap(obj, label) {
    obj[label] = []
    if (!((Number(label) + 1) in obj)) {
        fillGap(obj, Number(label) + 1)
    }
}

function typePieChart(json) {
    let values = Object.values(json)
        .filter(v => v.UserInfo.Status == "Finished" || v.UserInfo.Status == "Viewing")
    let data = Object.groupBy(values, i => i.EntryInfo.Type)

    let labels = Object.keys(data)
        .sort((a, b) => data[b].length - data[a].length)
    let counts = Array.from(labels, (v, k) => data[v].length)

    new Chart(typePieCtx, {
        type: 'pie',
        data: {
            labels: labels,
            datasets: [{
                label: "Types",
                data: counts,
                borderWidth: 1
            }]
        },
        options: {
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

    new Chart(rbyCtx, {
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
        .filter(v => v.UserInfo.Status == "Finished" && v.MetaInfo.ReleaseYear != 0  && v.EntryInfo.CopyOf == 0)

    let data = Object.groupBy(finishedValues, i => i.MetaInfo.ReleaseYear)
    let highestYear = Object.keys(data).sort((a, b) => b - a)[0]
    for (let year in data) {
        let yearInt = Number(year)
        if (highestYear == yearInt) break
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
                }
            }
        }
    });
}

fetch("http://localhost:8080/api/v1/list-tree")
    .then(res => res.json())
    .then(json => {
        byYearChart(json)
        typePieChart(json)
        ratingByYear(json)
    })
