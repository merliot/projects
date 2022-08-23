// Copyright 2021-2022 Scott Feldman (sfeldma@gmail.com). All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// in the LICENSE file.

var conn
var online = false

days = []
for (var i = 0; i < 7; i++) {
	days[i] = document.getElementById("day" + i)
}
startTime = document.getElementById("startTime")
gallons = document.getElementById("gallons")

function getState() {
	conn.send(JSON.stringify({Msg: "_GetState"}))
}

function getIdentity() {
	conn.send(JSON.stringify({Msg: "_GetIdentity"}))
}

function saveState(msg) {
	startTime.value = msg.StartTime
	for (var i = 0; i < days.length; i++) {
		days[i].checked = msg.StartDays[i]
	}
	gallons.innerHTML = msg.Gallons
}

function saveDay(msg) {
	days[msg.Day].checked = msg.State
}

function saveStartTime(msg) {
	startTime.value = msg.Time
}

function showAll() {
}

function changeStartTime() {
	conn.send(JSON.stringify({Msg: "StartTime",
		Time: startTime.value}))
}

function changeDay(box, day) {
	conn.send(JSON.stringify({Msg: "Day", Day: day,
		State: box.checked}))
}

function start() {
	conn.send(JSON.stringify({Msg: "Start"}))
}

function stop() {
	conn.send(JSON.stringify({Msg: "Stop"}))
}

function Run(ws) {

	function connect() {
		conn = new WebSocket(ws)

		conn.onopen = function(evt) {
			getIdentity()
		}

		conn.onclose = function(evt) {
			online = false
			showAll()
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
			case "Update":
				saveState(msg)
				showAll()
				break
			case "Day":
				saveDay(msg)
				break
			case "StartTime":
				saveStartTime(msg)
				break
			}
		}
	}

	connect()
}
