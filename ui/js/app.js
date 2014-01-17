var app = function() {

  var running = false
  var chart = d3.custom.pats.throughput(d3.select("#graph"))

  refresh = function() {
    d3.json(running, function(data) {
      chart(data.Items)
      setTimeout(refresh, 5000);
    })
  }

  $("#cmd1").click(function(){
    $('#console').append("App push started<br>running ...<br>");
    $.post( "/experiments/", function(data) {
      $('#console').append("Experiment started, URL is " + data.Location + "<br>");
      running = data.Location
      refresh()

    }, "json");
  });

  $("#cmd2").click(function(){
    $.get( "/cmd2", function(data) {
      alert("command was performed");
    });
  });

  var events = new EventSource("/events");
  events.onmessage = function(event) {
    $('#console').append(event.data + "<br>");
  };

  events.addEventListener("date", function(event) {
    $('#console').append(event.data + "<br>");
  }, false);

}
