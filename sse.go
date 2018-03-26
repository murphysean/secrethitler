package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// https://www.html5rocks.com/en/tutorials/eventsource/basics/
func ServerSentEventsHandler(w http.ResponseWriter, r *http.Request) {
	//TODO Find room?

	//Is this a flushable connection
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "webserver doesn't support flushing", http.StatusInternalServerError)
		return
	}
	cnot, ok := w.(http.CloseNotifier)
	if !ok {
		http.Error(w, "webserver doesn't support close notifications", http.StatusInternalServerError)
		return
	}
	cnotchan := cnot.CloseNotify()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, ": Getting Started\n\n")
	flusher.Flush()

	//TODO Stream events from the last event id specified
	//r.Header.Get("Last-Event-Id")

	//Subscribe to game events
	myChan := make(chan Event)
	uid := GenUUIDv4()
	theGame.m.Lock()
	theGame.subscribers[uid] = myChan
	theGame.m.Unlock()

	//Defer the removal of the chanel from the game on disconnect
	defer func() {
		theGame.m.Lock()
		delete(theGame.subscribers, uid)
		theGame.m.Unlock()
	}()

	//Loop on events coming out of the gameserver
	for {
		select {
		case e := <-myChan:
			b, err := json.Marshal(&e)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Fprintf(w, "id: %d\n", e.GetID())
			fmt.Fprintf(w, "event: %s\n", e.GetType())
			//TODO Before sending an event, filter it for the auth'd user
			fmt.Fprintf(w, "data: %s\n\n", b)
		case <-time.After(time.Minute * 5):
			fmt.Fprintf(w, ": keepalive\n\n")
		case <-cnotchan:
			return
		}
		//Optionally also include a seperate event sending the whole state for the client to sync on
		/*
			fmt.Fprintln(w, "event: state")
			//TODO Before sending the state, filter it for the auth'd user
			b, err := json.Marshal(&theGame.Game)
			if err != nil {
				fmt.Println(err)
			}
			fmt.Fprintf(w, "data: %s\n\n", b)
		*/
		//Flush the data down the pipe
		flusher.Flush()
	}

}
