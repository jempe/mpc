var Video = React.createClass({
	"getInitialState": function(event) {
		var initial_state = {
			"found": true
		};

		return initial_state;
	},
	"handleError": function(event) {
		this.setState({"found": false});
	},
	"render" : function(){
		var thumbURL = this.props.thumburl;

		var link_class = "video";
		if(this.props.index == selected_video)
		{
			link_class = "video selected";
		}

		if( ! this.state.found)
		{
			thumbURL = "images/file-video.jpg";
		}

		return(
			<article className={link_class}>
				<div data-href={this.props.file}>
					<figure>
						<img src={thumbURL} onError={this.handleError} />
						<figcaption>
							<span className="actors">{this.props.actors}</span>
							<span className="categories">{this.props.categories}</span>
						</figcaption>
					</figure>
				</div>
			</article>
		);
	}
});
var Videos = React.createClass({
	"render" : function(){
		var videos = this.props.videos.map(function(video, i){
			var thumb = "/videos/thumbs/" + video.id + ".jpg";
			var file = "/videos/" + video.file;

			var actors_list = [];
			for(var actor_index in video.actors)
			{
				var actor = video.actors[actor_index];
				actors_list.push(actor.name);
			}

			var actors = actors_list.join(", ");

			var categories_list = [];
			for(var category_index in video.categories)
			{
				var category = video.categories[category_index];
				categories_list.push(category.name);
			}

			var categories = categories_list.join(", ");


			return <Video key={'video_' + i} index={i} id={video.id} thumburl={thumb} file={file} actors={actors} categories={categories}  />
		});

		var videos_class = "";

		if(current_focus == "thumbs")
		{
			videos_class = "focused";
		}

		return(
			<main className={videos_class}>
				{videos}
			</main>
		);
	}
});
var VideoPreview = React.createClass({
	"getInitialState": function(event) {
		var initial_state = {
			"found": true
		};

		return initial_state;
	},
	"handleError": function(event) {
		this.setState({"found": false});
	},
	"render" : function(){
		var curtime = new Date().getTime();
		
		var videoThumbURL = "/videos/screenshots/" + this.props.video.id + "/20.jpg";
		var thumbURL = videoThumbURL + "?v=" + curtime;

		if(last_preview_image != videoThumbURL)
		{
			last_preview_image = videoThumbURL;
		}
		else if( ! this.state.found)
		{
			thumbURL = "images/file-video.jpg";
		}
		

		var preview_class = "preview";

		if(current_focus == "preview")
		{
			preview_class += " focused";
		}

		return(
			<div className={preview_class}>
				<figure>
					<img src={thumbURL} onError={this.handleError} />
				</figure>
				<div className="buttons">
					<span className="button play">PLAY</span>
				</div>
			</div>
		);
	}
});
var VideoPlayer = React.createClass({
	"getInitialState": function(event) {
		var initial_state = {
			"curtime": 0
		};

		return initial_state;
	},
	"render" : function(){
		var player_class = "show_video";
		var video_file = "videos/" + this.props.video.file;

		if(current_focus != "player")
		{
			player_class = "hide";
			video_file = "";
		}

		var elapsed_time = "0:00";
		var total_time = "0:00";

		var status_class = "play";

		if(document.getElementById('player') != null)
		{
			elapsed_time = format_time(document.getElementById('player').currentTime);
			total_time = format_time(document.getElementById('player').duration);

			if(document.getElementById('player').paused)
			{
				status_class = "pause";
			}
		}

		if(new Date().getTime() > show_time_until)
		{
			player_class += " hide_controls";
		}

		var screenshot_position = "";
		var screenshot_url = "videos/screenshots/" + videos_data[selected_video].id +  "/" + screenshot_time  + ".jpg";


		if(player_navigation)
		{
			player_class += " screenshot";
			var screenshot_percent = parseInt((screenshot_time/document.getElementById('player').duration) * 100);  
		}

		return(
			<div className={player_class}>
				<video id="player" width="100%" height="100%" src={video_file}></video>
				<span id="controls_container">
					<span id="control_status" className={status_class}></span>
					<span id="elapsed" className="time">{elapsed_time}</span>
					<span className="progress_container">
						<progress id="player_progress"></progress>
						<figure className="screenshot" style={{left: screenshot_percent + '%'}}><img src={screenshot_url} /></figure>
					</span>
					<span id="total" className="time">{total_time}</span>
				</span>
			</div>
		);
	}
});

var request = new XMLHttpRequest();

var videos_requested = false;
var videos_data;
var total_videos;
var selected_video = 0;
var thumb_width = 370;
var thumb_height = 300;
var videos_per_row = Math.floor(document.body.clientWidth / thumb_width);
var videos_per_column = Math.floor(window.innerHeight / thumb_height)- 2;

var videos_container = document.getElementById("videos_container");
videos_container.style.height = (window.innerHeight - (2 * thumb_height)) + "px";
videos_container.style.width = window.innerWidth + "px";

var current_focus = "thumbs";
var videos_component;
var preview_component;
var player_component;
var last_nav_attempt = 0;
var show_time_until = 0;
var last_preview_image;
var player_navigation = false;
var screenshot_time = 0;
document.onkeydown = checkKey;

function requestVideos() {
	videos_requested = true;
	request.onreadystatechange = function() {
		if(request.readyState === 4) {
			if(request.status === 200) {
				var videos_json = JSON.parse(request.responseText);
				videos_data = videos_json.videos
				total_videos = videos_data.length;
				showVideos();
			} else {

			}
		}
	}

	request.open('Get', '/videos.json?view=1000&offset=0');
	request.send();
}

function showVideos()
{
	videos_component = ReactDOM.render(
		<Videos videos={videos_data} />,
		document.getElementById('videos_container')
	);

	var sel_video = videos_data[selected_video];
	preview_component = ReactDOM.render(
		<VideoPreview video={sel_video} />,
		document.getElementById("video_preview")
	);
	preview_component.setState({found : true});

	player_component = ReactDOM.render(
		<VideoPlayer video={sel_video} />,
		document.getElementById("video_player")
	);
}

var security_code = [
	'39', '40', '40', '40', '39', '38', '37'	
];
var key_attempt = 0;
var granted = true;


function checkKey(e) {
	e = e || window.event;
	console.log(e.keyCode);

	if( ! videos_requested)
	{
		console.log(key_attempt + ":" + e.keyCode);
		if(e.keyCode != security_code[key_attempt])
		{
			granted = false;
		}

		key_attempt++;

		if(granted && key_attempt == security_code.length)
		{
			requestVideos();
		}
	}
	else
	{
		if(current_focus == "thumbs")
		{
			if(parseInt(e.keyCode) >= 37 && parseInt(e.keyCode) <= 40)
			{
				var current_selected_video = selected_video;
				if (e.keyCode == '38')
				{
					// up arrow
					select_video(- videos_per_row);
				}
				else if (e.keyCode == '40')
				{
					// down arrow
					select_video(videos_per_row);
				}
				else if (e.keyCode == '37')
				{
					// left arrow
					select_video(-1);
				}
				else if (e.keyCode == '39')
				{
					// right arrow
					select_video(1);
				}

				if(selected_video != current_selected_video)
				{
					showVideos();
				}
			}
			else if(e.keyCode == '13')
			{
				focus_preview();
				//window.location = "videos/" + videos_data[selected_video].file;
			}
		}
		else if(current_focus == "preview")
		{
			if(parseInt(e.keyCode) >= 37 && parseInt(e.keyCode) <= 40)
			{
				if (e.keyCode == '38')
				{
					// up arrow
				}
				else if (e.keyCode == '40')
				{
					// down arrow
					focus_thumbs();
				}
				else if (e.keyCode == '37')
				{
					// left arrow
				}
				else if (e.keyCode == '39')
				{
					// right arrow
				}

			}
			else if(e.keyCode == '13')
			{
				play_video();
			}
			else if(e.keyCode == '27')
			{
				focus_thumbs();
			}	
		}
		else if(current_focus == "player")
		{
			if(player_navigation)
			{
				if(parseInt(e.keyCode) >= 37 && parseInt(e.keyCode) <= 40)
				{
					if (e.keyCode == '38')
					{
						// up arrow
					}
					else if (e.keyCode == '40')
					{
						// down arrow
						hidePlayerNavigation();
					}
					else if (e.keyCode == '37')
					{
						// left arrow
						showPlayerScreenshot(-1);
					}
					else if (e.keyCode == '39')
					{
						// right arrow
						showPlayerScreenshot(1);
					}

				}
				else if(e.keyCode == '13')
				{
					// menu button
					document.getElementById('player').currentTime = screenshot_time;
					document.getElementById('player').play();
					hidePlayerNavigation();
				}

			}
			else
			{
				if(parseInt(e.keyCode) >= 37 && parseInt(e.keyCode) <= 40)
				{
					if (e.keyCode == '38')
					{
						// up arrow
					}
					else if (e.keyCode == '40')
					{
						// down arrow
						focus_preview();
					}
					else if (e.keyCode == '37')
					{
						// left arrow
						showPlayerNavigation();
					}
					else if (e.keyCode == '39')
					{
						// right arrow
						showPlayerNavigation();
					}

				}
			}

			if(e.keyCode == '32' || e.keyCode == '179')
			{
				if(document.getElementById('player').paused)
				{
					document.getElementById('player').play();
					showTime(5);
				}
				else
				{
					document.getElementById('player').pause();
					showTime(5);	
				}
			}
			else if(e.keyCode == '18')
			{
				// menu button

			}
			else if(e.keyCode == '228' || e.keyCode == '227')
			{
				var current_time = new Date().getTime();

				if((last_nav_attempt - current_time) > 5000)
				{
					var jump_time = 5;
				}
				else if((last_nav_attempt - current_time) > 2000)
				{
					var jump_time = 10;
				}
				else if((last_nav_attempt - current_time) > 1000)
				{
					var jump_time = 20;
				}
				else
				{
					var jump_time = 30;
				}

				if(e.keyCode == '228')
				{
					document.getElementById('player').currentTime += jump_time;
				}
				else
				{
					document.getElementById('player').currentTime -= jump_time;
				}
				showTime(5);
			}
			else if(e.keyCode == '27')
			{
				focus_preview();
			}
		}
	}
}
function showPlayerNavigation()
{
	showTime(1000);
	player_navigation = true;
	screenshot_time = document.getElementById('player').currentTime;
}
function showPlayerScreenshot(offset)
{
	var navigation_step = document.getElementById('player').duration / 30;
	screenshot_time += (offset * navigation_step);

	if(screenshot_time < 0)
	{
		screenshot_time = 0;
	}
	else if(screenshot_time > document.getElementById('player').duration)
	{
		screenshot_time = document.getElementById('player').duration - 1;
	}
}
function hidePlayerNavigation()
{
	showTime(5);
	player_navigation = false;
}
function play_video()
{
//	window.location = "videos/" + videos_data[selected_video].file;
	focus_player();
	document.getElementById('player').play();
	showTime(10);
}
function focus_player()
{
	current_focus = "player";
	player_navigation = false;
	update_components();
}
function focus_preview()
{
	current_focus = "preview";
	update_components();
}
function focus_thumbs()
{
	current_focus = "thumbs";
	update_components();
}
function update_components()
{
	document.body.className = current_focus;
	videos_component.forceUpdate();
	preview_component.forceUpdate();
	player_component.forceUpdate();
}
function select_video(add_to_current)
{
	var previous_selected = selected_video;
	var new_selected_video = selected_video + add_to_current;

	if(new_selected_video >= 0 && new_selected_video <= (total_videos - 1))
	{
		selected_video = new_selected_video;
	}

	document.querySelector("#videos_container > main").style.marginTop = (getRow(selected_video) * -1 * thumb_height) + "px";
}
function getRow(selected)
{
	return Math.floor(selected / videos_per_row);
}
var track_player_progress = setInterval(function()
{
	if(document.getElementById('player') != null && current_focus == "player")
	{
		if(player_navigation)
		{
			var current_progress = screenshot_time / document.getElementById('player').duration;
		}
		else
		{
			var current_progress = document.getElementById('player').currentTime / document.getElementById('player').duration;
		}

		document.getElementById('player_progress').value = current_progress;
		player_component.forceUpdate();
	}
}, 500);
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
function showTime(seconds)
{
	show_time_until = new Date().getTime() + (seconds * 1000);
}
