package nano_rp2040

import "github.com/merliot/merle"

//tinyjson:json
type nano_rp2040 struct {
	Msg     string
}

func New() *nano_rp2040 {
	return &nano_rp2040{Msg: merle.ReplyState}
}

func (n *nano_rp2040) Subscribers() merle.Subscribers {
	return merle.Subscribers{
		merle.CmdInit: merle.NoInit,
		merle.CmdRun:  n.run,
	}
}

const html = `
<html lang="en">
	<head>
		<meta name="viewport" content="width=device-width, initial-scale=1">
	</head>
	<body style="margin: 0">
		<b>Hello</b>

		<script>
			var conn
			var online = false

			function getState() {
				conn.send(JSON.stringify({Msg: "_GetState"}))
			}

			function connect() {
				conn = new WebSocket("{{.WebSocket}}")

				conn.onopen = function(evt) {
					getState()
				}

				conn.onclose = function(evt) {
					setTimeout(connect, 1000)
				}

				conn.onerror = function(err) {
					console.log('nano_rp2040', err)
					conn.close()
				}

				conn.onmessage = function(evt) {
					msg = JSON.parse(evt.data)
					console.log('nano_rp2040', msg)

					switch(msg.Msg) {
					case "_ReplyState":
						break
					}
				}
			}

			connect()
		</script>
	</body>
</html>`

func (n *nano_rp2040) Assets() merle.ThingAssets {
	return merle.ThingAssets{
		HtmlTemplateText: html,
	}
}
