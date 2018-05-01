package sh

import (
	"context"
	"testing"
	"time"
)

func TestValidatePlayerJoin(t *testing.T)            {}
func TestValidatePlayerReady(t *testing.T)           {}
func TestValidatePlayerAcknowledge(t *testing.T)     {}
func TestValidatePlayerNominate(t *testing.T)        {}
func TestValidatePlayerVote(t *testing.T)            {}
func TestValidatePlayerLegislate(t *testing.T)       {}
func TestValidatePlayerInvestigate(t *testing.T)     {}
func TestValidatePlayerSpecialElection(t *testing.T) {}
func TestValidatePlayerExecute(t *testing.T)         {}
func TestValidatePlayerMessage(t *testing.T)         {}

func TestValidateReact(t *testing.T) {
	now := time.Now()
	g := Game{Players: []Player{Player{
		ID: "id",
	}}}

	e := ReactEvent{
		BaseEvent: BaseEvent{
			Type:   TypeReactStatus,
			Moment: now,
		},
		PlayerID: "id",
	}
	ctx := context.Background()
	ctx = context.WithValue(ctx, "playerID", "id")
	err := g.Validate(ctx, e)
	if err != nil {
		t.Fatal("No last reaction should allow reaction", err)
	}

	g.Players[0].LastAction = now.Add(time.Millisecond * -500)
	err = g.Validate(ctx, e)
	if err == nil {
		t.Fatal("Should throw error when throttle limit exceeded")
	}

}

func TestValidateAssertPolicies(t *testing.T) {}
func TestValidateAssertParty(t *testing.T)    {}
func TestValidateOther(t *testing.T)          {}
