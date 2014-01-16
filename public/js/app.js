var app = function() {

    $("#cmd1").click(function(){
        $.get( "/push", function(data) {
            alert( "command was performed." + data );
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
