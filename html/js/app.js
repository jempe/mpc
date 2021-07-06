/* Init Variables*/
var request = new XMLHttpRequest();
var videos_requested = false;
var videos_data;
var total_videos;
var offset;
var app;
var pattern;
var grid_item_width = 350;
var grid_item_height = 195;
if(document.body.clientHeight <= 800 && document.body.clientWidth > document.body.clientHeight)
{
	grid_item_width = Math.floor(document.body.clientWidth / 3.2);
	grid_item_height = Math.floor(grid_item_width * 195 / 350);
}
var last_grid_row = 0;
var videos_per_row = Math.floor(document.body.clientWidth / grid_item_width);
var videos_margin = (document.body.clientWidth - (videos_per_row * grid_item_width)) / (videos_per_row + 1);
var selected_video;
var preview_element_type = "image";
var last_rewind_forward_press = 0;
var video_steps = 30;
var seed = parseInt(new Date().getTime() / 1000);
var sort_by = "random"; //title, titleDesc, duration, durationDesc, random
var filter = {
	"title" : "",
	"category" : "",
	"actor" : "",
	"quality" : "",
	"duration" : [0, 1000000000]
}

function requestVideos(request_offset) {
	videos_requested = true;
	request.onreadystatechange = function() {
		if(request.readyState === 4) {
			if(request.status === 200) {
				var videos_json = JSON.parse(request.responseText);
				videos_data = videos_json.videos
				total_videos = videos_json.total;
				offset = videos_json.offset;

				if(typeof(app) == "object")
				{
					app.setState({ "videos" : videos_data, "total_videos" : total_videos, "offset" : offset });
				}
				else
				{
					app = ReactDOM.render(
						<App videos={videos_data} total_videos={total_videos} offset={offset} videos_per_row={videos_per_row} videos_margin={videos_margin} preview_element={preview_element_type} />,
						document.getElementById('app_container')
					);
				}
			} 
			else
			{

			}
		}
	}

	request.open('POST', '/videos.json?view=' + (videos_per_row * 4)  + '&offset=' + request_offset + '&seed=' + seed + "&sort=" + sort_by);
	request.send(JSON.stringify(filter));
}

if(enter_pattern)
{
	pattern = ReactDOM.render(
		<Pattern />,
		document.getElementById('pattern_container')
	);
}
else
{
	requestVideos(0);
}

document.onkeydown = checkKey;

function checkKey(e) {
	e = e || window.event;
	console.log(e.keyCode);

	if(typeof(app) == "undefined" && typeof(pattern) == "object")
	{
		if (e.keyCode == '38')
		{
			// up arrow
			pattern.handleUpArrow();
		}
		else if (e.keyCode == '40')
		{
			// down arrow
			pattern.handleDownArrow();
		}
		else if (e.keyCode == '37')
		{
			// left arrow
			pattern.handleLeftArrow();
		}
		else if (e.keyCode == '39')
		{
			// right arrow
			pattern.handleRightArrow();
		}
		else if (e.keyCode == '13')
		{
			// enter key
			pattern.handleEnterKey();
		}

	}
	else
	{
		if (e.keyCode == '38')
		{
			// up arrow
			app.handleUpArrow();
		}
		else if (e.keyCode == '40')
		{
			// down arrow
			app.handleDownArrow();
		}
		else if (e.keyCode == '37')
		{
			// left arrow
			app.handleLeftArrow();
		}
		else if (e.keyCode == '39')
		{
			// right arrow
			app.handleRightArrow();
		}
		else if (e.keyCode == '13')
		{
			// enter key
			app.handleEnterKey();
		}
		else if (e.keyCode == '32' || e.keyCode == '179')
		{
			// space or play button
			app.handlePlayButton();
		}
		else if(e.keyCode == '228' || e.keyCode == '102')
		{
			app.handleForwardButton();
		}
		else if(e.keyCode == '227' || e.keyCode == '100')
		{
			app.handleRewindButton();
		}
	}
}
function previewScreenshots(duration, video_id)
{
	var preview_element = document.getElementById('preview_image');

	if(preview_element_type == "video")
	{
		var preview_element = document.getElementById('preview_video');
	}

	if(preview_element != null)
	{
		if(video_id == selected_video.id)
		{
			var selected_time = random_second(duration);

			if(preview_element_type == "video")
			{
				preview_element.currentTime = selected_time;
			}
			else
			{
				preview_element.src = "/videos/screenshots/" + video_id + "/" + selected_time + ".jpg" ;
			}

			setTimeout(function()
			{
				previewScreenshots(duration, video_id);
			}, 4000);
		}
	}
	else
	{
		setTimeout(function()
		{
			previewScreenshots(duration, video_id);
		}, 200);
	}
}
function random_second(duration)
{
	return calc_player_time(duration * Math.random());
}
function format_time(seconds)
{
	var output = "";

	var hours = Math.floor(seconds / 3600);
	seconds = seconds % 3600;

	if(hours > 0)
	{
		output += add_zeros(hours) + ":";
	}

	var minutes = Math.floor(seconds / 60);
	seconds = seconds % 60;

	output += add_zeros(minutes) + ":" + add_zeros(Math.floor(seconds));

	return output;
}
function add_zeros(number)
{
	if(number > 9)
	{
		return number;
	}
	else
	{
		return '0' + number;
	}
}
function forward_rewind_seconds()
{
	var cur_time = new Date().getTime();
	var seconds = 5;

	var time_offset = cur_time - last_rewind_forward_press;

	if(time_offset < 3000)
	{
		seconds = 10;

		if(time_offset < 500)
		{
			seconds = 60;	
		}
		else if(time_offset < 1000)
		{
			seconds = 30;
		}
		else
		{
			seconds = 20;
		}
	}


	last_rewind_forward_press = cur_time;

	console.log("move seconds:" + seconds);

	return seconds;
}
function calc_player_time(seconds)
{
	if(typeof(app) == "object")
	{
		var step = app.state.videos[app.state.selected - app.state.offset].step;

		return Math.floor(seconds / step) * step;
	}
	else
	{
		return 0;
	}
}
function updateVideo(e)
{
	var xPos = e.clientX - document.getElementById("progress").getBoundingClientRect().left;

	var video_time = document.getElementById("video_player").duration * (xPos / document.getElementById("progress").getBoundingClientRect().width);
	document.getElementById("video_player").currentTime = video_time;
}
function bindUpdateVideo()
{
	if(document.getElementById("progress") != null && document.getElementById("progress").dataset.binded != "true")
	{
		document.getElementById("progress").dataset.binded = "true";
		document.getElementById("progress").addEventListener("click", updateVideo, false);
	}
}
window.onpopstate = function(event) {
	if(typeof(app) == "object" && window.location.hash == "#preview" && app.state.view == "player")
	{
		app.setState({ "view" : "preview" });
	}
};
