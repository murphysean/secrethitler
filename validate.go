package sh

import (
	"context"
	"errors"
	"time"
)

//Validate ensures that an event is consistent with the current state and then
//sends it to the event log.
func (g Game) Validate(ctx context.Context, e Event) error {
	pid, _ := ctx.Value("playerID").(string)
	//Players must all be ready for game to start
	switch e.GetType() {
	case TypePlayerJoin:
		pje := e.(PlayerEvent)
		if pje.Player.ID != pid {
			return errors.New("PlayerID must match currently authenticated user")
		}
		if g.State != GameStateLobby {
			return errors.New("Players can only join while the game is in the lobby state")
		}
		if len(g.Players) >= 10 {
			return errors.New("Max of 10 players allowed")
		}
		for _, p := range g.Players {
			if p.ID == pje.Player.ID {
				return errors.New("Player has already joined")
			}
		}
	case TypePlayerReady:
		pre := e.(PlayerEvent)
		if pre.Player.ID != pid {
			return errors.New("PlayerID must match currently authenticated user")
		}
		if g.State != GameStateLobby {
			return errors.New("Players can only ready while the game is in the lobby state")
		}
		//If the player doesn't exist, or isn't authenticated, they can't become ready
		for _, p := range g.Players {
			if p.ID == pre.Player.ID {
				if p.Ready {
					errors.New("Player is already ready")
				}
				return nil
			}
		}
		return errors.New("No player found with matching ID")
	case TypePlayerAcknowledge:
		pae := e.(PlayerEvent)
		if pae.Player.ID != pid {
			return errors.New("PlayerID must match currently authenticated user")
		}
		if g.State != GameStateInit {
			return errors.New("Players can only ack while the game is in the init state")
		}
		for _, p := range g.Players {
			if p.ID == pae.Player.ID {
				if p.Ack {
					return errors.New("Player has already acknowledged")
				}
				if p.Party != pae.Player.Party {
					return errors.New("Player must acknowledge assigned party")
				}
				if p.Role != pae.Player.Role {
					return errors.New("Player must acknowledge assigned role")
				}
				return nil
			}
		}
		return errors.New("No player found with matching ID")
	case TypePlayerNominate:
		ope := e.(PlayerPlayerEvent)
		if ope.PlayerID != pid {
			return errors.New("PlayerID must match currently authenticated user")
		}
		if g.Round.State != RoundStateNominating {
			return errors.New("Players can only nominate while the round is in the nominating state")
		}
		if g.Round.PresidentID != ope.PlayerID {
			return errors.New("Must be the round president to nominate a chancellor")
		}
		if ope.PlayerID == ope.OtherPlayerID {
			return errors.New("Must nominate another player as chancellor")
		}
		if g.PreviousChancellorID == ope.OtherPlayerID {
			return errors.New("Nominated player was previous chancellor")
		}
		//If there are only 5 alive players
		playersLeft := 0
		found := false
		for _, p := range g.Players {
			if p.ExecutedBy == "" {
				playersLeft++
			}
			if p.ID == ope.OtherPlayerID {
				if p.ExecutedBy != "" {
					return errors.New("The proposed player has been executed")
				}
				found = true
			}
		}
		if !found {
			return errors.New("Other player invalid")
		}
		if playersLeft > 5 {
			if g.PreviousPresidentID == ope.OtherPlayerID {
				return errors.New("Nominated player was previous president")
			}
		}
	case TypePlayerVote:
		pve := e.(PlayerVoteEvent)
		if pve.PlayerID != pid {
			return errors.New("PlayerID must match currently authenticated user")
		}
		if g.Round.State != RoundStateVoting {
			return errors.New("Players can only vote while the round is in the voting state")
		}
		for _, v := range g.Round.Votes {
			if pve.PlayerID == v.PlayerID {
				return errors.New("Players can only vote once per round")
			}
		}
		found := false
		for _, p := range g.Players {
			if p.ID == pve.PlayerID {
				found = true
				if p.ExecutedBy != "" {
					return errors.New("Executed players can't vote")
				}
			}
		}
		if !found {
			return errors.New("Voting player not found")
		}
	case TypePlayerLegislate:
		ple := e.(PlayerLegislateEvent)
		if ple.PlayerID != pid {
			return errors.New("PlayerID must match currently authenticated user")
		}
		if g.Round.State != RoundStateLegislating {
			return errors.New("Players can only legislate while the round is in the legislating state")
		}
		if len(g.Round.Policies) == 3 {
			if g.Round.PresidentID != ple.PlayerID {
				return errors.New("Only the president can discard the first card in a round")
			}
		} else if len(g.Round.Policies) == 2 {
			if g.Round.ChancellorID != ple.PlayerID {
				return errors.New("Only the chancellor can discard the second card in a round")
			}
		} else if len(g.Round.Policies) == 1 {
			if g.Round.PresidentID != ple.PlayerID {
				return errors.New("Only the president can discard the last card with a veto")
			}
		}
		if g.Fascist < 5 && ple.Veto {
			return errors.New("Can only veto on the 5th fascist policy")
		} else if !ple.Veto {
			found := false
			for _, c := range g.Round.Policies {
				if c == ple.Discard {
					found = true
				}
			}
			if !found {
				return errors.New("Discarded policy must be one of the available options")
			}
		}
	case TypePlayerInvestigate:
		ope := e.(PlayerPlayerEvent)
		if ope.PlayerID != pid {
			return errors.New("PlayerID must match currently authenticated user")
		}
		if g.Round.State != RoundStateExecutiveAction {
			return errors.New("Players can only investigate while the round is in the executive_action state")
		}
		if g.Round.PresidentID != ope.PlayerID {
			return errors.New("Only the president can investigate as an executive action")
		}
		if g.Round.ExecutiveAction != ExecutiveActionInvestigate {
			return errors.New("The round did not result in an investigate executive action")
		}
		if ope.PlayerID == ope.OtherPlayerID {
			return errors.New("Must investigate another player")
		}
		if op, err := g.GetPlayerByID(ope.OtherPlayerID); err != nil {
			return errors.New("Other player invalid")
		} else {
			if op.InvestigatedBy != "" {
				return errors.New("This player has been previously investigated")
			}
		}
	case TypePlayerSpecialElection:
		ope := e.(PlayerPlayerEvent)
		if ope.PlayerID != pid {
			return errors.New("PlayerID must match currently authenticated user")
		}
		if g.Round.State != RoundStateExecutiveAction {
			return errors.New("Players can only call a special election while the round is in the executive_action state")
		}
		if g.Round.PresidentID != ope.PlayerID {
			return errors.New("Only the president can call a special election")
		}
		if g.Round.ExecutiveAction != ExecutiveActionSpecialElection {
			return errors.New("The round did not result in an special election executive action")
		}
		if ope.PlayerID == ope.OtherPlayerID {
			return errors.New("Must pick another player")
		}
		if _, err := g.GetPlayerByID(ope.OtherPlayerID); err != nil {
			return errors.New("Other player invalid")
		}
	case TypePlayerExecute:
		ope := e.(PlayerPlayerEvent)
		if ope.PlayerID != pid {
			return errors.New("PlayerID must match currently authenticated user")
		}
		if g.Round.State != RoundStateExecutiveAction {
			return errors.New("Players can only execute while the round is in the executive_action state")
		}
		if g.Round.PresidentID != ope.PlayerID {
			return errors.New("Only the president can execute as an executive action")
		}
		if g.Round.ExecutiveAction != ExecutiveActionExecute {
			return errors.New("The round did not result in an execute executive action")
		}
		if ope.PlayerID == ope.OtherPlayerID {
			return errors.New("Must execute another player")
		}
		if op, err := g.GetPlayerByID(ope.OtherPlayerID); err != nil {
			return errors.New("Other player invalid")
		} else {
			if op.ExecutedBy != "" {
				return errors.New("This player has been previously executed")
			}
		}
	case TypePlayerMessage:
		me := e.(MessageEvent)
		if me.PlayerID != pid {
			return errors.New("PlayerID must match currently authenticated user")
		}
		if p, _ := g.GetPlayerByID(pid); time.Now().Sub(p.LastAction) < time.Second {
			return errors.New("Throttle limit reached on messages")
		}
	case TypeReactPlayer:
		//TODO Other player must exist
		fallthrough
	case TypeReactEventID:
		//TODO React event id must be valid
		fallthrough
	case TypeReactStatus:
		re := e.(ReactEvent)
		if re.PlayerID != pid {
			return errors.New("PlayerID must match currently authenticated user")
		}
		if p, _ := g.GetPlayerByID(pid); re.Moment.Sub(p.LastAction) < time.Second {
			return errors.New("Throttle limit reached on reactions")
		}
	case TypeGuess:
		ge := e.(GuessEvent)
		if ge.PlayerID != pid {
			return errors.New("PlayerID must match currently authenticated user")
		}
		if p, _ := g.GetPlayerByID(pid); ge.Moment.Sub(p.LastAction) < time.Second {
			return errors.New("Throttle limit reached on guesses")
		}
	case TypeAssertPolicies:
		ae := e.(AssertEvent)
		if ae.PlayerID != pid {
			return errors.New("PlayerID must match currently authenticated user")
		}
		if g.Round.State == RoundStateLegislating && ae.PolicySource == TypeRequestLegislate {
			return errors.New("Can't reveal information during legislation")
		}
		//Token must validate
		t, err := validateToken(g.Secret, ae.Token)
		if err != nil {
			return err
		}
		if t.PlayerID != ae.PlayerID {
			return errors.New("PlayerID must match token")
		}
		if t.RoundID != ae.RoundID {
			return errors.New("RoundID must match token")
		}
		if t.Assertion != ae.PolicySource {
			return errors.New("Policy Source must match token")
		}
		if t.PolicyCount != len(ae.Policies) {
			return errors.New("Number of policies must match those revealed")
		}
	case TypeAssertParty:
		ae := e.(AssertEvent)
		if ae.PlayerID != pid {
			return errors.New("PlayerID must match currently authenticated user")
		}
		//Token must validate
		t, err := validateToken(g.Secret, ae.Token)
		if err != nil {
			return err
		}
		if t.PlayerID != ae.PlayerID {
			return errors.New("PlayerID must match token")
		}
		if t.RoundID != ae.RoundID {
			return errors.New("RoundID must match token")
		}
		if t.OtherPlayerID != ae.OtherPlayerID {
			return errors.New("OtherPlayerID must match token")
		}
	default:
		if pid != "admin" && pid != "engine" {
			return errors.New("Not Authorized")
		}
	}

	return nil
}
