package sh

import (
	cr "crypto/rand"
	"fmt"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UTC().UnixNano())
}

func nextIndex(len, idx int) int {
	if idx+1 >= len {
		return 0
	}
	return idx + 1
}

func genUUIDv4() string {
	u := make([]byte, 16)
	cr.Read(u)
	//Set the version to 4
	u[6] = (u[6] | 0x40) & 0x4F
	u[8] = (u[8] | 0x80) & 0xBF
	return fmt.Sprintf("%x-%x-%x-%x-%x", u[0:4], u[4:6], u[6:8], u[8:10], u[10:])
}

func (gs Game) createNextRound() []Event {
	ge := GameEvent{}
	if gs.Secret == "" {
		ge.Game.Secret = genUUIDv4()
	}
	ge.Type = TypeGameUpdate
	ge.Game.State = GameStateStarted
	ge.Game.Round.ID = gs.Round.ID + 1
	ge.Game.Round.State = RoundStateNominating
	ge.Game.Round.PresidentID = "-"
	ge.Game.Round.ChancellorID = "-"
	ge.Game.Round.EnactedPolicy = "-"
	ge.Game.Round.ExecutiveAction = "-"
	ge.Game.Round.Votes = []Vote{Vote{PlayerID: "-"}}
	ge.Game.Round.Policies = []string{"-"}

	//Is the next round a special election?
	if ge.Game.Round.ID == gs.SpecialElectionRoundID {
		ge.Game.Round.PresidentID = gs.SpecialElectionPresidentID
		ge.Game.NextPresidentID = gs.NextPresidentID
	} else {
		//Go to the next unexecuted president in the array
		//Next president is the next one in the array, that's not dead
		pi := -1
		for i, p := range gs.Players {
			if gs.NextPresidentID == p.ID {
				pi = i
				break
			}
		}
		for {
			//If president index is not dead, break
			if gs.Players[pi].ExecutedBy == "" {
				break
			} else {
				pi = nextIndex(len(gs.Players), pi)
			}
		}
		ge.Game.Round.PresidentID = gs.Players[pi].ID
		npi := nextIndex(len(gs.Players), pi)
		for {
			//If president index is not dead, break
			if gs.Players[npi].ExecutedBy == "" {
				break
			} else {
				npi = nextIndex(len(gs.Players), npi)
			}
		}
		ge.Game.NextPresidentID = gs.Players[npi].ID
	}

	return []Event{ge, RequestEvent{
		BaseEvent: BaseEvent{Type: TypeRequestNominate},
		PlayerID:  ge.Game.Round.PresidentID,
		RoundID:   ge.Game.Round.ID,
	}}
}

func executiveAction(numPlayers, numFascistPolicies int) string {
	switch numFascistPolicies {
	case 1:
		if numPlayers > 8 {
			return ExecutiveActionInvestigate
		}
	case 2:
		if numPlayers > 6 {
			return ExecutiveActionInvestigate
		}
	case 3:
		if numPlayers > 6 {
			return ExecutiveActionSpecialElection
		} else {
			return ExecutiveActionPeek
		}
	case 4:
		return ExecutiveActionExecute
	case 5:
		return ExecutiveActionExecute
	}
	return ""
}

//The engine will read the incoming event and process it to see if a new event
// should be created to update the game state. This function itself should not modify the game
// state in any way other than returning events that will.
func (g Game) Engine(e Event) ([]Event, error) {
	ret := []Event{}

	switch e.GetType() {
	case TypePlayerReady:
		allReady := false
		if len(g.Players) >= 5 {
			allReady = true
			for _, p := range g.Players {
				if !p.Ready {
					allReady = false
				}
			}
		}
		if allReady {
			ge := GameEvent{}
			ge.Type = TypeGameUpdate
			ge.Game.State = GameStateInit
			ge.Game.Draw = make([]string, 0)
			for i := 0; i < 11; i++ {
				ge.Game.Draw = append(ge.Game.Draw, PolicyFascist)
			}
			for i := 0; i < 6; i++ {
				ge.Game.Draw = append(ge.Game.Draw, PolicyLiberal)
			}
			rand.Shuffle(len(ge.Game.Draw), func(i, j int) {
				ge.Game.Draw[i], ge.Game.Draw[j] = ge.Game.Draw[j], ge.Game.Draw[i]
			})
			roles := []string{RoleLiberal, RoleLiberal, RoleLiberal, RoleHitler, RoleFascist}
			if len(g.Players) > 5 {
				roles = append(roles, RoleLiberal)
			}
			if len(g.Players) > 6 {
				roles = append(roles, RoleFascist)
			}
			if len(g.Players) > 7 {
				roles = append(roles, RoleLiberal)
			}
			if len(g.Players) > 8 {
				roles = append(roles, RoleFascist)
			}
			if len(g.Players) > 9 {
				roles = append(roles, RoleLiberal)
			}
			rand.Shuffle(len(roles), func(i, j int) {
				roles[i], roles[j] = roles[j], roles[i]
			})
			for i, p := range g.Players {
				p.Role = roles[i]
				if p.Role == RoleLiberal {
					p.Party = PartyLiberal
				} else {
					p.Party = PartyFascist
				}
				ge.Game.Players = append(ge.Game.Players, p)
			}
			ge.Game.NextPresidentID = g.Players[rand.Intn(len(g.Players)-1)].ID
			ret = append(ret, ge, RequestEvent{
				BaseEvent: BaseEvent{Type: TypeRequestAcknowledge},
				PlayerID:  PlayerIDAll,
			})
		}
	case TypePlayerAcknowledge:
		allAck := true
		for _, p := range g.Players {
			if !p.Ack {
				allAck = false
			}
		}
		if allAck {
			ret = append(ret, g.createNextRound()...)
		}
	case TypePlayerNominate:
		ret = append(ret, GameEvent{
			BaseEvent: BaseEvent{Type: TypeGameUpdate},
			Game: Game{
				Round: Round{
					State: RoundStateVoting,
				},
			},
		})
		ret = append(ret, RequestEvent{
			BaseEvent:    BaseEvent{Type: TypeRequestVote},
			PlayerID:     PlayerIDAll,
			RoundID:      g.Round.ID,
			PresidentID:  g.Round.PresidentID,
			ChancellorID: g.Round.ChancellorID,
		})
	case TypePlayerVote:
		//If all the votes are in...
		votesIn := make(map[string]bool)
		c := 0
		for _, v := range g.Round.Votes {
			votesIn[v.PlayerID] = true
			if v.Vote {
				c++
			}
		}
		allIn := true
		for _, p := range g.Players {
			if p.ExecutedBy == "" {
				if !votesIn[p.ID] {
					allIn = false
					break
				}
			}
		}
		if allIn {
			succeeded := ((float64(c) / float64(len(g.Round.Votes))) * 100) > 50.0
			//Send out an event
			ret = append(ret, VoteResultEvent{
				BaseEvent: BaseEvent{Type: TypeGameVoteResults},
				Succeeded: succeeded,
				RoundID:   g.Round.ID,
				Votes:     g.Round.Votes,
			})
			if succeeded {
				//If secret hitler is elected chancellor with 3 fascist polices down, fascists win
				if g.Fascist > 2 {
					for _, p := range g.Players {
						if p.ID == g.Round.ChancellorID {
							if p.Role == RoleHitler {
								ret = append(ret, GameEvent{
									BaseEvent: BaseEvent{Type: TypeGameUpdate},
									Game: Game{
										State:        GameStateFinished,
										WinningParty: PartyFascist,
									},
								}, FinishedEvent{
									BaseEvent:        BaseEvent{Type: TypeGameFinished},
									WinningCondition: ConditionHitlerChancellor,
									WinningParty:     PartyFascist,
								})
								return ret, nil
							}
						}
					}
				}
				//Start legislating
				newdraw := g.Draw[:len(g.Draw)-3]
				if len(newdraw) == 0 {
					newdraw = []string{"-"}
				}
				ret = append(ret, GameEvent{
					BaseEvent: BaseEvent{Type: TypeGameUpdate},
					Game: Game{
						Draw:                 newdraw,
						Discard:              g.Discard,
						PreviousPresidentID:  g.Round.PresidentID,
						PreviousChancellorID: g.Round.ChancellorID,
						Round: Round{
							Policies: g.Draw[len(g.Draw)-3:],
							State:    RoundStateLegislating,
						},
					},
				})
				ret = append(ret, RequestEvent{
					BaseEvent:    BaseEvent{Type: TypeRequestLegislate},
					PlayerID:     g.Round.PresidentID,
					RoundID:      g.Round.ID,
					Policies:     g.Draw[len(g.Draw)-3:],
					VetoPossible: g.Fascist > 4,
					Token: createToken(g.Secret, Token{
						EventID:     g.EventID,
						Assertion:   TypeRequestLegislate,
						PlayerID:    g.Round.PresidentID,
						RoundID:     g.Round.ID,
						PolicyCount: 3,
					}),
				})
			} else {
				//If the vote failed, enact a policy if failed votes = 3
				if g.ElectionTracker > 1 {
					ge := GameEvent{
						BaseEvent: BaseEvent{Type: TypeGameUpdate},
						Game: Game{
							ElectionTracker:      -1,
							PreviousPresidentID:  "-",
							PreviousChancellorID: "-",
						},
					}
					//Pop the top policy off the draw pile and enact it
					tp := g.Draw[len(g.Draw)-1]
					if tp == PolicyLiberal {
						ge.Game.Liberal = g.Liberal + 1
						ge.Game.PreviousEnactedPolicy = PolicyLiberal
					} else {
						ge.Game.Fascist = g.Fascist + 1
						ge.Game.PreviousEnactedPolicy = PolicyFascist
					}
					ge.Game.Draw = g.Draw[:len(g.Draw)-1]
					//Shuffle if there are < 3 policies in the draw pile
					if len(ge.Game.Draw) < 3 {
						ge.Game.Draw = append(ge.Game.Draw, g.Discard...)
						ge.Game.Discard = []string{"-"}
						rand.Shuffle(len(ge.Game.Draw), func(i, j int) {
							ge.Game.Draw[i], ge.Game.Draw[j] = ge.Game.Draw[j], ge.Game.Draw[i]
						})
					}
					over := false
					if ge.Game.Fascist > 5 {
						ge.Game.State = GameStateFinished
						ge.Game.WinningParty = PartyFascist
						over = true
					}
					if ge.Game.Liberal > 4 {
						ge.Game.State = GameStateFinished
						ge.Game.WinningParty = PartyLiberal
						over = true
					}
					ret = append(ret, ge)
					if !over {
						ret = append(ret, g.createNextRound()...)
					} else {
						ret = append(ret, FinishedEvent{
							BaseEvent:        BaseEvent{Type: TypeGameFinished},
							WinningCondition: ConditionPoliciesEnacted,
							WinningParty:     ge.Game.WinningParty,
						})
					}
				} else {
					ret = append(ret, GameEvent{
						BaseEvent: BaseEvent{Type: TypeGameUpdate},
						Game: Game{
							ElectionTracker: g.ElectionTracker + 1,
						},
					})
					//End the round now, start a new one
					ret = append(ret, g.createNextRound()...)
				}
			}
		}
	case TypePlayerLegislate:
		le := e.(PlayerLegislateEvent)
		ge := GameEvent{
			BaseEvent: BaseEvent{Type: TypeGameUpdate},
		}
		//If this was a veto vote, and no discard was selected, pick one
		if le.Discard == "" && len(g.Round.Policies) > 0 {
			le.Discard = g.Round.Policies[0]
		}

		chancellorVeto := false
		//If the chancellor sends a veto = true with a discard, the president will need to confirm
		if le.Veto && len(g.Round.Policies) == 2 {
			//Send out another request to the president
			ret = append(ret, RequestEvent{
				BaseEvent:    BaseEvent{Type: TypeRequestLegislate},
				PlayerID:     g.Round.PresidentID,
				RoundID:      g.Round.ID,
				VetoPossible: true,
				Veto:         true,
			})
			//At this point we will want to discard the second card, but will not want to play it
			chancellorVeto = true
		}

		if len(g.Round.Policies) > 1 {
			//First subtract the discarded policy from the round policies
			ge.Game.Round.Policies = removeElement(g.Round.Policies, le.Discard)
			//Second add it to the game discard pile
			ge.Game.Discard = append(g.Discard, le.Discard)
		} else {
			ge.Game.Round.Policies = g.Round.Policies
		}

		//Now if there is only one remaining, play it... unless this is the chancellor asking for a veto
		over := false
		if len(ge.Game.Round.Policies) == 1 && !chancellorVeto {
			//Is this because it's the president responding to a veto?
			if le.Veto {
				//Discard the last remaining tile
				ge.Game.Discard = append(g.Discard, ge.Game.Round.Policies[0])
				ge.Game.Round.Policies = []string{"-"}
				ge.Game.ElectionTracker = g.ElectionTracker + 1
			} else {
				ge.Game.Round.EnactedPolicy = ge.Game.Round.Policies[0]
				ge.Game.PreviousEnactedPolicy = ge.Game.Round.EnactedPolicy
				if ge.Game.Round.EnactedPolicy == PolicyLiberal {
					ge.Game.Liberal = g.Liberal + 1
				} else {
					ge.Game.Fascist = g.Fascist + 1
					//If a card was played on a fascist, trigger an executive action, or ea request
					ge.Game.Round.ExecutiveAction = executiveAction(len(g.Players), ge.Game.Fascist)
				}
				if ge.Game.Fascist > 5 {
					ge.Game.State = GameStateFinished
					ge.Game.WinningParty = PartyFascist
					over = true
				}
				if ge.Game.Liberal > 4 {
					ge.Game.State = GameStateFinished
					ge.Game.WinningParty = PartyLiberal
					over = true
				}
				ge.Game.ElectionTracker = -1
				ge.Game.Round.Policies = []string{"-"}
			}
			//Shuffle if there are < 3 policies in the draw pile
			if len(g.Draw) < 3 {
				ge.Game.Draw = append(g.Draw, ge.Game.Discard...)
				ge.Game.Discard = []string{"-"}
				rand.Shuffle(len(ge.Game.Draw), func(i, j int) {
					ge.Game.Draw[i], ge.Game.Draw[j] = ge.Game.Draw[j], ge.Game.Draw[i]
				})
			}
			//if failed votes is == 3, flip top policy
			if ge.Game.ElectionTracker > 2 {
				ge.Game.ElectionTracker = -1
				ge.Game.PreviousPresidentID = "-"
				ge.Game.PreviousChancellorID = "-"
				var tp string
				if len(ge.Game.Draw) > 0 {
					tp = ge.Game.Draw[len(g.Draw)-1]
					ge.Game.Draw = ge.Game.Draw[:len(ge.Game.Draw)-1]
				} else {
					tp = g.Draw[len(g.Draw)-1]
					ge.Game.Draw = g.Draw[:len(g.Draw)-1]
				}
				if tp == PolicyLiberal {
					ge.Game.Liberal = g.Liberal + 1
					ge.Game.PreviousEnactedPolicy = PolicyLiberal
				} else {
					ge.Game.Fascist = g.Fascist + 1
					ge.Game.PreviousEnactedPolicy = PolicyFascist
				}
				if ge.Game.Fascist > 5 {
					ge.Game.State = GameStateFinished
					ge.Game.WinningParty = PartyFascist
					over = true
				}
				if ge.Game.Liberal > 4 {
					ge.Game.State = GameStateFinished
					ge.Game.WinningParty = PartyLiberal
					over = true
				}
			}
		}

		ret = append(ret, ge)
		if over {
			ret = append(ret, FinishedEvent{
				BaseEvent:        BaseEvent{Type: TypeGameFinished},
				WinningCondition: ConditionPoliciesEnacted,
				WinningParty:     ge.Game.WinningParty,
			})
			return ret, nil
		}
		//Trigger an executive action if round policies are empty
		if len(ge.Game.Round.Policies) == 1 && ge.Game.Round.Policies[0] == "-" {
			if ge.Game.Round.EnactedPolicy == PolicyFascist {
				switch ge.Game.Round.ExecutiveAction {
				case ExecutiveActionInvestigate:
					ret = append(ret, RequestEvent{
						BaseEvent:       BaseEvent{Type: TypeRequestExecutiveAction},
						PlayerID:        g.Round.PresidentID,
						RoundID:         g.Round.ID,
						ExecutiveAction: ExecutiveActionInvestigate,
					})
				case ExecutiveActionPeek:
					var pp []string
					if len(g.Draw) > 2 {
						pp = g.Draw[len(g.Draw)-3:]
					} else {
						pp = ge.Game.Draw[len(ge.Game.Draw)-3:]
					}
					ret = append(ret, InformationEvent{
						BaseEvent: BaseEvent{Type: TypeGameInformation},
						PlayerID:  g.Round.PresidentID,
						RoundID:   g.Round.ID,
						Policies:  pp,
						Token: createToken(g.Secret, Token{
							PlayerID:    g.Round.PresidentID,
							EventID:     g.EventID,
							RoundID:     g.Round.ID,
							Assertion:   ExecutiveActionPeek,
							PolicyCount: 3,
						}),
					})
					ret = append(ret, g.createNextRound()...)
				case ExecutiveActionSpecialElection:
					ret = append(ret, RequestEvent{
						BaseEvent:       BaseEvent{Type: TypeRequestExecutiveAction},
						PlayerID:        g.Round.PresidentID,
						RoundID:         g.Round.ID,
						ExecutiveAction: ExecutiveActionSpecialElection,
					})
				case ExecutiveActionExecute:
					ret = append(ret, RequestEvent{
						BaseEvent:       BaseEvent{Type: TypeRequestExecutiveAction},
						PlayerID:        g.Round.PresidentID,
						RoundID:         g.Round.ID,
						ExecutiveAction: ExecutiveActionExecute,
					})
				default:
					//If no exeutive action, start a new round
					ret = append(ret, g.createNextRound()...)
				}
			} else {
				ret = append(ret, g.createNextRound()...)
			}
		}
		if len(ge.Game.Round.Policies) > 1 {
			//Trigger a legislate chancellor with the remaining cards
			ret = append(ret, RequestEvent{
				BaseEvent:    BaseEvent{Type: TypeRequestLegislate},
				PlayerID:     g.Round.ChancellorID,
				RoundID:      g.Round.ID,
				Policies:     ge.Game.Round.Policies,
				VetoPossible: g.Fascist > 4,
				Token: createToken(g.Secret, Token{
					EventID:     g.EventID,
					Assertion:   TypeRequestLegislate,
					PlayerID:    g.Round.ChancellorID,
					RoundID:     g.Round.ID,
					PolicyCount: 2,
				}),
			})
		}
	case TypePlayerInvestigate:
		//Give out the information!
		te := e.(PlayerPlayerEvent)
		party := PartyMasked
		for _, p := range g.Players {
			if p.ID == te.OtherPlayerID {
				party = p.Party
			}
		}
		ret = append(ret, InformationEvent{
			BaseEvent:     BaseEvent{Type: TypeGameInformation},
			PlayerID:      g.Round.PresidentID,
			OtherPlayerID: te.OtherPlayerID,
			RoundID:       g.Round.ID,
			Party:         party,
			Token: createToken(g.Secret, Token{
				PlayerID:      g.Round.PresidentID,
				OtherPlayerID: te.OtherPlayerID,
				EventID:       g.EventID,
				RoundID:       g.Round.ID,
				Assertion:     ExecutiveActionInvestigate,
			}),
		})
		ret = append(ret, g.createNextRound()...)
	case TypePlayerSpecialElection:
		ret = append(ret, g.createNextRound()...)
	case TypePlayerExecute:
		//If hitler is assasinated, game over for fascists
		for _, p := range g.Players {
			if p.Role == RoleHitler && p.ExecutedBy != "" {
				ret = append(ret, GameEvent{
					BaseEvent: BaseEvent{Type: TypeGameUpdate},
					Game: Game{
						State:        GameStateFinished,
						WinningParty: PartyLiberal,
					},
				}, FinishedEvent{
					BaseEvent:        BaseEvent{Type: TypeGameFinished},
					WinningCondition: ConditionHitlerExecuted,
					WinningParty:     PartyLiberal,
				})
				return ret, nil
			}
		}
		ret = append(ret, g.createNextRound()...)
	}
	return ret, nil
}

func removeElement(a []string, e string) []string {
	i := -1
	for c, v := range a {
		if v == e {
			i = c
			break
		}
	}
	if i >= 0 {
		a[i] = a[len(a)-1]
		a = a[:len(a)-1]
	}
	return a
}

func removeAtIndex(s []string, i int) []string {
	s[len(s)-1], s[i] = s[i], s[len(s)-1]
	return s[:len(s)-1]
}
