d3_throughput = function() {
  const barWidth = 30;
  var margin = {top: 50, right: 30, bottom: 30, left: 30};  
  var svgWidth, svgHeight, jqObj, d3Obj, drawArea;
  var x, y, xAxis, yAxis, svg, outerBody, graphBox;

  var d3Graph = document.createElement('div');
  d3Graph.width = "100%";
  d3Graph.height = "100%";

  // size and draw svg elements onto DOM
  var initDOM = function(el) {

    var d3Obj = d3.select(el),
    jqObj = $(el);

    svgWidth = jqObj.width() - margin.left - margin.right;
    svgHeight = jqObj.height() - margin.top - margin.bottom;
    x = d3.scale.linear().range([0, svgWidth], 1);
    y = d3.scale.linear().range([0, svgHeight], 1);
    color = d3.scale.category10();

    xAxis = d3.svg.axis()
      .scale(x)
      .orient("bottom")
      .tickSize(-svgHeight);  

    el.appendChild(d3Graph);

    svg = d3.select(d3Graph)
      .append("svg")
        .attr("width", jqObj.width())
        .attr("height", jqObj.height())
        .attr("class", "throughput")
      .append("g")
        .attr("transform", "translate(" + margin.left + "," + margin.top + ")");

    outerBody = svg.append("g");

    svg.append("defs").append("clipPath")
      .attr("id", "throughputclip2132")
    .append("rect")
      .attr("x", 0)
      .attr("y", 0)
      .attr("width", svgWidth)
      .attr("height", svgHeight);

    var chartBody = svg.append("g")
      .attr("clip-path", "url(#" + "throughputclip");    
    chartBody.append("rect")
      .attr("width","100%")
      .attr("height","100%")
      .attr("style","fill:none;pointer-events: all;");

    outerBody.append("g")
      .attr("class", "x axis")
      .attr("transform", "translate(0," + svgHeight + ")");

    graphBox = chartBody.append("g")
      .attr("transform", "translate(0,0)");  
    svg.append("text")
      .attr("x", svgWidth - 60)
      .attr("y", svgHeight + 25)
      .text("throughput / sec")
      .attr("text-anchor","middle");
    svg.append("text")
      .attr("x", svgWidth / 2)
      .attr("y", -20)
      .text("Commands Throughput / sec")
      .attr("style", "text-anchor: middle; font-size: 15pt; fill: #888;");   

    drawArea = graphBox.node();  
  } //end initDOM

  // functions to be called for graph plotting
  var drawGraph = function(data) {
    if (!data[0]) return;

    var cmds = flattenJSON(data);
    color.domain(cmds.map(function (d) { return d.cmd; }));

    var len = cmds.length;
    y = y.domain( [1, len, 1] ).range([barWidth + 1, len * (barWidth + 1)]);
    x = x.domain([0, d3.max(cmds, function(d) { return d.throughput } )]).range([0, svgWidth]);
    var bars = graphBox.selectAll("rect.bar").data(cmds);
    var labels = graphBox.selectAll("text").data(cmds);

    bars.transition()
      .attr("x", x(0))
      .attr("y", function(d, i) { return y(i) + 10 } )  
      .attr("width", function(d) { return x(d.throughput) } );

    labels.transition()
      .attr("x", function(d) { return x(d.throughput) - 20} )
      .attr("y", function(d, i) { return y(i) + (barWidth / 2 ) + 4 } )
      .text(function(d){ return (d.cmd + ": " + d.throughput.toFixed(3) + " / sec") });

    bars.enter()
      .append("rect")
        .attr("x", x(0) )  
        .attr("y", function(d, i) { return y(i) + 10 })        
        .attr("width", function(d) { return x(d.throughput) } )
        .attr("height", barWidth)
        .attr("class", "bar")
        .style("fill", function (d) { return color(d.cmd) })
      .transition()        
        .duration(1000)
        .attr("y", function(d, i) { return y(i) + 10 })        
        .attr("width", function(d) { return x(d.throughput) } )

    bars.enter()
      .append("text")
        .attr("x", function(d) { return x(d.throughput) -20 })
        .attr("y", function(d, i) { return y(i) + (barWidth / 2 ) + 4 } )
        .attr("dy", ".7em")
        .text(function(d){ return (d.cmd + ": " + d.throughput.toFixed(3) + " / sec") })
        .attr("style", "text-anchor: end; font-size: 12pt; fill: #fff")       

    outerBody.select(".x.axis").call(xAxis);   

    bars.exit().remove();   
    labels.exit().remove();     

    function flattenJSON(data) {
      var cmds = [];
      for (var command in data[data.length - 1].Commands) {
        cmds.push({"cmd": command, "throughput": data[data.length - 1].Commands[command].Throughput})
      }
      return cmds;
    }  

  } //end drawGraph

  var changeState = function(fn) { 
    fn(d3Graph) 
  }

  // Public Properties
  return {        
    init: function(el){ //draw DOM elements and return function 'drawGraph' for graph drawing
      initDOM(el);
      return drawGraph;
    },
    changeState: changeState,
    totalBars: function() { return $(d3Graph).find("rect.bar").length },
    xAxisMax: function() { return xAxis.scale().domain()[1] },
    display: function() { return $(d3Graph).css('display') },
  } //end return

}()


