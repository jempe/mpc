var admin;

admin = ReactDOM.render(
	<Admin />,
	document.getElementById('admin_container')
);

function requestVideos(offset, rows, seed, sort_by, filter) {
	var request = new XMLHttpRequest();
	request.onreadystatechange = function() {
		if(request.readyState === 4) {
			if(request.status === 200) {
				var videos_json = JSON.parse(request.responseText);
				var videos_data = videos_json.videos
				var total_videos = videos_json.total;
				var offset = videos_json.offset;

			} 
			else
			{

			}
		}
	}

	request.open('POST', '/videos.json?view=' + rows  + '&offset=' + offset + '&seed=' + seed + "&sort=" + sort_by);
	request.send(JSON.stringify(filter));
}
requestVideos(0, 10, new Date().getTime(), {});


