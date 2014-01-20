pat = {}

pat.experiment = function(refreshRate) {

  var infoUrl
  function exports() {}

  exports.state = ko.observable("")
  exports.csvUrl = ko.observable("")
  exports.data = ko.observableArray()

  exports.refresh = function() {
    $.get(infoUrl, function(data) {
      exports.data(data.Items.filter(function(d) { return d.Type === 0 }))
      exports.waitAndRefreshOnce()
    })
  }

  exports.waitAndRefreshOnce = function() {
    setTimeout(exports.refresh, refreshRate)
  }

  exports.run = function() {
    exports.state("running")
    $.post( "/experiments/", { "pushes": 10, "concurrency": 3 }, function(data) {
      infoUrl = data.Location
      exports.csvUrl(data.CsvLocation)
      exports.refresh()
    })
  }

  return exports
}

ko.bindingHandlers.chart = {
  c: {},
  init: function(element, valueAccessor) {
    //ko.bindingHandlers.chart.c = d3.custom.pats.throughput(d3.select(element))
    ko.bindingHandlers.chart.b = barchart;
  },
  update: function(element, valueAccessor) {
    //ko.bindingHandlers.chart.c(ko.unwrap(valueAccessor()))
    ko.bindingHandlers.chart.b(ko.unwrap(valueAccessor()))
  }
}

  //belong to bar.js
  function barchart(data) {
    if (data.length === 0) return;

    var xOffset = 50, yOffset = 50, barWidth = 30;
    var svg = d3.select(".barchart");
    var h = $('.barchart').height();
    //Bug(simon) Range is hard-coded for now
    var x = d3.scale.linear().domain([0, 10]).range([0, $('.barchart').width() - (xOffset*2)], 1);
    var y = d3.scale.linear().domain([5, 0]).range([0, h - (yOffset*2)]);

    var xAxis = d3.svg.axis().scale(x).orient("bottom");
    var yAxis = d3.svg.axis().scale(y).orient("left");
    svg.append("g").attr("class", "x axis").attr("transform", "translate(" + xOffset + "," + (h-yOffset) + ")").call(xAxis);
    svg.append("g").attr("class", "y axis").attr("transform", "translate(" + xOffset + ","+ yOffset + ")").call(yAxis);


    data.forEach(function(d){
      svg.append("rect").attr("x",x(d.Total) + xOffset - (barWidth/2)).attr("y",  y(d.LastResult / 1000000000) + yOffset ).attr("width", barWidth).attr("height", h - y(d.LastResult / 1000000000) - (yOffset * 2));
      svg.append("text").attr("x",x(d.Total) + xOffset ).attr("y", y(d.LastResult / 1000000000) + yOffset - 10).attr("dy", ".75em").text((d.LastResult / 1000000000).toFixed(2) + " sec");
    });
  }

pat.view = function(experiment) {
  var self = this

  //Todo: move to bar.js - setup SVG to draw barchart
  d3.select("#graph").append("svg").attr("class","barchart").attr("width", $("#graph").width()-20).attr("height", $("#graph").height()-20);

  this.redirectTo = function(location) { window.location = location }

  this.start = function() { experiment.run() }
  this.stop = function() { alert("Not implemented") }
  this.downloadCsv = function() { self.redirectTo(experiment.csvUrl()) }

  this.canStart = ko.computed(function() { return experiment.state() !== "running" })
  this.canStop = ko.computed(function() { return experiment.state() === "running" })
  this.canDownloadCsv = ko.computed(function() { return experiment.csvUrl() !== "" })
  this.noExperimentRunning = ko.computed(function() { return self.canStart() })
  this.data = experiment.data
}
