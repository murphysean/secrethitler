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

	TypeGuess = "guess"

	TypeRequestAcknowledge     = "request.acknowledge"
	TypeRequestVote            = "request.vote"
	TypeRequestNominate        = "request.nominate"
	TypeRequestLegislate       = "request.legislate"
	TypeRequestExecutiveAction = "request.executive_action"

	TypeGameVoteResults = "game.vote_results"
	TypeGameInformation = "game.information"
	TypeGameUpdate      = "game.update"
	TypeGameFinished    = "game.finished"
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
		if e.Moment.IsZero() {
			e.Moment = time.Now()
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
		if e.Moment.IsZero() {
			e.Moment = time.Now()
		}
		return e, nil
	case TypePlayerVote:
		e := PlayerVoteEvent{}
		err = json.Unmarshal(b, &e)
		if err != nil {
			return bt, err
		}
		if e.Moment.IsZero() {
			e.Moment = time.Now()
		}
		return e, nil
	case TypePlayerLegislate:
		e := PlayerLegislateEvent{}
		err = json.Unmarshal(b, &e)
		if err != nil {
			return bt, err
		}
		if e.Moment.IsZero() {
			e.Moment = time.Now()
		}
		return e, nil
	case TypePlayerMessage:
		e := MessageEvent{}
		err = json.Unmarshal(b, &e)
		if err != nil {
			return bt, err
		}
		if e.Moment.IsZero() {
			e.Moment = time.Now()
		}
		return e, nil
	case TypeAssertPolicies:
		fallthrough
	case TypeAssertParty:
		e := AssertEvent{}
		err = json.Unmarshal(b, &e)
		if err != nil {
			return bt, err
		}
		if e.Moment.IsZero() {
			e.Moment = time.Now()
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
		if e.Moment.IsZero() {
			e.Moment = time.Now()
		}
		return e, nil
	case TypeGuess:
		e := GuessEvent{}
		err = json.Unmarshal(b, &e)
		if err != nil {
			return bt, err
		}
		if e.Moment.IsZero() {
			e.Moment = time.Now()
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
		if e.Moment.IsZero() {
			e.Moment = time.Now()
		}
		return e, nil
	case TypeGameVoteResults:
		e := VoteResultEvent{}
		err = json.Unmarshal(b, &e)
		if err != nil {
			return bt, err
		}
		if e.Moment.IsZero() {
			e.Moment = time.Now()
		}
		return e, nil
	case TypeGameInformation:
		e := InformationEvent{}
		err = json.Unmarshal(b, &e)
		if err != nil {
			return bt, err
		}
		if e.Moment.IsZero() {
			e.Moment = time.Now()
		}
		return e, nil
	case TypeGameFinished:
		e := FinishedEvent{}
		err = json.Unmarshal(b, &e)
		if err != nil {
			return bt, err
		}
		if e.Moment.IsZero() {
			e.Moment = time.Now()
		}
		return e, nil
	case TypeGameUpdate:
		e := GameEvent{}
		err = json.Unmarshal(b, &e)
		if err != nil {
			return bt, err
		}
		if e.Moment.IsZero() {
			e.Moment = time.Now()
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
	pid, _ := ctx.Value("playerID").(string)
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
	PlayerID      string `json:"playerId"`
	OtherPlayerID string `json:"otherPlayerId"`
}

func (e PlayerPlayerEvent) Filter(ctx context.Context) Event { return e }

type PlayerVoteEvent struct {
	BaseEvent
	PlayerID string `json:"playerId"`
	Vote     bool   `json:"vote"`
}

func (e PlayerVoteEvent) Filter(ctx context.Context) Event {
	pid, _ := ctx.Value("playerId").(string)
	if pid != "admin" && pid != "engine" && pid != e.PlayerID {
		e.Vote = false
	}
	return e
}

type PlayerLegislateEvent struct {
	BaseEvent
	PlayerID string `json:"playerId"`
	Discard  string `json:"discard"`
	Veto     bool   `json:"veto"`
}

func (e PlayerLegislateEvent) Filter(ctx context.Context) Event {
	pid, _ := ctx.Value("playerId").(string)
	if pid != "admin" && pid != "engine" && pid != e.PlayerID {
		e.Discard = PolicyMasked
		e.Veto = false
	}
	return e
}

type MessageEvent struct {
	BaseEvent
	PlayerID string `json:"playerId"`
	Message  string `json:"message"`
}

func (e MessageEvent) Filter(ctx context.Context) Event { return e }

type VoteResultEvent struct {
	BaseEvent
	RoundID   int    `json:"roundId"`
	Succeeded bool   `json:"succeeded"`
	Votes     []Vote `json:"votes"`
}

func (e VoteResultEvent) Filter(ctx context.Context) Event { return e }

type GameEvent struct {
	BaseEvent
	Game Game `json:"game"`
}

func (e GameEvent) Filter(ctx context.Context) Event {
	pid, _ := ctx.Value("playerID").(string)
	if pid != "admin" && pid != "engine" {
		e.Game = e.Game.Filter(ctx)
	}
	return e
}

type InformationEvent struct {
	BaseEvent
	PlayerID      string   `json:"playerId"`
	RoundID       int      `json:"roundId"`
	OtherPlayerID string   `json:"otherPlayerId,omitempty"`
	Policies      []string `json:"policies,omitempty"`
	Party         string   `json:"party,omitempty"`
	Token         string   `json:"token"`
}

func (e InformationEvent) Filter(ctx context.Context) Event {
	pid, _ := ctx.Value("playerID").(string)
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

type FinishedEvent struct {
	BaseEvent
	WinningCondition string `json:"winningCondition"`
	WinningParty     string `json:"winningParty"`
}

func (e FinishedEvent) Filter(ctx context.Context) Event {
	return e
}

type RequestEvent struct {
	BaseEvent
	PlayerID        string   `json:"playerId"`
	RoundID         int      `json:"roundId"`
	PresidentID     string   `json:"presidentId,omitempty"`
	ChancellorID    string   `json:"chancellorId,omitempty"`
	ExecutiveAction string   `json:"executiveAction,omitempty"`
	Policies        []string `json:"policies,omitempty"`
	VetoPossible    bool     `json:"vetoPossible,omitempty"`
	Veto            bool     `json:"veto,omitempty"`
	Token           string   `json:"token,omitempty"`
}

func (e RequestEvent) Filter(ctx context.Context) Event {
	pid, _ := ctx.Value("playerID").(string)
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
	PlayerID      string `json:"playerId"`
	ReactPlayerID string `json:"reactPlayerIdomitempty"`
	ReactEventID  int    `json:"reactEventId,omitempty"`
	Reaction      string `json:"reaction"`
}

func (e ReactEvent) Filter(ctx context.Context) Event { return e }

type AssertEvent struct {
	BaseEvent
	PlayerID      string   `json:"playerId"`
	RoundID       int      `json:"roundId"`
	Token         string   `json:"token"`
	PolicySource  string   `json:"policySource,omitempty"`
	Policies      []string `json:"policies,omitempty"`
	OtherPlayerID string   `json:"otherPlayerId,omitempty"`
	Party         string   `json:"party,omitempty"`
}

func (e AssertEvent) Filter(ctx context.Context) Event {
	pid, _ := ctx.Value("playerID").(string)
	if pid != "admin" && pid != "engine" && pid != e.PlayerID {
		e.Token = "masked"
	}
	return e
}

// GuessEvent is an event a player can send to make a prediction or guess as to outcomes of the game
type GuessEvent struct {
	BaseEvent
	PlayerID       string   `json:"playerId"`
	Facists        []string `json:"facists,omitempty"`
	SecretHitlerID string   `json:"secretHitlerId,omitempty"`
	WinningParty   string   `json:"winningParty,omitempty"`
	CallEventID    string   `json:"callEventId,omitempty"`
}

func (e GuessEvent) Filter(ctx context.Context) Event {
	pid, _ := ctx.Value("playerID").(string)
	if pid != "admin" && pid != "engine" && pid != e.PlayerID {
		e.PlayerID = "masked"
		nf := []string{}
		for _, _ = range e.Facists {
			nf = append(nf, "masked")
		}
		e.Facists = nf
		if e.SecretHitlerID != "" {
			e.SecretHitlerID = "masked"
		}
		if e.WinningParty != "" {
			e.WinningParty = "masked"
		}
		if e.CallEventID != "" {
			e.CallEventID = "masked"
		}
	}
	return e
}
