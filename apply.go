package sh

import (
	"time"
)

//Apply mutates the game state by applying the given event.
func (g Game) Apply(e Event) (Game, Event, error) {
	//Increment the event counter
	g.EventID = g.EventID + 1

	//Assign the event id to the event
	switch e.GetType() {
	//REQUEST EVENTS
	case TypeRequestAcknowledge:
		ne := e.(RequestEvent)
		ne.ID = g.EventID
		ne.Moment = time.Now()
		e = ne
		g.State = GameStateStarted
	case TypeRequestVote:
		ne := e.(RequestEvent)
		ne.ID = g.EventID
		ne.Moment = time.Now()
		e = ne
		//Set the round state to voting
		g.Round.State = RoundStateVoting
	case TypeRequestNominate:
		ne := e.(RequestEvent)
		ne.ID = g.EventID
		ne.Moment = time.Now()
		e = ne
		g.Round.State = RoundStateNominating
	case TypeRequestLegislate:
		ne := e.(RequestEvent)
		ne.ID = g.EventID
		ne.Moment = time.Now()
		e = ne
		g.Round.State = RoundStateLegislating
	case TypeRequestExecutiveAction:
		ne := e.(RequestEvent)
		ne.ID = g.EventID
		ne.Moment = time.Now()
		e = ne
		g.Round.State = RoundStateExecutiveAction

		//PLAYER EVENTS
	case TypePlayerJoin:
		ne := e.(PlayerEvent)
		ne.ID = g.EventID
		ne.Moment = time.Now()
		e = ne
		g.Players = append(g.Players, ne.Player)
	case TypePlayerReady:
		ne := e.(PlayerEvent)
		ne.ID = g.EventID
		ne.Moment = time.Now()
		e = ne
		for i, p := range g.Players {
			if p.ID == ne.Player.ID {
				g.Players[i].Ready = true
				break
			}
		}
	case TypePlayerAcknowledge:
		ne := e.(PlayerEvent)
		ne.ID = g.EventID
		ne.Moment = time.Now()
		e = ne
		//Switch the given users ack attribute to true
		for i, p := range g.Players {
			if p.ID == ne.Player.ID {
				g.Players[i].Ack = true
				break
			}
		}
	case TypePlayerVote:
		ne := e.(PlayerVoteEvent)
		ne.ID = g.EventID
		ne.Moment = time.Now()
		e = ne
		//Add the given vote to the rounds vote array
		g.Round.Votes = append(g.Round.Votes, Vote{ne.PlayerID, ne.Vote})
	case TypePlayerNominate:
		ne := e.(PlayerPlayerEvent)
		ne.ID = g.EventID
		ne.Moment = time.Now()
		e = ne
		//Add the chancelor to the round object
		g.Round.ChancellorID = ne.OtherPlayerID
	case TypePlayerLegislate:
		ne := e.(PlayerLegislateEvent)
		ne.ID = g.EventID
		ne.Moment = time.Now()
		e = ne
	case TypePlayerInvestigate:
		ne := e.(PlayerPlayerEvent)
		ne.ID = g.EventID
		ne.Moment = time.Now()
		e = ne
		for i, p := range g.Players {
			if p.ID == ne.OtherPlayerID {
				g.Players[i].InvestigatedBy = ne.PlayerID
			}
		}
	case TypePlayerSpecialElection:
		ne := e.(PlayerPlayerEvent)
		ne.ID = g.EventID
		ne.Moment = time.Now()
		e = ne
		g.SpecialElectionPresidentID = ne.OtherPlayerID
		g.SpecialElectionRoundID = g.Round.ID + 1
	case TypePlayerExecute:
		ne := e.(PlayerPlayerEvent)
		ne.ID = g.EventID
		ne.Moment = time.Now()
		e = ne
		for i, p := range g.Players {
			if p.ID == ne.OtherPlayerID {
				g.Players[i].ExecutedBy = ne.PlayerID
			}
		}
		//GAME EVENTS
	case TypeGameUpdate:
		ne := e.(GameEvent)
		ne.ID = g.EventID
		ne.Moment = time.Now()
		e = ne
		//The event data, set the discard and draw pile accordingly
		if ne.Game.State == "-" {
			g.State = ""
		} else if ne.Game.State != "" {
			g.State = ne.Game.State
		}
		if ne.Game.WinningParty == "-" {
			g.WinningParty = ""
		} else if ne.Game.WinningParty != "" {
			g.WinningParty = ne.Game.WinningParty
		}
		if len(ne.Game.Draw) == 1 && ne.Game.Draw[0] == "-" {
			g.Draw = []string{}
		} else if len(ne.Game.Draw) > 0 {
			g.Draw = ne.Game.Draw
		}
		if len(ne.Game.Discard) == 1 && ne.Game.Discard[0] == "-" {
			g.Discard = []string{}
		} else if len(ne.Game.Discard) > 0 {
			g.Discard = ne.Game.Discard
		}
		if ne.Game.FailedVotes > 0 {
			g.FailedVotes = ne.Game.FailedVotes
		} else if ne.Game.FailedVotes == -1 {
			g.FailedVotes = 0
		}
		if ne.Game.Liberal > 0 {
			g.Liberal = ne.Game.Liberal
		} else if ne.Game.Liberal == -1 {
			g.Liberal = 0
		}
		if ne.Game.Facist > 0 {
			g.Facist = ne.Game.Facist
		} else if ne.Game.Facist == -1 {
			g.Facist = 0
		}
		if ne.Game.NextPresidentID == "-" {
			g.NextPresidentID = ""
		} else if ne.Game.NextPresidentID != "" {
			g.NextPresidentID = ne.Game.NextPresidentID
		}
		if len(ne.Game.Players) == 1 && ne.Game.Players[0].ID == "-" {
			g.Players = []Player{}
		} else if len(ne.Game.Players) > 0 {
			g.Players = ne.Game.Players
		}
		if ne.Game.PreviousPresidentID == "-" {
			g.PreviousPresidentID = ""
		} else if ne.Game.PreviousPresidentID != "" {
			g.PreviousPresidentID = ne.Game.PreviousPresidentID
		}
		if ne.Game.PreviousChancellorID == "-" {
			g.PreviousChancellorID = ""
		} else if ne.Game.PreviousChancellorID != "" {
			g.PreviousChancellorID = ne.Game.PreviousChancellorID
		}
		if ne.Game.SpecialElectionPresidentID == "-" {
			g.SpecialElectionPresidentID = ""
		} else if ne.Game.SpecialElectionPresidentID != "" {
			g.SpecialElectionPresidentID = ne.Game.SpecialElectionPresidentID
		}
		if ne.Game.SpecialElectionRoundID > 0 {
			g.SpecialElectionRoundID = ne.Game.SpecialElectionRoundID
		} else if ne.Game.SpecialElectionRoundID == -1 {
			g.SpecialElectionRoundID = 0
		}
		//Round Updates
		if ne.Game.Round.ID > 0 {
			g.Round.ID = ne.Game.Round.ID
		} else if ne.Game.Round.ID == -1 {
			g.Round.ID = 0
		}
		if ne.Game.Round.State == "-" {
			g.Round.State = ""
		} else if ne.Game.Round.State != "" {
			g.Round.State = ne.Game.Round.State
		}
		if ne.Game.Round.PresidentID == "-" {
			g.Round.PresidentID = ""
		} else if ne.Game.Round.PresidentID != "" {
			g.Round.PresidentID = ne.Game.Round.PresidentID
		}
		if ne.Game.Round.ChancellorID == "-" {
			g.Round.ChancellorID = ""
		} else if ne.Game.Round.ChancellorID != "" {
			g.Round.ChancellorID = ne.Game.Round.ChancellorID
		}
		if ne.Game.Round.EnactedPolicy == "-" {
			g.Round.EnactedPolicy = ""
		} else if ne.Game.Round.EnactedPolicy != "" {
			g.Round.EnactedPolicy = ne.Game.Round.EnactedPolicy
		}
		if ne.Game.Round.ExecutiveAction == "-" {
			g.Round.ExecutiveAction = ""
		} else if ne.Game.Round.ExecutiveAction != "" {
			g.Round.ExecutiveAction = ne.Game.Round.ExecutiveAction
		}
		if len(ne.Game.Round.Votes) == 1 && ne.Game.Round.Votes[0].PlayerID == "-" {
			g.Round.Votes = []Vote{}
		} else if len(ne.Game.Round.Votes) > 0 {
			g.Round.Votes = ne.Game.Round.Votes
		}
		if len(ne.Game.Round.Policies) == 1 && ne.Game.Round.Policies[0] == "-" {
			g.Round.Policies = []string{}
		} else if len(ne.Game.Round.Policies) > 1 {
			g.Round.Policies = ne.Game.Round.Policies
		}
	}

	return g, e, nil
}
