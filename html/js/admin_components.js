var Admin = React.createClass({
	"getInitialState": function(event) {
		return { "logged_in" : false};
	},
	"render" : function()
	{
		return (
			<div id="container">
				<LoginForm />
			</div>
		);
	}
});

var LoginForm = React.createClass({
	"render" : function() {
		return (
			<div className="login_overlay">
				<form method="POST" onSubmit={this.Login}>
					<span className="row">
						<label>Username</label>
						<input type="text" id="username" name="username" required />
					</span>
					<span className="row">
						<label>Password</label>
						<input type="password" id="password" name="password" required />
					</span>
					<span className="row">
						<input type="submit" value="Login" />
					</span>
				</form>
			</div>
		);
	},
	"Login" : function(event) {
		event.preventDefault();
		var request = new XMLHttpRequest();
		request.onreadystatechange = function() {
			if(request.readyState === 4) {
				if(request.status === 200) {
					console.log(request.responseText);
				} 
				else
				{

				}
			}
		}

		request.open('POST', '/login', true);
		request.setRequestHeader("Content-type", "application/x-www-form-urlencoded");
		request.send("username=" + document.getElementById("username").value + "&password=" + document.getElementById("password").value);
	}
});	

