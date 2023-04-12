var state
var conn
var online = false

var opts = {
	angle: -0.2, // The span of the gauge arc
	lineWidth: 0.2, // The line thickness
	radiusScale: 1, // Relative radius
	pointer: {
		length: 0.6, // // Relative to gauge radius
		strokeWidth: 0.035, // The thickness
		color: '#000000' // Fill color
	},
	limitMax: true,      // If false, max value increases automatically if value > maxValue
	limitMin: false,     // If true, the min value of the gauge will be fixed
	highDpiSupport: true,     // High resolution support
	staticZones: [
		{strokeStyle: "#30B32D", min:      0, max:  650000}, // Green
		{strokeStyle: "#FFDD00", min: 650000, max:  700000}, // Yellow
		{strokeStyle: "#F03E3E", min: 700000, max: 1000000}  // Red
	],
}
var bh1750= document.getElementById('bh1750')
var gauge = new Gauge(bh1750).setOptions(opts)

gauge.maxValue = 1000000
gauge.setMinValue(0)
gauge.animationSpeed = 32
gauge.set(0)

function showSystem() {
	let system = document.getElementById("system")
	system.value = ""
	system.value += "CPU Frequency:   " + state.CPUFreq + "Mhz\r\n"
	system.value += "MAC Address:     " + state.Mac + "\r\n"
	system.value += "IP Address:      " + state.Ip + "\r\n"
	system.value += "Temperature:     " + state.TempC + "(C)\r\n"
}

function showBH1750() {
	gauge.set(state.Lux)
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

