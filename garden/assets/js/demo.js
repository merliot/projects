function showChild(id) {
	document.getElementById("child").src = "/" + encodeURIComponent(id)
}

function clearScreen() {
	document.getElementById("child").src = ""
}

function saveState(msg) {
	if (msg.ChildId != "") {
		showChild(msg.ChildId)
	}
}

function update(msg) {
	if (msg.Online) {
		showChild(msg.Id)
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

			switch(msg.Msg) {
			case "_ReplyState":
				saveState(msg)
				break
			case "_EventStatus":
				update(msg)
				break
			}
		}
	}

	connect()
}
