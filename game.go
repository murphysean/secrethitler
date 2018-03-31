package sh

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
)

const (
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
	ec := make(chan Event)
	ret.subscribers["engine"] = ec
	go func() {
		for {
			select {
			case e := <-ec:
				//TODO If the game is over, then return
				if nes, err := ret.Engine(e); err == nil {
					fmt.Println("engine: Produced:", nes)
					for _, ne := range nes {
						err = ret.SubmitEvent(context.TODO(), ne)
						if err != nil {
							fmt.Println("Apply Error:", err)
						}
					}
				}
			}
		}
		fmt.Println("Exiting game engine loop")
	}()
	return ret
}

type SecretHitler struct {
	Game

	Log *os.File
	m   sync.RWMutex

	//Make the engine a subscriber
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

type Game struct {
	EventID                    int      `json:"eventID,omitempty"`
	State                      string   `json:"state,omitempty"`
	Draw                       []string `json:"draw,omitempty"`
	Discard                    []string `json:"discard,omitempty"`
	Liberal                    int      `json:"liberal,omitempty"`
	Facist                     int      `json:"facist,omitempty"`
	FailedVotes                int      `json:"failedVotes,omitempty"`
	Players                    []Player `json:"players,omitempty"`
	Round                      Round    `json:"round,omitempty"`
	NextPresidentID            string   `json:"nextPresidentID,omitempty"`
	PreviousPresidentID        string   `json:"previousPresidentID,omitempty"`
	PreviousChancellorID       string   `json:"previousChancellorID,omitempty"`
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
	ID             string `json:"id,omitempty"`
	Name           string `json:"name,omitempty"`
	Party          string `json:"party,omitempty"`
	Role           string `json:"role,omitempty"`
	Ready          bool   `json:"ready,omitempty"`
	Ack            bool   `json:"ack,omitempty"`
	ExecutedBy     string `json:"executedBy,omitempty"`
	InvestigatedBy string `json:"investigatedBy,omitempty"`
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
