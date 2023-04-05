var lastImg
var shown = false

function showChild(id) {
	var iframe = document.getElementById("child")
	var img = document.getElementById(id)

	iframe.src = "/" + encodeURIComponent(id) + "/"

	img.style.border = "2px dashed blue"
	if (typeof lastImg !== 'undefined') {
		if (lastImg != img) {
			lastImg.style.border = "none"
		}
	}

	lastImg = img
	shown = true
}

function iconName(child) {
	var status = "offline"
	if (child.Online) {
		status = "online"
	}
	return "images/" + status + ".jpg"
}

function newIcon(child) {
	var children = document.getElementById("children")
	var newdiv = document.createElement("div")
	var newpre = document.createElement("pre")
	var newimg = document.createElement("img")

	newpre.innerText = child.Id
	newpre.id = "pre-" + child.Id

	newimg.src = iconName(child)
	newimg.onclick = function (){showChild(child.Id)}
	newimg.id = child.Id

	newdiv.appendChild(newpre)
	newdiv.appendChild(newimg)
	children.appendChild(newdiv)
}

function addChild(child) {
	newIcon(child)
	if (!shown) {
		showChild(child.Id)
	}
}

function clearScreen() {
	var children = document.getElementById("children")
	var iframe = document.getElementById("child")

	iframe.src = ""
	while (children.firstChild) {
		children.removeChild(children.firstChild)
	}
	shown = false
}

function saveState(msg) {
	for (const id in msg.Children) {
		child = msg.Children[id]
		addChild(child)
	}
}

function update(child) {
	var img = document.getElementById(child.Id)
	var pre = document.getElementById("pre-" + child.Id)

	if (img == null) {
		addChild(child)
	} else {
		img.src = iconName(child)
	}
}

function Run(ws) {

	var conn

	function connect() {
		conn = new WebSocket(ws)

		conn.onopen = function(evt) {
			clearScreen()
			conn.send(JSON.stringify({Path: "get/state"}))
		}

		conn.onclose = function(evt) {
			clearScreen()
			setTimeout(connect, 1000)
		}

		conn.onerror = function(err) {
			conn.close()
		}

		conn.onmessage = function(evt) {
			var msg = JSON.parse(evt.data)

			console.log('hub', msg)

			switch(msg.Path) {
			case "state":
				saveState(msg)
				break
			case "connected":
			case "disconnected":
				update(msg)
				break
			}
		}
	}

	connect()
}
