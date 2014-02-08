var barchart = function(el) {
  const d3_id = "d3_workload",
        d3_container = "d3_workload_container";
        
  var margin = {top: 30, right: 30, bottom: 30, left: 30},
      d3Obj = d3.select(el),
      jqObj = $(el);

  self = this;
  this.barWidth = 30;
  this.clipMarginRight = 40;
  this.svgWidth = jqObj.width() - margin.left - margin.right;
  this.svgHeight = jqObj.height() - margin.top - margin.bottom;
  this.x = d3.scale.linear().range([0, this.svgWidth], 1),
  this.y = d3.scale.linear().range([0, this.svgHeight], 1);

  this.xAxis = d3.svg.axis()
    .scale(this.x)
    .orient("bottom")
    .ticks(0);
  this.yAxis = d3.svg.axis()
    .scale(this.y)
    .orient("right")
    .tickSize(-this.svgWidth + 30);  
  this.zoom = d3.behavior.zoom()
    .x(this.x)
    .scaleExtent([1, 10])
    .on("zoom", function(){
      self.barCon.attr("transform", "translate(" + d3.event.translate[0] + ",0)scale(1, 1)");
    })

  d3Obj.append("div")
    .attr("id",d3_container)
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

  this.drawBoxWidth = parseInt(this.svgWidth - this.clipMarginRight);
  this.svg.append("defs").append("clipPath")
    .attr("id", "clip")
  .append("rect")
    .attr("x", 0)
    .attr("y", 0)
    .attr("width", this.drawBoxWidth)
    .attr("height", this.svgHeight + 30);

  var chartBody = this.svg.append("g")
    .attr("clip-path", "url(#clip)")
    .call(this.zoom);

  chartBody.append("rect")
    .attr("width","100%")
    .attr("height","100%")
    .attr("style","fill:none;pointer-events: all;");

  this.outerBody.append("g")
    .attr("class", "x axis")
    .attr("transform", "translate(0," + this.svgHeight + ")");
  this.outerBody.append("g")  
    .attr("class", "y axis")
    .attr("transform", "translate(" + (this.svgWidth - 20) + ", 0)");

  this.barCon = chartBody.append("g")
    .attr("id", "chart_box")
    .attr("transform", "translate(0,0)");  

  this.svg.append("text")
    .attr("x", -20)
    .attr("y", -10)
    .attr("transform", "rotate(-90)")
    .text("Seconds");
  this.svg.append("text")
    .attr("x", this.svgWidth - 60)
    .attr("y", this.svgHeight + 25)
    .text("Task #")
    .attr("text-anchor","middle");

  var exports = function(data) {
    const second = 1000000000;
    var len = data.length;
    var x = self.x.domain( [1, len, 1] ).range([self.barWidth + 1, len * (self.barWidth + 1)]);      
    var y = self.y.domain([d3.max(data, function(d) { return d.LastResult / second} ), 0] );
    var bars = self.barCon.selectAll("rect.bar").data(data);
    var labels = self.barCon.selectAll("text").data(data);

    panToViewable();
    
    bars.transition()
      .attr("x", function(d, i) { return x(i) })
      .attr("y", function(d) { return y(d.LastResult / second) } )  
      .attr("height", function(d) { return (self.svgHeight - y(d.LastResult / second)) } );

    labels.transition()
      .attr("x", function(d, i) { return (x(i) + (self.barWidth / 2)) })
      .attr("y", self.svgHeight + 3 );

    bars.enter()
      .append("rect")
        .attr("x", function(d, i) { return x(i) + 10 })
        .attr("y", function(d) { return y(d.LastResult / second) } )  
        .attr("height", function(d) { return (self.svgHeight - y(d.LastResult / second)) } )
        .attr("width", self.barWidth)
        .attr("class", "bar")
        .attr("data-shift", function(){ 
            var bodyWidth = len * (self.barWidth + 1);
            if (bodyWidth + getTranslateX(self.barCon) > self.drawBoxWidth) {    
              transformChart(self.barCon, self.drawBoxWidth - bodyWidth  - 25, 0);
              self.zoom = adjustZoomX(self.zoom, self.drawBoxWidth - bodyWidth  - 25, bodyWidth);
            }
        })
      .transition()        
        .duration(1000)
        .attr("x", function(d, i) { return x(i) })
        .attr("y", function(d) { return y(d.LastResult / second) })     
        .attr("height", function(d) { return (self.svgHeight - y(d.LastResult / second) ) } );

    bars.enter()
      .append("text")
        .attr("x", function(d, i) { return (x(i) + (self.barWidth / 2)) })
        .attr("y", self.svgHeight + 20 )
        .attr("dy", ".7em")
        .text(function(d, i){ return i + 1 })
      .transition()
        .attr("y", self.svgHeight + 3 );

    self.outerBody.select(".x.axis").call(self.xAxis);    
    self.outerBody.select(".y.axis").call(self.yAxis);

    bars.exit().remove();   
    labels.exit().remove();      

  } //end exports

  exports.yAxisMax = function() { return self.yAxis.scale().domain()[0]; };
  exports.xAxisMax = function() { return self.xAxis.scale().domain()[0]; }
  exports.drawBoxWidth = function() { return self.drawBoxWidth; }

  function panToViewable() {
    var bodyWidth = getNodeWidth(self.barCon[0][0]);
    var barsPan = getTranslateX(self.barCon);      

    if (bodyWidth + barsPan <= 0 && bodyWidth > 0) {
      transformChart(self.barCon, ( 0 - bodyWidth) + self.drawBoxWidth / 3, 0);      
      self.zoom = adjustZoomX(self.zoom, ((0 - bodyWidth) + self.drawBoxWidth / 3), bodyWidth);      
    } else if (barsPan > self.drawBoxWidth) {
      transformChart(self.barCon, self.drawBoxWidth - (self.drawBoxWidth / 3), 0);      
      self.zoom = adjustZoomX(self.zoom, self.drawBoxWidth - (self.drawBoxWidth / 3), bodyWidth);    
    }
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

  return exports;

}; //end barchart 






