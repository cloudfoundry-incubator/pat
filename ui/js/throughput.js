var throughput = function(el) {
 const d3_id = "d3_throughput",
       d3_container = d3_id + "_container";
   
  var self = this;      
  var margin = {top: 50, right: 30, bottom: 30, left: 30},
      d3Obj = d3.select(el),
      jqObj = $(el);

  this.barWidth = 30;
  this.svgWidth = jqObj.width() - margin.left - margin.right;
  this.svgHeight = jqObj.height() - margin.top - margin.bottom;
  this.x = d3.scale.linear().range([0, this.svgWidth], 1);
  this.y = d3.scale.linear().range([0, this.svgHeight], 1);
  this.color = d3.scale.category10();

  this.xAxis = d3.svg.axis()
    .scale(this.x)
    .orient("bottom")
    .tickSize(-this.svgHeight);  

  d3Obj.append("div")
    .attr("id", d3_container)
    .attr("width", "100%")
    .attr("height", "100%")
  .append("svg")
    .attr("id", d3_id)
    .attr("width", jqObj.width())
    .attr("height", jqObj.height());

  this.svg = d3.select("#" + d3_id)
    .append("g")
      .attr("transform", "translate(" + margin.left + "," + margin.top + ")");

  this.outerBody = this.svg.append("g");

  this.svg.append("defs").append("clipPath")
    .attr("id", d3_id + "clip")
  .append("rect")
    .attr("x", 0)
    .attr("y", 0)
    .attr("width", this.svgWidth)
    .attr("height", this.svgHeight);

  var chartBody = this.svg.append("g")
    .attr("clip-path", "url(#" + d3_id + "clip)");    
  chartBody.append("rect")
    .attr("width","100%")
    .attr("height","100%")
    .attr("style","fill:none;pointer-events: all;");

  this.outerBody.append("g")
    .attr("class", "x axis")
    .attr("transform", "translate(0," + this.svgHeight + ")");

  this.graphBox = chartBody.append("g")
    .attr("id", d3_id + "_box")
    .attr("transform", "translate(0,0)");  
  this.svg.append("text")
    .attr("x", this.svgWidth - 60)
    .attr("y", this.svgHeight + 25)
    .text("Commands / sec")
    .attr("text-anchor","middle");
  this.svg.append("text")
    .attr("x", this.svgWidth / 2)
    .attr("y", -20)
    .text("Commands Throughput / sec")
    .attr("style", "text-anchor: middle; font-size: 15pt; fill: #000;");

  var exports = function(data) {
    if (!data[0]) return;

    var cmds = flattenJSON(data);
    self.color.domain(cmds.map(function (d) { return d.cmd; }));

    var len = cmds.length;
    var y = self.y.domain( [1, len, 1] ).range([self.barWidth + 1, len * (self.barWidth + 1)]);
    var x = self.x.domain([0, d3.max(cmds, function(d) { return d.throughput } )]).range([0, self.svgWidth]);
    var bars = self.graphBox.selectAll("rect.bar").data(cmds);
    var labels = self.graphBox.selectAll("text").data(cmds);

    bars.transition()
      .attr("x", x(0))
      .attr("y", function(d, i) { return y(i) + 10 } )  
      .attr("width", function(d) { return x(d.throughput) } );

    labels.transition()
      .attr("x", function(d) { return x(d.throughput) - 20} )
      .attr("y", function(d, i) { return y(i) + (self.barWidth / 2 ) + 4 } )
      .text(function(d){ return (d.cmd + ": " + d.throughput.toFixed(3) + " / sec") });

    bars.enter()
      .append("rect")
        .attr("x", x(0) )  
        .attr("y", function(d, i) { return y(i) + 10 })        
        .attr("width", function(d) { return x(d.throughput) } )
        .attr("height", self.barWidth)
        .attr("class", "bar")
        .style("fill", function (d) { return self.color(d.cmd) })
      .transition()        
        .duration(1000)
        .attr("y", function(d, i) { return y(i) + 10 })        
        .attr("width", function(d) { return x(d.throughput) } )

    bars.enter()
      .append("text")
        .attr("x", function(d) { return x(d.throughput) -20 })
        .attr("y", function(d, i) { return y(i) + (self.barWidth / 2 ) + 4 } )
        .attr("dy", ".7em")
        .text(function(d){ return (d.cmd + ": " + d.throughput.toFixed(3) + " / sec") })
        .attr("style", "text-anchor: end; font-size: 12pt; fill: #fff")       

    self.outerBody.select(".x.axis").call(self.xAxis);   

    bars.exit().remove();   
    labels.exit().remove();       

  } //end exports

  exports.xAxisMax = function() { return self.xAxis.scale().domain()[1]; }  

  function flattenJSON(data) {
    var cmds = [];
    for (var command in data[data.length - 1].Commands) {
      cmds.push({"cmd": command, "throughput": data[data.length - 1].Commands[command].Throughput})
    }
    return cmds;
  }

  return exports;

}; //end throughput

