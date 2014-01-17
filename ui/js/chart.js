d3.custom = {}
d3.custom.pats = {}
d3.custom.pats.throughput = function my(selection) {
  var width = 300
  var height = 300
  var svg

  function exports(dataset) {
    if (!svg) {
      svg = selection.append("svg")
      .attr("width", width)
      .attr("height", height)
      .append("g")
    }

    svg.selectAll("circle")
    .data(dataset)
    .enter()
      .append("circle")
        .style("fill", "steelblue")
        .attr("cx", function(data, index) { return index })
        .attr("cy", function(data) { return data.TotalTime })
        .attr("r", 3);
  }

  exports.width = function() { return width }
  exports.height = function() { return height }

  return exports
}
