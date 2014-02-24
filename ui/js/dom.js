var DOM = function() {

	var d3Chart = "d3_workload_container";
	
	var contentIn = function(id) {	
		contentOut();
		var el = document.getElementById(id);
		el.style.display="block";
		el.style.position = "relative";		
		el.parentNode.appendChild(document.getElementById(id));
		var e = document.getElementById(id);
		e.className = "slide_content contentIn";
		d3Chart = id;

		function contentOut() {
			var e = document.getElementById(d3Chart);
			e.style.position = "absolute";
			e.className = "contentOut slide_content";
		}
	}

	var hideContent = function(id) {
		var e;
		id.forEach(function(d){
			e = document.getElementById(d);		
			e.className = "slide_content";
		});	
	}
	
	return {
		contentIn: contentIn,
		hideContent: hideContent
	}
}