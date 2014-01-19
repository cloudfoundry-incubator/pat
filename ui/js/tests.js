describe("The view", function() {
  var experiment = { run: function() {}, state: ko.observable(""), csvUrl: ko.observable("") }

  beforeEach(function() {
    spyOn(experiment, "run")
    v = new pat.view(experiment)
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

describe("Running an experiment", function my() {

  var replyUrl = "foo/bar/baz"

  describe("Calling the endpoint", function() {

    beforeEach(function() {
      spyOn($, "post").andCallFake(function(url, data, callback) { callback({ "Location": replyUrl }) })
      spyOn($, "get").andCallFake(function(url, callback) {  })
      var experiment = pat.experiment()
      experiment.run()
    })

    it("sends a POST to the /experiments/ endpoint", function() {
      expect($.post).toHaveBeenCalledWith("/experiments/", jasmine.any(Object), jasmine.any(Function))
    })

    it("sends a GET to the tracking URL", function() {
      expect($.get).toHaveBeenCalledWith(replyUrl, jasmine.any(Function))
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
