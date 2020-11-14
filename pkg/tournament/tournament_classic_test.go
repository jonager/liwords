package tournament

import (
	"errors"
	"fmt"
	"github.com/matryer/is"
	"math/rand"
	"strings"
	"testing"

	"github.com/domino14/liwords/pkg/entity"

	realtime "github.com/domino14/liwords/rpc/api/proto/realtime"
)

var playerStrings = []string{"Will", "Josh", "Conrad", "Jesse"}
var playersOddStrings = []string{"Will", "Josh", "Conrad", "Jesse", "Matt"}
var rounds = 2
var defaultFirsts = []entity.FirstMethod{entity.ManualFirst, entity.ManualFirst}
var defaultGamesPerRound = []int{1, 1}

func TestClassicDivisionRandom(t *testing.T) {
	// This test attempts to cover the basic
	// functions of a Classic Tournament

	is := is.New(t)

	// Tournaments must have at least two players
	tc, err := NewClassicDivision([]string{"Sad"}, rounds, defaultGamesPerRound, []entity.PairingMethod{entity.Random, entity.Random}, defaultFirsts)
	is.True(err != nil)

	// Tournaments must have at least 1 round
	tc, err = NewClassicDivision(playerStrings, 0, defaultGamesPerRound, []entity.PairingMethod{entity.Random, entity.Random}, defaultFirsts)
	is.True(err != nil)

	// Tournaments must have an equal number of rounds and pairing methods
	tc, err = NewClassicDivision(playerStrings, rounds, defaultGamesPerRound, []entity.PairingMethod{entity.Random}, defaultFirsts)
	is.True(err != nil)

	// Tournaments must have an equal number of rounds and first methods
	tc, err = NewClassicDivision(playerStrings, rounds, defaultGamesPerRound, []entity.PairingMethod{entity.Random}, []entity.FirstMethod{})
	is.True(err != nil)

	tc, err = NewClassicDivision(playerStrings, rounds, defaultGamesPerRound, []entity.PairingMethod{entity.Random, entity.Random}, defaultFirsts)
	is.NoErr(err)
	is.True(tc != nil)

	is.NoErr(validatePairings(tc, 0))

	// Test getting a nonexistent round
	_, err = tc.GetPlayerRoundInfo("Josh", 9)
	is.True(err != nil)

	// Test getting a nonexistent player
	_, err = tc.GetPlayerRoundInfo("No one", 1)
	is.True(err != nil)

	playerPairings := getPlayerPairings(tc.Players, tc.Matrix[0])
	player1 := playerPairings[0]
	player2 := playerPairings[1]
	player3 := playerPairings[2]
	player4 := playerPairings[3]

	pri1, err := tc.GetPlayerRoundInfo(player1, 0)
	is.NoErr(err)
	pri2, err := tc.GetPlayerRoundInfo(player3, 0)
	is.NoErr(err)

	expectedpri1 := newPlayerRoundInfo(tc, player1, player2, tc.GamesPerRound[0], 0)
	expectedpri2 := newPlayerRoundInfo(tc, player3, player4, tc.GamesPerRound[0], 0)

	// Submit result for an unpaired round
	err = tc.SubmitResult(1, player1, player2, 10000, -40, realtime.TournamentGameResult_WIN,
		realtime.TournamentGameResult_LOSS, realtime.GameEndReason_STANDARD, false, 0)
	is.True(err != nil)

	// The result and record should remain unchanged
	is.NoErr(equalPRI(expectedpri1, pri1))

	// Submit result for players that didn't player each other
	err = tc.SubmitResult(1, player1, player3, 10000, -40, realtime.TournamentGameResult_WIN,
		realtime.TournamentGameResult_LOSS, realtime.GameEndReason_STANDARD, false, 0)
	is.True(err != nil)

	// The result and record should remain unchanged
	is.NoErr(equalPRI(expectedpri1, pri1))

	// Submit a result for a paired round
	err = tc.SubmitResult(0, player1, player2, 10000, -40, realtime.TournamentGameResult_WIN,
		realtime.TournamentGameResult_LOSS, realtime.GameEndReason_STANDARD, false, 0)
	is.NoErr(err)

	// The result and record should have changed
	expectedpri1.Pairing.Games[0].Results =
		[]realtime.TournamentGameResult{realtime.TournamentGameResult_WIN, realtime.TournamentGameResult_LOSS}
	expectedpri1.Pairing.Games[0].Scores[0] = 10000
	expectedpri1.Pairing.Games[0].Scores[1] = -40
	expectedpri1.Pairing.Games[0].GameEndReason = realtime.GameEndReason_STANDARD
	expectedpri1.Pairing.Outcomes[0] = realtime.TournamentGameResult_WIN
	expectedpri1.Pairing.Outcomes[1] = realtime.TournamentGameResult_LOSS
	is.NoErr(equalPRI(expectedpri1, pri1))

	// Attempt to submit the same result
	err = tc.SubmitResult(0, player1, player2, 10000, -40, realtime.TournamentGameResult_WIN,
		realtime.TournamentGameResult_LOSS, realtime.GameEndReason_STANDARD, false, 0)
	is.True(err != nil)

	// The result and record should remain unchanged
	is.NoErr(equalPRI(expectedpri1, pri1))

	// Round 2 should not have been paired,
	// so attempting to submit a result for
	// it will throw an error.
	err = tc.SubmitResult(1, player1, player2, 10000, -40,
		realtime.TournamentGameResult_WIN,
		realtime.TournamentGameResult_LOSS,
		realtime.GameEndReason_STANDARD, false, 0)
	is.True(err != nil)

	// The result and record should remain unchanged
	is.NoErr(equalPRI(expectedpri1, pri1))

	// Amend the result
	err = tc.SubmitResult(0, player1, player2, 30, 900,
		realtime.TournamentGameResult_LOSS,
		realtime.TournamentGameResult_WIN,
		realtime.GameEndReason_STANDARD, true, 0)
	is.NoErr(err)

	// The result and record should be amended
	expectedpri1.Pairing.Games[0].Results =
		[]realtime.TournamentGameResult{realtime.TournamentGameResult_LOSS,
			realtime.TournamentGameResult_WIN}
	expectedpri1.Pairing.Games[0].Scores[0] = 30
	expectedpri1.Pairing.Games[0].Scores[1] = 900
	expectedpri1.Pairing.Outcomes[0] = realtime.TournamentGameResult_LOSS
	expectedpri1.Pairing.Outcomes[1] = realtime.TournamentGameResult_WIN
	is.NoErr(equalPRI(expectedpri1, pri1))

	// Submit the final result for round 1
	err = tc.SubmitResult(0, player3, player4, 1, 1,
		realtime.TournamentGameResult_DRAW,
		realtime.TournamentGameResult_DRAW,
		realtime.GameEndReason_ABANDONED, false, 0)
	is.NoErr(err)

	expectedpri2.Pairing.Games[0].Results =
		[]realtime.TournamentGameResult{realtime.TournamentGameResult_DRAW,
			realtime.TournamentGameResult_DRAW}
	expectedpri2.Pairing.Games[0].Scores[0] = 1
	expectedpri2.Pairing.Games[0].Scores[1] = 1
	expectedpri2.Pairing.Outcomes[0] = realtime.TournamentGameResult_DRAW
	expectedpri2.Pairing.Outcomes[1] = realtime.TournamentGameResult_DRAW
	expectedpri2.Pairing.Games[0].GameEndReason = realtime.GameEndReason_ABANDONED
	is.NoErr(equalPRI(expectedpri2, pri2))

	roundIsComplete, err := tc.IsRoundComplete(0)
	is.NoErr(err)
	is.True(roundIsComplete)

	// Set pairings to test more easily
	err = tc.SetPairing(player1, player2, 1)
	is.NoErr(err)
	err = tc.SetPairing(player3, player4, 1)
	is.NoErr(err)

	pri1, err = tc.GetPlayerRoundInfo(player1, 1)
	is.NoErr(err)
	pri2, err = tc.GetPlayerRoundInfo(player3, 1)
	is.NoErr(err)

	expectedpri1 = newPlayerRoundInfo(tc, player1, player2, tc.GamesPerRound[1], 1)
	expectedpri2 = newPlayerRoundInfo(tc, player3, player4, tc.GamesPerRound[1], 1)

	// Round 2 should have been paired,
	// submit a result

	err = tc.SubmitResult(1, player1, player2, 0, 0,
		realtime.TournamentGameResult_FORFEIT_LOSS,
		realtime.TournamentGameResult_FORFEIT_LOSS,
		realtime.GameEndReason_ABANDONED, false, 0)
	is.NoErr(err)

	expectedpri1.Pairing.Games[0].Scores[0] = 0
	expectedpri1.Pairing.Games[0].Scores[1] = 0
	expectedpri1.Pairing.Games[0].GameEndReason = realtime.GameEndReason_ABANDONED
	expectedpri1.Pairing.Outcomes[0] = realtime.TournamentGameResult_FORFEIT_LOSS
	expectedpri1.Pairing.Outcomes[1] = realtime.TournamentGameResult_FORFEIT_LOSS
	expectedpri1.Pairing.Games[0].Results =
		[]realtime.TournamentGameResult{realtime.TournamentGameResult_FORFEIT_LOSS,
			realtime.TournamentGameResult_FORFEIT_LOSS}
	is.NoErr(equalPRI(expectedpri1, pri1))

	// Submit the final tournament results
	err = tc.SubmitResult(1, player3, player4, 50, 50,
		realtime.TournamentGameResult_BYE,
		realtime.TournamentGameResult_BYE,
		realtime.GameEndReason_ABANDONED, false, 0)
	is.NoErr(err)

	expectedpri2.Pairing.Games[0].Results =
		[]realtime.TournamentGameResult{realtime.TournamentGameResult_BYE, realtime.TournamentGameResult_BYE}
	expectedpri2.Pairing.Games[0].Scores[0] = 50
	expectedpri2.Pairing.Games[0].Scores[1] = 50
	expectedpri2.Pairing.Games[0].GameEndReason = realtime.GameEndReason_ABANDONED
	expectedpri2.Pairing.Outcomes[0] = realtime.TournamentGameResult_BYE
	expectedpri2.Pairing.Outcomes[1] = realtime.TournamentGameResult_BYE
	is.NoErr(equalPRI(expectedpri2, pri2))

	roundIsComplete, err = tc.IsRoundComplete(1)
	is.NoErr(err)
	is.True(roundIsComplete)

	tournamentIsFinished, err := tc.IsFinished()
	is.NoErr(err)
	is.True(tournamentIsFinished)

	// Attempt to get the Standings
	// for an out of range round number
	_, err = tc.GetStandings(8)
	is.True(err != nil)

	// Standings are tested in the
	// King of the Hill Classic Tournament test.

	// Get the standings for round 1
	_, err = tc.GetStandings(0)
	is.NoErr(err)

	// Get the standings for round 2
	_, err = tc.GetStandings(1)
	is.NoErr(err)

	// Check that pairings are correct with an odd number of players
	tc, err = NewClassicDivision(playersOddStrings, rounds, defaultGamesPerRound, []entity.PairingMethod{entity.Random, entity.Random}, defaultFirsts)
	is.NoErr(err)
	is.True(tc != nil)

	is.NoErr(validatePairings(tc, 0))
}

func TestClassicDivisionKingOfTheHill(t *testing.T) {
	// This test is used to ensure that the standings are
	// calculated correctly and that King of the Hill
	// pairings are correct

	is := is.New(t)

	tc, err := NewClassicDivision(playerStrings, rounds, defaultGamesPerRound, []entity.PairingMethod{entity.KingOfTheHill, entity.KingOfTheHill}, defaultFirsts)
	is.NoErr(err)
	is.True(tc != nil)

	is.NoErr(validatePairings(tc, 0))

	// Tournament should not be over

	player1 := playerStrings[0]
	player2 := playerStrings[1]
	player3 := playerStrings[2]
	player4 := playerStrings[3]

	tournamentIsFinished, err := tc.IsFinished()
	is.NoErr(err)
	is.True(!tournamentIsFinished)

	playerPairings := getPlayerPairings(tc.Players, tc.Matrix[0])
	for i := 0; i < len(playerPairings); i++ {
		is.True(playerPairings[i] == playerStrings[i])
	}

	// Submit results for the round
	err = tc.SubmitResult(0, player1, player2, 550, 400,
		realtime.TournamentGameResult_WIN,
		realtime.TournamentGameResult_LOSS,
		realtime.GameEndReason_STANDARD, false, 0)
	is.NoErr(err)

	err = tc.SubmitResult(0, player3, player4, 300, 700,
		realtime.TournamentGameResult_LOSS,
		realtime.TournamentGameResult_WIN,
		realtime.GameEndReason_STANDARD, false, 0)
	is.NoErr(err)

	// Get the standings for round 1
	standings, err := tc.GetStandings(0)
	is.NoErr(err)

	expectedstandings := []*entity.Standing{&entity.Standing{Player: player4, Wins: 1, Losses: 0, Draws: 0, Spread: 400},
		&entity.Standing{Player: player1, Wins: 1, Losses: 0, Draws: 0, Spread: 150},
		&entity.Standing{Player: player2, Wins: 0, Losses: 1, Draws: 0, Spread: -150},
		&entity.Standing{Player: player3, Wins: 0, Losses: 1, Draws: 0, Spread: -400},
	}

	is.NoErr(equalStandings(expectedstandings, standings))

	// The next round should have been paired

	// Tournament should not be over

	tournamentIsFinished, err = tc.IsFinished()
	is.NoErr(err)
	is.True(!tournamentIsFinished)

	// Submit results for the round
	err = tc.SubmitResult(1, player1, player4, 670, 400,
		realtime.TournamentGameResult_WIN,
		realtime.TournamentGameResult_LOSS,
		realtime.GameEndReason_STANDARD, false, 0)
	is.NoErr(err)

	err = tc.SubmitResult(1, player3, player2, 700, 700,
		realtime.TournamentGameResult_DRAW,
		realtime.TournamentGameResult_DRAW,
		realtime.GameEndReason_STANDARD, false, 0)
	is.NoErr(err)

	// Get the standings for round 2
	standings, err = tc.GetStandings(1)
	is.NoErr(err)

	expectedstandings = []*entity.Standing{&entity.Standing{Player: player1, Wins: 2, Losses: 0, Draws: 0, Spread: 420},
		&entity.Standing{Player: player4, Wins: 1, Losses: 1, Draws: 0, Spread: 130},
		&entity.Standing{Player: player2, Wins: 0, Losses: 1, Draws: 1, Spread: -150},
		&entity.Standing{Player: player3, Wins: 0, Losses: 1, Draws: 1, Spread: -400},
	}

	is.NoErr(equalStandings(expectedstandings, standings))

	// Tournament should be over

	tournamentIsFinished, err = tc.IsFinished()
	is.NoErr(err)
	is.True(tournamentIsFinished)

	// Check that pairings are correct with an odd number of players
	tc, err = NewClassicDivision(playersOddStrings, rounds, defaultGamesPerRound, []entity.PairingMethod{entity.KingOfTheHill, entity.KingOfTheHill}, defaultFirsts)
	is.NoErr(err)
	is.True(tc != nil)

	is.NoErr(validatePairings(tc, 0))

	// The last player should have a bye
	l := len(tc.Players) - 1
	lastPlayer := tc.Players[l]
	is.True(tc.Matrix[0][l].Pairing.Players[0] == lastPlayer)
	is.True(tc.Matrix[0][l].Pairing.Players[1] == lastPlayer)
}

func TestClassicDivisionRoundRobinAlgorithm(t *testing.T) {
	// This test is used to ensure that round robin
	// pairings work correctly. See the function in tournament.go
	// for more details about the algorithm.

	is := is.New(t)

	roundRobinPlayers := []string{"1", "2", "3", "4", "5", "6", "7", "8"}

	is.NoErr(equalRoundRobinPairings([]string{"1", "8", "2", "7", "3", "6", "4", "5"},
		getRoundRobinPairings(roundRobinPlayers, 0)))
	is.NoErr(equalRoundRobinPairings([]string{"1", "7", "8", "6", "2", "5", "3", "4"},
		getRoundRobinPairings(roundRobinPlayers, 1)))
	is.NoErr(equalRoundRobinPairings([]string{"1", "6", "7", "5", "8", "4", "2", "3"},
		getRoundRobinPairings(roundRobinPlayers, 2)))
	is.NoErr(equalRoundRobinPairings([]string{"1", "5", "6", "4", "7", "3", "8", "2"},
		getRoundRobinPairings(roundRobinPlayers, 3)))
	is.NoErr(equalRoundRobinPairings([]string{"1", "4", "5", "3", "6", "2", "7", "8"},
		getRoundRobinPairings(roundRobinPlayers, 4)))
	is.NoErr(equalRoundRobinPairings([]string{"1", "3", "4", "2", "5", "8", "6", "7"},
		getRoundRobinPairings(roundRobinPlayers, 5)))
	is.NoErr(equalRoundRobinPairings([]string{"1", "2", "3", "8", "4", "7", "5", "6"},
		getRoundRobinPairings(roundRobinPlayers, 6)))

	// Modulus operation should repeat the pairings past the first round robin

	is.NoErr(equalRoundRobinPairings([]string{"1", "8", "2", "7", "3", "6", "4", "5"},
		getRoundRobinPairings(roundRobinPlayers, 7)))
	is.NoErr(equalRoundRobinPairings([]string{"1", "7", "8", "6", "2", "5", "3", "4"},
		getRoundRobinPairings(roundRobinPlayers, 8)))
	is.NoErr(equalRoundRobinPairings([]string{"1", "6", "7", "5", "8", "4", "2", "3"},
		getRoundRobinPairings(roundRobinPlayers, 9)))
	is.NoErr(equalRoundRobinPairings([]string{"1", "5", "6", "4", "7", "3", "8", "2"},
		getRoundRobinPairings(roundRobinPlayers, 10)))
	is.NoErr(equalRoundRobinPairings([]string{"1", "4", "5", "3", "6", "2", "7", "8"},
		getRoundRobinPairings(roundRobinPlayers, 11)))
	is.NoErr(equalRoundRobinPairings([]string{"1", "3", "4", "2", "5", "8", "6", "7"},
		getRoundRobinPairings(roundRobinPlayers, 12)))
	is.NoErr(equalRoundRobinPairings([]string{"1", "2", "3", "8", "4", "7", "5", "6"},
		getRoundRobinPairings(roundRobinPlayers, 13)))

	// Test first pairing of third round robin just to be sure

	is.NoErr(equalRoundRobinPairings([]string{"1", "8", "2", "7", "3", "6", "4", "5"},
		getRoundRobinPairings(roundRobinPlayers, 14)))
}

func TestClassicDivisionRoundRobin(t *testing.T) {
	// This test is used to ensure that round robin
	// pairings work correctly

	is := is.New(t)

	tc, err := NewClassicDivision(playerStrings, 6, []int{1, 1, 1, 1, 1, 1}, []entity.PairingMethod{entity.RoundRobin,
		entity.RoundRobin,
		entity.RoundRobin,
		entity.RoundRobin,
		entity.RoundRobin,
		entity.RoundRobin},
		[]entity.FirstMethod{entity.ManualFirst,
			entity.ManualFirst,
			entity.ManualFirst,
			entity.ManualFirst,
			entity.ManualFirst,
			entity.ManualFirst})

	is.NoErr(err)
	is.True(tc != nil)

	is.NoErr(validatePairings(tc, 0))
	is.NoErr(validatePairings(tc, 1))
	is.NoErr(validatePairings(tc, 2))
	is.NoErr(validatePairings(tc, 3))
	is.NoErr(validatePairings(tc, 4))
	is.NoErr(validatePairings(tc, 5))

	// In a double round robin with 4 players,
	// everyone should have played everyone else twice.
	for i, player := range playerStrings {
		m := make(map[string]int)
		m[player] = 2

		for k := 0; k < len(tc.Matrix); k++ {
			opponent, err := opponentOf(tc.Matrix[k][i].Pairing, player)
			is.NoErr(err)
			m[opponent]++
		}
		for _, opponent := range playerStrings {
			var err error = nil
			if m[opponent] != 2 {
				err = errors.New(fmt.Sprintf("Player %s didn't play %s exactly twice!", player, opponent))
			}
			is.NoErr(err)
		}
	}

	// Test Round Robin with an odd number of players (a bye)

	tc, err = NewClassicDivision(playersOddStrings, 10, []int{1, 1, 1, 1, 1, 1, 1, 1, 1, 1}, []entity.PairingMethod{entity.RoundRobin,
		entity.RoundRobin,
		entity.RoundRobin,
		entity.RoundRobin,
		entity.RoundRobin,
		entity.RoundRobin,
		entity.RoundRobin,
		entity.RoundRobin,
		entity.RoundRobin,
		entity.RoundRobin}, []entity.FirstMethod{entity.ManualFirst,
		entity.ManualFirst,
		entity.ManualFirst,
		entity.ManualFirst,
		entity.ManualFirst,
		entity.ManualFirst,
		entity.ManualFirst,
		entity.ManualFirst,
		entity.ManualFirst,
		entity.ManualFirst})
	is.NoErr(err)
	is.True(tc != nil)

	is.NoErr(validatePairings(tc, 0))
	is.NoErr(validatePairings(tc, 1))
	is.NoErr(validatePairings(tc, 2))
	is.NoErr(validatePairings(tc, 3))
	is.NoErr(validatePairings(tc, 4))
	is.NoErr(validatePairings(tc, 5))
	is.NoErr(validatePairings(tc, 6))
	is.NoErr(validatePairings(tc, 7))
	is.NoErr(validatePairings(tc, 8))
	is.NoErr(validatePairings(tc, 9))

	// In a double round robin with 5 players,
	// everyone should have played everyone else twice
	// and everyone should have two byes
	for i, player := range playerStrings {
		m := make(map[string]int)
		// We don't assign the player as having played themselves
		// twice in this case because the bye will do that.

		for k := 0; k < len(tc.Matrix); k++ {
			opponent, err := opponentOf(tc.Matrix[k][i].Pairing, player)
			is.NoErr(err)
			m[opponent]++
		}
		for _, opponent := range playerStrings {
			var err error = nil
			if m[opponent] != 2 {
				err = errors.New(fmt.Sprintf("Player %s didn't play %s exactly twice!", player, opponent))
			}
			is.NoErr(err)
		}
	}
}

func TestClassicDivisionManual(t *testing.T) {
	is := is.New(t)

	tc, err := NewClassicDivision(playerStrings, rounds, defaultGamesPerRound, []entity.PairingMethod{entity.Manual, entity.Manual}, defaultFirsts)
	is.NoErr(err)
	is.True(tc != nil)

	player1 := playerStrings[0]
	player2 := playerStrings[1]
	player3 := playerStrings[2]
	player4 := playerStrings[3]

	// Check that round 1 is not paired
	for _, pri := range tc.Matrix[0] {
		pairing := pri.Pairing
		is.True(pairing == nil)
	}

	// Pair round 1
	err = tc.SetPairing(player1, player2, 0)
	is.NoErr(err)
	err = tc.SetPairing(player3, player4, 0)
	is.NoErr(err)

	is.NoErr(validatePairings(tc, 0))

	// Amend a pairing
	err = tc.SetPairing(player2, player3, 0)
	is.NoErr(err)

	// Confirm that players 1 and 4 are now unpaired
	is.True(tc.Matrix[0][tc.PlayerIndexMap[player1]].Pairing == nil)
	is.True(tc.Matrix[0][tc.PlayerIndexMap[player4]].Pairing == nil)

	// Complete the round 1 pairings
	err = tc.SetPairing(player1, player4, 0)
	is.NoErr(err)

	is.NoErr(validatePairings(tc, 0))

	// Submit results for round 1

	// FIXME
	// Here, amend is set to true, but it should not matter.
	// Maybe at some point we want to be stricter and
	// reject submissions that think they are amending
	// when really there is no result.
	err = tc.SubmitResult(0, player2, player3, 400, 500,
		realtime.TournamentGameResult_LOSS,
		realtime.TournamentGameResult_WIN,
		realtime.GameEndReason_STANDARD, true, 0)
	is.NoErr(err)
	err = tc.SubmitResult(0, player1, player4, 200, 450,
		realtime.TournamentGameResult_LOSS,
		realtime.TournamentGameResult_WIN,
		realtime.GameEndReason_STANDARD, true, 0)
	is.NoErr(err)

	roundIsComplete, err := tc.IsRoundComplete(0)
	is.NoErr(err)
	is.True(roundIsComplete)

	// Get the standings for round 1
	standings, err := tc.GetStandings(0)
	is.NoErr(err)

	expectedstandings := []*entity.Standing{&entity.Standing{Player: player4, Wins: 1, Losses: 0, Draws: 0, Spread: 250},
		&entity.Standing{Player: player3, Wins: 1, Losses: 0, Draws: 0, Spread: 100},
		&entity.Standing{Player: player2, Wins: 0, Losses: 1, Draws: 0, Spread: -100},
		&entity.Standing{Player: player1, Wins: 0, Losses: 1, Draws: 0, Spread: -250},
	}

	is.NoErr(equalStandings(expectedstandings, standings))

	// Amend a result
	err = tc.SubmitResult(0, player1, player4, 500, 450,
		realtime.TournamentGameResult_WIN,
		realtime.TournamentGameResult_LOSS,
		realtime.GameEndReason_STANDARD, true, 0)
	is.NoErr(err)

	roundIsComplete, err = tc.IsRoundComplete(0)
	is.NoErr(err)
	is.True(roundIsComplete)

	// Get the standings for round 1 again
	standings, err = tc.GetStandings(0)
	is.NoErr(err)

	expectedstandings = []*entity.Standing{&entity.Standing{Player: player3, Wins: 1, Losses: 0, Draws: 0, Spread: 100},
		&entity.Standing{Player: player1, Wins: 1, Losses: 0, Draws: 0, Spread: 50},
		&entity.Standing{Player: player4, Wins: 0, Losses: 1, Draws: 0, Spread: -50},
		&entity.Standing{Player: player2, Wins: 0, Losses: 1, Draws: 0, Spread: -100},
	}

	is.NoErr(equalStandings(expectedstandings, standings))
}

func TestClassicDivisionElimination(t *testing.T) {
	is := is.New(t)

	// Try and make an elimination tournament with too many rounds
	tc, err := NewClassicDivision(playerStrings, 3, []int{3, 3, 3}, []entity.PairingMethod{entity.Elimination,
		entity.Elimination,
		entity.Elimination}, defaultFirsts)
	is.True(err != nil)

	// Try and make an elimination tournament with other types
	// of pairings
	tc, err = NewClassicDivision(playerStrings, 3, []int{3, 3, 3}, []entity.PairingMethod{entity.Elimination,
		entity.Random,
		entity.Elimination}, defaultFirsts)
	is.True(err != nil)

	tc, err = NewClassicDivision(playerStrings, 2, []int{3, 3}, []entity.PairingMethod{entity.Elimination,
		entity.Elimination}, defaultFirsts)
	is.NoErr(err)
	is.True(tc != nil)

	is.NoErr(validatePairings(tc, 0))

	player1 := playerStrings[0]
	player2 := playerStrings[1]
	player3 := playerStrings[2]
	player4 := playerStrings[3]

	pri1, err := tc.GetPlayerRoundInfo(player1, 0)
	is.NoErr(err)
	pri2, err := tc.GetPlayerRoundInfo(player3, 0)
	is.NoErr(err)

	expectedpri1 := newPlayerRoundInfo(tc, player1, player2, tc.GamesPerRound[0], 0)
	expectedpri2 := newPlayerRoundInfo(tc, player3, player4, tc.GamesPerRound[0], 0)

	// Get the initial standings
	standings, err := tc.GetStandings(0)
	is.NoErr(err)

	// Ensure standings for Elimination are correct
	expectedstandings := []*entity.Standing{&entity.Standing{Player: player1, Wins: 0, Losses: 0, Draws: 0, Spread: 0},
		&entity.Standing{Player: player2, Wins: 0, Losses: 0, Draws: 0, Spread: 0},
		&entity.Standing{Player: player3, Wins: 0, Losses: 0, Draws: 0, Spread: 0},
		&entity.Standing{Player: player4, Wins: 0, Losses: 0, Draws: 0, Spread: 0},
	}

	is.NoErr(equalStandings(expectedstandings, standings))

	// The match is decided in two games
	err = tc.SubmitResult(0, player1, player2, 500, 490,
		realtime.TournamentGameResult_WIN,
		realtime.TournamentGameResult_LOSS,
		realtime.GameEndReason_STANDARD, false, 0)
	is.NoErr(err)

	// The games should have changed
	expectedpri1.Pairing.Games[0].Results =
		[]realtime.TournamentGameResult{realtime.TournamentGameResult_WIN,
			realtime.TournamentGameResult_LOSS}
	expectedpri1.Pairing.Games[0].Scores[0] = 500
	expectedpri1.Pairing.Games[0].Scores[1] = 490
	expectedpri1.Pairing.Games[0].GameEndReason = realtime.GameEndReason_STANDARD
	is.NoErr(equalPRI(expectedpri1, pri1))

	err = tc.SubmitResult(0, player1, player2, 50, 0,
		realtime.TournamentGameResult_FORFEIT_WIN,
		realtime.TournamentGameResult_FORFEIT_LOSS,
		realtime.GameEndReason_ABANDONED, false, 1)
	is.NoErr(err)

	// The outcomes should now be set
	expectedpri1.Pairing.Games[1].Results =
		[]realtime.TournamentGameResult{realtime.TournamentGameResult_FORFEIT_WIN, realtime.TournamentGameResult_FORFEIT_LOSS}
	expectedpri1.Pairing.Games[1].Scores[0] = 50
	expectedpri1.Pairing.Games[1].Scores[1] = 0
	expectedpri1.Pairing.Games[1].GameEndReason = realtime.GameEndReason_ABANDONED
	expectedpri1.Pairing.Outcomes[0] = realtime.TournamentGameResult_WIN
	expectedpri1.Pairing.Outcomes[1] = realtime.TournamentGameResult_ELIMINATED
	is.NoErr(equalPRI(expectedpri1, pri1))

	roundIsComplete, err := tc.IsRoundComplete(0)
	is.NoErr(err)
	is.True(!roundIsComplete)

	// The match is decided in three games
	err = tc.SubmitResult(0, player3, player4, 500, 400,
		realtime.TournamentGameResult_WIN,
		realtime.TournamentGameResult_LOSS,
		realtime.GameEndReason_STANDARD, false, 0)
	is.NoErr(err)

	// The spread and games should have changed
	expectedpri2.Pairing.Games[0].Results =
		[]realtime.TournamentGameResult{realtime.TournamentGameResult_WIN, realtime.TournamentGameResult_LOSS}
	expectedpri2.Pairing.Games[0].Scores[0] = 500
	expectedpri2.Pairing.Games[0].Scores[1] = 400
	expectedpri2.Pairing.Games[0].GameEndReason = realtime.GameEndReason_STANDARD
	is.NoErr(equalPRI(expectedpri2, pri2))

	err = tc.SubmitResult(0, player3, player4, 400, 400,
		realtime.TournamentGameResult_DRAW,
		realtime.TournamentGameResult_DRAW,
		realtime.GameEndReason_STANDARD, false, 1)
	is.NoErr(err)

	// The spread and games should have changed
	expectedpri2.Pairing.Games[1].Results =
		[]realtime.TournamentGameResult{realtime.TournamentGameResult_DRAW,
			realtime.TournamentGameResult_DRAW}
	expectedpri2.Pairing.Games[1].Scores[0] = 400
	expectedpri2.Pairing.Games[1].Scores[1] = 400
	expectedpri2.Pairing.Games[1].GameEndReason = realtime.GameEndReason_STANDARD
	is.NoErr(equalPRI(expectedpri2, pri2))

	err = tc.SubmitResult(0, player3, player4, 450, 400,
		realtime.TournamentGameResult_WIN,
		realtime.TournamentGameResult_LOSS,
		realtime.GameEndReason_STANDARD, false, 2)
	is.NoErr(err)

	// The spread and games should have changed
	// The outcome and record should have changed
	expectedpri2.Pairing.Games[2].Results =
		[]realtime.TournamentGameResult{realtime.TournamentGameResult_WIN, realtime.TournamentGameResult_LOSS}
	expectedpri2.Pairing.Games[2].Scores[0] = 450
	expectedpri2.Pairing.Games[2].Scores[1] = 400
	expectedpri2.Pairing.Games[2].GameEndReason = realtime.GameEndReason_STANDARD
	expectedpri2.Pairing.Outcomes[0] = realtime.TournamentGameResult_WIN
	expectedpri2.Pairing.Outcomes[1] = realtime.TournamentGameResult_ELIMINATED
	is.NoErr(equalPRI(expectedpri2, pri2))

	roundIsComplete, err = tc.IsRoundComplete(0)
	is.NoErr(err)
	is.True(roundIsComplete)

	// Get the standings for round 1
	standings, err = tc.GetStandings(0)
	is.NoErr(err)

	// Elimination standings are based on wins and player order only
	// Losses are not recorded in Elimination standings
	expectedstandings = []*entity.Standing{&entity.Standing{Player: player1, Wins: 1, Losses: 0, Draws: 0, Spread: 60},
		&entity.Standing{Player: player3, Wins: 1, Losses: 0, Draws: 0, Spread: 150},
		&entity.Standing{Player: player2, Wins: 0, Losses: 1, Draws: 0, Spread: -60},
		&entity.Standing{Player: player4, Wins: 0, Losses: 1, Draws: 0, Spread: -150},
	}

	is.NoErr(equalStandings(expectedstandings, standings))

	pri1, err = tc.GetPlayerRoundInfo(player1, 1)
	is.NoErr(err)
	pri2, err = tc.GetPlayerRoundInfo(player4, 1)
	is.NoErr(err)

	expectedpri1 = newPlayerRoundInfo(tc, player1, player3, tc.GamesPerRound[1], 1)

	// Half of the field should be eliminated

	// There should be no changes to the PRIs of players still
	// in the tournament. The Record gets carried over from
	// last round in the usual manner.
	is.NoErr(equalPRI(expectedpri1, pri1))

	// The usual pri comparison method will fail since the
	// Games and Players are nil for elimianted players
	is.True(pri2.Pairing.Outcomes[0] == realtime.TournamentGameResult_ELIMINATED)
	is.True(pri2.Pairing.Outcomes[1] == realtime.TournamentGameResult_ELIMINATED)
	is.True(pri2.Pairing.Games == nil)
	is.True(pri2.Pairing.Players == nil)

	// The match is decided in three games
	err = tc.SubmitResult(1, player1, player3, 500, 400,
		realtime.TournamentGameResult_WIN,
		realtime.TournamentGameResult_LOSS,
		realtime.GameEndReason_STANDARD, false, 0)
	is.NoErr(err)

	expectedpri1.Pairing.Games[0].Results =
		[]realtime.TournamentGameResult{realtime.TournamentGameResult_WIN, realtime.TournamentGameResult_LOSS}
	expectedpri1.Pairing.Games[0].Scores[0] = 500
	expectedpri1.Pairing.Games[0].Scores[1] = 400
	expectedpri1.Pairing.Games[0].GameEndReason = realtime.GameEndReason_STANDARD
	is.NoErr(equalPRI(expectedpri1, pri1))

	err = tc.SubmitResult(1, player1, player3, 400, 600,
		realtime.TournamentGameResult_LOSS,
		realtime.TournamentGameResult_WIN,
		realtime.GameEndReason_STANDARD, false, 1)
	is.NoErr(err)

	expectedpri1.Pairing.Games[1].Results =
		[]realtime.TournamentGameResult{realtime.TournamentGameResult_LOSS,
			realtime.TournamentGameResult_WIN}
	expectedpri1.Pairing.Games[1].Scores[0] = 400
	expectedpri1.Pairing.Games[1].Scores[1] = 600
	expectedpri1.Pairing.Games[1].GameEndReason = realtime.GameEndReason_STANDARD
	is.NoErr(equalPRI(expectedpri1, pri1))

	err = tc.SubmitResult(1, player1, player3, 450, 450,
		realtime.TournamentGameResult_DRAW,
		realtime.TournamentGameResult_DRAW,
		realtime.GameEndReason_STANDARD, false, 2)
	is.NoErr(err)

	expectedpri1.Pairing.Games[2].Results =
		[]realtime.TournamentGameResult{realtime.TournamentGameResult_DRAW, realtime.TournamentGameResult_DRAW}
	expectedpri1.Pairing.Games[2].Scores[0] = 450
	expectedpri1.Pairing.Games[2].Scores[1] = 450
	expectedpri1.Pairing.Games[2].GameEndReason = realtime.GameEndReason_STANDARD
	expectedpri1.Pairing.Outcomes[0] = realtime.TournamentGameResult_ELIMINATED
	expectedpri1.Pairing.Outcomes[1] = realtime.TournamentGameResult_WIN
	is.NoErr(equalPRI(expectedpri1, pri1))

	roundIsComplete, err = tc.IsRoundComplete(1)
	is.NoErr(err)
	is.True(roundIsComplete)

	// Amend a result
	err = tc.SubmitResult(1, player1, player3, 451, 450,
		realtime.TournamentGameResult_WIN,
		realtime.TournamentGameResult_LOSS,
		realtime.GameEndReason_STANDARD, true, 2)
	is.NoErr(err)

	expectedpri1.Pairing.Games[2].Results =
		[]realtime.TournamentGameResult{realtime.TournamentGameResult_WIN,
			realtime.TournamentGameResult_LOSS}
	expectedpri1.Pairing.Games[2].Scores[0] = 451
	expectedpri1.Pairing.Games[2].Scores[1] = 450
	expectedpri1.Pairing.Games[2].GameEndReason = realtime.GameEndReason_STANDARD
	expectedpri1.Pairing.Outcomes[0] = realtime.TournamentGameResult_WIN
	expectedpri1.Pairing.Outcomes[1] = realtime.TournamentGameResult_ELIMINATED
	is.NoErr(equalPRI(expectedpri1, pri1))

	roundIsComplete, err = tc.IsRoundComplete(1)
	is.NoErr(err)
	is.True(roundIsComplete)

	tournamentIsFinished, err := tc.IsFinished()
	is.NoErr(err)
	is.True(tournamentIsFinished)

	// Test ties and submitting tiebreaking results
	// Since this test is copied from above, the usual
	// validations are skipped, since they would be redundant.

	tc, err = NewClassicDivision(playerStrings, 2, []int{3, 3}, []entity.PairingMethod{entity.Elimination,
		entity.Elimination}, defaultFirsts)
	is.NoErr(err)
	is.True(tc != nil)

	player1 = playerStrings[0]
	player2 = playerStrings[1]
	player3 = playerStrings[2]
	player4 = playerStrings[3]

	pri2, err = tc.GetPlayerRoundInfo(player3, 0)
	is.NoErr(err)

	expectedpri1 = newPlayerRoundInfo(tc, player1, player2, tc.GamesPerRound[0], 0)
	expectedpri2 = newPlayerRoundInfo(tc, player3, player4, tc.GamesPerRound[0], 0)

	// The match is decided in two games
	err = tc.SubmitResult(0, player1, player2, 500, 490,
		realtime.TournamentGameResult_WIN,
		realtime.TournamentGameResult_LOSS,
		realtime.GameEndReason_STANDARD, false, 0)
	is.NoErr(err)

	err = tc.SubmitResult(0, player1, player2, 50, 0,
		realtime.TournamentGameResult_FORFEIT_WIN,
		realtime.TournamentGameResult_FORFEIT_LOSS,
		realtime.GameEndReason_ABANDONED, false, 1)
	is.NoErr(err)

	// The next match ends up tied at 1.5 - 1.5
	// with both players having the same spread.
	err = tc.SubmitResult(0, player3, player4, 500, 400,
		realtime.TournamentGameResult_WIN,
		realtime.TournamentGameResult_LOSS,
		realtime.GameEndReason_STANDARD, false, 0)
	is.NoErr(err)

	err = tc.SubmitResult(0, player3, player4, 400, 500,
		realtime.TournamentGameResult_LOSS,
		realtime.TournamentGameResult_WIN,
		realtime.GameEndReason_STANDARD, false, 1)
	is.NoErr(err)

	err = tc.SubmitResult(0, player3, player4, 500, 500,
		realtime.TournamentGameResult_DRAW,
		realtime.TournamentGameResult_DRAW,
		realtime.GameEndReason_STANDARD, false, 2)
	is.NoErr(err)

	expectedpri2.Pairing.Games[0].Results =
		[]realtime.TournamentGameResult{realtime.TournamentGameResult_WIN,
			realtime.TournamentGameResult_LOSS}
	expectedpri2.Pairing.Games[1].Results =
		[]realtime.TournamentGameResult{realtime.TournamentGameResult_LOSS,
			realtime.TournamentGameResult_WIN}
	expectedpri2.Pairing.Games[2].Results =
		[]realtime.TournamentGameResult{realtime.TournamentGameResult_DRAW,
			realtime.TournamentGameResult_DRAW}
	expectedpri2.Pairing.Games[0].Scores[0] = 500
	expectedpri2.Pairing.Games[1].Scores[0] = 400
	expectedpri2.Pairing.Games[2].Scores[0] = 500
	expectedpri2.Pairing.Games[0].Scores[1] = 400
	expectedpri2.Pairing.Games[1].Scores[1] = 500
	expectedpri2.Pairing.Games[2].Scores[1] = 500
	expectedpri2.Pairing.Games[0].GameEndReason = realtime.GameEndReason_STANDARD
	expectedpri2.Pairing.Games[1].GameEndReason = realtime.GameEndReason_STANDARD
	expectedpri2.Pairing.Games[2].GameEndReason = realtime.GameEndReason_STANDARD
	expectedpri2.Pairing.Outcomes[0] = realtime.TournamentGameResult_NO_RESULT
	expectedpri2.Pairing.Outcomes[1] = realtime.TournamentGameResult_NO_RESULT
	is.NoErr(equalPRI(expectedpri2, pri2))

	// Round should not be over
	roundIsComplete, err = tc.IsRoundComplete(0)
	is.NoErr(err)
	is.True(!roundIsComplete)

	// Submit a tiebreaking result, unfortunately, it's another draw
	err = tc.SubmitResult(0, player3, player4, 500, 500,
		realtime.TournamentGameResult_DRAW,
		realtime.TournamentGameResult_DRAW,
		realtime.GameEndReason_STANDARD, false, 3)
	is.NoErr(err)

	expectedpri2.Pairing.Games =
		append(expectedpri2.Pairing.Games,
			&entity.TournamentGame{Scores: []int{500, 500},
				Results: []realtime.TournamentGameResult{realtime.TournamentGameResult_DRAW,
					realtime.TournamentGameResult_DRAW}})
	expectedpri2.Pairing.Games[3].GameEndReason = realtime.GameEndReason_STANDARD
	is.NoErr(equalPRI(expectedpri2, pri2))

	// Round should still not be over
	roundIsComplete, err = tc.IsRoundComplete(0)
	is.NoErr(err)
	is.True(!roundIsComplete)

	// Attempt to submit a tiebreaking result, unfortunately, the game index is wrong
	err = tc.SubmitResult(0, player3, player4, 500, 500,
		realtime.TournamentGameResult_DRAW,
		realtime.TournamentGameResult_DRAW,
		realtime.GameEndReason_STANDARD, false, 5)
	is.True(err != nil)

	// Still wrong! Silly director (and definitely not code another layer up)
	err = tc.SubmitResult(0, player3, player4, 500, 500,
		realtime.TournamentGameResult_DRAW,
		realtime.TournamentGameResult_DRAW,
		realtime.GameEndReason_STANDARD, false, 2)
	is.True(err != nil)

	// Round should still not be over
	roundIsComplete, err = tc.IsRoundComplete(0)
	is.NoErr(err)
	is.True(!roundIsComplete)

	// The players finally reach a decisive result
	err = tc.SubmitResult(0, player3, player4, 600, 300,
		realtime.TournamentGameResult_WIN,
		realtime.TournamentGameResult_LOSS,
		realtime.GameEndReason_STANDARD, false, 4)
	is.NoErr(err)

	// Round is finally over
	roundIsComplete, err = tc.IsRoundComplete(0)
	is.NoErr(err)
	is.True(roundIsComplete)

	// Get the standings for round 1
	standings, err = tc.GetStandings(0)
	is.NoErr(err)

	expectedstandings = []*entity.Standing{&entity.Standing{Player: player1, Wins: 1, Losses: 0, Draws: 0, Spread: 60},
		&entity.Standing{Player: player3, Wins: 1, Losses: 0, Draws: 0, Spread: 300},
		&entity.Standing{Player: player2, Wins: 0, Losses: 1, Draws: 0, Spread: -60},
		&entity.Standing{Player: player4, Wins: 0, Losses: 1, Draws: 0, Spread: -300},
	}

	is.NoErr(equalStandings(expectedstandings, standings))
}

func TestClassicDivisionFirsts(t *testing.T) {
	// Test
	//   Manual sets the correct firsts
	//   Random at least works
	//   Automatic pairs correctly when
	//     firsts aren't tied
	//     firsts are tied and seconds aren't tied
	//     firsts and seconds are tied
	//   Byes, forfeits do not change the first/second values
	is := is.New(t)

	firstRounds := 10

	pairingMethods := []entity.PairingMethod{}
	for i := 0; i < firstRounds; i++ {
		pairingMethods = append(pairingMethods, entity.Manual)
	}

	tc, err := NewClassicDivision(playerStrings, firstRounds, []int{1, 1, 1, 1, 1, 1, 1, 1, 1, 1}, pairingMethods, []entity.FirstMethod{entity.ManualFirst,
		entity.ManualFirst,
		entity.AutomaticFirst,
		entity.AutomaticFirst,
		entity.ManualFirst,
		entity.ManualFirst,
		entity.AutomaticFirst,
		entity.AutomaticFirst,
		entity.AutomaticFirst,
		entity.RandomFirst})
	is.NoErr(err)
	is.True(tc != nil)

	player1 := playerStrings[0]
	player2 := playerStrings[1]
	player3 := playerStrings[2]
	player4 := playerStrings[3]

	// Pair round 0

	playerOrder := []string{player1, player2, player3, player4}
	fs := []int{1, 0, 0, 1, 1, 0, 0, 1}
	round := 0
	is.NoErr(runFirstMethodRound(tc, playerOrder, fs, round, false))

	// Pair round 1

	playerOrder = []string{player1, player2, player3, player4}
	fs = []int{2, 0, 0, 2, 2, 0, 0, 2}
	round = 1
	is.NoErr(runFirstMethodRound(tc, playerOrder, fs, round, false))

	// Next two rounds should even out firsts and seconds
	// since they are automatic
	// Pair round 2

	playerOrder = []string{player2, player1, player4, player3}
	fs = []int{1, 2, 2, 1, 1, 2, 2, 1}
	round = 2
	is.NoErr(runFirstMethodRound(tc, playerOrder, fs, round, false))

	// Pair round 3

	playerOrder = []string{player2, player1, player4, player3}
	fs = []int{2, 2, 2, 2, 2, 2, 2, 2}
	round = 3
	is.NoErr(runFirstMethodRound(tc, playerOrder, fs, round, false))

	// Make a bye so we can setup the same number of firsts
	// but a different number of seconds
	// Pair round4

	playerOrder = []string{player1, player2, player3, player4}
	fs = []int{3, 2, 2, 3, 2, 2, 2, 2}
	round = 4
	is.NoErr(runFirstMethodRound(tc, playerOrder, fs, round, true))

	// Pair round 5

	playerOrder = []string{player1, player3, player2, player4}
	fs = []int{4, 2, 2, 3, 2, 3, 2, 2}
	round = 5
	is.NoErr(runFirstMethodRound(tc, playerOrder, fs, round, true))

	// Round 6 tests that if firsts are tied, seconds are a tiebreaker
	// Pair round 6

	playerOrder = []string{player2, player1, player4, player3}
	fs = []int{3, 3, 4, 3, 2, 3, 3, 3}
	round = 6
	is.NoErr(runFirstMethodRound(tc, playerOrder, fs, round, false))

	// Pair round 7

	playerOrder = []string{player2, player4, player3, player1}
	fs = []int{3, 4, 3, 3, 3, 3, 4, 3}
	round = 7
	is.NoErr(runFirstMethodRound(tc, playerOrder, fs, round, true))
}

func runFirstMethodRound(tc *ClassicDivision, playerOrder []string, fs []int, round int, useByes bool) error {
	err := tc.SetPairing(playerOrder[0], playerOrder[1], round)

	if err != nil {
		return err
	}

	if useByes {
		err = tc.SetPairing(playerOrder[2], playerOrder[2], round)

		if err != nil {
			return err
		}
		err = tc.SetPairing(playerOrder[3], playerOrder[3], round)

		if err != nil {
			return err
		}
	} else {
		err = tc.SetPairing(playerOrder[2], playerOrder[3], round)

		if err != nil {
			return err
		}
	}

	err = completeManualRound(tc, round, playerOrder[0], playerOrder[1], playerOrder[2], playerOrder[3], useByes)

	if err != nil {
		return err
	}

	return checkFirsts(tc, playerOrder, fs, round)
}

func completeManualRound(tc *ClassicDivision, round int, player1 string, player2 string, player3 string, player4 string, useByes bool) error {

	err := tc.SubmitResult(round, player1, player2, 400, 500,
		realtime.TournamentGameResult_LOSS,
		realtime.TournamentGameResult_WIN,
		realtime.GameEndReason_STANDARD, true, 0)

	if err != nil {
		return err
	}

	if useByes {
		err = tc.SubmitResult(round, player3, player3, 0, 0,
			realtime.TournamentGameResult_BYE,
			realtime.TournamentGameResult_BYE,
			realtime.GameEndReason_STANDARD, false, 0)

		if err != nil {
			return err
		}

		err = tc.SubmitResult(round, player4, player4, 0, 0,
			realtime.TournamentGameResult_BYE,
			realtime.TournamentGameResult_BYE,
			realtime.GameEndReason_STANDARD, false, 0)

		if err != nil {
			return err
		}
	} else {
		err = tc.SubmitResult(round, player3, player4, 200, 450,
			realtime.TournamentGameResult_LOSS,
			realtime.TournamentGameResult_WIN,
			realtime.GameEndReason_STANDARD, true, 0)

		if err != nil {
			return err
		}
	}

	roundIsComplete, err := tc.IsRoundComplete(round)
	if err != nil {
		return err
	}

	if !roundIsComplete {
		return errors.New(fmt.Sprintf("Round %d is not complete.", round))
	}

	return err
}

func checkFirsts(tc *ClassicDivision, players []string, fs []int, round int) error {

	actualfs := []int{}

	for i := 0; i < len(players); i++ {
		playerfs := getPlayerFS(tc, players[i], round)
		actualfs = append(actualfs, playerfs...)
	}

	for i := 0; i < len(actualfs); i++ {
		if actualfs[i] != fs[i] {
			actualfsString := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(actualfs)), ", "), "[]")
			fsString := strings.Trim(strings.Join(strings.Fields(fmt.Sprint(fs)), ", "), "[]")
			return errors.New(fmt.Sprintf("Firsts and Seconds are not equal in round %d:\n%s\n%s\n", round, fsString, actualfsString))
		}
	}
	return nil
}

func TestClassicDivisionRandomData(t *testing.T) {
	is := is.New(t)

	is.NoErr(runRandomTournaments(entity.Random, false))
	is.NoErr(runRandomTournaments(entity.RoundRobin, false))
	is.NoErr(runRandomTournaments(entity.KingOfTheHill, false))
	is.NoErr(runRandomTournaments(entity.Elimination, false))
	// Randomize the pairing method for each round
	// Given pairing method is irrelevant
	is.NoErr(runRandomTournaments(entity.Manual, true))
}

func runRandomTournaments(method entity.PairingMethod, randomizePairings bool) error {
	for numberOfPlayers := 2; numberOfPlayers <= 512; numberOfPlayers++ {
		var numberOfRounds int
		if method == entity.Elimination {
			// numberOfRounds will be -1 if number of players
			// is not of the form 2 ^ n
			numberOfRounds = logTwo(numberOfPlayers)
			if numberOfRounds < 0 {
				continue
			}
		} else {
			numberOfRounds = rand.Intn(10) + 10
		}

		playersRandom := []string{}
		for i := 0; i < numberOfPlayers; i++ {
			playersRandom = append(playersRandom, fmt.Sprintf("%d", i))
		}

		firsts := []entity.FirstMethod{}
		if method == entity.Elimination {
			// Switching firsts and seconds can lead to incomplete
			// elimination rounds.
			for i := 0; i < numberOfRounds; i++ {
				firsts = append(firsts, entity.ManualFirst)
			}
		} else {
			for i := 0; i < numberOfRounds; i++ {
				firsts = append(firsts, entity.FirstMethod(rand.Intn(3)))
			}
		}

		methods := []entity.PairingMethod{}
		if randomizePairings {
			for i := 0; i < numberOfRounds; i++ {
				methods = append(methods, entity.PairingMethod(rand.Intn(3)))
			}
		} else {
			for i := 0; i < numberOfRounds; i++ {
				methods = append(methods, method)
			}
		}

		gpr := []int{}
		if method == entity.Elimination {
			// Even-numbered games per rounds was leading to ties
			// which give inconclusive results, therefor not ending the round
			for i := 0; i < numberOfRounds; i++ {
				gpr = append(gpr, (rand.Intn(5)*2)+1)
			}
		} else {
			for i := 0; i < numberOfRounds; i++ {
				gpr = append(gpr, 1)
			}
		}

		tc, err := NewClassicDivision(playersRandom, numberOfRounds, gpr, methods, firsts)
		if err != nil {
			return err
		}

		for round := 0; round < numberOfRounds; round++ {

			err = validatePairings(tc, round)

			if err != nil {
				return err
			}

			pairings := getPlayerPairings(tc.Players, tc.Matrix[round])
			for game := 0; game < tc.GamesPerRound[round]; game++ {
				for l := 0; l < len(pairings); l += 2 {

					// The outcome might already be decided in an elimination tournament, skip the submission
					if method == entity.Elimination &&
						tc.Matrix[round][tc.PlayerIndexMap[pairings[l]]].Pairing.Outcomes[0] != realtime.TournamentGameResult_NO_RESULT &&
						tc.Matrix[round][tc.PlayerIndexMap[pairings[l]]].Pairing.Outcomes[1] != realtime.TournamentGameResult_NO_RESULT {
						continue
					}

					// For Elimination tournaments, force a decisive result
					// otherwise the round may not be over when we check for it
					var res1 realtime.TournamentGameResult
					var res2 realtime.TournamentGameResult
					if method == entity.Elimination {
						if rand.Intn(2) == 0 {
							res1 = realtime.TournamentGameResult_WIN
							res2 = realtime.TournamentGameResult_LOSS
						} else {
							res1 = realtime.TournamentGameResult_LOSS
							res2 = realtime.TournamentGameResult_WIN
						}
					} else {
						res1 = realtime.TournamentGameResult(rand.Intn(6) + 1)
						res2 = realtime.TournamentGameResult(rand.Intn(6) + 1)
					}

					err = tc.SubmitResult(round,
						pairings[l],
						pairings[l+1],
						rand.Intn(300)+300,
						rand.Intn(300)+300,
						res1,
						res2,
						realtime.GameEndReason_STANDARD,
						false,
						game)
					if err != nil {
						fmt.Printf("(%d) Error on round %d game %d pairing (%s, %s)\n",
							numberOfPlayers,
							round,
							game,
							pairings[l],
							pairings[l+1])
						return err
					}
				} // Pairings
			} // Games

			// Skip testing amendments for elimination events here.
			// Because of the random data a match may have gone from
			// decided to undecided based on the amendment. This will
			// cause a failure when the round is checked for completion.
			if method != entity.Elimination {
				numberOfAmendments := rand.Intn(5)
				for l := 0; l < numberOfAmendments; l++ {
					randPairing := rand.Intn(len(pairings)/2) * 2
					err = tc.SubmitResult(round,
						pairings[randPairing],
						pairings[randPairing+1],
						rand.Intn(300)+300,
						rand.Intn(300)+300,
						realtime.TournamentGameResult(rand.Intn(6)+1),
						realtime.TournamentGameResult(rand.Intn(6)+1),
						realtime.GameEndReason_STANDARD,
						true,
						rand.Intn(tc.GamesPerRound[round]))
					if err != nil {
						return err
					}
				} // Amendments
			}

			roundIsComplete, err := tc.IsRoundComplete(round)
			if err != nil {
				return err
			}
			if !roundIsComplete {
				return errors.New(fmt.Sprintf("(%d) Round %d is not complete (%d, %d)\n",
					numberOfPlayers, round, method, numberOfPlayers))
			}

			_, err = tc.GetStandings(round)
			if err != nil {
				return err
			}

		} // Tournament
		tournamentIsFinished, err := tc.IsFinished()
		if err != nil {
			return err
		}
		if !tournamentIsFinished {
			return errors.New(fmt.Sprintf("Tournament is not complete (%d, %d)\n",
				method, numberOfPlayers))
		}
		if tc.PairingMethods[0] == entity.Elimination {
			standings, err := tc.GetStandings(numberOfRounds - 1)
			if err != nil {
				return err
			}
			bottomHalfSize := numberOfPlayers / 2
			eliminationPlayerIndex := numberOfPlayers - 1
			eliminatedInRound := 0
			for bottomHalfSize > 0 {
				for i := 0; i < bottomHalfSize; i++ {
					if standings[eliminationPlayerIndex].Wins != eliminatedInRound {
						return errors.New(fmt.Sprintf("Player has incorrect number of wins (%d, %d, %d)\n",
							eliminationPlayerIndex,
							eliminatedInRound,
							standings[eliminationPlayerIndex].Wins))
					}
					eliminationPlayerIndex--
				}
				eliminatedInRound++
				bottomHalfSize = bottomHalfSize / 2
			}
		}
	} // Number of players
	return nil
}

func validatePairings(tc *ClassicDivision, round int) error {
	// For each pairing, check that
	//   - Player's opponent is nonnull
	//   - Player's opponent's opponent is the player

	if round < 0 || round >= len(tc.Matrix) {
		return errors.New(fmt.Sprintf("Round number out of range: %d\n", round))
	}

	for i, pri := range tc.Matrix[round] {
		pairing := pri.Pairing
		if pairing.Players == nil {
			// Some pairings can be nil for Elimination tournaments
			if tc.PairingMethods[0] != entity.Elimination {
				return errors.New(fmt.Sprintf("Player %d is unpaired", i))
			} else {
				continue
			}
		}
		// Check that the pairing refs are correct
		opponent, err := opponentOf(pairing, tc.Players[i])
		if err != nil {
			return err
		}
		opponentOpponent, err := opponentOf(tc.Matrix[round][tc.PlayerIndexMap[opponent]].Pairing, opponent)
		if err != nil {
			return err
		}
		if tc.Players[i] != opponentOpponent {
			return errors.New(
				fmt.Sprintf("Player %s's opponent's (%s) opponent (%s) is not themself.",
					tc.Players[i],
					opponent,
					opponentOpponent))
		}
	}
	return nil
}

func equalStandings(sa1 []*entity.Standing, sa2 []*entity.Standing) error {

	if len(sa1) != len(sa2) {
		return errors.New(fmt.Sprintf("Length of the standings are not equal: %d != %d\n", len(sa1), len(sa2)))
	}

	for i := 0; i < len(sa1); i++ {
		s1 := sa1[i]
		s2 := sa2[i]
		err := equalStandingsRecord(s1, s2)
		if err != nil {
			return err
		}
	}
	return nil
}

func equalStandingsRecord(s1 *entity.Standing, s2 *entity.Standing) error {
	if s1.Player != s2.Player ||
		s1.Wins != s2.Wins ||
		s1.Losses != s2.Losses ||
		s1.Draws != s2.Draws ||
		s1.Spread != s2.Spread {
		return errors.New(fmt.Sprintf("Standings do not match: (%s, %d, %d, %d, %d) != (%s, %d, %d, %d, %d)",
			s1.Player, s1.Wins, s1.Losses, s1.Draws, s1.Spread,
			s2.Player, s2.Wins, s2.Losses, s2.Draws, s2.Spread))
	}
	return nil
}

func getPlayerPairings(players []string, pris []*entity.PlayerRoundInfo) []string {
	m := make(map[string]int)
	for _, player := range players {
		m[player] = 0
	}

	playerPairings := []string{}
	for _, pri := range pris {
		// An eliminated player could have nil for Players, skip them
		if pri.Pairing.Players != nil && m[pri.Pairing.Players[0]] == 0 {
			playerPairings = append(playerPairings, pri.Pairing.Players[0])
			playerPairings = append(playerPairings, pri.Pairing.Players[1])
			m[pri.Pairing.Players[0]] = 1
			m[pri.Pairing.Players[1]] = 1
		}
	}
	return playerPairings
}

func newPlayerRoundInfo(tc *ClassicDivision, playerOne string, playerTwo string, gamesPerRound int, round int) *entity.PlayerRoundInfo {
	return &entity.PlayerRoundInfo{Pairing: newClassicPairing(tc, playerOne, playerTwo, round)}
}

func equalPRI(pri1 *entity.PlayerRoundInfo, pri2 *entity.PlayerRoundInfo) error {
	err := equalPairing(pri1.Pairing, pri2.Pairing)
	if err != nil {
		return err
	}
	return nil
}

func equalPairing(p1 *entity.Pairing, p2 *entity.Pairing) error {
	// We are not concerned with ordering
	// Firsts and seconds are tested independently
	if (p1.Players[0] != p2.Players[0] && p1.Players[0] != p2.Players[1]) ||
		(p1.Players[1] != p2.Players[0] && p1.Players[1] != p2.Players[1]) {
		return errors.New(fmt.Sprintf("Players are not the same: (%s, %s) != (%s, %s)",
			p1.Players[0],
			p1.Players[1],
			p2.Players[0],
			p2.Players[1]))
	}
	if p1.Outcomes[0] != p2.Outcomes[0] || p1.Outcomes[1] != p2.Outcomes[1] {
		return errors.New(fmt.Sprintf("Outcomes are not the same: (%d, %d) != (%d, %d)",
			p1.Outcomes[0],
			p1.Outcomes[1],
			p2.Outcomes[0],
			p2.Outcomes[1]))
	}
	if len(p1.Games) != len(p2.Games) {
		return errors.New(fmt.Sprintf("Number of games are not the same: %d != %d", len(p1.Games), len(p2.Games)))
	}
	for i := 0; i < len(p1.Games); i++ {
		err := equalTournamentGame(p1.Games[i], p2.Games[i], i)
		if err != nil {
			return err
		}
	}
	return nil
}

func equalTournamentGame(t1 *entity.TournamentGame, t2 *entity.TournamentGame, i int) error {
	if t1.Scores[0] != t2.Scores[0] || t1.Scores[1] != t2.Scores[1] {
		return errors.New(fmt.Sprintf("Scores are not the same at game %d: (%d, %d) != (%d, %d)",
			i,
			t1.Scores[0],
			t1.Scores[1],
			t2.Scores[0],
			t2.Scores[1]))
	}
	if t1.Results[0] != t2.Results[0] || t1.Results[1] != t2.Results[1] {
		return errors.New(fmt.Sprintf("Results are not the same at game %d: (%d, %d) != (%d, %d)",
			i,
			t1.Results[0],
			t1.Results[1],
			t2.Results[0],
			t2.Results[1]))
	}
	if t1.GameEndReason != t2.GameEndReason {
		return errors.New(fmt.Sprintf("Game end reasons are not the same for game %d: %d != %d", i, t1.GameEndReason, t2.GameEndReason))
	}
	return nil
}

func equalRoundRobinPairings(s1 []string, s2 []string) error {
	if len(s1) != len(s2) {
		return errors.New(fmt.Sprintf("Pairing lengths do not match: %d != %d\n", len(s1), len(s2)))
	}
	for i := 0; i < len(s1); i++ {
		if s1[i] != s2[i] {
			return errors.New(fmt.Sprintf("Pairings are not equal:\n%s\n%s\n", strings.Join(s1, ", "), strings.Join(s2, ", ")))
		}
	}
	return nil
}

func printPriPairings(pris []*entity.PlayerRoundInfo) {
	for _, pri := range pris {
		fmt.Println(pri.Pairing)
	}
}

func printStandings(standings []*entity.Standing) {
	for _, standing := range standings {
		fmt.Println(standing)
	}
}

func logTwo(n int) int {
	res := 0
	for n > 1 {
		if n%2 != 0 {
			return -1
		}
		res++
		n = n / 2
	}
	return res
}