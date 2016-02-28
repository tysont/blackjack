package blackjack

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// main starts the blackjack service and then continues to run indefinitely.
func main() {

	StartService("localhost", 8080)
	for true {

	}
}

// nextId is the next available identifier to give to a game.
var nextId uint64

// Games maps identifiers to games that have either concluded or are in progress.
var Games map[uint64]*Game

// serviceHost is the host that the web service is running on.
var serviceHost string

// servicePort is the port that the web service is listening on.
var servicePort int

// isInitialized is whether or not blackjack has been initialized.
var isInitialized bool = false

// Cards maps the list of card values to their numeric representation.
var Cards = map[int]string{
	1:  "A",
	2:  "2",
	3:  "3",
	4:  "4",
	5:  "5",
	6:  "6",
	7:  "7",
	8:  "8",
	9:  "9",
	10: "10",
	11: "J",
	12: "Q",
	13: "K",
}

// Game represents an individual game/round/hand of blackjack.
type Game struct {

	// Player holds the numeric representation of cards in the player's hand.
	Player []int `json:"player"`

	// Dealer holds the numeric representation of cards in the dealer's hand.
	Dealer []int `json:"dealer"`

	// Deck holds the numeric representation of cards remaining in the deck.
	Deck []int `json:"deck"`

	// Double is whether or not the player has doubled down on the current hand.
	Double bool `json:"double"`

	// Complete is whether or not the current game is complete.
	Complete bool `json:"complete"`

	// Payout is the player's payout, which currently always assumes a bet of 10.
	Payout int `json:"payout"`
}

// String prints the game state as a string.
func (game *Game) String() string {

	return fmt.Sprintf("%v:%v", readable(game.Player, false), readable(game.Dealer, !game.Complete))
}

// Initialize
func Initiatlize() {

	Games = make(map[uint64]*Game)
	nextId = 0
	isInitialized = true
}

// GetNextId gets the next available id to assign to a game.
func GetNextId() uint64 {

	nextId++
	return nextId
}

// getGameIds gets the list of allocated game identifiers.
func getGameIds() []uint64 {

	u := make([]uint64, len(Games))
	i := 0
	for k, _ := range Games {
		u[i] = k
		i++
	}

	return u
}

// Deal creates a new game and returns the identifier for the game.
func Deal(id uint64) uint64 {

	if !isInitialized {
		Initiatlize()
	}

	deck := shuffle()
	game := new(Game)
	game.Double = false

	player := make([]int, 0)
	dealer := make([]int, 0)
	var k int

	k, deck = Draw(deck)
	player = append(player, k)

	k, deck = Draw(deck)
	dealer = append(dealer, k)

	k, deck = Draw(deck)
	player = append(player, k)

	k, deck = Draw(deck)
	dealer = append(dealer, k)

	game.Deck = deck
	game.Player = player
	game.Dealer = dealer

	game.Complete = false
	game.Payout = 0

	Games[id] = game
	return id
}

// Hit causes the player to draw an additional card, and then evaluates the game.
func Hit(id uint64) *Game {

	game := Peek(id)
	k, deck := Draw(game.Deck)
	if game.Complete {
		return game
	}

	game.Player = append(game.Player, k)
	game.Deck = deck

	p, _ := Evaluate(game.Player)
	if p > 21 {
		game = Stand(id)
	}

	return game
}

// Stand means that the player is done drawing, and then evaluates and completes the game.
func Stand(id uint64) *Game {

	game := Peek(id)
	if game.Complete {
		return game
	}

	d, s := Evaluate(game.Dealer)
	for (d < 18) || ((d == 17) && (s == true)) {

		k, deck := Draw(game.Deck)
		game.Dealer = append(game.Dealer, k)
		game.Deck = deck
		d, s = Evaluate(game.Dealer)
	}

	game.Complete = true
	game.Payout = payout(game)
	return game
}

// Double means that the player takes one card and doubles their initial bet.
func Double(id uint64) *Game {

	game := Peek(id)
	if game.Complete {
		return game
	}

	game.Double = true
	game = Hit(id)
	game = Stand(id)
	return game
}

// Peek returns the game with the given identifier.
func Peek(id uint64) *Game {

	return Games[id]
}

// Shuffle creates a numeric representation of a randomized deck.
func shuffle() []int {

	hand := make([]int, 52)
	for k := 1; k <= 13; k++ {
		for i := (k - 1) * 4; i < k*4; i++ {
			hand[i] = k
		}
	}

	return hand
}

// Draw takes the top card off of a numeric representation of a deck, and returns the card and deck.
func Draw(deck []int) (int, []int) {

	k := rand.Intn(len(deck))
	n := deck[k]
	deck = append(deck[:k], deck[k+1:]...)
	return n, deck
}

// Evaluate returns the score of a numeric representation of a hand, and whether the player doubled.
func Evaluate(hand []int) (int, bool) {

	e := 0
	s := false

	for _, k := range hand {
		if (k >= 2) && (k <= 10) {
			e += k
		}
		if (k >= 11) && (k <= 13) {
			e += 10
		}
	}

	for _, k := range hand {
		if k == 1 {
			if e <= 10 {
				e += 11
				s = true
			} else {
				e += 1
			}
		}
	}

	return e, s
}

// readable converts the numeric representation of a hand to the list of card values.
func readable(hand []int, hide bool) []string {

	s := make([]string, len(hand))
	for i, k := range hand {
		if (i == 0) || (!hide) {
			s[i] = Cards[k]
		} else {
			s[i] = "X"
		}
	}

	return s
}

// payout calculates the amount that a player should receive for a game.
func payout(game *Game) int {

	//fmt.Println(game)
	p, _ := Evaluate(game.Player)
	d, _ := Evaluate(game.Dealer)

	if (p == 21) && (len(game.Player) == 2) {

		return 15
	}

	if (p > 21) || ((d <= 21) && (d > p)) {

		if game.Double {
			return -20
		}

		return -10

	} else if (d > 21) || (p > d) {

		if game.Double {

			return 20
		}

		return 10
	}

	return 0
}

// StartService starts the blackjack web service.
func StartService(host string, port int) {

	if !isInitialized {
		Initiatlize()
	}

	serviceHost = host
	servicePort = port

	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/blackjack", showGames)
	router.HandleFunc("/blackjack/{id}", showGame)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%v", servicePort), router))
}

// ShowGames writes the list of current games to the writer for the web service.
func showGames(w http.ResponseWriter, r *http.Request) {

	fmt.Fprintf(w, "<a href=\"http://%v:%v/blackjack/%v\">Deal</a><br />", serviceHost, servicePort, GetNextId())

	for _, k := range getGameIds() {

		fmt.Fprintf(w, "<a href=\"http://%v:%v/blackjack/%v\">Play Hand %v</a><br />", serviceHost, servicePort, k, k)
	}
}

// ShowGames writes the game status to the writer for the web service.
func showGame(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	i, _ := strconv.ParseUint(vars["id"], 10, 64)

	if _, x := Games[i]; !x {
		Deal(i)
	}

	a := r.URL.Query().Get("action")
	if a == "hit" {
		Hit(i)
	} else if a == "double" {
		Double(i)
	} else if a == "stand" {
		Stand(i)
	}

	g := Games[i]

	if !g.Complete {

		fmt.Fprintf(w, "<a href=\"http://%v:%v/blackjack/%v?action=hit\">Hit</a><br />", serviceHost, servicePort, i)
		fmt.Fprintf(w, "<a href=\"http://%v:%v/blackjack/%v?action=double\">Double</a><br />", serviceHost, servicePort, i)
		fmt.Fprintf(w, "<a href=\"http://%v:%v/blackjack/%v?action=stand\">Stand</a><br />", serviceHost, servicePort, i)

	} else {

		fmt.Fprintf(w, "<a href=\"http://%v:%v/blackjack\">Back</a><br />", serviceHost, servicePort)
		fmt.Fprintf(w, "Payout was %v.<br />", g.Payout)

	}

	fmt.Fprintf(w, "%v", g)
}
