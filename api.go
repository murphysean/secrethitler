package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func APIStateHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		//Get the current game state
		w.Header().Set("Content-Type", "application/json")
		e := json.NewEncoder(w)
		//TODO Filter it for the authenticated user
		e.Encode(theGame.Game)
	case http.MethodPut:
		//TODO Only admins can do this
		g := Game{}
		d := json.NewDecoder(r.Body)
		err := d.Decode(&g)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			fmt.Println(err)
		}
		//Fire a game update event with the changes
		err = theGame.SubmitEvent(r.Context(), GameEvent{
			BaseEvent: BaseEvent{Type: TypeGameUpdate},
			Game:      g,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			fmt.Println(err)
		}
		w.Header().Set("Content-Type", "application/json")
		e := json.NewEncoder(w)
		//TODO Filter it for the authenticated user
		e.Encode(theGame.Game)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func APIEventHandler(w http.ResponseWriter, r *http.Request) {
	//Read the whole body into a buffer (to be read twice)
	b, err := ioutil.ReadAll(r.Body)
	//First determine the type of event being posted
	bt := BaseEvent{}
	err = json.Unmarshal(b, &bt)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		fmt.Println(err)
		return
	}
	//Second serialize it into that type of event object
	var e Event
	//Set the playerID on the event from authenticated context
	switch bt.GetType() {
	case TypePlayerJoin:
		fallthrough
	case TypePlayerReady:
		fallthrough
	case TypePlayerAcknowledge:
		pe := PlayerEvent{}
		err = json.Unmarshal(b, &pe)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			fmt.Println(err)
			return
		}
		e = pe
	case TypePlayerVote:
		pe := PlayerVoteEvent{}
		err = json.Unmarshal(b, &pe)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			fmt.Println(err)
			return
		}
		e = pe
	case TypePlayerNominate:
		pe := PlayerPlayerEvent{}
		err = json.Unmarshal(b, &pe)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			fmt.Println(err)
			return
		}
		e = pe
	case TypePlayerLegislate:
		pe := PlayerLegislateEvent{}
		err = json.Unmarshal(b, &pe)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			fmt.Println(err)
			return
		}
		e = pe
	case TypePlayerInvestigate:
		pe := PlayerPlayerEvent{}
		err = json.Unmarshal(b, &pe)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			fmt.Println(err)
			return
		}
		e = pe
	case TypePlayerSpecialElection:
		pe := PlayerPlayerEvent{}
		err = json.Unmarshal(b, &pe)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			fmt.Println(err)
			return
		}
		e = pe
	case TypePlayerExecute:
		pe := PlayerPlayerEvent{}
		err = json.Unmarshal(b, &pe)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			fmt.Println(err)
			return
		}
		e = pe
	default:
		http.Error(w, "Unrecognized Event Type", http.StatusBadRequest)
		fmt.Println(err)
		return
	}
	if e == nil {
		http.Error(w, "Nil Event", http.StatusBadRequest)
		return
	}
	//Validate & submit the event against the game state
	err = theGame.SubmitEvent(r.Context(), e)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		fmt.Println(err)
		return
	}

	w.WriteHeader(http.StatusAccepted)
}
