d3.custom.barchart = function(el) {
  var d3Obj = d3.select("#" + el);
  var jqObj = $("#" + el);
  var xAxis, yAxis
  d3Obj.append("svg").attr("class","barchart").attr("width", jqObj.width() - 20).attr("height", jqObj.height() - 20);

  function barchart(data) {
    if (data.length === 0) return;
    const second = 1000000000;
    var xRange = 10;
    var yRange = 0;
    $(".barchart").html("");

    data.forEach(function(d) {
      if (d.LastResult > yRange) yRange = d.LastResult;
    });
    yRange = yRange / second;
    if (data.length > xRange) {
      xRange = data.length
    }

    var xOffset = 50, yOffset = 50, barWidth = 30;
    var h = $('.barchart').height();

    var x = d3.scale.linear().domain([0, xRange]).range([0, $('.barchart').width() - (xOffset*2)], 1);
    var y = d3.scale.linear().domain([yRange, 0]).range([0, h - (yOffset*2)]);
    xAxis = d3.svg.axis().scale(x).orient("bottom");
    yAxis = d3.svg.axis().scale(y).orient("left");

    var svg = d3.select(".barchart");
    svg.append("g").attr("class", "x axis").attr("transform", "translate(" + xOffset + "," + (h-yOffset) + ")").call(xAxis);
    svg.append("g").attr("class", "y axis").attr("transform", "translate(" + xOffset + ","+ yOffset + ")").call(yAxis);
    svg.append("text").attr("x",30).attr("y", 30).attr("dy", ".85em").text("Seconds");
    svg.append("text").attr("x",$('.barchart').width() - xOffset).attr("y", h - 20).attr("dy", ".85em").text("App Pushes");

    var bars = svg.selectAll("rect.bar").data(data)
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
//        .attr("height", h - y(d.LastResult / 1000000000) - (yOffset * 2)).attr("class","bar");
//      svg.append("text").attr("x",x(d.Total) + xOffset ).attr("y", y(d.LastResult / 1000000000) + yOffset - 10).attr("dy", ".75em").text((d.LastResult / 1000000000).toFixed(2) + " sec");
//    });
  }

  barchart.xAxis_max = function() { return xAxis.scale().domain()[1]; }
  barchart.yAxis_max = function() { return yAxis.scale().domain()[0]; }

  return barchart;
}
