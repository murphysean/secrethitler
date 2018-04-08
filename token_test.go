package sh

import (
	"testing"
)

func TestCreateToken(t *testing.T) {
	token := createToken("testingtesting123", Token{
		Assertion:     "blah",
		EventID:       2,
		PlayerID:      "sean",
		OtherPlayerID: "ryan",
		RoundID:       10,
	})
	if token != "eyJhbGciOiJIUzI1NiJ9.eyJldmVudElEIjoyLCJwbGF5ZXJJRCI6InNlYW4iLCJhc3NlcnRpb24iOiJibGFoIiwicm91bmRJRCI6MTAsIm90aGVyUGxheWVySUQiOiJyeWFuIn0.jn_SLMx31mIAiLde_8VahHFoktC7XJgwbn_r8lODpN0" {
		t.Fatal(token)
	}
}

func TestValidateToken(t *testing.T) {
	token, err := validateToken("testingtesting123", "eyJhbGciOiJIUzI1NiJ9.eyJldmVudElEIjoyLCJwbGF5ZXJJRCI6InNlYW4iLCJhc3NlcnRpb24iOiJibGFoIiwicm91bmRJRCI6MTAsIm90aGVyUGxheWVySUQiOiJyeWFuIn0.jn_SLMx31mIAiLde_8VahHFoktC7XJgwbn_r8lODpN0")
	if err != nil {
		t.Fatal(err)
	}
	if token.Assertion != "blah" {
		t.Fatal("wrong assertion")
	}
}
