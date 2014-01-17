d3.custom = {}
d3.custom.pats = {}
d3.custom.pats.throughput = function my() {
  var width = 300
  var height = 300
  var svg

  function exports(dataset) {
    if (!svg) {
      svg = d3.select("body").append("svg")
      .attr("width", width)
      .attr("height", height)
      .append("g")
    };

    svg.selectAll("circle")
      .data(dataset)
      .enter().append("circle")
  };
  exports.width = function() { return width }
  exports.height = function() { return height }

  return exports
}
