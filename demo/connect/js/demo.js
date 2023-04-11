var state
var conn
var online = false

function showSystem() {
	let system = document.getElementById("system")
	system.value = ""
	system.value += "CPU Frequency:   " + state.CPUFreq + "Mhz\r\n"
	system.value += "MAC Address:     " + state.Mac + "\r\n"
	system.value += "IP Address:      " + state.Ip + "\r\n"
	system.value += "Temperature:     " + state.TempC + "(C)\r\n"
}

function showBH1750() {
	let bh1750 = document.getElementById("bh1750")
	bh1750.value = ""
	bh1750.value += "Lux:   " + state.Lux + "\r\n"
}

function show() {
	overlay = document.getElementById("overlay")
	overlay.style.display = online ? "none" : "block"
	showSystem()
	showBH1750()
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
		console.log('connect', msg)

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

