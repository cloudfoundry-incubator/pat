pat = {}

pat.experiment = function(onDataCallbacks, refreshRate) {

  var infoUrl
  function exports() {}

  exports.state = ko.observable("")
  exports.csvUrl = ko.observable("")

  exports.refresh = function() {
    $.get(infoUrl, function(data) {
      onDataCallbacks.forEach(function(cb) {
        cb(data.Items.filter(function(d) { return d.Type === 0 }))
      })
      setTimeout(exports.refresh, refreshRate)
    })
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

pat.table = function(nodes) {
  return function table(data) {
    tr = nodes.selectAll("tr").data(data.filter(function(d) { return d.Type === 0 })).enter().append("tr")
    tr.append("td").text(function(d) { return d.WallTime })
    tr.append("td").text(function(d) { return d.LastResult })
    tr.append("td").text(function(d) { return d.Average })
    tr.append("td").text(function(d) { return d.TotalTime })
  }
}

pat.view = function(experiment) {
  var self = this
  var chart = d3.custom.pats.throughput(d3.select("#graph"))

  this.redirectTo = function(location) { window.location = location }

  this.start = function() { experiment.run() }
  this.stop = function() { alert("Not implemented") }
  this.downloadCsv = function() { self.redirectTo(experiment.csvUrl()) }

  this.canStart = ko.computed(function() { return experiment.state() !== "running" })
  this.canStop = ko.computed(function() { return experiment.state() === "running" })
  this.canDownloadCsv = ko.computed(function() { return experiment.csvUrl() !== "" })
  this.noExperimentRunning = ko.computed(function() { return self.canStart() })
}
