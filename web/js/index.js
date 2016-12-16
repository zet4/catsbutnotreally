document.body.onload = addElement;

function addElement() {
	var evtSource = new EventSource("/events");
	var totalCount = document.createElement("h1")
	document.body.appendChild(totalCount)
	evtSource.addEventListener("stats", function(e) {
		var obj = JSON.parse(e.data);
		totalCount.innerHTML = obj.total;
		document.body.style.backgroundImage = "url('" +  obj.last_image + "')";
	}, false);
}