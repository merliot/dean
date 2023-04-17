var state
var conn
var online = false

function showSystem() {
	let system = document.getElementById("system")
	system.value = ""
	system.value += "CPU Frequency:   " + state.CPUFreq + "Mhz\r\n"
	system.value += "MAC Address:     " + state.Mac + "\r\n"
	system.value += "IP Address:      " + state.Ip
}

function showRx() {
	let rx = document.getElementById("rx")
	rx.value = ""
	rx.value += "Last Received:      " + state.Rx
}

function show() {
	overlay = document.getElementById("overlay")
	overlay.style.display = online ? "none" : "block"
	showSystem()
	showRx()
}

function reset() {
	conn.send(JSON.stringify({Path: "reset"}))
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
		console.log('matrix', msg)

		switch(msg.Path) {
		case "state":
			online = true
			state = msg
			show()
			break
		case "rx":
			state.Packet = msg.Packet
			showRx()
			break
		}
	}
}

