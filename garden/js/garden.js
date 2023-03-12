var conn
var online = false
var timeDiff

days = []
for (var i = 0; i < 7; i++) {
	days[i] = document.getElementById("day" + i)
}

startTime = document.getElementById("startTime")
gallons = document.getElementById("gallons")
gallonsGoal = document.getElementById("gallonsGoal")
startButton = document.getElementById("start")
stopButton = document.getElementById("stop")
bar = document.getElementById("bar")
dateTime = document.getElementById("date-time")

function sendDateTime() {
	now = new Date()
	conn.send(JSON.stringify({Msg: "DateTime",
		DateTime: now, ZoneOffsetMinutes: now.getTimezoneOffset()}))
}

function getState() {
	conn.send(JSON.stringify({Msg: "_GetState"}))
}

function getIdentity() {
	conn.send(JSON.stringify({Msg: "_GetIdentity"}))
}

function update(msg) {
	gallons.innerHTML = msg.Gallons.toFixed(2)
	startButton.disabled = msg.Running
	stopButton.disabled = !msg.Running
	progress = parseInt(msg.Gallons / gallonsGoal.value * 100.0)
	bar.style.width = progress + "%";
	bar.innerHTML = progress + "%";
}

function saveState(msg) {

	nowRemote = new Date(msg.Now)
	nowLocal = new Date()
	timeDiff = nowLocal - nowRemote

	startTime.value = msg.StartTime
	for (var i = 0; i < days.length; i++) {
		days[i].checked = msg.StartDays[i]
	}
	gallonsGoal.value = msg.GallonsGoal
	update(msg)
}

function saveDay(msg) {
	days[msg.Day].checked = msg.State
}

function saveStartTime(msg) {
	startTime.value = msg.Time
}

function saveGallonsGoal(msg) {
	gallonsGoal.value = msg.GallonsGoal
}

function changeStartTime() {
	conn.send(JSON.stringify({Msg: "StartTime",
		Time: startTime.value}))
}

function changeDay(box, day) {
	conn.send(JSON.stringify({Msg: "Day", Day: day,
		State: box.checked}))
}

function changeGallonsGoal() {
	conn.send(JSON.stringify({Msg: "GallonsGoal",
		GallonsGoal: parseInt(gallonsGoal.value)}))
}

function start() {
	conn.send(JSON.stringify({Msg: "Start"}))
}

function stop() {
	conn.send(JSON.stringify({Msg: "Stop"}))
}

function showNow() {
	now = new Date(new Date() - timeDiff)
	dateTime.innerHTML = now.toLocaleString('en-US', {
		weekday: 'long',
		hour: '2-digit',
		minute: '2-digit',
		timeZoneName: "short",
	});
	setTimeout('showNow()', (60 - (now.getSeconds())) * 1000)
}

function Run(ws) {

	function connect() {
		conn = new WebSocket(ws)

		conn.onopen = function(evt) {
			sendDateTime()
			getIdentity()
		}

		conn.onclose = function(evt) {
			online = false
			setTimeout(connect, 1000)
		}

		conn.onerror = function(err) {
			conn.close()
		}

		conn.onmessage = function(evt) {
			msg = JSON.parse(evt.data)
			console.log('garden', msg)

			switch(msg.Msg) {
			case "_ReplyIdentity":
				online = msg.Online
				getState()
				break
			case "_EventStatus":
				online = msg.Online
				break
			case "_ReplyState":
				saveState(msg)
				showNow()
				break
			case "Update":
				update(msg)
				break
			case "Day":
				saveDay(msg)
				break
			case "StartTime":
				saveStartTime(msg)
				break
			case "GallonsGoal":
				saveGallonsGoal(msg)
				break
			}
		}
	}

	connect()
}
