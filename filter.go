package sh

import (
	"context"
)

func (g Game) Filter(ctx context.Context) Game {
	playerID := ctx.Value("playerID").(string)
	if playerID == "admin" || g.State == GameStateFinished {
		return g
	}
	me, _ := g.GetPlayerByID(playerID)
	//Filter the game secret
	if g.Secret != "" {
		g.Secret = "masked"
	}
	//Filter the draw and dscard pile
	g.Draw = maskedPolicies(len(g.Draw))
	g.Discard = maskedPolicies(len(g.Discard))
	//Filter the player roles
	nps := []Player{}
	for _, p := range g.Players {
		np := Player{
			ID:             p.ID,
			Name:           p.Name,
			Ready:          p.Ready,
			Ack:            p.Ack,
			Party:          PartyMasked,
			Role:           RoleMasked,
			InvestigatedBy: p.InvestigatedBy,
			ExecutedBy:     p.ExecutedBy,
		}
		if me.ID == p.ID {
			np.Party = p.Party
			np.Role = p.Role
		}
		if me.ID == p.InvestigatedBy {
			np.Party = p.Party
		}
		if me.Role == RoleFacist || (len(g.Players) < 7 && me.Role == RoleHitler) {
			np.Party = p.Party
			np.Role = p.Role
		}
		nps = append(nps, np)
	}
	g.Players = nps
	//Filter the round votes
	if g.Round.State == RoundStateVoting {
		vs := make([]Vote, len(g.Round.Votes))
		for i, v := range g.Round.Votes {
			if me.ID == v.PlayerID {
				vs[i] = v
			} else {
				vs[i].PlayerID = v.PlayerID
				vs[i].Vote = false
			}
		}
		g.Round.Votes = vs
	}
	//Filter the round policies
	if me.ID != g.Round.PresidentID && me.ID != g.Round.ChancellorID {
		g.Round.Policies = maskedPolicies(len(g.Round.Policies))
	} else if me.ID == g.Round.ChancellorID && len(g.Round.Policies) > 2 {
		g.Round.Policies = maskedPolicies(len(g.Round.Policies))
	}

	return g
}

func maskedPolicies(len int) []string {
	ret := make([]string, len)
	for i := 0; i < len; i++ {
		ret[i] = PolicyMasked
	}
	return ret
}
