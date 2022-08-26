
$( document ).ready(function(){

    var colorHash = new ColorHash();
    var xlabel = [];
    for (let i=0; i<54; i+=1){ xlabel.push(i); };

    var myChart;

    async function drawChart(elementId, title, xlabel, datasets) {
        var ctx = document.getElementById(elementId).getContext('2d');
        myChart = new Chart(ctx, {
            type: 'line',
            data: {
                labels: [...xlabel],
                datasets: datasets
            },
            options: {
                title: {display: true, text: title, position: "left"},
                hover: {mode: 'nearest', intersect: false},
                tooltips: {mode: 'nearest', intersect: false},
                elements: {point: {radius: 1}},
                maintainAspectRatio: false,
                responsive: true,
                scales: {yAxes: [{ticks: {beginAtZero: true, precision: 0}}]}
            }
        });
        myChart.canvas.parentNode.style.height = '180px';
    }

    async function getDatasets(url){
        const response = await fetch(url, {headers: {'Content-Type': 'application/json'}});
        const userData = await response.json();

        var datasets = []

        Object.keys(userData).forEach(function(user, idx) {
            let dataset = {
                label: user,
                data: userData[user],
                fill: false,
                backgroundColor: colorHash.hex(user),
                borderColor: colorHash.hex(user),
                borderWidth: 1,
            }
            datasets.push(dataset)
        });
        return datasets
    }


    async function createWeeklyDistancesChart() {
        const datasets = await getDatasets("/dashboard/weekly/distances")
        drawChart('distance-chart', "Weekly Distance (Km)", xlabel, datasets)
    };

    async function createCumulativeDistancesChart() {
        const datasets = await getDatasets("/dashboard/weekly/cumulative-distances")
        drawChart('cumulative-distance-chart', "Overall Distance (Km)", xlabel, datasets)
    };

    async function createWeeklyCountsChart() {
        const datasets = await getDatasets("/dashboard/weekly/counts")
        drawChart('counts-chart', "Weekly Run Activities", xlabel, datasets)
    };


    async function createCumulativeCountsChart() {
        const datasets = await getDatasets("/dashboard/weekly/cumulative-counts")
        drawChart('cumulative-counts-chart', "Overall Run Activities", xlabel, datasets)
    };

    createWeeklyDistancesChart();
    createCumulativeDistancesChart();
    createWeeklyCountsChart();
    createCumulativeCountsChart();
});


async function fillHtmlDiv(selector,spinnerSelector, url) {
    const response = await fetch(url, {headers: {'Content-Type': 'text/html'}});
    const body = await response.text();
    $(spinnerSelector).addClass("d-none");
    $(selector).html(body);
}

$("#nav-other-top-tab").on("shown.bs.tab", function (e) {
    // Fetch if div is empty
    if ($("#other-tops-content").html() == ""){
        fillHtmlDiv("#other-tops-content", "#nav-other-top-spinner", "/dashboard/other-tops")
    }
});