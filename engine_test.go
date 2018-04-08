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
			//TODO Every player should have an assigned role
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
