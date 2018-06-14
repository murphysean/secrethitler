package sh

import (
	"testing"
)

func TestGameStart(t *testing.T) {
	g := Game{
		Players: []Player{
			Player{ID: "1", Ready: true},
			Player{ID: "2", Ready: true},
			Player{ID: "3", Ready: true},
			Player{ID: "4", Ready: true},
			Player{ID: "5", Ready: true},
		}}
	e := PlayerEvent{
		BaseEvent: BaseEvent{
			Type: TypePlayerReady,
		},
		Player: Player{ID: "5", Ready: true},
	}
	events, err := g.Engine(e)
	if err != nil {
		t.Fatal(err)
	}
	updateEventFound := false
	for _, e := range events {
		if e.GetType() == TypeGameUpdate {
			ge := e.(GameEvent)
			updateEventFound = true
			//Every player should have an assigned role
			allHaveParty := true
			allHaveRoles := true
			for _, p := range ge.Game.Players {
				if p.Party == "" {
					allHaveParty = false
				}
				if p.Role == "" {
					allHaveRoles = false
				}
			}
			if !allHaveParty {
				t.Fatal("Not all players have party")
			}
			if !allHaveRoles {
				t.Fatal("Not all players have a role")
			}
			//Should be 11+6 policies in the draw pile
			if len(ge.Game.Draw) != 17 {
				t.Fatal("Not 17 policies")
			}
			//Should be a nextPresident defined
			if ge.Game.NextPresidentID == "" {
				t.Fatal("No Next President defined")
			}
		}
	}
	if !updateEventFound {
		t.Fatal("No Game update event")
	}
}

func TestVeto(t *testing.T) {
	g := Game{
		ID:     "1",
		Secret: "secret",
		State:  GameStateStarted,
		Players: []Player{
			Player{ID: "1", Party: PartyLiberal, Role: RoleLiberal},
			Player{ID: "2", Party: PartyLiberal, Role: RoleLiberal},
			Player{ID: "3", Party: PartyLiberal, Role: RoleLiberal},
			Player{ID: "4", Party: PartyFascist, Role: RoleFascist},
			Player{ID: "5", Party: PartyFascist, Role: RoleHitler},
		},
		PreviousPresidentID:  "4",
		PreviousChancellorID: "5",
		NextPresidentID:      "2",
		ElectionTracker:      2,
		Liberal:              4,
		Fascist:              5,
		Draw:                 []string{"draw", "draw", "draw"},
		Discard:              []string{"discard"},
		Round: Round{
			ID:           10,
			PresidentID:  "1",
			ChancellorID: "2",
			Policies:     []string{PolicyFascist},
			State:        RoundStateLegislating,
		},
	}
	e := PlayerLegislateEvent{
		BaseEvent: BaseEvent{Type: TypePlayerLegislate},
		PlayerID:  "1",
		Veto:      true,
	}
	events, err := g.Engine(e)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range events {
		t.Log(e)
	}
}
