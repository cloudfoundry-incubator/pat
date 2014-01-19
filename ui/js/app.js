pat = {}

pat.experiment = function(onDataCallback, refreshRate) {

  var infoUrl

  function exports() {
  }

  exports.onCsvUrlChanged = function() {}
  exports.onStateChanged = function() {}

  exports.refresh = function() {
    $.get(infoUrl, function(data) {
      onDataCallback(data.Items.filter(function(d) { return d.Type === 0 }))
      setTimeout(exports.refresh, refreshRate)
    })
  }

  exports.run = function() {
    $.post( "/experiments/", { "pushes": 10, "concurrency": 3 }, function(data) {
      infoUrl = data.Location
      exports.onCsvUrlChanged(data.CsvLocation)
      exports.onStateChanged("running")
      exports.refresh()
    })
  }

  return exports
}

var app = function() {
  var chart = d3.custom.pats.throughput(d3.select("#graph"))

  // fixme(jz) - move to table.js and test
  function table(data) {
    tr = d3.select("#data").selectAll("tr").data(data.filter(function(d) { return d.Type === 0 })).enter().append("tr")
    tr.append("td").text(function(d) { return d.WallTime })
    tr.append("td").text(function(d) { return d.LastResult })
    tr.append("td").text(function(d) { return d.Average })
    tr.append("td").text(function(d) { return d.TotalTime })
  }

  var experiment = pat.experiment(function(data) {
    chart(data)
    table(data)
  }, 800)

  experiment.onCsvUrlChanged = function(location) {
    $('#csvbtn').prop('disabled', false)
    $('#csvbtn').click(function() {
      window.location = location
    })
  }

  experiment.onStateChanged = function(state) {
    $('#startbtn').hide()
    $('#stopbtn').show()
  }

  $("#startbtn").click(function() {
    $('#console').append("App push started<br>running ...<br>");
    $("#data tr").remove()
    experiment.run()
  });

  $("#stopbtn").click(function() {
    alert("Not implemented")
  })
}
