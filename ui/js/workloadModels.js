patWorkload = function(){

	// workload declaration
	var restTarget = {
		name: "rest:target",
		html: "<span class='glyphicon glyphicon-plus-sign'></span> rest:target",
		requires: [],
		requiredBy: ["rest:login", "rest:push"],
		args: ["rest:target"],
		click: workloadClick
	}

	var restLogin = {
		name: "rest:login",
		html: "<span class='glyphicon glyphicon-plus-sign'></span> rest:login",
		requires: ["rest:target"],
		requiredBy: ["rest:push"],
		args: ["rest:username",	"rest:password", "rest:space"],
		click: workloadClick
	}

	var restPush = {
		name: "rest:push",
		html: "<span class='glyphicon glyphicon-plus-sign'></span> rest:push",
		requires: ["rest:target", "rest:login"],
		requiredBy: [],
		args: [],
		click: workloadClick
	}

	var gcfPush = {
		name: "gcf:push",
		html: "<span class='glyphicon glyphicon-plus-sign'></span> gcf:push",
		requires: [],
		requiredBy: [],
		args: [],
		click: workloadClick
	}

	var dummy = {
		name: "dummy",
		html: "<span class='glyphicon glyphicon-plus-sign'></span> dummy",
		requires: [],
		requiredBy: [],
		args: [],
		click: workloadClick
	}

	var dummyWithErrors = {
		name: "dummyWithErrors",
		html: "<span class='glyphicon glyphicon-plus-sign'></span> dummyWithErrors",
		requires: [],
		requiredBy: [],
		args: [],
		click: workloadClick
	}

	var selectedItem = {
		name: "",
		html: "<span class='glyphicon glyphicon-minus-sign'></span> ",
		click: null
	}

	// Argument declaration
	var cfTarget =
	{	argName: "CF Target",
		forCmd: "rest:target",
		value: ko.observable("http://api.example.com"),			
		display: ko.observable("none"),
		regex: /^(?:https?:\/\/(?:www\.)?|www\.)[a-z0-9]+(?:[-.][a-z0-9]+)*\.[a-z]{2,5}(?::[0-9]{1,5})?(?:\/\S*)?$/,
		requiredBy: ["rest:target"],
	}
	
	var cfUser = 
	{	argName: "CF Username",
		forCmd: "rest:username",
		value: ko.observable("cfUser"),			
		display: ko.observable("none"),
		regex: /^[a-zA-Z0-9-_,@.]+$/,
		requiredBy: ["rest:login"]			
	}

	var cfPass =
	{	argName: "CF Password",
		forCmd: "rest:password",
		value: ko.observable("cfPass"),			
		display: ko.observable("none"),
		regex: /^[a-zA-Z0-9-_,]+$/,
		requiredBy: ["rest:login"]
	}

	var cfSpace = 
 	{	argName: "CF Space",
 		forCmd: "rest:space",
 		value: ko.observable("dev"),			
 		display: ko.observable("none"),
 		regex: /^[a-zA-Z0-9-_,]+$/,
 		requiredBy: ["rest:login"]
 	}

	function workloadClick() {
		var self = this
		
		// add command dependencies
		self.requires.forEach(function(d){
			if (findSelectedItemIndex(d) == -1) {							
				selectedCmds.push(generateSelectedNode(d));					
 				displayNeededArgument(availableCmds()[findWorkloadItemIndex(d)]);
			}
		})
		
		selectedCmds.push(generateSelectedNode(self.name));
		displayNeededArgument(self);

		updateWorkloadStr();
	}

	function selectedClick() {
		var self = this
		var warning = ""
		var duplicatedCmd = (findSelectedItemIndex(this.name, 0, selectedCmds.indexOf(self)) == -1)? false : true;

		var tmpItem = {}

		if (!duplicatedCmd){
			tmpItem = JSON.parse(JSON.stringify(availableCmds()[findWorkloadItemIndex(self.name)]));			

			tmpItem.requiredBy.forEach(function(d) {				
				if (findSelectedItemIndex(d, selectedCmds.indexOf(self)) != -1) {
					warning = (warning == "")? d : warning + ", " + d;							
				}
			})
		}

		if (warning != "") {
			alert("Cannot remove item\nIt is required by [" + warning + "]");
			return;
		} 

		if (!duplicatedCmd){								
			removeUnneededArgument(availableCmds()[findSelectedItemIndex(self.name)])			
		}
		selectedCmds.remove(self)			

		updateWorkloadStr();
	}

	function displayNeededArgument(node) {
		node.args.forEach(function(cmd) {			
			reqArguments()[findArgumentIndex(cmd)].display("inherit")		
		});
	}

	function removeUnneededArgument(node) {			
		node.args.forEach(function(cmd) {			
			reqArguments()[findArgumentIndex(cmd)].display("none")
		});
	}

	function findArgumentIndex(cmd) {			
		for (var i = 0; i < reqArguments().length; i++) {
			if (reqArguments()[i].forCmd == cmd) {
				return i;					
			}
		}
	}

	function generateSelectedNode(cmd) {
		var node = JSON.parse(JSON.stringify(selectedItem));
		node.name = cmd;
		node.html = node.html + cmd;
		node.click = selectedClick
		return node
	}

	function findWorkloadItemIndex (cmd, start, end) {
		start = (start >= 0)? start : 0;
		end = ( end >= 0)? end : availableCmds().length;
		for (var i = start; i < end; i++) {
			if (availableCmds()[i].name == cmd) return i;
		}
		return -1;
	}

	function findSelectedItemIndex (cmd, start, end) {
		start = (start >= 0)? start : 0;
		end = ( end >= 0)? end : selectedCmds().length;
		for (var i = start; i < end; i++) {
			if (selectedCmds()[i].name == cmd) return i;
		}
		return -1;
	}

	function updateWorkloadStr() {
		var str = ""
		selectedCmds().forEach (function(d) {
			str = str + ((str == "")? d.name : "," + d.name)
		})
		workloadStr(str)
	}

	function isTargetError() {
		if (cfTarget.display() == "none") return false;
		var value = cfTarget.value();			
		if (value.trim() == "") return true;
		return !(cfTarget.regex.test(value))
	}
	function isUsernameError() {
		if (cfUser.display() == "none") return false;
		var value = cfUser.value();			
		if (value.trim() == "") return true;
		return !(cfUser.regex.test(value))
	}
	function isPasswordError() {
		if (cfPass.display() == "none") return false;
		var value = cfPass.value();			
		if (value.trim() == "") return true;
		return !(cfPass.regex.test(value))
	}

	function isSpaceError() {
 		if (cfSpace.display() == "none") return false;
 		var value = cfSpace.value();			
 		if (value.trim() == "") return true;
 		return !(cfSpace.regex.test(value))
 	}

	var workloadStr = ko.observable("");
	var worklistHasError = ko.computed( function() {
			return (workloadStr() == "")
		}
	);

	cfTarget.errCheckFn = ko.computed( isTargetError )
 	cfUser.errCheckFn = ko.computed( isUsernameError )
 	cfPass.errCheckFn = ko.computed( isPasswordError )
 	cfSpace.errCheckFn = ko.computed( isSpaceError )

 	var validation = {} 	
 	validation.HasError = function() {
 		return (cfTarget.errCheckFn() || cfUser.errCheckFn() || cfPass.errCheckFn() || cfSpace.errCheckFn() || worklistHasError())
 	}

	// construct models
	var availableCmds = ko.observableArray([restTarget, restLogin, restPush, gcfPush, dummy, dummyWithErrors])
	var selectedCmds = ko.observableArray([])
	var reqArguments = ko.observableArray([cfTarget, cfUser, cfPass, cfSpace]);

	var showSelectedCaption = ko.computed(function(){
		if (selectedCmds().length > 0) {
			return true
		} else {
			return false
		}
	})

	var showArgumentCaption = ko.computed(function(){			
		for (var i = 0; i < reqArguments().length; i ++) {
			if (reqArguments()[i].display() != "none") return true
		}
		return false
	})

	return {
		workloads: workloadStr,
		availableCmds: availableCmds,
		selectedCmds: selectedCmds,
		shouldShowSelectedCaption: showSelectedCaption,
		shouldShowArgumentCaption: showArgumentCaption,		
		reqArguments: reqArguments,
		validation: validation,
		cfTarget: reqArguments()[0].value,
		cfUsername: reqArguments()[1].value,
		cfPassword: reqArguments()[2].value,
		cfSpace: reqArguments()[3].value
	}
	
}

