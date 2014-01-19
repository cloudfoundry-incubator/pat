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

    $('.barchart').html("");
    var svg = d3.select(".barchart");

    var h = $('.barchart').height();
    //var x = d3.scale.linear().domain([0, d3.max(data, function(d){return d.LastResult/1000000000})]).range([0, h]);
    var x = d3.scale.linear().domain([0, 6]).range([0, h]);

    data.forEach(function(d){
      svg.append("rect").attr("x",50 * d.Total).attr("y",h - x(d.LastResult / 1000000000)).attr("width",30).attr("height", x(d.LastResult / 1000000000));
      svg.append("text").attr("x", 50 * d.Total + 15 ).attr("y", h - 10 - x(d.LastResult / 1000000000)).attr("dy", ".75em").text((d.LastResult / 1000000000).toFixed(2) + " sec");
    });
  }

pat.view = function(experiment) {
  var self = this

  //Todo: move to bar.js - setup SVG to draw barchart
  d3.select("#graph").append("svg").attr("class","barchart").attr("width", $("#graph").width()).attr("height", $("#graph").height());

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
