
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

    var data = { update: function() {} }
    var obj = { onCsvCallback: function() {}, onStateChangedCallback: function() {} }
    var refreshRate = 800
    var csvUrl   = "foo/bar/baz.csv"

    beforeEach(function() {
      jasmine.Clock.useMock();

      a = {"Type": 0, "name": "a"}
      b = {"Type": 1, "name": "b"}
      spyOn($, "post").andCallFake(function(url, data, callback) { callback({ "Location": replyUrl, "CsvLocation": csvUrl }) })
      spyOn($, "get").andCallFake(function(url, callback) {callback({ "Items": [a, b] }) })
      spyOn(obj, "onCsvCallback")
      spyOn(obj, "onStateChangedCallback")
      spyOn(data, "update")

      var experiment = pat.experiment(data.update, refreshRate)
      experiment.onCsvUrlChanged = obj.onCsvCallback
      experiment.onStateChanged = obj.onStateChangedCallback

      experiment.run()
    })

    it("calls the onData function", function() {
      expect(data.update).toHaveBeenCalledWith([a])
    })

    it("only calls with Type = result", function() {
      expect(data.update).not.toHaveBeenCalledWith([a, b])
    })

    it("refreshes the data at the refresh rate", function() {
      jasmine.Clock.tick(refreshRate + 1)
      expect(data.update.calls.length).toBe(2)
      jasmine.Clock.tick(refreshRate + 1)
      expect(data.update.calls.length).toBe(3)
    })

    it("calls the onCsvUrlChanged callback", function() {
      expect(obj.onCsvCallback).toHaveBeenCalledWith(csvUrl)
    })

    it("calls onStateChanged callback", function() {
      expect(obj.onStateChangedCallback).toHaveBeenCalledWith("running")
    })
  })
})
