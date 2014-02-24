var DOM = function() {

	var d3Chart;
	
	var contentIn = function(node) {		
		contentOut();
		var el = node;
		el.style.display="block";
		el.style.position = "relative";		
		el.parentNode.appendChild(node);
		var e = node;
		e.className = "slide_content contentIn";
		d3Chart = node;

		function contentOut() {				
			var e = d3Chart;
			e.style.position = "absolute";
			e.className = "contentOut slide_content";
			e.style.display = "none";
		}
	}

	var hideContent = function(node) {				
		$(node).css('display', 'none')
	}

	var showGraph = function(node) {				
		$(node).css('display', 'block')
		d3Chart = node
	}
	
	return {
		contentIn: contentIn,
		hideContent: hideContent,
		showGraph: showGraph
	}
}