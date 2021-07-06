var Pattern = React.createClass({
	"getInitialState" : function()
	{
		return { "selected" : null, "draw_line" : false, "sequence" : [] };
	},
	"render" : function()
	{
		var dots = [];

	 	for (var i=0; i < 9; i++) {
			var dot_selected = false;

			if(this.state.selected == i)
			{
				dot_selected = true;
			}

			var dot_row = this.calcRow(i); 
			var dot_column = this.calcColumn(i); 

			var dot_clicked = false;

			if(this.state.sequence.indexOf(i) != -1)
			{
				dot_clicked = true;
			}

    			dots.push(<Dot key={'dot' + i} selected={dot_selected} row={dot_row} column={dot_column} clicked={dot_clicked} index={i} />);
		}

		return (
			<div className="dots">
				{dots}
			</div>
		)
	},
	"handleLeftArrow" : function()
	{
		if(this.state.selected == null)
		{
			pattern.selectDot(0);
		}
		else
		{
			this.moveDot(-1, 0);
		}
	},
	"handleRightArrow" : function()
	{
		if(this.state.selected == null)
		{
			pattern.selectDot(0);
		}
		else
		{
			this.moveDot(1, 0);
		}
	},
	"handleUpArrow" : function()
	{
		if(this.state.selected == null)
		{
			pattern.selectDot(0);
		}
		else
		{
			this.moveDot(0, -1);
		}
	},
	"handleDownArrow" : function()
	{
		if(this.state.selected == null)
		{
			pattern.selectDot(0);
		}
		else
		{
			this.moveDot(0, 1);
		}
	},
	"handleEnterKey" : function()
	{
		if(this.state.selected == null)
		{
			pattern.selectDot(0);
		}
		else
		{
			this.setState({ "draw_line" : true });
			this.selectDot(this.state.selected);
		}
	},
	"selectDot" : function(dot_index)
	{
		var sequence = this.state.sequence;

		if(this.state.draw_line)
		{
			sequence.push(dot_index);
		}
		this.setState({ "selected" : dot_index, "sequence" : sequence });

		if(this.state.sequence.length == 6 && this.state.sequence.join(",") == "1,4,7,8,5,4")
		{
			document.getElementById("pattern_container").style.display = "none";
			requestVideos(0);
		}
	},
	"calcRow" : function(dot_index)
	{
		return Math.floor(dot_index / 3); 
	},
	"calcColumn" : function(dot_index)
	{
		return Math.floor(dot_index % 3);
	},
	"moveDot" : function(column_offset, row_offset)
	{
		var current_row = this.calcRow(this.state.selected);
		var current_column = this.calcColumn(this.state.selected);

		var row = current_row + row_offset;
		var column = current_column + column_offset;

		if(row < 0)
		{
			row = 0;
		}

		if(column < 0)
		{
			column = 0;
		}

		if(row > 2)
		{
			row = 2;
		}

		if(column > 2)
		{
			column = 2;
		}

		var selected_dot = (row * 3) + column;

		this.selectDot(selected_dot);
	}
});
var Dot = React.createClass({
	"render" : function()
	{
		if(isTV())
		{
			return (
				<div className="dot" data-selected={this.props.selected} data-row={this.props.row} data-column={this.props.column} data-clicked={this.props.clicked}>

				</div>
			)

		}
		else
		{
			return (
				<div className="dot" data-selected={this.props.selected} data-row={this.props.row} data-column={this.props.column} data-clicked={this.props.clicked} onClick={this.selectThisDot}>

				</div>
			)
		}
	},
	"selectThisDot" : function()
	{
		if(pattern.state.selected == null)
		{
			pattern.selectDot(this.props.index);
		}
		pattern.setState({ "draw_line" : true });
		pattern.selectDot(this.props.index);
	}
});
