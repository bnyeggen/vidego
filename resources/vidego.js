//Effectively global
var filterEl = document.getElementById("filter");
var tableHeaderEl = document.getElementById("table_header");
var tableBodyEl = document.getElementById("table_body");

function rowIdToMovieId(s){
	return parseInt(s.substring(6,s.length))
}

//Important to sync up this, movieToTableRow, and tableRowToMovie
var fields = ["title", "director", "year", "watched", "added_date"]

function makeYearPicker(yearSelected){
	var sel = document.createElement("select");
	
	var option = document.createElement("option");
	option.value="0";
	if(yearSelected==0){
		option.selected = true;
	}
	sel.add(option);
	
	for(var i=1920; i<=new Date().getFullYear(); i++){
		var option = document.createElement("option");
		option.value = i.toString();
		option.innerHTML = i.toString();
		if(yearSelected == i){
			option.selected = true;
		}
		sel.add(option);
	}
	return sel;
}

function movieToTableRow(movie){
	var row = document.createElement("tr");
	row.setAttribute("id", "movie_" + movie.Id);

	var titleCol = document.createElement("td");
	titleCol.setAttribute("contentEditable", true);
	titleCol.innerHTML = movie.Title;

	var directorCol = document.createElement("td");
	directorCol.setAttribute("contentEditable", true);
	directorCol.innerHTML = movie.Director;
	
	var yearCol = document.createElement("td");
	yearCol.appendChild(makeYearPicker(movie.Year));
	
	var watchedCol = document.createElement("td");
	var watchedCheck = document.createElement("input");
	watchedCheck.type="checkbox";
	watchedCheck.checked = movie.Watched;
	watchedCol.appendChild(watchedCheck);

	var addedCol = document.createElement("td");
	addedCol.innerHTML = movie.Added_date.substring(0,10);
	
	var pathCol = document.createElement("td");
	pathCol.innerHTML = movie.Path
	
	row.appendChild(titleCol);
	row.appendChild(directorCol);
	row.appendChild(yearCol);
	row.appendChild(watchedCol);
	row.appendChild(addedCol);
	row.appendChild(pathCol);

	return row
}

//Approximate mapping based on displayed/embedded info
function tableRowToMovie(tr){
	var yearIdx = tr.cells[2].children[0].selectedIndex;
	return {"title": tr.cells[0].innerHTML,
			"director": tr.cells[1].innerHTML,
			"year": tr.cells[2].children[0].options[yearIdx].value,
			"watched": tr.cells[3].checked,
			"added_date": tr.cells[4].innerHTML,
			"id": rowIdToMovieId(tr.id)}
}

//Grabs initial data & renders
function renderTable(sorter) {
	var req = new XMLHttpRequest();
	req.overrideMimeType("application/json");
	req.onload = function() {
		while (tableBodyEl.hasChildNodes()) {
			tableBodyEl.removeChild(tableBodyEl.lastChild);
		}
		//Un-var this to make a global for debugging
		var jsondata = JSON.parse(this.responseText);
		if(sorter != null) {
			sorter(jsondata);
		}
		for(i=0; i<jsondata.length; i++){
			row = movieToTableRow(jsondata[i]);
			tableBodyEl.appendChild(row)
		}
	};
	req.open("GET","/json",true);
	req.send();
}
renderTable()

//Returns a function that when called, sorts the array by the field, first in
//one direction, then the other.
var sorterStates = {}
function makeSorter(col){
	sorterStates[col] = true
	return function(ar){
		ar.sort(function(i1,i2){
			return i1[col].toString().localeCompare(i2[col].toString());
		});
		if (!sorterStates[col]){
			ar.reverse();
		}
		sorterStates[col] = !sorterStates[col]
	}
}

directorSorter = makeSorter("Director");
titleSorter = makeSorter("Title");
watchedSorter = makeSorter("Watched");
yearSorter = makeSorter("Year");
addedSorter = makeSorter("Added_date");

document.getElementById("title_col").addEventListener("click",function(e){renderTable(titleSorter)});
document.getElementById("director_col").addEventListener("click",function(e){renderTable(directorSorter)});
document.getElementById("year_col").addEventListener("click",function(e){renderTable(yearSorter)});
document.getElementById("watched_col").addEventListener("click",function(e){renderTable(watchedSorter)});
document.getElementById("added_col").addEventListener("click",function(e){renderTable(addedSorter)});

function fireUpdateRequest(id, field, val){
	var uri = "/update?id=" + id + "&field=" + field + "&val=" + encodeURIComponent(val);
	var req = new XMLHttpRequest();
	req.open("PUT", uri, true);
	req.send()
}

function handleSelect(e){
	var el = e.target;
	if (el.nodeName !== "SELECT"){
		return;
	}
	var colN = el.parentNode.cellIndex;
	var field = fields[colN];
	var val = el.options[el.selectedIndex].value;
	var row = el.parentNode.parentNode
	var id = rowIdToMovieId(row.id)
	
	fireUpdateRequest(id, field, val);
}

function handleInput(e){
	var el = e.target;
	if (el.nodeName !== "INPUT"){
		return;
	}

	var colN = el.parentNode.cellIndex;
	var field = fields[colN];
	var val = (field==="watched") ? el.checked : el.value;
	var row = el.parentNode.parentNode
	var id = rowIdToMovieId(row.id)
	
	fireUpdateRequest(id, field, val);

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
	var field = fields[colN]
	var row = el.parentNode
	var id = rowIdToMovieId(row.id)
	el.innerHTML = nukeBreaks(el.textContent)
	var val = el.innerHTML
	
	fireUpdateRequest(id, field, val);
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
	var filter_txt = filterEl.value;
	//Needed for when we delete the last filter
	if(filter_txt===""){
		for(var j=0; j<tableBodyEl.rows.length; j++){
			tableBodyEl.rows[j].style.display = "";
		}
	}
	//Consider each token as separate filter to be and-joined
	splits = filter_txt.split(" ");
	for(var i=0; i<splits.length; i++){
		var term = splits[i].toLowerCase();
		if(term==="") {
			continue;
		}
		for(var j=0; j<tableBodyEl.rows.length; j++){
			var thisRow = tableBodyEl.rows[j];
			var thisMovie = tableRowToMovie(thisRow);
			if(thisMovie["title"].toLowerCase().indexOf(term)!==-1 
			|| thisMovie["director"].toLowerCase().indexOf(term)!==-1
			|| thisMovie["year"].toString().toLowerCase().indexOf(term)!==-1){
				thisRow.style.display = "";
			} else {
				thisRow.style.display = "none";
			}
		}
	}
}

//Filters table
filterEl.addEventListener("keyup", handleFilterTable);

//Handles changes to actual input elements
tableBodyEl.addEventListener("change", handleInput);
//And loss-of-focus for raw (presumably contenteditable) td elements
tableBodyEl.addEventListener("blur", handleContentEditable, true);
//And selects
tableBodyEl.addEventListener("change", handleSelect);

//Prevent "enter" from embedding newlines
tableBodyEl.addEventListener("keydown", handleTableEnterKey);

