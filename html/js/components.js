/*
My Private Collection HTML5 Client

*/

/*
App Component
*/
var App = React.createClass({
	"getInitialState": function(event) {
		return { "selected" : 0, "videos" : this.props.videos, "total_videos" : this.props.total_videos, "view" : "grid", "offset" : this.props.offset, "preview_action" : "play", "show_controls" : true, "show_controls_until" : 5, "show_screenshots" : false, "player_screenshot" : 0};
	},
	"render" : function()
	{
		if(this.state.videos == null)
		{
			return (
				<div id="container" data-view={this.state.view}>
					<p>There are no videos with that criteria</p>
				</div>
			);
			
		}
		else
		{
			var preview_video = this.state.videos[this.state.selected - this.state.offset];
			selected_video = preview_video;
			var current_row = Math.floor(this.state.selected / this.props.videos_per_row);
			var videos_position = grid_item_height * current_row;

			if(this.state.view == "player")
			{
				return (
					<div id="container" data-view={this.state.view}>
						<Player video={preview_video} controls={this.state.show_controls} show_screenshots={this.state.show_screenshots} screenshot={this.state.player_screenshot} />
					</div>
				);
			}
			else
			{
				window.location.hash = this.state.view;

				return (
					<div id="container" data-view={this.state.view}>
						<span id="total_videos">{this.state.total_videos} Videos</span>
						<Preview video={preview_video} action={this.state.preview_action} element={this.props.preview_element} />
						<Grid videos={this.state.videos} offset={this.state.offset} selected={this.state.selected} total={this.state.total_videos} videos_per_row={this.props.videos_per_row} videos_margin={this.props.videos_margin} videos_position={videos_position} />
					</div>
				);
			}
		}
	},
	"previewVideo" : function(index) {
		var current_row = Math.floor(this.state.selected / this.props.videos_per_row);
		var next_row = Math.floor(index / this.props.videos_per_row);
		var new_offset = (next_row * this.props.videos_per_row) - this.props.videos_per_row;

		console.log("selected video:" + index);

		this.setState({"selected" : index, "preview_action" : "play"});
	
		if(next_row != current_row && new_offset >= 0)
		{
			this.state.selected = index;
			requestVideos(new_offset);
		}
	},
	"handleRightArrow" : function(event) {
		if(this.state.view == "grid")
		{
			if(this.state.selected < (this.props.total_videos - 1))
			{
				this.previewVideo(this.state.selected + 1);
			}
		}
		else if(this.state.view == "player")
		{
			if(this.state.show_screenshots)
			{
				this.changePlayerScreenshot(1);
			}
			else
			{
				this.showScreenshots();
			}
		}
	},
	"handleLeftArrow" : function(event) {
		if(this.state.view == "grid")
		{
			if(this.state.selected > 0)
			{
				this.previewVideo(this.state.selected - 1);
			}
		}
		else if(this.state.view == "player")
		{
			if(this.state.show_screenshots)
			{
				this.changePlayerScreenshot(-1);
			}
			else
			{
				this.showScreenshots();
			}
		}
	},
	"handleUpArrow" : function(event) {
		if(this.state.view == "grid")
		{
			if(this.state.selected >= this.props.videos_per_row)
			{
				this.previewVideo(this.state.selected - this.props.videos_per_row);
			}
		}
		else if(this.state.view == "preview")
		{
			if(this.state.preview_action == "resume")
			{
				this.setState({ "preview_action" : "play"});
			}	
		}
	},
	"handleDownArrow" : function(event) {
		if(this.state.view == "grid")
		{
			if(this.state.selected < this.state.total_videos - this.props.videos_per_row)
			{
				this.previewVideo(this.state.selected + this.props.videos_per_row);
			}
		}
		else if(this.state.view == "preview")
		{
			if(this.state.preview_action == "play")
			{
				this.setState({ "preview_action" : "resume" });
			}
			else if(this.state.preview_action == "resume")
			{
				this.setState({ "view" : "grid" });
			}
		}
		else if(this.state.view == "player")
		{
			if(this.state.show_screenshots)
			{
				this.hideScreenshots();
			}
			else
			{
				this.setState({ "view" : "preview" });
			}
		}
	},
	"handleEnterKey" : function(event) {
		if(this.state.view == "grid")
		{
			this.setState({ "view" : "preview" });
		}
		else if(this.state.view == "preview")
		{
			if(this.state.preview_action == "play" || this.state.preview_action == "resume")
			{
				this.playVideo();
			}
		}
		else if(this.state.view == "player")
		{
			if(this.state.show_screenshots)
			{
				this.jumpToScreenshot();
			}
			else
			{
				this.showScreenshots();
			}
		}
	},
	"handlePlayButton" : function(event) {
		if(this.state.view == "player")
		{
			if(document.getElementById('video_player').paused)
			{
				document.getElementById('video_player').play();
			}
			else
			{
				document.getElementById('video_player').pause();
			}	
		}
	},
	"handleForwardButton" : function(event) {
		if(this.state.view == "player")
		{
			document.getElementById('video_player').currentTime += forward_rewind_seconds();
			this.showControls();
		}
	},
	"handleRewindButton" : function(event) {
		if(this.state.view == "player")
		{
			document.getElementById('video_player').currentTime -= forward_rewind_seconds();
			this.showControls();
		}
	},
	"showControls" : function()
	{
		var show_until = document.getElementById('video_player').currentTime + 5;
		this.setState({ "show_controls" : true , "show_controls_until" : show_until });
	},
	"hideControls" : function()
	{
		this.setState({ "show_controls" : false });
	},
	"playVideo" : function()
	{
		this.setState({ "view" : "player" });
	},
	"showScreenshots" : function()
	{
		this.showControls();
		this.setState({ "show_screenshots" : true, "player_screenshot" : calc_player_time(document.getElementById("video_player").currentTime) });
	},
	"hideScreenshots" : function()
	{
		this.hideControls();
		this.setState({ "show_screenshots" : false });
	},
	"changePlayerScreenshot" : function(steps)
	{
		var new_screenshot_time = this.state.player_screenshot + (steps * (this.state.videos[this.state.selected - this.state.offset].step));

		if(new_screenshot_time < 0)
		{
			new_screenshot_time = 0;
		}

		if(new_screenshot_time > document.getElementById("video_player").duration)
		{
			new_screenshot_time = document.getElementById("video_player").duration;
		}

		this.setState({ "player_screenshot" : calc_player_time(new_screenshot_time) });
	},
	"jumpToScreenshot" : function()
	{
		document.getElementById("video_player").currentTime = this.state.player_screenshot;
		this.hideScreenshots();
	}
});
/*
Preview Component
*/
var Preview = React.createClass({
	"render" : function()
	{
		var screenshot = "/videos/screenshots/" + this.props.video.id + "/" + random_second(this.props.video.duration)  + ".jpg";

		var categoriesList = [];
		if(this.props.video.categories != null)
		{
			categoriesList = this.props.video.categories;
		}

		var categories = categoriesList.map(function(category, i){

			return <li key={'prev_cat_' + i}>{category.name}</li>;
		});

		var actorsList = [];
		if(this.props.video.actors != null)
		{
			actorsList = this.props.video.actors;
		}

		var actors = actorsList.map(function(actor, i){

			return <li key={'prev_actor_' + i}>{actor.name}</li>;
		});

		var play_button_status = "inactive";
		var resume_button_status = "inactive";

		if(this.props.action == "play")
		{
			play_button_status = "active";
		}

		if(this.props.action == "resume")
		{
			resume_button_status = "active";
		}

		var play_button = <Button status={play_button_status} name="PLAY" action="play" />;
		var resume_button = <Button status={resume_button_status} name="RESUME" action="play" />;
		var back_button = <Button status="inactive" name="BACK" action="back" />;

		var preview_element = (
			<img src={screenshot} id="preview_image" />
		);

		if(this.props.element == "video")
		{
			var video_file = "/videos/" + this.props.video.file;
			preview_element = (
				<video src={video_file} id="preview_video" />
			);
		}

		previewScreenshots(this.props.video.duration, this.props.video.id);

		return (
			<div id="preview">
				<div className="info">
					<h1>{this.props.video.title}</h1>
					<p>{this.props.video.description}</p>
				{categoriesList.length > 0 &&
					<div className="list categories">
						<ul>
							{categories}
						</ul>
					</div>
				}
				{actorsList.length > 0 &&
					<div className="list actors">
						<ul>
							{actors}
						</ul>
					</div>
				}
					<div className="buttons">
						{play_button}
						{resume_button}
						{back_button}
					</div>
				</div>
				<figure>
					{preview_element}
				</figure>
			</div>
		);	
	}
});
/*
Videos Grid
*/
var Grid = React.createClass({
	"render" : function()
	{
		var selectedVideoIndex = this.props.selected - offset;
		var videoMargin = this.props.videos_margin;
		var voffset = this.props.offset;

		var videos = this.props.videos.map(function(video, i){
			var thumb = "/videos/thumbs/" + video.id + ".jpg";

			var selected = false;

			if(i == selectedVideoIndex)
			{
				selected = true;
			}

			var video_id = 'vid_' + (i + voffset)
			var video_key = 'video_' + (i + voffset)
			var found_state = true;

			return <Video key={video_key} id={video_id} index={i + voffset} thumburl={thumb} selected={selected} margin={videoMargin} found={found_state} />
		});

		var videosPadding = Math.floor(offset / videos_per_row) * grid_item_height;

		return (
			<div id="grid" style={{paddingLeft : this.props.videos_margin + "px"}}>
				<div id="grid_videos" style={{ marginTop : (this.props.videos_position * -1) + 'px', paddingTop : videosPadding + 'px' }}>
				{videos}
				</div>
			</div>
		);
	}
});
/*
Video
*/
var Video = React.createClass({
	"getInitialState": function() {
		return { "found" : this.props.found };
	},
	"handleError": function(event) {
		this.setState({"found": false});
	},
	"clickVideo" : function() {
		app.previewVideo(this.props.index);
		app.setState({ "view" : "preview" });
	},
	"resizeImage": function() {
		var image = document.getElementById(this.props.id);
		
		var container_ratio = image.parentNode.clientWidth / image.parentNode.clientHeight;
		var image_ratio = image.width / image.height;

		if(image_ratio > container_ratio)
		{
			image.style.height = image.parentNode.clientHeight + "px";
		}
		else
		{
			image.style.width = image.parentNode.clientWidth + "px";
		}
	},
	"render" : function(){
		var thumbURL = this.props.thumburl;

		var link_class = "video";
		if(this.props.selected)
		{
			link_class = "video selected";
		}

		if( ! this.state.found)
		{
			thumbURL = "images/file-video.jpg";
		}

		if(isTV())
		{
			return(
				<article className={link_class} style={{marginRight: this.props.margin + 'px', width: grid_item_width + 'px', height: grid_item_height + 'px'}}>
					<img id={this.props.id} src={thumbURL} onError={this.handleError} onLoad={this.resizeImage} />
				</article>
			);

		}
		else
		{
			return(
				<article className={link_class} style={{marginRight: this.props.margin + 'px', width: grid_item_width + 'px', height: grid_item_height + 'px'}} onClick={this.clickVideo}>
					<img id={this.props.id} src={thumbURL} onError={this.handleError} onLoad={this.resizeImage} />
				</article>
			);
		}
	}
	
});
/*
Button
*/
var Button = React.createClass({
	"render" : function() {
		if(isTV())
		{
			return(
				<span className="button" data-status={this.props.status}>{this.props.name}</span>
			)
		}
		else
		{
			return(
				<span className="button" data-status={this.props.status} onClick={this.clickButton}>{this.props.name}</span>
			)
		}
	},
	"clickButton" : function() {
		if(this.props.action == "back")
		{
			app.setState({ "view" : "grid" });
		}
		else if(this.props.action == "play")
		{
			location.hash = "player";
			app.playVideo();
		}
	}
});
/*
Player Component
*/
var Player = React.createClass({
	"getInitialState" : function()
	{
		return { "status" : "play", "elapsed" : 0, "duration" : 0 };
	},
	"render" : function() {

		var video_file = "/videos/" + this.props.video.id + ".mp4";

		var elapsed_time = format_time(this.state.elapsed);
		var total_time = format_time(this.state.duration);
		var progress_percent = (this.state.elapsed / this.state.duration) * 100;
		var screenshot_file = "videos/screenshots/" + this.props.video.id + "/" + this.props.screenshot + ".jpg";

		if(this.props.show_screenshots)
		{
			progress_percent = (this.props.screenshot / this.state.duration) * 100;
		}	

		
		var videoEl;
		var playButton;
		var progressBar;

		if(isTV())
		{
			videoEl = <video id="video_player" src={video_file} autoPlay="true" onPlay={this.handlePlayAction} onPause={this.handlePauseAction} onDurationChange={this.handleDurationChange} onTimeUpdate={this.handleProgress} />
			playButton = <span id="control_status" className={this.state.status}></span> 
			progressBar = (
					<span className="progress" id="progress">
						<span className="progress_bar" style={{ width : progress_percent + "%" }}></span>
					</span>
			)
		}
		else
		{
			videoEl = <video id="video_player" src={video_file} autoPlay="true" onPlay={this.handlePlayAction} onPause={this.handlePauseAction} onDurationChange={this.handleDurationChange} onTimeUpdate={this.handleProgress} onClick={this.showControls} />
			playButton = <span id="control_status" className={this.state.status} onClick={this.clickPlay}></span> 
			progressBar = (
					<span className="progress" id="progress" onLoad={bindUpdateVideo()}>
						<span className="progress_bar" style={{ width : progress_percent + "%" }}></span>
					</span>
			)

		}

		return(
			<div id="player_container" data-controls={this.props.controls}>
				{videoEl}
				<span id="controls_container">
					{playButton}
					<span id="elapsed" className="time">{elapsed_time}</span>
					<span className="progress_container">
							{progressBar}
						<figure className="screenshot" data-show={this.props.show_screenshots} style={{ left : progress_percent + "%" }}>
							<img src={screenshot_file} />
						</figure>
					</span>
					<span id="total" className="time">{total_time}</span>
				</span>
			</div>
		);
	},
	"handlePlayAction" : function()
	{
		app.showControls();
		this.setState({ "status" : "play" });
	},
	"handlePauseAction" : function()
	{
		app.showControls();
		this.setState({ "status" : "pause" });
	},
	"handleDurationChange" : function()
	{
		this.setState({ "duration" : document.getElementById("video_player").duration });
	},
	"handleProgress" : function()
	{
		if( ! this.props.show_screenshots)
		{
			if(document.getElementById("video_player").currentTime >  app.state.show_controls_until)
			{
				app.hideControls();
			}
			this.setState({ "elapsed" : document.getElementById("video_player").currentTime  });
		}
	},
	"showControls" : function()
	{
		app.showControls();
	},
	"clickPlay" : function()
	{
		if(document.getElementById("video_player").paused)
		{
			document.getElementById("video_player").play();
		}
		else
		{
			document.getElementById("video_player").pause();
		}
	}
});

