describe("The view", function() {
	var experiment
	var experimentList

	beforeEach(function() {
		experiment = { run: function() {}, url: ko.observable(""), state: ko.observable(""), view: function() {}, csvUrl: ko.observable(""), config: { pushes: ko.observable(1), concurrency: ko.observable(1)  } }
		experimentList = { experiments: [], refresh: function(){} }
		spyOn(experimentList, "refresh")
		spyOn(experiment, "view")
		spyOn(experiment, "run")
		v = new pat.view(experimentList, experiment)
		spyOn(v, "redirectTo").andReturn()
		v.start()
	})

	describe("clicking start", function() {
		it("runs the experiment", function() {
			expect(experiment.run).toHaveBeenCalled()
		})
	})

	describe("when the state of the experiment changes to running", function() {
		beforeEach(function() { experiment.state("running") })

		it("sets canStart to false", function() {
			expect(v.canStart()).toBe(false)
		})

		it("sets canStop to true", function() {
			expect(v.canStop()).toBe(true)
		})

		it("sets noExperimentRunning to false", function() {
			expect(v.noExperimentRunning()).toBe(false)
		})

		it("refreshes the experiments list", function() {
			expect(experimentList.refresh).toHaveBeenCalled()
		})
	})

	describe("validation", function() {
		it("prevents pushes being <= 0", function() {
			v.numPushes(-1)
			v.numConcurrent(1)
			expect(v.numPushesHasError()).toBe(true)
			expect(v.numConcurrentHasError()).toBe(false)
			expect(v.formHasNoErrors()).toBe(false)
		})

		it("prevents concurrency being <= 0", function() {
			v.numConcurrent(-1)
			v.numPushes(1)
			expect(v.numPushesHasError()).toBe(false)
			expect(v.numConcurrentHasError()).toBe(true)
			expect(v.formHasNoErrors()).toBe(false)
		})
	})

	describe("hash urls", function() {
		it("does nothing if the hash is empty", function() {
			v.onHashChange("#")
			expect(experiment.view).not.toHaveBeenCalledWith()
		})

		it("views an experiment when the url hash changes", function() {
			v.onHashChange("#foo.csv")
			expect(experiment.view).toHaveBeenCalledWith("foo.csv")
		})
	})

	describe("when the experiment has an associated CSV URL", function() {
		beforeEach(function() { experiment.csvUrl("some-url.csv") })

		it("sets canDownloadCsv to true", function() {
			expect(v.canDownloadCsv()).toBe(true)
		})

		describe("clicking downloadCsv", function() {
			it("redirects to the csv URL", function() {
				v.downloadCsv()
				expect(v.redirectTo).toHaveBeenCalledWith("some-url.csv")
			})
		})
	})
})

describe("Throughput chart", function() {
	var chart

	beforeEach(function() {
		chart = d3.custom.pats.throughput(d3.select("#target"))
	})

	it("should have a default width (300) and height (500)", function() {
		expect(chart.width()).toBe(800)
		expect(chart.height()).toBe(400)
	})

	it("should create a point for each element", function() {
		chart([1, 2, 3])
		expect(d3.selectAll('circle').size()).toBe(3)
	})
})

describe("The experiment list", function() {

	var self = this
	var experiments
	var list

	describe("refreshing", function() {
		beforeEach(function() {
			self.experiments = [ { name: "a" }, { name: "b" } ]
			spyOn($, "get").andCallFake(function(url, callback) { console.log("ex", self.experiments); callback({ "Items": self.experiments }) })
			list = pat.experimentList()
		})

		it("adds all the items to the experiments array", function() {
			self.experiments = [1, 2, 3]
			list.refresh()
			expect(list.experiments()).toEqual(self.experiments)
		})

		it("refreshes on startup", function() {
			expect(list.experiments()).toEqual(self.experiments)
		})
	})
})

describe("Running an experiment", function my() {

	var replyUrl = "foo/bar/baz"

	describe("Calling the endpoint", function() {

		var pushes = 3
		var concurrency = 5
		var experiment

		beforeEach(function() {
			spyOn($, "post").andCallFake(function(url, data, callback) { callback({ "Location": replyUrl }) })
			spyOn($, "get").andCallFake(function(url, callback) {  })
			experiment = pat.experiment()
			experiment.config.pushes(pushes)
			experiment.config.concurrency(concurrency)
			experiment.data([1,2,3])
			experiment.run()
		})

		it("sends a POST to the /experiments/ endpoint", function() {
			expect($.post).toHaveBeenCalledWith("/experiments/", jasmine.any(Object), jasmine.any(Function))
		})

		it("sends the pushes and concurrency in the POST body", function() {
			expect($.post.mostRecentCall.args[1].pushes).toBe(3)
			expect($.post.mostRecentCall.args[1].concurrency).toBe(5)
		})

		it("sends a GET to the tracking URL", function() {
			expect($.get).toHaveBeenCalledWith(replyUrl, jasmine.any(Function))
		})

		it("clears any existing data", function() {
			expect(experiment.data().length).toEqual(0)
		})
	})

	describe("When results are returned", function() {

		var refreshRate = 800
		var csvUrl   = "foo/bar/baz.csv"
		var experiment
		var results

		beforeEach(function() {
			a = {"Type": 0, "name": "a"}
			b = {"Type": 1, "name": "b"}
			spyOn($, "post").andCallFake(function(url, data, callback) { callback({ "Location": replyUrl, "CsvLocation": csvUrl }) })
			spyOn($, "get").andCallFake(function(url, callback) {
				callback({ "Items": [a,b] })
			})

			experiment = pat.experiment(refreshRate)
			spyOn(experiment, "refresh").andCallThrough()
			spyOn(experiment, "waitAndRefreshOnce") //mocked because jasmine.Clock was being painful
			experiment.run()
		})

		it("updates with the latest data", function() {
			expect(experiment.data()).toEqual([a])
		})

		it("only includes data of type 0 (ResultSample)", function() {
			expect(experiment.data()).not.toBe([a, b])
		})

		it("refreshes the data at the refresh rate after data is returned", function() {
			expect(experiment.waitAndRefreshOnce.callCount).toEqual(1)
			$.get.mostRecentCall.args[1]({"Items": [{"Type": 0}]})
			expect(experiment.waitAndRefreshOnce.callCount).toEqual(2)
		})

		it("updates the csv url", function() {
			expect(experiment.csvUrl()).toBe(csvUrl)
		})

		it("updates the state to 'running'", function() {
			expect(experiment.state()).toBe("running")
		})
	})
})
