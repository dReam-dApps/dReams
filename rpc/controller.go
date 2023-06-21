package rpc

type displayStrings struct {
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

type predictionValues struct {
	Init   bool
	Amount uint64
	Buffer int64
}

type signals struct {
	Startup bool
}
