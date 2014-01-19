d3.custom = {}
d3.custom.pats = {}
d3.custom.pats.throughput = function my(selection) {
	var width = 800
	var height = 400
	var padding = 12
	var svg

	function exports(dataset) {
		if (!svg) {
			svg = selection.append("svg")
			.attr("width", width)
			.attr("height", height)
			.append("g")
		}

		var x = d3.scale.linear().range([width-padding, padding]);
		x.domain(d3.extent(dataset, function(d) { return d.WallTime; }));

		var y = d3.scale.linear().range([padding, height-padding]);
		y.domain(d3.extent(dataset, function(d) { return d.LastResult; }));

		var circles = svg.selectAll("circle")
		.data(dataset);

		circles.enter()
		.append("circle")
		.style("fill", "steelblue")
		.attr("r", 3);

		circles.exit().remove();

		circles.attr("cx", function(d) { return x(d.WallTime) })
		circles.attr("cy", function(d) { return y(d.LastResult) })
	}

	exports.width = function() { return width }
	exports.height = function() { return height }

	return exports
}
