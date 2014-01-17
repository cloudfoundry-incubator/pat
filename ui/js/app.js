var app = function() {

    $("#cmd1").click(function(){
        $('#console').append("App push started<br>running ...<br>");
        $.post( "/experiments/", function(data) {
            $('#console').append("Experiment started, URL is " + data.Location + "<br>");

						$.get(data["Location"], function(data) {
							$('#data').val(""+data.Items.map(function(el) {
								return JSON.stringify(el)
							}).join("\n"))
							// repeat.. (keep polling for new results, and graph..)
						}, "json")
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
