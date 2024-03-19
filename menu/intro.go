package menu

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/dReam-dApps/dReams/bundle"
)

// Menu tree items for dApp intros
type IntroText struct {
	name    string
	content []string
}

// Create menu tree items for dApps
func MakeMenuIntro(items map[string][]string) (entries []IntroText) {
	var menu_entry IntroText
	for name, e := range items {
		menu_entry.name = name
		menu_entry.content = e
		entries = append(entries, menu_entry)
	}

	return
}

// Menu instruction tree
func IntroTree(intros []IntroText) fyne.CanvasObject {
	list := map[string][]string{
		"":                        {"Welcome to dReams"},
		"Welcome to dReams":       {"Get Started", "dApps", "Assets", "Market"},
		"Get Started":             {"Visit dero.io for daemon and wallet download info", "Connecting", "FAQ"},
		"Connecting":              {"Daemon", "Wallet"},
		"FAQ":                     {"Can't connect", "How to resync Gnomon DB", "Can't see any tables, contracts or market info", "How to see terminal log", "Visit dreamdapps.io for further documentation"},
		"Can't connect":           {"Using a local daemon will yield the best results", "If you are using a remote daemon, try changing daemons", "Any connection errors can be found in terminal log"},
		"How to resync Gnomon DB": {"Go to Gnomon options in Menu", "If Gnomon is running you will be prompted to shut it down to make changes", "Click the delete DB button", "Reconnect to a daemon to resync", "Any sync errors can be found in terminal log"},

		"Can't see any tables, contracts or market info": {"Make sure daemon, wallet and Gnomon indicators are lit up solid", "If you've added new dApps to your dReams, a Gnomon resync will add them to your index", "Look in the asset tab for number of indexed SCIDs, it should be above 0", "Make sure your collection or dApp is enabled", "Try resyncing", "Any errors can be found in terminal log"},

		"How to see terminal log": {"Windows", "Mac", "Linux"},
		"Windows":                 {"Open powershell or command prompt", "Navigate to dReams directory", `Start dReams using       .\dReams-windows-amd64.exe`},
		"Mac":                     {"Open a terminal", "Navigate to dReams directory", `Start dReams using       ./dReams-macos-amd64`},
		"Linux":                   {"Open a terminal", "Navigate to dReams directory", `Start dReams using       ./dReams-linux-amd64`},
		"Daemon":                  {"Using local daemon will give best performance while using dReams", "Remote daemon options are available in drop down if a local daemon is not available", "Enter daemon address and the D light in top right will light up if connection is successful", "Once daemon is connected Gnomon will start up, the Gnomon indicator light will have a stripe in middle"},
		"Wallet":                  {"Set up and register a Dero wallet", "Your wallet will need to be running rpc server", "Using cli, start your wallet with flags --rpc-server --rpc-login=user:pass", "With Engram, turn on cyberdeck to start rpc server", "In dReams enter your wallet rpc address and rpc user:pass", "Press connect and the W light in top right will light up if connection is successful", "Once wallet is connected and Gnomon is running, Gnomon will sync with wallet", "The Gnomon indicator will turn solid when this is complete, everything is now connected"},

		"dApps":         {"Loading dApps", "Holdero", "Baccarat", "Predictions", "Sports", "dService", "Iluma", "Asset Duels", "Grokked", "dDice", "Contract Ratings"},
		"Loading dApps": {"You can add or remove dApps in the dApps tab", "Loading changes will disconnect your wallet", "Gnomon will continue to run, but may need to be resynced to index any new dApps added", "Your dApp preferences will be saved in local config file", "Loading only the dApps you are using will increase Gnomon and dReams performance"},
	}

	for i := range intros {
		list[intros[i].name] = intros[i].content
	}

	list["Contract Ratings"] = []string{
		"dReams has a public rating store on chain for multiplayer contracts",
		"Players can rate other contracts positively or negatively",
		"Four rating tiers, tier two being the starting tier for all contracts",
		"Each rating transaction is weight based by its Dero value",
		"Contracts that fall below tier one will no longer populate in the public index"}

	list["Assets"] = []string{
		"Enabling assets collections",
		"View any owned assets held in wallet",
		"Put owned assets up for auction or for sale",
		"Send assets privately to another wallet",
		"Headers, set the Gnomon SC headers to a owned contract",
		"Profile, change background themes, set your global name and assets for in use in dApps",
		"Indexer, add custom contracts to your index and search current index DB"}

	list["Enabling assets collections"] = []string{
		"You can enable or disable indexing of any asset collection in the Asset/Index tab",
		"Changes will require Gnomon DB to be resynced to take effect",
		"Your collection preferences will be saved in local config file",
		"Loading only the asset collections you are using will increase Gnomon and dReams performance"}

	list["Market"] = []string{
		"View any in game assets up for auction or sale",
		"Search all NFAs",
		"Bid on or buy assets",
		"Cancel or close out any existing listings",
		"Create NFA charity auctions and sales"}

	tree := widget.NewTreeWithStrings(list)

	tree.OnBranchClosed = func(uid widget.TreeNodeID) {
		tree.UnselectAll()
		if uid == "Welcome to dReams" {
			tree.CloseAllBranches()
		}
	}

	tree.OnBranchOpened = func(uid widget.TreeNodeID) {
		tree.Select(uid)
	}

	tree.OpenBranch("Welcome to dReams")

	alpha120 := canvas.NewRectangle(color.RGBA{0, 0, 0, 120})
	if bundle.AppColor == color.White {
		alpha120 = canvas.NewRectangle(color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x55})
	}

	return container.NewStack(alpha120, tree)
}
