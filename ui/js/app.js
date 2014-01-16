var app = function() {

    $("#cmd1").click(function(){
        $('#console').append("App push started<br>running ...<br>");
        $.get( "/experiments/push", function(data) {
            $('#console').append("Total time ran: " + data.TotalTime + "<br>");
        });
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
