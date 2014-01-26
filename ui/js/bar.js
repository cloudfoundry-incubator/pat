d3.custom.barchart = function(el) {
  el.append("svg").attr("class","barchart").attr("width", $("#graph").width()-20).attr("height", $("#graph").height()-20);

  function barchart(data) {
    if (data.length === 0) return;
    const second = 1000000000;
    var xRange = 10;
    var yRange = 0;
    data.forEach(function(d) {
      if (d.LastResult > yRange) yRange = d.LastResult;
    });
    yRange = yRange / second;
    if (data.length > xRange) {
      xRange = data.length
    }
    $(".barchart").html("");
    var xOffset = 50, yOffset = 50, barWidth = 30;
    var h = $('.barchart').height();
    var svg = d3.select(".barchart");
    var x = d3.scale.linear().domain([0, xRange]).range([0, $('.barchart').width() - (xOffset*2)], 1);
    var y = d3.scale.linear().domain([yRange, 0]).range([0, h - (yOffset*2)]);
    var xAxis = d3.svg.axis().scale(x).orient("bottom");
    var yAxis = d3.svg.axis().scale(y).orient("left");
    svg.append("g").attr("class", "x axis").attr("transform", "translate(" + xOffset + "," + (h-yOffset) + ")").call(xAxis);
    svg.append("g").attr("class", "y axis").attr("transform", "translate(" + xOffset + ","+ yOffset + ")").call(yAxis);
    svg.append("text").attr("x",30).attr("y", 30).attr("dy", ".85em").text("Seconds");
    svg.append("text").attr("x",$('.barchart').width() - xOffset).attr("y", h - 20).attr("dy", ".85em").text("App Pushes");

		bars = svg.selectAll("rect.bar").data(data)
		bars.enter().append("rect")
			.attr("width", barWidth)
			.attr("class", "bar")
		bars.exit().remove()
		bars
			.attr("x", function(d) { return x(d.Total) + xOffset - (barWidth/2) })
			.attr("y", function(d) { return y(d.LastResult / second) + yOffset })
			.attr("height", function(d) { return h - y(d.LastResult / second) - (yOffset * 2) })

//    data.forEach(function(d){
//      svg.append("rect").attr("x",x(d.Total) + xOffset - (barWidth/2)).attr("y",  y(d.LastResult / 1000000000) + yOffset ).attr("width", barWidth)
//			.attr("height", h - y(d.LastResult / 1000000000) - (yOffset * 2)).attr("class","bar");
//      svg.append("text").attr("x",x(d.Total) + xOffset ).attr("y", y(d.LastResult / 1000000000) + yOffset - 10).attr("dy", ".75em").text((d.LastResult / 1000000000).toFixed(2) + " sec");
//    });
  }

  return barchart;
}
