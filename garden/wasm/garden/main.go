//go:build js && wasm

package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"syscall/js"
	"time"

	"github.com/merliot/projects/garden/msg"
	"github.com/merliot/projects/garden/state"
)

type wasm struct {
	ws       js.Value
	again    chan bool
	fn       js.Func
	timeDiff time.Duration
}

type mymsg struct {
	Msg string
}

func getElementById(id string) js.Value {
	return js.Global().Get("document").Call("getElementById", id)
}

func getElementValue(id string, p string) js.Value {
	return getElementById(id).Get(p)
}

func setValue(id string, p string, x any) {
	getElementById(id).Set(p, x)
}

func setStyle(id string, p string, s string) {
	getElementValue(id, "style").Call("setProperty", p, s)
}

func getString(id string, p string) string {
	return getElementValue(id, p).String()
}

func getFloat(id string, p string) float64 {
	return getElementValue(id, p).Float()
}

func getType(id string, p string) string {
	return getElementValue(id, p).Type().String()
}

func addEventListener(id string, event string, fn js.Func) {
	getElementById(id).Call("addEventListener", event, fn)
}

func removeEventListener(id string, event string, fn js.Func) {
	getElementById(id).Call("removeEventListener", event, fn)
}

func (w wasm) getIdentity() {
	msg, _ := json.Marshal(mymsg{Msg: "_GetIdentity"})
	w.ws.Call("send", string(msg))
}

func (w wasm) getState() {
	msg, _ := json.Marshal(mymsg{Msg: "_GetState"})
	w.ws.Call("send", string(msg))
}

func (w wasm) update(msg state.State) {
	setValue("gallons", "innerHTML", fmt.Sprintf("%.2f", msg.Gallons))
	setValue("start", "disabled", msg.Running)
	setValue("stop", "disabled", !msg.Running)
	gallonsGoal, _ := strconv.ParseFloat(getString("gallonsGoal", "value"), 64)
	progress := int(msg.Gallons / gallonsGoal * 100.0)
	percent := fmt.Sprintf("%d%%", progress)
	setStyle("bar", "width", percent)
	setValue("bar", "innerHTML", percent)
}

func (w wasm) saveState(msg state.State) {
	w.timeDiff = time.Now().Sub(msg.Now)
	setValue("startTime", "value", msg.StartTime)
	for i := 0; i < 7; i++ {
		setValue("day" + strconv.Itoa(i), "checked", msg.StartDays[i])
	}
	setValue("gallonsGoal", "value", msg.GallonsGoal)
	w.update(msg)
}

func (w wasm) changeStartTime(this js.Value, args []js.Value) any {
	startTime := getString("startTime", "value")
	msg, err := json.Marshal(msg.StartTime{Msg: "StartTime", Time: startTime})
	fmt.Println(string(msg), err)
	w.ws.Call("send", string(msg))
	return nil
}

func (w wasm) open(this js.Value, args []js.Value) any {
	fmt.Println("open")
	w.fn = js.FuncOf(w.changeStartTime)
	addEventListener("startTime", "change", w.fn)
	w.getIdentity()
	return nil
}

func (w wasm) close(this js.Value, args []js.Value) any {
	fmt.Println("close")
	removeEventListener("startTime", "change", w.fn)
	w.fn.Release()
	w.again<-true
	return nil
}

func (w wasm) error(this js.Value, args []js.Value) any {
	fmt.Println("error")
	w.ws.Call("close")
	return nil
}

func (w wasm) message(this js.Value, args []js.Value) any {
	var msg mymsg
	data := []byte(args[0].Get("data").String())
	json.Unmarshal(data, &msg)
	fmt.Println(string(data))
	switch msg.Msg {
	case "_ReplyIdentity":
		w.getState()
	case "_ReplyState":
		var msg state.State
		json.Unmarshal(data, &msg)
		w.saveState(msg)
	}
	return nil
}

func (w wasm) run(this js.Value, args []js.Value) any {
	url := args[0].String()
	fmt.Println("Opening WebSocket:", url)

	w.again = make(chan bool)

	for {
		w.ws = js.Global().Get("WebSocket").New(url)
		w.ws.Call("addEventListener", "open", js.FuncOf(w.open))
		w.ws.Call("addEventListener", "close", js.FuncOf(w.close))
		w.ws.Call("addEventListener", "error", js.FuncOf(w.error))
		w.ws.Call("addEventListener", "message", js.FuncOf(w.message))
		select {
		case <-w.again:
		}
	}

	return nil
}

func (w wasm) Run() {
	js.Global().Set("Run", js.FuncOf(w.run))
	select{}
}

func main() {
	wasm{}.Run()
}
