var state
var conn
var online = false

function showSystem() {
	var system = document.getElementById("system")
	system.value = ""
	system.value += "CPU Frequency: " + state.Freq
	system.value += "MAC Address: " + state.Mac
	system.value += "IP Address: " + state.Ip
}

function show() {
	overlay = document.getElementById("overlay")
	overlay.style.display = online ? "none" : "block"
	showSystem()
}

function run(ws) {

	console.log('connecting...')
	conn = new WebSocket(ws)

	conn.onopen = function(evt) {
		console.log("open")
		conn.send(JSON.stringify({Path: "get/state"}))
	}

	conn.onclose = function(evt) {
		console.log("close")
		online = false
		show()
		setTimeout(run(ws), 1000)
	}

	conn.onerror = function(err) {
		console.log("error", err)
		conn.close()
	}

	conn.onmessage = function(evt) {
		msg = JSON.parse(evt.data)
		console.log('garden', msg)

		switch(msg.Path) {
		case "state":
			online = true
			// fall-thru
		case "update":
			state = msg
			show()
			break
		}
	}
}

