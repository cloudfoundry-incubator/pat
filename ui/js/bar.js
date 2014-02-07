d3.custom.barchart = function(el, observableArray) {
  const d3_id = "d3_workload",
        d3_container = "d3_workload_container";
        
  var margin = {top: 30, right: 30, bottom: 30, left: 30},
      d3Obj = d3.select(el),
      jqObj = $(el);

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
    .on("zoom", this.zoomed());

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

  this.outterBody = this.svg.append("g");

  this.svg.append("defs").append("clipPath")
    .attr("id", "clip")
  .append("rect")
    .attr("x", 15)
    .attr("y", 1)
    .attr("width", this.svgWidth - this.clipMarginRight)
    .attr("height", this.svgHeight + 30);

 var chartBody = this.svg.append("g")
    .attr("clip-path", "url(#clip)")
    .call(this.zoom);
  chartBody.append("rect")
    .attr("width","100%")
    .attr("height","100%")
    .attr("style","fill:none;pointer-events: all;");

  this.chartShift = chartBody.append("g")
    .attr("transform", "translate(0, 0)");

  this.outterBody.append("g")
    .attr("class", "x axis")
    .attr("transform", "translate(0," + this.svgHeight + ")");
  this.outterBody.append("g")  
    .attr("class", "y axis")
    .attr("transform", "translate(" + (this.svgWidth - 20) + ", 0)");

  this.barCon = this.chartShift.append("g")
    .attr("transform", "translate(30,0)");

  var self = this;
  observableArray.subscribe(function() {
      if (!self.data) {
          self.svg.append("text")
            .attr("x", -40)
            .attr("y", 0)
            .attr("transform", "rotate(-90)")
            .text("Seconds");
          self.svg.append("text")
            .attr("x", self.svgWidth - 60)
            .attr("y", self.svgHeight + 25)
            .text("Task #")
            .attr("text-anchor","middle");
          self.data = observableArray();
          self.refresh()();
      } else {
        self.data = observableArray();
        self.refresh()();
      }
  });
};

d3.custom.barchart.prototype.refresh = function() {
  const second = 1000000000;
  var self = this;

  return function(){
    var len = self.data.length;
    var x = self.x.domain( [1, len, 1] ).range([self.barWidth + 1, len * (self.barWidth + 1)]);
    var y = self.y.domain([d3.max(self.data, function(d) { return d.LastResult / second}), 0] );
    var bars = self.barCon.selectAll("rect.bar").data(self.data);
    var labels = self.barCon.selectAll("text").data(self.data);

    var w = len * (self.barWidth + 1);
    var dw = self.svgWidth - self.clipMarginRight;
    var transf = self.barCon.attr("transform");
    var splitted = transf.split(",");
    var barsPan = getTranslateX(self.barCon);      
    var shiftPan = getTranslateX(self.chartShift);
    if (shiftPan + barsPan + w < 0) {
      self.chartShift.transition()
        .attr("transform", "translate(0, 0)" );
      self.barCon.transition()
        .attr("transform", "translate(30, 0)" );            
      self.zoom.translate([0,0]).scale(1);      
    } 

    if ((w) > dw) {        
      self.chartShift.transition()
        .attr("transform", "translate(" + (dw - w  - 25 ) + ", 0)" );    
    }

    if (shiftPan + barsPan > dw) {
      self.chartShift.transition()
        .attr("transform", "translate(0, 0)" );
      self.barCon.transition()
        .attr("transform", "translate(30, 0)" );      
      self.zoom.translate([0,0]);
    }
    
    bars.transition()
      .attr("x", function(d, i) { return x(i) })
      .attr("y", function(d) { return y(d.LastResult / second) } )
      .attr("height", function(d) { return (self.svgHeight - y(d.LastResult / second)) } );

    labels.transition()
      .attr("x", function(d, i) { return (x(i) + (self.barWidth / 2)) })
      .attr("y", self.svgHeight + 3 );

    bars.enter()
      .append("rect")
        .attr("x", function(d, i) { return x(i) + 20 })
        .attr("y", function(d) { return y(d.LastResult / second) } )
        .attr("height", function(d) { return (self.svgHeight - y(d.LastResult / second)) } )
        .attr("width", self.barWidth)
        .attr("class", "bar")
      .transition()        
        .duration(1000)
        .attr("x", function(d, i) { return x(i) })
        .attr("y", function(d) { return y(d.LastResult / second) })
        .attr("height", function(d) { return (self.svgHeight - y(d.LastResult / second) ) } );

    bars.enter()
      .append("text")
        .attr("x", function(d, i) { return (x(i) + (self.barWidth / 2)) })
        .attr("y", self.svgHeight + 30 )
        .attr("dy", ".7em")
        .text(function(d, i){ return i + 1 })
      .transition()
        .attr("y", self.svgHeight + 3 );
      

    self.outterBody.select(".x.axis").call(self.xAxis);    
    self.outterBody.select(".y.axis").call(self.yAxis);

    bars.exit().remove();   
    labels.exit().remove();      
  }
}

d3.custom.barchart.prototype.zoomed = function() {
  var self = this;
  return function() {    
    var svg = self.svg;
    self.barCon.attr("transform", "translate(" + d3.event.translate[0] + ",0)scale(1, 1)");
  }
}

d3.custom.barchart.prototype.yAxis_max = function() {
  return this.yAxis.scale().domain()[0];
}

d3.custom.barchart.prototype.xAxis_max = function() {
  return this.xAxis.scale().domain()[0];
}

function getTranslateX(node) {
  var splitted = node.attr("transform").split(",");
  return parseInt(splitted [0].split("(")[1]);
}
