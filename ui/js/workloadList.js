patWorkload = function(workloadNode, selectedNode, commandArgsNode){

	const workloadItems = {
	"rest:target": {
		requires: [],
		requiredBy: [
			"rest:login",
			"rest:push"
		],
		args: [
			"rest:target"
		]	
	},
	"rest:login": {		
		requires: [
			"rest:target"
		],
		requiredBy: [
			"rest:push"
		],
		args: [
			"rest:username",
			"rest:password"
		]
	},
	"rest:push": {
		requires: [
			"rest:target",
			"rest:login"
		],
		requiredBy: [],
		args: []
	},
	"gcf:push": {
		requires: [],
		requiredBy: [],
		args: []	
	},
	"dummy": {
		requires: [],
		requiredBy: [],
		args: []	
	},
	"dummyWithErrors": {
		requires: [],
		requiredBy: [],
		args: []	
	}};

	const argumentItems = {
	"rest:target": {
			koBind: "",
			errCheckFn: "",
			regex: /^(?:https?:\/\/(?:www\.)?|www\.)[a-z0-9]+(?:[-.][a-z0-9]+)*\.[a-z]{2,5}(?::[0-9]{1,5})?(?:\/\S*)?$/,
			requiredBy: ["rest:target"],
			default: "http://xio.10.xx.xx.xx"
	},
	"rest:username": {
			koBind: "",
			errCheckFn: "",
			regex: /^[a-zA-Z0-9-_,]+$/,
			default: "admin",
			requiredBy: ["rest:login"]			
	},
	"rest:password": {
			koBind: "",
			errCheckFn: "",
			regex: /^[a-zA-Z0-9-_,]+$/,
			default: "admin",
			requiredBy: ["rest:login"]
	}};

	var workloadBtns  = {};
	var commandArgs = {};
	var selectedWorkloads  = [];
	var workloadsObservableFn;

	$(workloadNode).html('')
	$(selectedNode).html('')
	$(commandArgsNode).html('')

	// draw workload buttons
	for (var cmd in workloadItems) {
		workloadBtns[cmd] = $("<button>", {
			"type": "button",
			"id": "workloadItem-" + cmd,
			"click": function(e) { addWorkloadBtn(e) }, 
			"class": "btn btn-default",			
			"html": '<span class="glyphicon glyphicon-plus-sign"></span> ' + cmd
		})									
		workloadBtns[cmd].appendTo($(workloadNode))
		$(workloadNode).append(" ")		
	}
	
	var workloads = function() {
		var str = ""
		selectedWorkloads.forEach (function(d) {			
			str = str + ((str == "")? d.cmd : "," + d.cmd)
		})		
		return str;
	}

	var importKoBindingsVars = function(target, targetErrFn, username, usernameErrFn, password, passwordErrFn) {
		argumentItems["rest:target"].koBind = target;
		argumentItems["rest:username"].koBind = username;
		argumentItems["rest:password"].koBind = password;
		argumentItems["rest:target"].errCheckFn = targetErrFn;
		argumentItems["rest:username"].errCheckFn = usernameErrFn;
		argumentItems["rest:password"].errCheckFn = passwordErrFn;
		drawArgumentInputs()
	}

	var isArgumentError = function(arg, value) {
		if (commandArgs[arg].css("display") == "none") return false;
		
		if (value.trim() == "") return true;

		return !(argumentItems[arg].regex.test(value))
	}

	var workloadsObservable = function(ob) {
		workloadsObservableFn = ob;
	}

	return {		
		workloads: workloads,		
		workloadsObservable: workloadsObservable,
		importKoBindingsVars: importKoBindingsVars,
		isArgumentError: isArgumentError,
		workloadItems: workloadItems
	};	

	function drawArgumentInputs() {
		for (var cmd in argumentItems) {
			var input = $("<div>", {
					"class": "col-sm-8",
					"html" : '<input type="text" class="form-control" data-bind="value: ' + argumentItems[cmd].koBind + '">'
				})
				var label = $("<label>", {					
					"class" : "col-sm-3 control-label",
					"html"  : cmd
				})
				commandArgs[cmd] = $("<div>", {
					"class" : "form-group",
					"data-bind": "css: { 'has-error': " + argumentItems[cmd].errCheckFn + " }",
					"style": "display: none"
				})

				label.appendTo(commandArgs[cmd])
				input.appendTo(commandArgs[cmd])
				commandArgs[cmd].appendTo($(commandArgsNode))
		}
	}

	function addWorkloadBtn(e) {
		var cmd = (e.target.innerText.trim()=="")? e.target.parentNode.innerText.trim() : e.target.innerText.trim();
		// first add the required commands
		workloadItems[cmd].requires.forEach(function(d){
			if (findIndexByCommand(d) == -1) appendWorkloadBtnToDom(d)
		})
		appendWorkloadBtnToDom(cmd)
	}

	function appendWorkloadBtnToDom(cmd) {
		var btn = {};				
		btn = $("<button>", {
			"type" : "button", 			
			"click": function(e) { removeWorkloadItem(btn) },
			"class": "btn btn-link " + cmd,
			"html" : '<span class="glyphicon glyphicon-minus-sign"></span> ' + cmd
		});
		btn.cmd = cmd;					
		btn.appendTo( $(selectedNode) );
		selectedWorkloads.push(btn);

		workloadItems[cmd].args.forEach(function(d) {
			commandArgs[d].css("display", "inherit")
		})

		if (workloadsObservableFn) workloadsObservableFn(workloads());
	}

	function addCommandArgs(cmd) {
		if (commandArgs[cmd] != null) return;
			var input = $("<div>", {
				"class": "col-sm-8",
				"html" : '<input type="text" placeHolder="' + cmd + '" class="form-control" data-bind="value: cfTarget">'
			})
			var label = $("<label>", {
				"for"   : "args:" + cmd,
				"class" : "col-sm-3 control-label",
				"html"  : cmd
			})
			commandArgs[cmd] = $("<div>", {
				"class" : "form-group",
				"data-bind": "css: { 'has-error': targetHasError }"
			})

			label.appendTo(commandArgs[cmd])
			input.appendTo(commandArgs[cmd])
			commandArgs[cmd].appendTo($(commandArgsNode))
	}

	function removeWorkloadItem(btn) {
		var warning = "";
		var duplicatedCmd = (findIndexByCommand(btn.cmd, 0, findIndexByElement(btn)) == -1)? false : true;
		
		//first find any requiredBy commands
		if (!duplicatedCmd) {
			workloadItems[btn.cmd].requiredBy.forEach(function(d) {
				if (findIndexByCommand(d, findIndexByElement(btn)) != -1) {
					warning = (warning == "")? d : warning + ", " + d;
				}
			})			
		}

		if (warning != "") {
			alert("Cannot remove item\nIt is required by [" + warning + "]");
			return;
		} 

		var index = findIndexByElement(btn);
		selectedWorkloads[index].remove();
		selectedWorkloads.splice(index, 1);

		if (!duplicatedCmd) {
			//need to hide required arguments when item is removed				
			workloadItems[btn.cmd].args.forEach(function(d) {
				commandArgs[d].css("display", "none")
			})
		}

		if (workloadsObservableFn) workloadsObservableFn(workloads());
	}

	function findIndexByElement (btn) {
		for (var i=0; i < selectedWorkloads.length; i++) {
			if (selectedWorkloads[i] === btn) return i;
		}
		return -1;
	}

	function findIndexByCommand (cmd, start, end) {
		start = (start >= 0)? start : 0;
		end = ( end >= 0)? end : selectedWorkloads.length;
		for (var i = start; i < end; i++) {
			if (selectedWorkloads[i].cmd == cmd) return i;
		}
		return -1;
	}
	
}


