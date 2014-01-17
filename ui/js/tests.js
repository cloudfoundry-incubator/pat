
describe("throughput chart", function() {
  var chart
  beforeEach(function() {
    chart = d3.custom.pats.throughput()
  })

  it("should have a default width (300) and height (500)", function() {
    expect(chart.width()).toBe(300)
    expect(chart.height()).toBe(300)
  })

  it("should create a point for each element", function() {
    chart([1, 2, 3])
    expect(d3.selectAll('circle').size()).toBe(3)
  })
})
