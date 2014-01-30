pat = {}

pat.experiment = function(refreshRate) {

  function exports() {}

  exports.state = ko.observable("")
  exports.url = ko.observable("")
  exports.csvUrl = ko.observable("")
  exports.data = ko.observableArray()
  exports.config = { pushes: ko.observable(1), concurrency: ko.observable(1) }

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
    if ($("#cmdSelect").val() == "Simple Push") {
      $.post( "/experiments/", { "pushes": exports.config.pushes(), "concurrency": exports.config.concurrency() }, function(data) {
      exports.url(data.Location)
      exports.csvUrl(data.CsvLocation)
      exports.refreshNow()
      })
    } else {
      $.post( "/experiments/", { "pushes": exports.config.pushes(), "concurrency": exports.config.concurrency(), "workload": $("#cmdSelect").val() }, function(data) {
      exports.url(data.Location)
      exports.csvUrl(data.CsvLocation)
      exports.refreshNow()
      })
    }
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
    ko.bindingHandlers.chart.b = d3.custom.barchart(element);
  },
  update: function(element, valueAccessor) {
    var data = ko.unwrap(valueAccessor())
    data.forEach(function(obj) {
      for (k in obj) {
        if (k == "Average" || k == "WallTime" || k == "LastResult" || k == "TotalTime") obj[k + '_fmt'] = (obj[k] / 1000000000).toFixed(2) + " sec";
      }
    });
    ko.bindingHandlers.chart.b(data)
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
  this.numPushes = experiment.config.pushes
  this.numPushesHasError = ko.computed(function() { return experiment.config.pushes() <= 0 })
  this.numConcurrent = experiment.config.concurrency
  this.numConcurrentHasError = ko.computed(function() { return experiment.config.concurrency() <= 0 })
  this.formHasNoErrors = ko.computed(function() { return ! ( this.numPushesHasError() | this.numConcurrentHasError() ) }, this)
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
