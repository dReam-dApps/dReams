package rpc

type displayStrings struct {
	Total_w  string
	Player_w string
	Banker_w string
	Ties     string
	BaccMax  string
	BaccMin  string
	BaccRes  string

	Prediction string
	P_feed     string

	Game    string
	S_count string
	League  string
	S_end   string
	TeamA   string
	TeamB   string

	Balance       map[string]string
	Wallet_height string
}

type baccValues struct {
	P_card1  int
	P_card2  int
	P_card3  int
	B_card1  int
	B_card2  int
	B_card3  int
	CHeight  int
	MinBet   float64
	MaxBet   float64
	AssetID  string
	Contract string
	Last     string
	Found    bool
	Display  bool
	Notified bool
}

type predictionValues struct {
	Init   bool
	Amount uint64
	Buffer int64
}

type signals struct {
	Startup bool
}
