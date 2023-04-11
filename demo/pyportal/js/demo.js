var state
var conn
var online = false

function neo(letter, value) {
	state.NeoColor[letter] = parseInt(value)
	conn.send(JSON.stringify({Path: "neo", NeoColor: state.NeoColor}))
}

function showSystem() {
	let system = document.getElementById("system")
	system.value = ""
	system.value += "CPU Frequency:   " + state.CPUFreq + "Mhz\r\n"
	system.value += "MAC Address:     " + state.Mac + "\r\n"
	system.value += "IP Address:      " + state.Ip + "\r\n"
	system.value += "Light Intensity: " + state.Light + "\r\n"
}

function showNeoPixel() {
	let circle = document.getElementById("neopixel");
	let r = state.NeoColor["R"]
	let g = state.NeoColor["G"]
	let b = state.NeoColor["B"]
	let a = state.NeoColor["A"] / 255.0
	let colorString = "rgba(" + r + "," + g + "," + b + "," + a + ")";
	circle.setAttribute("fill", colorString);
}

function showNeo() {
	let neoR = document.getElementById("neoR")
	let neoG = document.getElementById("neoG")
	let neoB = document.getElementById("neoB")
	neoR.value = state.NeoColor["R"]
	neoG.value = state.NeoColor["G"]
	neoB.value = state.NeoColor["B"]
	showNeoPixel()
}

function show() {
	overlay = document.getElementById("overlay")
	overlay.style.display = online ? "none" : "block"
	showSystem()
	showNeo()
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

