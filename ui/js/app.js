pat = {}

pat.experiment = function(refreshRate) {

  function exports() {}

  exports.state = ko.observable("")
  exports.url = ko.observable("")
  exports.csvUrl = ko.observable("")
  exports.data = ko.observableArray()
  exports.config = { iterations: ko.observable(1), concurrency: ko.observable(1), interval: ko.observable(0), stop: ko.observable(0) }

  var timer = null

  exports.refresh = function() {
    $.get(exports.url(), function(data) {
      exports.data(data.Items.filter(function(d) { return d.Type === 0 }))
      exports.waitAndRefreshOnce()
    })
  }

  exports.refreshNow = function() {
    if(timer) { clearTimeout(timer) }
    exports.refresh()
  }

  exports.waitAndRefreshOnce = function() {
    timer = setTimeout(exports.refresh, refreshRate)
  }

  exports.run = function() {
    exports.state("running")
    exports.data([])
		$.post( "/experiments/", { "iterations": exports.config.iterations(), "concurrency": exports.config.concurrency(), "interval": exports.config.interval(), "stop": exports.config.stop(),  "workload": $("#cmdSelect").val() }, function(data) {
			exports.url(data.Location)
			exports.csvUrl(data.CsvLocation)
			exports.refreshNow()
		})
  }

  exports.view = function(url) {
    exports.state("running")
    exports.url(url)
    exports.csvUrl("")
    exports.refreshNow()
  }

  exports.url.subscribe(function(u) {
    $(document).trigger("experimentChanged", u)
  })

  return exports
}

pat.experimentList = function() {
  var exports = {}

  var self = this
  var timer = null
  self.active = ko.observable()

  exports.experiments = ko.observable()
  exports.refresh = function() {
    $.get("/experiments/", function(data) {
      // fixme(jz) be better to do an append here, when server supports it
      data.Items.forEach(function(d) {
        d.active = ko.computed(function() { return self.active() == d.Location })
      })
      exports.experiments(data.Items.reverse())
      timer = setTimeout(exports.refresh, 1000 * 10)
    })
  }

  exports.refreshNow = function() {
    if(timer) { clearTimeout(timer) }
    exports.refresh()
  }

  $(document).on("experimentChanged", function(e, url) {
    self.active(url)
  })

  exports.refresh()

  return exports
}

ko.bindingHandlers.chart = {
  c: {},
  init: function(element, valueAccessor) {    
    ko.bindingHandlers.chart.b = new barchart(element);
    ko.bindingHandlers.chart.t = new throughput(element);
  },
  update: function(element, valueAccessor) {
    var data = ko.unwrap(valueAccessor())
    data.forEach(function(obj) {
      for (k in obj) {
        if (k == "Average" || k == "WallTime" || k == "LastResult" || k == "TotalTime") obj[k + '_fmt'] = (obj[k] / 1000000000).toFixed(2) + " sec";
      }
    });
    ko.bindingHandlers.chart.b(data);
    ko.bindingHandlers.chart.t(data);
    console.log(JSON.stringify(data))
  }
}

pat.view = function(experimentList, experiment) {
  var self = this

  this.redirectTo = function(location) { window.location = location }

  this.start = function() { experiment.run() }
  this.stop = function() { alert("Not implemented") }
  this.downloadCsv = function() { self.redirectTo(experiment.csvUrl()) }

  this.canStart = ko.computed(function() { return experiment.state() !== "running" })
  this.canStop = ko.computed(function() { return experiment.state() === "running" })
  this.canDownloadCsv = ko.computed(function() { return experiment.csvUrl() !== "" })
  this.noExperimentRunning = ko.computed(function() { return self.canStart() })
  this.numIterations = experiment.config.iterations
  this.numIterationsHasError = ko.computed(function() { return experiment.config.iterations() <= 0 })
  this.numConcurrent = experiment.config.concurrency
  this.numConcurrentHasError = ko.computed(function() { return experiment.config.concurrency() <= 0 })
  this.numInterval = experiment.config.interval
  this.numIntervalHasError = ko.computed(function() { return experiment.config.interval() < 0 })
  this.numStop = experiment.config.stop
  this.numStopHasError = ko.computed(function() { return experiment.config.stop() < 0 })
  this.formHasNoErrors = ko.computed(function() { return ! ( this.numIterationsHasError() | this.numConcurrentHasError() | this.numIntervalHasError() | this.numStopHasError() ) }, this)
  this.previousExperiments = experimentList.experiments
  this.data = experiment.data

  experiment.url.subscribe(function(url) {
    window.location.hash = "#" + url
  })

  experiment.state.subscribe(function() {
    experimentList.refreshNow()
  })

  this.onHashChange = function(hash) {
    if(hash.length > 1) {
      experiment.view(hash.slice(1));
    }
  }

  $(document).ready(function() { self.onHashChange(window.location.hash) })
  $(window).on('hashchange', function() { self.onHashChange(window.location.hash) })
}
