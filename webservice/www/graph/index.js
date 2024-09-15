const canv = /**@type {HTMLCanvasElement}*/(document.getElementById("graph"))
const ctx = canv.getContext("2d")

function fillGap(obj, label) {
    obj[label] = []
    if(!((Number(label) + 1) in obj)) {
        fillGap(obj, Number(label) + 1)
    }
}

fetch("http://localhost:8080/api/v1/list-tree")
    .then(res => res.json())
    .then(json => {
        let finishedValues = Object.values(json)
            .filter(v => v.UserInfo.Status == "Finished" && v.MetaInfo.ReleaseYear != 0)

        let data = Object.groupBy(finishedValues, i => i.MetaInfo.ReleaseYear)
        let highestYear = Object.keys(data).sort((a, b) => b - a)[0]
        for(let year in data) {
            let yearInt = Number(year)
            if(highestYear == yearInt) break
            if(!((yearInt + 1) in data)) {
                fillGap(data, yearInt + 1)
            }
        }

        delete data["0"]

        const years = Object.keys(data)
        const counts = Object.values(data).map(v => v.length)

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
                scales: {
                    y: {
                        beginAtZero: true
                    }
                }
            }
        });

    })
