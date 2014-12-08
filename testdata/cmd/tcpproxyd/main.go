package main

import (
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/websocket"
	"honnef.co/go/js/console"
	"honnef.co/go/js/dom"
	"net/url"
	"strings"
)

var (
	doc = dom.GetWindow().Document()
	ws  *websocket.WebSocket
)

func main() {
	form := doc.GetElementByID("startws")
	form.AddEventListener("submit", false, handleClick)
}

func handleClick(ev dom.Event) {
	ev.PreventDefault()
	if ws != nil {
		console.Log("connection already open")
		return
	}
	epUrl := doc.GetElementByID("url").Underlying().Get("value").Str()
	if len(epUrl) == 0 {
		epUrl = doc.GetElementByID("url").GetAttribute("placeholder")
		doc.GetElementByID("url").Underlying().Set("value", epUrl)
	}

	wsUrl := url.URL{}
	wsUrl.Scheme = "ws"
	wsUrl.Host = "localhost:8081"
	wsUrl.Path = "/wstcp/"
	q := wsUrl.Query()
	q.Set("endpoint", epUrl)
	wsUrl.RawQuery = q.Encode()
	console.Log("url: ", wsUrl.String())
	ws = websocket.New(wsUrl.String())
	ws.OnOpen(handleOpen)
	ws.OnMessage(handleMessage)
	ws.OnClose(handleClose)
}

func handleOpen(obj js.Object) {
	console.Log("open")
}

func handleMessage(obj *dom.MessageEvent) {
	if obj.Data.Str() == "OK " {
		console.Log("connection done. say hi")
		// connection completed
		ws.Send("hi")
	} else if strings.HasPrefix(obj.Data.Str(), "ERR") {
		console.Log("error creating connection")
		// error creating connection
	} else {
		console.Log("data: ", obj.Data)
	}
}

func handleClose(obj js.Object) {
	console.Log("closed", obj)
}
