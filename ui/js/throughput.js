d3_throughput = function() {
  var margin = {top: 50, right: 30, bottom: 30, left: 30};  
  var svgWidth, svgHeight, jqObj, d3Obj, drawArea;
  var x, y, xAxis, yAxis, svg, graphBox, color;

  var d3Graph = document.createElement('div');
  d3Graph.className = "throughputContainer";
  d3Graph.width = "100%";
  d3Graph.height = "100%";

  // size and draw svg elements onto DOM
  var initDOM = function(el) {
    var d3Obj = d3.select(el),
    jqObj = $(el);

    svgWidth = jqObj.width() - margin.left - margin.right;
    svgHeight = jqObj.height() - margin.top - margin.bottom;
    x = d3.scale.linear().range([0, svgWidth], 1);
    y = d3.scale.linear().domain([1,0]).range([10, svgHeight], 1);
    color = d3.scale.category10();

    xAxis = d3.svg.axis()
      .scale(x)
      .orient("bottom")
      .tickFormat(d3.format("d"))
      .tickSize(-svgHeight);  
    yAxis = d3.svg.axis()
      .scale(y)
      .orient("left")      
      .tickSize(-svgWidth);  

    el.appendChild(d3Graph);

    svg = d3.select(d3Graph)
      .append("svg")
        .attr("width", jqObj.width())
        .attr("height", jqObj.height())
        .attr("class", "throughput")
      .append("g")
        .attr("transform", "translate(" + margin.left + "," + margin.top + ")");

    svg.append("defs").append("clipPath")
      .attr("id", "throughputclip")
    .append("rect")
      .attr("x", 0)
      .attr("y", -3)
      .attr("width", svgWidth + 3)
      .attr("height", svgHeight + 3);

    svg.append("g")
      .attr("class", "x axis")
      .attr("transform", "translate(0," + svgHeight + ")")
      .call(xAxis);
    svg.append("g")
      .attr("class", "y axis")
      .call(yAxis)  

    var chartBody = svg.append("g")
      .attr("clip-path", "url(#" + "throughputclip)");    
    chartBody.append("rect")
      .attr("width","100%")
      .attr("height","100%")
      .attr("style","fill:none;pointer-events: all;");

    graphBox = chartBody.append("g")
      .attr("transform", "translate(0,0)");  
    svg.append("text")
      .attr("x", svgWidth - 15)
      .attr("y", svgHeight + 25)
      .text("Iterations")
      .attr("text-anchor","end");
    svg.append("text")
      .attr("x", 50)
      .attr("y", -svgWidth - 10)
      .attr("transform", "rotate(90)")
      .text("Command Throughput / sec")
    svg.append("text")
      .attr("x", svgWidth / 2)
      .attr("y", -10)
      .text("Command Throughput")
      .attr("style", "text-anchor: middle; font-size: 15pt; fill: #888;");        

    drawArea = graphBox.node();  
  } //end initDOM

  // functions to be called for graph plotting
  var drawGraph = function(data) {
    if (!data[0]) return;

    var commands = flattenJSON(data); 
    var lineChart = graphBox.selectAll("path.line").data(commands)
    var legend = graphBox.selectAll("g.tplegend").data(commands)

    color.domain(commands.map(function (d) { return d.cmd; }));
    
    y = y.domain( [findMaxThroughput(commands), 0] );
    x = x.domain([0, commands[0].throughput.length -1]).range([0, svgWidth]);
    svg.select(".x.axis").call(xAxis);
    svg.select(".y.axis").call(yAxis);

    var line = d3.svg.line()    
      .interpolate("monotone")
      .x(function(d, i) { return x(i) })
      .y(function(d) { return y(d) })    

    lineChart.transition()
      .attr("class", "line")
      .attr("d", function(d) {return line(d.throughput); })
      .style("stroke", function (d) { return color(d.cmd) })      
      
    lineChart.enter()
      .append("path")
        .attr("class", "line")
        .attr("d", function(d, i) {return line(d.throughput); })
        .style("stroke", function (d) { return color(d.cmd) })  
        .on("mouseover", function(d){ drawToolTip(d, color(d.cmd)) })      
        .on("mouseout", function(d) { svg.selectAll("g.data" + d.cmd.replace(/[^a-zA-Z0-9]/g,"_")).remove() })

    var l = legend.enter()
      .append("g")
        .attr("class", "tplegend")

    legend.transition()
      .select("g.tplegend text")
        .text(function(d){ return d.cmd }) 

    l.append("rect")
      .attr("x", 30)
      .attr("y", function(d, i) { return i * 15 + 2} )
      .attr("height", 10)
      .attr("width", 55)
      .style("fill", function(d) { return color(d.cmd) })
    l.append("text")
      .attr("x", 90 )
      .attr("y", function(d, i) { return i * 15 + 3 } )
      .attr("dy", ".7em")
      .style("stroke", function(d) { return color(d.cmd) })
      .attr("stroke-width", "1px")
      .attr("style", "text-anchor: start;")
      .text(function(d){ return d.cmd }) 

    lineChart.exit().remove();
    legend.exit().remove();    

    function flattenJSON(data) {
      var list = [];
      var throughput
      for (var k in data[0].Commands) {
        throughput = [0];
        data.forEach(function(d){
          throughput.push(d.Commands[k].Throughput)
        })
        list.push({"cmd": k, "throughput": throughput})
      }
        
      return list
    }  

    function findMaxThroughput(data) {
      var max = 0;
      data.forEach(function(row){
        row.throughput.forEach(function(d){
          if (d > max) max = d
        })
      })
      return max
    }

  } //end drawGraph

  function drawToolTip(d, c) {
    const pointRadis = 13;
    const yOffset = 4;
    var className = d.cmd.replace(/[^a-zA-Z0-9]/g,"_");
    var tooptip = svg.selectAll("g.data" + className).data(d.throughput).enter()
      .append("g")
      .attr("class", "data" + className)
    tooptip.append("circle")        
        .style("fill", c)
        .attr("class", className)
        .attr("cx", function(d, i){ return x(i) })
        .attr("cy", function(d){ return y(d) })
        .attr("r", pointRadis);
    tooptip.append("text")
        .attr("x", function(d,i){ return x(i) })
        .attr("y", function(d){ return y(d) + yOffset })
        .style("fill", "white")
        .text(function(d){ return parseFloat(d).toFixed(2) })
  }

  var changeState = function(fn) { 
    fn(d3Graph) 
  }

  // Public Properties
  return {        
    init: function(el){ //draw DOM elements and return function 'drawGraph' for graph drawing
      initDOM(el);
      return drawGraph;
    },
    changeState: changeState    
  } //end return

}()