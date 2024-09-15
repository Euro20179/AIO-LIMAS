const canv = /**@type {HTMLCanvasElement}*/(document.getElementById("graph"))
const ctx = canv.getContext("2d")

fetch("http://localhost:8080/api/v1/list-tree")
    .then(res => res.json())
    .then(json => {
        let finishedValues = Object.values(json)
            .filter(v => v.UserInfo.Status == "Finished")

        let data = Object.groupBy(finishedValues, i => i.MetaInfo.ReleaseYear)
        console.log(data)

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
