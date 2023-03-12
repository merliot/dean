log = document.getElementById("log")
child = document.getElementById("child")
stateBtn = document.getElementById("stateBtn")
overlay = document.getElementById("overlay")

var childId = ""

function state() {
	if (stateBtn.value == "Show State") {
		stateBtn.value = "Show UI"
	} else {
		stateBtn.value = "Show State"
	}
	showChild()
}

function showChild() {
	if (childId == "") {
		child.src = ""
		return
	}

	if (stateBtn.value == "Show State") {
		child.src = "/" + encodeURIComponent(childId)
	} else {
		child.src = "/" + encodeURIComponent(childId) + "/state"
	}
}

function clearScreen() {
	childId = ""
	showChild()
	log.innerHTML = ""
}

function saveState(msg) {
	childId = msg.ChildId
	showChild()
}

function update(msg) {
	if (msg.Online) {
		saveState(msg)
	} else {
		clearScreen()
	}
}

function Run(ws) {

	var conn

	function connect() {
		conn = new WebSocket(ws)

		conn.onopen = function(evt) {
			clearScreen()
			conn.send(JSON.stringify({Msg: "_GetState"}))
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

			console.log('garden_demo', msg)
			log.innerHTML += JSON.stringify(msg) + '\r\n'
			log.scrollTop = log.scrollHeight

			switch(msg.Msg) {
			case "_ReplyState":
				saveState(msg)
				break
			case "_EventStatus":
				update(msg)
				break
			case "Start":
				overlay.style.display = "block"
				break
			case "Stop":
				overlay.style.display = "none"
				break
			}
		}
	}

	connect()
}
