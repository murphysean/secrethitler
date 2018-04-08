package sh

import (
	"encoding/json"
	"errors"
	"time"
)

const (
	TypePlayerJoin            = "player.join"
	TypePlayerReady           = "player.ready"
	TypePlayerAcknowledge     = "player.acknowledge"
	TypePlayerNominate        = "player.nominate"
	TypePlayerVote            = "player.vote"
	TypePlayerLegislate       = "player.legislate"
	TypePlayerInvestigate     = "player.investigate"
	TypePlayerSpecialElection = "player.special_election"
	TypePlayerExecute         = "player.execute"

	TypeRequestAcknowledge     = "request.acknowledge"
	TypeRequestVote            = "request.vote"
	TypeRequestNominate        = "request.nominate"
	TypeRequestLegislate       = "request.legislate"
	TypeRequestExecutiveAction = "request.executive_action"

	TypeGameInformation = "game.information"
	TypeGameUpdate      = "game.update"
)

type Event interface {
	GetID() int
	GetType() string
}

func UnmarshalEvent(b []byte) (Event, error) {
	//read it as a base event
	bt := BaseEvent{}
	err := json.Unmarshal(b, &bt)
	if err != nil {
		return nil, err
	}
	//finally read it as it's real event
	switch bt.GetType() {
	case TypePlayerJoin:
		fallthrough
	case TypePlayerReady:
		fallthrough
	case TypePlayerAcknowledge:
		e := PlayerEvent{}
		err = json.Unmarshal(b, &e)
		if err != nil {
			return bt, err
		}
		return e, nil
	case TypePlayerNominate:
		fallthrough
	case TypePlayerSpecialElection:
		fallthrough
	case TypePlayerExecute:
		fallthrough
	case TypePlayerInvestigate:
		e := PlayerPlayerEvent{}
		err = json.Unmarshal(b, &e)
		if err != nil {
			return bt, err
		}
		return e, nil
	case TypePlayerVote:
		e := PlayerVoteEvent{}
		err = json.Unmarshal(b, &e)
		if err != nil {
			return bt, err
		}
		return e, nil
	case TypePlayerLegislate:
		e := PlayerLegislateEvent{}
		err = json.Unmarshal(b, &e)
		if err != nil {
			return bt, err
		}
		return e, nil
	case TypeRequestAcknowledge:
		fallthrough
	case TypeRequestVote:
		fallthrough
	case TypeRequestNominate:
		fallthrough
	case TypeRequestLegislate:
		fallthrough
	case TypeRequestExecutiveAction:
		e := RequestEvent{}
		err = json.Unmarshal(b, &e)
		if err != nil {
			return bt, err
		}
		return e, nil
	case TypeGameInformation:
		e := InformationEvent{}
		err = json.Unmarshal(b, &e)
		if err != nil {
			return bt, err
		}
		return e, nil
	case TypeGameUpdate:
		e := GameEvent{}
		err = json.Unmarshal(b, &e)
		if err != nil {
			return bt, err
		}
		return e, nil
	default:
		return bt, errors.New("Unknown Event Type")
	}
}

type BaseEvent struct {
	ID     int       `json:"id"`
	Type   string    `json:"type"`
	Moment time.Time `json:"moment"`
}

func (e BaseEvent) GetID() int      { return e.ID }
func (e BaseEvent) GetType() string { return e.Type }

type PlayerEvent struct {
	BaseEvent
	Player Player `json:"player"`
}

type PlayerPlayerEvent struct {
	BaseEvent
	PlayerID      string `json:"playerID"`
	OtherPlayerID string `json:"otherPlayerID"`
}

type PlayerVoteEvent struct {
	BaseEvent
	PlayerID string `json:"playerID"`
	Vote     bool   `json:"vote"`
}

type PlayerLegislateEvent struct {
	BaseEvent
	PlayerID string
	Discard  string
	Veto     bool
}

type GameEvent struct {
	BaseEvent
	Game Game `json:"game"`
}

type InformationEvent struct {
	BaseEvent
	PlayerID      string   `json:"playerID"`
	OtherPlayerID string   `json:"otherPlayerID,omitempty"`
	Policies      []string `json:"policies,omitempty"`
	Party         string   `json:"party,omitempty"`
}

type RequestEvent struct {
	BaseEvent
	PlayerID        string   `json:"playerID"`
	PresidentID     string   `json:"presidentID,omitempty"`
	ChancellorID    string   `json:"chancellorID,omitempty"`
	ExecutiveAction string   `json:"executiveAction,omitempty"`
	Policies        []string `json:"policies,omitempty"`
}
