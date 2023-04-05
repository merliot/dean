var state
var conn
var online = false

function sendUpdate(path) {
	state.Path = path
	conn.send(JSON.stringify(state))
}

function makeDiv(cls, id, parent) {
	div = document.createElement('div')
	div.setAttribute('class', cls)
	div.setAttribute('id', id)
	parent.appendChild(div)
	return div
}

function makeP(cls, id, text, parent) {
	p = document.createElement('p')
	p.setAttribute('id', id)
	p.textContent = text
	parent.appendChild(p)
	return p
}

function makeInput(cls, id, parent, f) {
	input = document.createElement('input')
	input.setAttribute('class', cls)
	input.setAttribute('id', id)
	input.onchange = f
	parent.appendChild(input)
	return input
}

function makeInputTime(cls, id, parent, f) {
	input = makeInput(cls, id, parent, f)
	input.type = "time"
	input.onchange = f
	return input
}

function makeButton(cls, id, text, parent, f) {
	button = document.createElement('button')
	button.setAttribute('class', cls)
	button.setAttribute('id', id)
	button.textContent = text
	button.onclick = f
	parent.appendChild(button)
	return button
}

function makeCheckbox(cls, id, text, parent, f) {
	cb = makeInput(cls, id, parent, f)
	cb.type = "checkbox"
	label = document.createElement('label')
	label.for = id
	label.textContent = text
	parent.appendChild(label)
}

const all = document.getElementById('all')

function makeZone(zone, i) {
	zoneDiv = makeDiv('box', `zone-${i}`, all)

	zoneHdrDiv = makeDiv('heading', '', zoneDiv)
	makeP('zone-name', `zone-name-${i}`, '', zoneHdrDiv)
	zoneProgressDiv = makeDiv('progress', '', zoneHdrDiv)
	makeDiv('progress-bar', `zone-progress-bar-${i}`, zoneProgressDiv)

	zoneGGDiv = makeDiv('heading line', '', zoneDiv)
	makeP('', '', 'Goal Gallons', zoneGGDiv)
	makeInput('gallons-goal', `zone-goal-gallons-${i}`, zoneGGDiv,
		function () {
			state.Zones[i].GallonsGoal = parseInt(this.value)
			sendUpdate("update")
		})

	zoneRTMDiv = makeDiv('heading', '', zoneDiv)
	makeP('', '', 'Maximum Run Time (minutes)', zoneRTMDiv)
	makeInput('running-time-max', `zone-running-time-max-${i}`, zoneRTMDiv,
		function () {
			state.Zones[i].RunningTimeMax = parseInt(this.value)
			sendUpdate("update")
		})

	zoneManualDiv = makeDiv('heading line', '', zoneDiv)
	makeP('', '', 'Manual', zoneManualDiv)
	zoneManualDivDiv = makeDiv('buttons', '', zoneManualDiv)
	makeButton('start', `zone-start-${i}`, 'Start', zoneManualDivDiv,
		function () {
			conn.send(JSON.stringify({Path: "startzone", Zone: i}))
		})
	makeButton('stop', `zone-stop-${i}`, 'Stop', zoneManualDivDiv,
		function () {
			conn.send(JSON.stringify({Path: "stopzone", Zone: i}))
		})
}

function showZones() {
	state.Zones.forEach((zone, i) => {
		zoneDiv = document.getElementById(`zone-${i}`)
		if (!zoneDiv) {
			makeZone(zone, i)
		}

		zoneName = document.getElementById(`zone-name-${i}`)
		zoneName.textContent = `${zone.Name}`

		input = document.getElementById(`zone-goal-gallons-${i}`)
		input.value = zone.GallonsGoal

		input = document.getElementById(`zone-running-time-max-${i}`)
		input.value = zone.RunningTimeMax

		button = document.getElementById(`zone-start-${i}`)
		button.disabled = zone.Running
		button = document.getElementById(`zone-stop-${i}`)
		button.disabled = !zone.Running

		progress = 0
		if (zone.GallonsGoal > 0) {
			progress = parseInt(zone.GallonsSoFar * 100 / zone.GallonsGoal)
			if (progress > 100) {
				progress = 100
			}
		}
		bar = document.getElementById(`zone-progress-bar-${i}`)
		bar.style.width = progress + '%'
		bar.innerHTML = progress + '%'
	})
}

const days = ['Sunday', 'Monday', 'Tuesday', 'Wednesday',
	'Thursday', 'Friday', 'Saturday']

function makeHeader() {
	hdrDiv = makeDiv('box', 'header', all)

	systemTimeDiv = makeDiv('heading', '', hdrDiv)
	makeP('', '', 'System Time', systemTimeDiv)
	makeP('', 'system-time', '', systemTimeDiv)

	startTimeDiv = makeDiv('heading line', '', hdrDiv)
	makeP('', '', 'Start Time', startTimeDiv)
	makeInputTime('', 'start-time', startTimeDiv,
		function () {
			state.StartTime = this.value
			sendUpdate("starttime")
		})

	startDaysDiv = makeDiv('heading', 'start-days', hdrDiv)
	makeP('', '', 'Start Days', startDaysDiv)
	weekdayDiv = makeDiv('', 'weekdays', startDaysDiv)

	days.forEach((day, i) => {
		dayDiv = makeDiv('', '', weekdayDiv)
		makeCheckbox('', `day-${i}`, day, dayDiv,
			function () {
				state.StartDays[i] = this.checked
				sendUpdate("update")
			})
	})
}

function showNow() {
	now = new Date()
	systemTime = document.getElementById("system-time")
	systemTime.innerText = now.toLocaleString('en-US', {
		weekday: 'long',
		hour: '2-digit',
		minute: '2-digit',
		timeZoneName: "short",
	});
	setTimeout('showNow()', (60 - (now.getSeconds())) * 1000)
}

function showHeader() {
	header = document.getElementById("header")
	if (!header) {
		makeHeader()
		showNow()
	}

	startTime = document.getElementById("start-time")
	startTime.value = state.StartTime

	days.forEach((day, i) => {
		cb = document.getElementById(`day-${i}`)
		cb.checked = state.StartDays[i]
	})
}

function show() {
	overlay = document.getElementById("overlay")
	overlay.style.display = online ? "none" : "block"
	showHeader()
	showZones()
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
		case "starttime":
			// fall-thru
		case "update":
			state = msg
			show()
			break
		}
	}
}
