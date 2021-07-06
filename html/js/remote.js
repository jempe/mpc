var socket;

if (window["WebSocket"])
{
	start_remote();
}

function start_remote()
{
	socket = new WebSocket("ws://" + location.host + "/remote");
	socket.onclose = function() {
		console.log("Connection has been closed.");
		setTimeout(function()
		{
			start_remote();
		}, 500);
	}
	socket.onmessage = function(e) {
		filter = JSON.parse(e.data);
		app.setState({"selected" : 0, "offset" : 0});
		requestVideos(0);
	}	
}

