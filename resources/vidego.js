//Effectively global
var filterEl = document.getElementById("filter");
var tableBodyEl = document.getElementById("table_body");

function escapeHtml(unsafe) {
	return unsafe
		.replace(/&/g, "&amp;")
		.replace(/</g, "&lt;")
		.replace(/>/g, "&gt;")
		.replace(/"/g, "&quot;")
		.replace(/'/g, "&#039;");
 }

//*****/
//MODEL/
//*****/

/*This is actually the most straightforward part - we don't really have a lot
  of "model" besides an array of POJOs.*/

//Constructor.
function MovieModel(){
	if(false === (this instanceof MovieModel)) {
        return new MovieModel();
    }
	this.fetchDataSync()
}
MovieModel.prototype.data = null
//Synchronous; unclear if async logic is worthwhile to maintain when this is 
//the core dependency of any other operation
MovieModel.prototype.fetchDataSync = function(){
	var that = this;
	var req = new XMLHttpRequest();
	req.overrideMimeType("application/json");
	req.onload = function(){
		that.data = JSON.parse(this.responseText);
	}
	req.open("GET", "/json", false);
	req.send()
}

MovieModel.prototype.updateLocalData = function(id, field, val){
	for(var i=0; i<this.data.length; i++){
		var row = this.data[i]
		if(row.Id == id){
			row[field] = val;
			return;
		}
	}
}

//For filtering, only looks at semantically meaningful entries
//Ideally this would be a method of a Movie object, but then we have to use
//Movie objects rather than passing the raw JSON-decoded objects
function movieContainsTxt(movie, filterTxt){
	if(filterTxt==="") {
		return true;
	}
	var allTerms = filterTxt.split(" ");
	for(var i=0; i<allTerms.length; i++){
		var term = allTerms[i].trim();
		if(term===""){
			continue;
		}
		if(movie.Title.toLowerCase().indexOf(term) === -1
		&& movie.Director.toLowerCase().indexOf(term) === -1
		&& movie.Year.toString().indexOf(term) === -1){
			return false;
		}
	}
	return true;
}

//****/
//VIEW/
//****/

/*This gets a bit conceptually tricky.  Essentially this is the stuff that
  generates the underlying DOM elements, and also symmetrically represents the
  correspondence between rendered DOM elements and the backing data (eg, which
  column corresponds to what field of the movie objects).  In that sense it's
  more of a ViewModel, with the classical View being the DOM itself.

  OTOH, it doesn't really declare the actual mechanisms for handling input /
  view modifications and propogating them back to the model or server(and
  thence to the view again) or modifying the view directly (eg, by sorting),
  just the metadata necessary to do so through other means.  In that sense it
  really is more of a view layer with a separate controller, just one that
  actually needs to represent some state because of the crappiness of the DOM.
*/

//Renders a movie as a single tr element - conceptually part of view object,
//but doesn't depend on any state besides that of its argument.
//If we were librarifying this, this would check the Model for column
//definitions and render as appropriate, but each of these columns has pretty
//special handling so we do it manually.  The order should correspond w/ the
//column layout declared in the view protoype though.
function movieToTableRow(movie){
	var row = document.createElement("tr");
	row.setAttribute("id", "movie_" + movie.Id);

	var titleCol = document.createElement("td");
	titleCol.setAttribute("contentEditable", true);
	titleCol.innerHTML = movie.Title;
	//titleCol.classList.add("title_col");

	var directorCol = document.createElement("td");
	directorCol.setAttribute("contentEditable", true);
	directorCol.innerHTML = movie.Director;
	//directorCol.classList.add("director_col");
	
	//Right now FF 28 is the only modern browser that has issues w/ this, and
	//the fix is coming in FF 29.
	var yearCol = document.createElement("td");
	var yearInp = document.createElement("input");
	yearInp.type = "number";
	yearInp.min = "1920";
	yearInp.max = "2014";
	if (movie.Year != 0) {
		yearInp.defaultValue = movie.Year.toString();
	}
	//yearInp.value = movie.Year;
	yearCol.appendChild(yearInp);

	var watchedCol = document.createElement("td");
	var watchedCheck = document.createElement("input");
	watchedCheck.type="checkbox";
	//If we just set checked = true, it doesn't affect the HTML
	if(movie.Watched){
		watchedCheck.setAttribute("checked","checked");
	}
	watchedCol.appendChild(watchedCheck);

	var addedCol = document.createElement("td");
	addedCol.innerHTML = movie.Added_date.substring(0,10);
	
	var pathCol = document.createElement("td");
	var asLink = document.createElement("a");
	asLink.href = "/movies" + movie.Path;
	asLink.innerHTML = escapeHtml(movie.Path);
	pathCol.appendChild(asLink);
	
	row.appendChild(titleCol);
	row.appendChild(directorCol);
	row.appendChild(yearCol);
	row.appendChild(watchedCol);
	row.appendChild(addedCol);
	row.appendChild(pathCol);

	return row
}

//Extract numeric ID from text ID of row DOM element
function rowIdToMovieId(s){
	return parseInt(s.substring(6,s.length))
}

//Constructor
function MovieView(model, parentElement){
	if(false === (this instanceof MovieView)) {
        return new MovieView(model, parentElement);
    }
	this.model = model;
	this.parentElement = parentElement;
	this.filteredRows = [];
	for(var i=0; i<model.data.length; i++){
		this.filteredRows.push(i);
	}
	this.sorterStates = {}
}
//We check this in the controller when we're deciding what to send to the server, and when attaching sort handlers
//In the future this could turn into enhanced metadata about the type & rendering of the column
MovieView.prototype.fieldLayout = ["Title", "Director", "Year", "Watched", "Added_date", "Path"]
//Source of backing data
MovieView.prototype.model = null;
//This element has its innerHTML replaced on render
MovieView.prototype.parentElement = null;
//Array of indexes in the model data that are actually valid according to the filter
MovieView.prototype.filteredRows = null;
//The prior filter to be applied
MovieView.prototype.prevFilterTxt = "";
//The count of rows that have presently been attached to the DOM, for managing lazy render
MovieView.prototype.renderedRows = 0;

//Method to filter internal list of rows that pass filter
MovieView.prototype.filter = function(filterTxt){
	var filterTxtLC = filterTxt.toLowerCase();
	var newFilteredRows = [];
	//Avoids re-computing on arrows, etc.
	if(filterTxtLC === this.prevFilterTxt){
		return;
	}
	//Is filter purely additive?
	if(filterTxtLC.indexOf(this.prevFilterTxt) !== -1){
		//If so, filter in place
		for(var i=0; i<this.filteredRows.length; i++){
			var movieIdx = this.filteredRows[i];
			var thisMovie = this.model.data[movieIdx];
			if(movieContainsTxt(thisMovie,filterTxtLC)){
				newFilteredRows.push(movieIdx);
			}
		}
	} else {
		//Otherwise, filter whole list
		for(var i=0; i<this.model.data.length; i++){
			var thisMovie = this.model.data[i]
			if(movieContainsTxt(thisMovie,filterTxtLC)){
				newFilteredRows.push(i);
			}
		}
	}
	//Update present filter
	this.filteredRows = newFilteredRows;
	//Record filter
	this.prevFilterTxt = filterTxtLC;
	//And re-render
	this.renderFromScratch()
}

//Map from "column" name in the model data to a boolean indicating last sort order
//This allows you to alternate between ascending and descending
MovieView.prototype.sorterStates = null;
MovieView.prototype.sortModelDataBy = function(sortCol){
	if(sortCol != null){
		if(!(sortCol in this.sorterStates)){
			this.sorterStates[sortCol] = true;
		}
		var data = this.model.data;

		//We only need to sort our local filtered index list, since that's what we actually render
		this.filteredRows.sort(function(i1, i2){
			var c1 = data[i1][sortCol];
			var c2 = data[i2][sortCol];
			return (c1<c2?-1:(c1>c2?1:0));
		});

		//Reverse if we're inverse sorting
		if(!this.sorterStates[sortCol]){
			this.filteredRows.reverse();
		}

		this.sorterStates[sortCol] = !this.sorterStates[sortCol];
	}
}

//Blow away all previously rendered rows, and start from scratch
MovieView.prototype.renderFromScratch = function(){	
	var allRowHTML = "";
	var i=0;
	//Render until we have >= N rows and enough to fill the screen, or until we run out of rows
	while(i < this.filteredRows.length){
		var rowData = this.model.data[this.filteredRows[i]];
		allRowHTML += movieToTableRow(rowData).outerHTML;
		i++;
	}
	this.renderedRows = i;
	this.parentElement.innerHTML = allRowHTML;
}

//*************/
//INSTANTIATION
//*************/

/*Get whatever we actually need to attach controllers to*/


function scan(){
	var req = new XMLHttpRequest();
	req.open("GET", "/scan", false);
	req.send()
}

scan();

var attachmentPoint = document.getElementById("table_body");
//Also fetches initial data
var mainModel       = new MovieModel();
var mainView        = new MovieView(mainModel, attachmentPoint);
var filterElement   = document.getElementById("filter");

//**********/
//CONTROLLER
//**********/

/*This is mostly event handling.  As mentioned above, we rely of the view to
  maintain the correspondence between data layout & visual layout*/

//To generate closure
function makeSorter(col){
	return function(e){
		mainView.sortModelDataBy(col);
		mainView.renderFromScratch();
	}
}

//Columns sort on click
for(var i=0; i<mainView.fieldLayout.length; i++){
	var col = mainView.fieldLayout[i];
	document.getElementById(col+"_header").addEventListener("click", makeSorter(col));
}

//Conform to backend API
function fireUpdateRequest(id, field, val){
	//Update backend
	var uri = "/update?id=" + id + "&field=" + field.toLowerCase() + "&val=" + encodeURIComponent(val);
	var req = new XMLHttpRequest();
	req.open("PUT", uri, true);
	req.send()
}

function handleInput(e){
	var el = e.target;
	if (el.nodeName !== "INPUT"){
		return;
	}
	var colN = el.parentNode.cellIndex;
	var field = mainView.fieldLayout[colN];
	var val = (field==="Watched") ? el.checked : el.value;
	var row = el.parentNode.parentNode
	var id = rowIdToMovieId(row.id)
	fireUpdateRequest(id, field, val);
	mainModel.updateLocalData(id, field, val);
}

function nukeBreaks(s){
	var re = /\s*<br.*>(?:\s*<\/br\s*>)?\s*/
	return s.replace(re, "")
}

function handleContentEditable(e){
	//Only fire on td level; this also prevents embedded inputs from firing
	var el = e.target;
	if (el.nodeName !== "TD"){
		return;
	}
	var colN = el.cellIndex
	var field = mainView.fieldLayout[colN]
	var row = el.parentNode
	var id = rowIdToMovieId(row.id)
	el.innerHTML = nukeBreaks(el.textContent)
	var val = el.innerHTML
	
	fireUpdateRequest(id, field, val);
	mainModel.updateLocalData(id, field, val);
}

function handleTableEnterKey(e){
	if (e.which === 13){
		//Prevents embedding newlines
		e.preventDefault()
		//Triggers blur event, naturally
		e.target.blur()
	}
}

function handleFilterTable(e){
	var filterTxt = filterEl.value;
	mainView.filter(filterTxt);
}

//Filters table
filterEl.addEventListener("keyup", handleFilterTable);

//Handles changes to actual input elements
attachmentPoint.addEventListener("change", handleInput);
//And loss-of-focus for raw (presumably contenteditable) td elements
attachmentPoint.addEventListener("blur", handleContentEditable, true);

//Prevent "enter" from embedding newlines
attachmentPoint.addEventListener("keydown", handleTableEnterKey);

//Sort by Added_date descending
mainModel.data.sort(function(i1, i2){
			var c1 = i1["Added_date"];
			var c2 = i2["Added_date"];
			return (c1<c2?-1:(c1>c2?1:0));
		});

mainModel.data.reverse();
mainView.renderFromScratch();
