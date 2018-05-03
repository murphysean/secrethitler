package sh

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"
)

const (
	PlayerIDEngine = "engine"
	PlayerIDAdmin  = "admin"
	PlayerIDAll    = "all"

	PolicyFacist  = "facist"
	PolicyLiberal = "liberal"
	PolicyMasked  = "masked"

	RoleFacist  = "facist"
	RoleLiberal = "liberal"
	RoleHitler  = "hitler"
	RoleMasked  = "masked"

	PartyFacist  = "facist"
	PartyLiberal = "liberal"
	PartyMasked  = "masked"

	GameStateLobby    = ""
	GameStateInit     = "init"
	GameStateStarted  = "started"
	GameStateFinished = "finished"

	RoundStateNominating      = "nominating"
	RoundStateVoting          = "voting"
	RoundStateFailed          = "failed"
	RoundStateLegislating     = "legislating"
	RoundStateExecutiveAction = "executive_action"
	RoundStateFinished        = "finished"

	ExecutiveActionInvestigate     = "investigate"
	ExecutiveActionPeek            = "peek"
	ExecutiveActionSpecialElection = "special_election"
	ExecutiveActionExecute         = "execute"
)

func NewSecretHitler() *SecretHitler {
	ret := new(SecretHitler)
	ret.subscribers = make(map[string]chan<- Event)
	ec := make(chan Event, 10)
	//Make the engine a subscriber
	ret.subscribers["engine"] = ec
	go func() {
		for {
			select {
			case e := <-ec:
				if e == nil {
					fmt.Println("Exiting game engine loop via nil read")
					return
				}
				if nes, err := ret.Engine(e); err == nil {
					for _, ne := range nes {
						ctx := context.Background()
						ctx = context.WithValue(ctx, "playerID", PlayerIDEngine)
						err = ret.SubmitEvent(ctx, ne)
						if err != nil {
							fmt.Println("engine:Submit Error:", err)
						}
					}
				}
				//If the game is over, then return
				if ret.Game.State == GameStateFinished {
					ret.m.Lock()
					for k, v := range ret.subscribers {
						delete(ret.subscribers, k)
						close(v)
					}
					ret.m.Unlock()
					break
				}
			}
		}
		fmt.Println("Exiting game engine loop via loop break")
	}()
	return ret
}

type SecretHitler struct {
	Game

	Log io.Writer
	m   sync.RWMutex

	subscribers map[string]chan<- Event
}

func (sh *SecretHitler) SubmitEvent(ctx context.Context, e Event) error {
	sh.m.Lock()
	defer sh.m.Unlock()
	//Do the validate here
	err := sh.Validate(ctx, e)
	if err != nil {
		return err
	}
	g, ne, err := sh.Apply(e)
	if err != nil {
		return err
	}
	sh.Game = g
	//Persist the event to a file
	if sh.Log != nil {
		enc := json.NewEncoder(sh.Log)
		err := enc.Encode(ne)
		if err != nil {
			return err
		}
	}
	go func() {
		sh.BroadcastEvent(ne)
	}()
	return nil
}

func (sh *SecretHitler) AddSubscriber(key string, channel chan<- Event) {
	if sh.Game.State == GameStateFinished {
		return
	}
	sh.m.Lock()
	sh.subscribers[key] = channel
	sh.m.Unlock()
}

func (sh *SecretHitler) RemoveSubscriber(key string) {
	sh.m.Lock()
	delete(sh.subscribers, key)
	sh.m.Unlock()
}

func (sh *SecretHitler) BroadcastEvent(e Event) {
	sh.m.RLock()
	defer sh.m.RUnlock()
	for k, _ := range sh.subscribers {
		sh.subscribers[k] <- e
	}
}

//ReadEventLog will read the associated event log and publish all the events to the included channel
func ReadEventLog(r io.Reader, c chan<- Event) error {
	//Read in a byte slice
	d := json.NewDecoder(r)
	var err error
	for err == nil {
		var rm json.RawMessage
		err = d.Decode(&rm)
		if err != nil {
			close(c)
			return err
		}
		e, err := UnmarshalEvent(rm)
		if err != nil {
			close(c)
			return err
		}
		c <- e
	}
	close(c)
	return nil
}

type Token struct {
	EventID       int    `json:"eventID"`
	PlayerID      string `json:"playerID"`
	Assertion     string `json:"assertion"`
	RoundID       int    `json:"roundID"`
	OtherPlayerID string `json:"otherPlayerID,omitempty"`
	PolicyCount   int    `json:"policyCount,omitempty"`
}

type Game struct {
	ID                         string   `json:"id,omitempty"`
	Secret                     string   `json:"secret,omitempty"`
	EventID                    int      `json:"eventID,omitempty"`
	State                      string   `json:"state,omitempty"`
	Draw                       []string `json:"draw,omitempty"`
	Discard                    []string `json:"discard,omitempty"`
	Liberal                    int      `json:"liberal,omitempty"`
	Facist                     int      `json:"facist,omitempty"`
	ElectionTracker            int      `json:"electionTracker,omitempty"`
	Players                    []Player `json:"players,omitempty"`
	Round                      Round    `json:"round,omitempty"`
	NextPresidentID            string   `json:"nextPresidentID,omitempty"`
	PreviousPresidentID        string   `json:"previousPresidentID,omitempty"`
	PreviousChancellorID       string   `json:"previousChancellorID,omitempty"`
	PreviousEnactedPolicy      string   `json:"previousEnactedPolicy,omitempty"`
	SpecialElectionRoundID     int      `json:"specialElectionRoundID,omitempty"`
	SpecialElectionPresidentID string   `json:"specialElectionPresidentID,omitempty"`
	WinningParty               string   `json:"winningParty,omitempty"`
}

func (g Game) GetPlayerByID(id string) (Player, error) {
	for _, p := range g.Players {
		if p.ID == id {
			return p, nil
		}
	}
	return Player{}, errors.New("Not Found")
}

type Player struct {
	ID             string    `json:"id,omitempty"`
	Party          string    `json:"party,omitempty"`
	Role           string    `json:"role,omitempty"`
	Ready          bool      `json:"ready,omitempty"`
	Ack            bool      `json:"ack,omitempty"`
	ExecutedBy     string    `json:"executedBy,omitempty"`
	InvestigatedBy string    `json:"investigatedBy,omitempty"`
	LastAction     time.Time `json:"lastAction,omitempty"`
	Status         string    `json:"status,omitempty"`
}

type Round struct {
	ID              int      `json:"id,omitempty"`
	PresidentID     string   `json:"presidentID,omitempty"`
	ChancellorID    string   `json:"chancellorID,omitempty"`
	State           string   `json:"state,omitempty"`
	Votes           []Vote   `json:"votes,omitempty"`
	Policies        []string `json:"policies,omitempty"`
	EnactedPolicy   string   `json:"enactedPolicy,omitempty"`
	ExecutiveAction string   `json:"executiveAction,omitempty"`
}

type Vote struct {
	PlayerID string `json:"playerID,omitempty"`
	Vote     bool   `json:"vote,omitempty"`
}
