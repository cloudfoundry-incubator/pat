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
    ko.bindingHandlers.chart.c = d3.custom.pats.throughput(d3.select(element))
  },
  update: function(element, valueAccessor) {
    ko.bindingHandlers.chart.c(ko.unwrap(valueAccessor()))
  }
}

pat.view = function(experiment) {
  var self = this

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
