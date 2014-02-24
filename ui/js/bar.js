d3_workload = function() {
  const barWidth = 30,
        clipMarginRight = 40;

  var margin = {top: 50, right: 30, bottom: 30, left: 30};
  var svgWidth, svgHeight, jqObj, d3Obj, drawArea, drawBoxWidth;
  var x, y, xAxis, yAxis, svg, outerBody, barCon;

  var d3Graph = document.createElement('div');
  d3Graph.width = "100%";
  d3Graph.height = "100%";

  var initDOM = function(el) {

    d3Obj = d3.select(el);
    jqObj = $(el);

    svgWidth = jqObj.width() - margin.left - margin.right;
    svgHeight = jqObj.height() - margin.top - margin.bottom;
    x = d3.scale.linear().range([0, svgWidth], 1),
    y = d3.scale.linear().range([0, svgHeight], 1);

    xAxis = d3.svg.axis()
      .scale(x)
      .orient("bottom")
      .ticks(0);
    yAxis = d3.svg.axis()
      .scale(y)
      .orient("right")
      .tickSize(-svgWidth + 30);
    zoom = d3.behavior.zoom()
      .x(x)
      .scaleExtent([1, 10])
      .on("zoom", function(){
        barCon.attr("transform", "translate(" + d3.event.translate[0] + ",0)scale(1, 1)");
      })

    el.appendChild(d3Graph);

    svg = d3.select(d3Graph)
      .append("svg")      
        .attr("width", jqObj.width())
        .attr("height", jqObj.height())
        .attr("class", "workload")
      .append("g")
        .attr("transform", "translate(" + margin.left + "," + margin.top + ")");

    outerBody = svg.append("g");

    drawBoxWidth = parseInt(svgWidth - clipMarginRight);
    svg.append("defs").append("clipPath")
      .attr("id", "workloadclip6235")
    .append("rect")
      .attr("x", 0)
      .attr("y", 0)
      .attr("width", drawBoxWidth)
      .attr("height", svgHeight + 30);

    var chartBody = svg.append("g")
      .attr("clip-path", "url(#workloadclip)")
      .call(zoom);

    chartBody.append("rect")
      .attr("width","100%")
      .attr("height","100%")
      .attr("style","fill:none;pointer-events: all;");

    outerBody.append("g")
      .attr("class", "x axis")
      .attr("transform", "translate(0," + svgHeight + ")");
    outerBody.append("g")
      .attr("class", "y axis")
      .attr("transform", "translate(" + (svgWidth - 20) + ", 0)");

    barCon = chartBody.append("g")    
      .attr("transform", "translate(0,0)");

    drawArea = barCon.node();

    svg.append("text")
      .attr("x", 20)
      .attr("y", -svgWidth - 10)
      .attr("transform", "rotate(90)")
      .text("Seconds");
    svg.append("text")
      .attr("x", 20)
      .attr("y", svgHeight + 25)
      .text("Task #")
      .attr("text-anchor","middle");
    svg.append("text")
      .attr("x", svgWidth / 2)
      .attr("y", -10)
      .text("Experiment Duration (seconds)")
      .attr("style", "text-anchor: middle; font-size: 15pt; fill: #888;");  

  } //end initDOM

  var drawGraph = function(data) {
    const second = 1000000000;
    var len = data.length;
    x = x.domain( [1, len, 1] ).range([barWidth + 1, len * (barWidth + 1)]);
    y = y.domain([d3.max(data, function(d) { return d.LastResult / second} ), 0] );
    var bars = barCon.selectAll("rect.bar").data(data);
    var labels = barCon.selectAll("text").data(data);

    panToViewable();

    bars.transition()
      .attr("x", function(d, i) { return x(i) })
      .attr("y", function(d) { return y(d.LastResult / second) } )
      .attr("height", function(d) { return (svgHeight - y(d.LastResult / second)) } );

    labels.transition()
      .attr("x", function(d, i) { return (x(i) + (barWidth / 2)) })
      .attr("y", svgHeight + 3 );

    bars.enter()
      .append("rect")
        .attr("x", function(d, i) { return x(i) + 10 })
        .attr("y", function(d) { return y(d.LastResult / second) } )
        .attr("height", function(d) { return (svgHeight - y(d.LastResult / second)) } )
        .attr("width", barWidth)
        .attr("class", "bar")
        .attr("data-shift", function(){
            var bodyWidth = len * (barWidth + 1);
            if (bodyWidth + getTranslateX(barCon) > drawBoxWidth) {
              transformChart(barCon, drawBoxWidth - bodyWidth  - 25, 0);
              zoom = adjustZoomX(zoom, drawBoxWidth - bodyWidth  - 25, bodyWidth);
            }
        })
      .transition()
        .duration(1000)
        .attr("x", function(d, i) { return x(i) })
        .attr("y", function(d) { return y(d.LastResult / second) })
        .attr("height", function(d) { return (svgHeight - y(d.LastResult / second) ) } );

    bars.enter()
      .append("text")
        .attr("x", function(d, i) { return (x(i) + (barWidth / 2)) })
        .attr("y", svgHeight + 20 )
        .attr("dy", ".7em")
        .text(function(d, i){ return i + 1 })
      .transition()
        .attr("y", svgHeight + 3 );

    outerBody.select(".x.axis").call(xAxis);
    outerBody.select(".y.axis").call(yAxis);

    hightlightErrors();
   
    bars.exit().remove();
    labels.exit().remove();

  } //end drawGraph

  function panToViewable() {
    var bodyWidth = getNodeWidth(barCon[0][0]);
    var barsPan = getTranslateX(barCon);

    if (bodyWidth + barsPan <= 0 && bodyWidth > 0) {
      transformChart(barCon, ( 0 - bodyWidth) + drawBoxWidth / 3, 0);
      zoom = adjustZoomX(zoom, ((0 - bodyWidth) + drawBoxWidth / 3), bodyWidth);
    } else if (barsPan > drawBoxWidth) {
      transformChart(barCon, drawBoxWidth - (drawBoxWidth / 3), 0);
      zoom = adjustZoomX(zoom, drawBoxWidth - (drawBoxWidth / 3), bodyWidth);
    }
  }

  function hightlightErrors() {
    var errorCount = 0;
    var bars = d3.select(drawArea).selectAll("rect.bar");
    bars.each(function(d,i) {
      if (d.TotalErrors > errorCount) {
        var bar = d3.select(this);     
        bar.classed("error", true)
          .attr("data-toggle","tooltip")
          .attr("title", "Error: " + d.LastError);
          
        errorCount += 1;
        $(this).tooltip({
          "placement": "top",
          "container": "body"
        });
      } else {
        d3.select(this).classed("error", false);
      }
    })
  }

  function getTranslateX(node) {
    var splitted = node.attr("transform").split(",");
    return parseInt(splitted [0].split("(")[1]);
  }

  function getNodeWidth(node) {
    return node.getBBox().width;
  }

  function transformChart(node, x, y) {
    node.transition()
      .attr("transform", "translate(" + x + ", " + y + ")" );
  }

  function adjustZoomX(zoom, x, range) {
    zoom.x(d3.scale.linear().range([1, range]));
    zoom.translate([x, 0]);
    return zoom;
  }

  var changeState = function(fn) {
    fn(d3Graph)
  }

  return {
    init: function(el){
      initDOM(el);
      return drawGraph;
    },
    node: d3Graph,
    changeState: changeState,
    totalBars: function() { return $(d3Graph).find("rect.bar").length },
    yAxisMax: function() { return yAxis.scale().domain()[0] },
    chartAreaWidth: function() { return drawBoxWidth },
    chartArea: function(){ return drawArea },
    display: function() { return $(d3Graph).css('display') },
  }

}()

