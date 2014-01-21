pat = {}

pat.experiment = function(refreshRate) {

	var infoUrl
	function exports() {}

	exports.state = ko.observable("")
	exports.csvUrl = ko.observable("")
	exports.data = ko.observableArray()
	exports.config = { pushes: ko.observable(1), concurrency: ko.observable(1) }

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
		$.post( "/experiments/", { "pushes": exports.config.pushes(), "concurrency": exports.config.concurrency() }, function(data) {
			infoUrl = data.Location
			exports.csvUrl(data.CsvLocation)
			exports.refresh()
		})
	}

	exports.view = function(url) {
		exports.state("running")
		infoUrl = url
		exports.csvUrl("")
		exports.refresh()
	}

	return exports
}

pat.experimentList = function() {
	var exports = {}

	exports.experiments = ko.observableArray()
	exports.refresh = function() {
		$.get("/experiments/", function(data) {
			exports.experiments.removeAll()
			data.Items.forEach(function(d) { exports.experiments.push(d) })
			setTimeout(exports.refresh, 1000 * 10)
		})
	}

	exports.refresh()

	return exports
}

ko.bindingHandlers.chart = {
  c: {},
  init: function(element, valueAccessor) {
    ko.bindingHandlers.chart.b = d3.custom.barchart(d3.select(element));
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

	experiment.state.subscribe(function() {
		experimentList.refresh()
	})

	this.onHashChange = function(hash) {
		if(hash.length > 1) {
			experiment.view(hash.slice(1));
		}
	}

	$(document).ready(function() { self.onHashChange(window.location.hash) })
	$(window).on('hashchange', function() { self.onHashChange(window.location.hash) })

}
