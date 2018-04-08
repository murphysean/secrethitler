package sh

import (
	"context"
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
	TypePlayerMessage         = "player.message"

	TypeAssertPolicies = "assert.policies"
	TypeAssertParty    = "assert.party"

	TypeReactPlayer  = "react.player"
	TypeReactEventID = "react.event_id"
	TypeReactStatus  = "react.status"

	TypeGuessPlayer = "guess.player"
	TypeGuessLie    = "guess.lie"

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
	Filter(context.Context) Event
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
	case TypePlayerMessage:
		e := MessageEvent{}
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
	case TypeReactPlayer:
		fallthrough
	case TypeReactEventID:
		fallthrough
	case TypeReactStatus:
		e := ReactEvent{}
		err = json.Unmarshal(b, &e)
		if err != nil {
			return bt, err
		}
		return e, nil
	case TypeAssertParty:
		fallthrough
	case TypeAssertPolicies:
		e := AssertEvent{}
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

func (e BaseEvent) GetID() int                       { return e.ID }
func (e BaseEvent) GetType() string                  { return e.Type }
func (e BaseEvent) Filter(ctx context.Context) Event { return e }

type PlayerEvent struct {
	BaseEvent
	Player Player `json:"player"`
}

func (e PlayerEvent) Filter(ctx context.Context) Event {
	pid := ctx.Value("playerID").(string)
	if pid != "admin" && pid != "engine" && pid != e.Player.ID {
		if e.Player.Party != "" {
			e.Player.Party = PartyMasked
		}
		if e.Player.Role != "" {
			e.Player.Role = RoleMasked
		}
	}
	return e
}

type PlayerPlayerEvent struct {
	BaseEvent
	PlayerID      string `json:"playerID"`
	OtherPlayerID string `json:"otherPlayerID"`
}

func (e PlayerPlayerEvent) Filter(ctx context.Context) Event { return e }

type PlayerVoteEvent struct {
	BaseEvent
	PlayerID string `json:"playerID"`
	Vote     bool   `json:"vote"`
}

func (e PlayerVoteEvent) Filter(ctx context.Context) Event {
	pid := ctx.Value("playerID").(string)
	if pid != "admin" && pid != "engine" && pid != e.PlayerID {
		e.Vote = false
	}
	return e
}

type PlayerLegislateEvent struct {
	BaseEvent
	PlayerID string
	Discard  string
	Veto     bool
}

func (e PlayerLegislateEvent) Filter(ctx context.Context) Event {
	pid := ctx.Value("playerID").(string)
	if pid != "admin" && pid != "engine" && pid != e.PlayerID {
		e.Discard = PolicyMasked
		e.Veto = false
	}
	return e
}

type MessageEvent struct {
	BaseEvent
	PlayerID string `json:"playerID"`
	Message  string `json:"message"`
}

func (e MessageEvent) Filter(ctx context.Context) Event { return e }

type GameEvent struct {
	BaseEvent
	Game Game `json:"game"`
}

func (e GameEvent) Filter(ctx context.Context) Event {
	pid := ctx.Value("playerID").(string)
	if pid != "admin" && pid != "engine" {
		e.Game = e.Game.Filter(ctx)
	}
	return e
}

type InformationEvent struct {
	BaseEvent
	PlayerID      string   `json:"playerID"`
	OtherPlayerID string   `json:"otherPlayerID,omitempty"`
	Policies      []string `json:"policies,omitempty"`
	Party         string   `json:"party,omitempty"`
	Token         string   `json:"token"`
}

func (e InformationEvent) Filter(ctx context.Context) Event {
	pid := ctx.Value("playerID").(string)
	if pid != "admin" && pid != "engine" && pid != e.PlayerID {
		if e.Policies != nil {
			np := []string{}
			for range e.Policies {
				np = append(np, PolicyMasked)
			}
			e.Policies = np
		}
		e.Party = PartyMasked
	}
	return e
}

type RequestEvent struct {
	BaseEvent
	PlayerID        string   `json:"playerID"`
	PresidentID     string   `json:"presidentID,omitempty"`
	ChancellorID    string   `json:"chancellorID,omitempty"`
	ExecutiveAction string   `json:"executiveAction,omitempty"`
	Policies        []string `json:"policies,omitempty"`
	Token           string   `json:"token,omitempty"`
}

func (e RequestEvent) Filter(ctx context.Context) Event {
	pid := ctx.Value("playerID").(string)
	if pid != "admin" && pid != "engine" && pid != e.PlayerID && pid != "all" {
		if e.Policies != nil {
			np := []string{}
			for range e.Policies {
				np = append(np, PolicyMasked)
			}
			e.Policies = np
		}
	}
	return e
}

type ReactEvent struct {
	BaseEvent
	PlayerID      string `json:"playerID"`
	ReactPlayerID string `json:"reactPlayerIDomitempty"`
	ReactEventID  int    `json:"reactEventID,omitempty"`
	Reaction      string `json:"reaction"`
}

func (e ReactEvent) Filter(ctx context.Context) Event { return e }

type AssertEvent struct {
	BaseEvent
	PlayerID string   `json:"playerID"`
	Token    string   `json:"token"`
	Policies []string `json:"policies,omitempty"`
	Party    string   `json:"party,omitempty"`
}

func (e AssertEvent) Filter(ctx context.Context) Event {
	pid := ctx.Value("playerID").(string)
	if pid != "admin" && pid != "engine" && pid != e.PlayerID {
		e.Token = "masked"
	}
	return e
}
