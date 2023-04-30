package menu

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image/color"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/SixofClubsss/dReams/bundle"
	"github.com/SixofClubsss/dReams/dwidget"
	"github.com/SixofClubsss/dReams/holdero"
	"github.com/SixofClubsss/dReams/rpc"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/validation"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type marketObjects struct {
	Tab           string
	Entry         *dwidget.DeroAmts
	Name          *canvas.Text
	Type          *canvas.Text
	Collection    *canvas.Text
	Description   *canvas.Text
	Creator       *canvas.Text
	Owner         *canvas.Text
	Owner_update  *canvas.Text
	Start_price   *canvas.Text
	Art_fee       *canvas.Text
	Royalty       *canvas.Text
	Bid_count     *canvas.Text
	Buy_price     *canvas.Text
	Current_bid   *canvas.Text
	Bid_price     *canvas.Text
	End_time      *canvas.Text
	Loading       *canvas.Text
	Market_button *widget.Button
	Cancel_button *widget.Button
	Close_button  *widget.Button
	Auction_list  *widget.List
	Buy_list      *widget.List
	Icon          canvas.Image
	Cover         canvas.Image
	Details_box   fyne.Container
	Market_box    fyne.Container
	Confirming    bool
	DreamsFilter  bool
	Buy_amt       uint64
	Bid_amt       uint64
	Viewing       string
	Viewing_coll  string
	Auctions      []string
	Buy_now       []string
	Filters       []string
}

type assetObjects struct {
	Dreams_bal    *canvas.Text
	Dero_bal      *canvas.Text
	Dero_price    *canvas.Text
	Wall_height   *canvas.Text
	Daem_height   *canvas.Text
	Gnomes_height *canvas.Text
	Gnomes_sync   *canvas.Text
	Gnomes_index  *canvas.Text
	Index_entry   *widget.Entry
	Index_button  *widget.Button
	Index_search  *widget.Button
	Asset_list    *widget.List
	Assets        []string
	Asset_map     map[string]string
	Name          *canvas.Text
	Collection    *canvas.Text
	Descrption    *canvas.Text
	Icon          canvas.Image
	Stats_box     fyne.Container
	Header_box    fyne.Container
}

var Assets assetObjects
var Market marketObjects

// NFA market amount entry
func MarketEntry() fyne.CanvasObject {
	Market.Entry = dwidget.DeroAmtEntry("", 0.1, 1)
	Market.Entry.ExtendBaseWidget(Market.Entry)
	Market.Entry.SetText("0.0")
	Market.Entry.PlaceHolder = "Dero:"
	Market.Entry.Validator = validation.NewRegexp(`^\d{1,}\.\d{1,5}$|^[^0.]\d{0,}$`, "Int or float required")
	Market.Entry.OnChanged = func(s string) {
		if Market.Entry.Validate() != nil {
			Market.Entry.SetText("0.0")
		}
	}

	return Market.Entry
}

// Confirm a bid or buy action of listed NFA
//   - amt of Dero in atomic units
//   - b defines auction or sale
//   - Pass main window obj to reset to
func BidBuyConfirm(scid string, amt uint64, b int, obj *container.Split, reset fyne.CanvasObject) fyne.CanvasObject {
	var text, coll, name string
	f := float64(amt)
	amt_str := fmt.Sprintf("%.5f", f/100000)
	switch b {
	case 0:
		listing := checkNfaAuctionListing(scid)
		split := strings.Split(listing, "   ")
		if len(split) == 4 {
			coll = split[0]
			name = split[1]
		}
		text = fmt.Sprintf("Bidding on SCID:\n\n%s\n\nAsset: %s\n\nCollection: %s\n\nBid amount: %s Dero\n\nConfirm bid", scid, name, coll, amt_str)
	case 1:
		listing := checkNfaBuyListing(scid)
		split := strings.Split(listing, "   ")
		if len(split) == 4 {
			coll = split[0]
			name = split[1]
		}
		text = fmt.Sprintf("Buying SCID:\n\n%s\n\nAsset: %s\n\nCollection: %s\n\nAmount: %s Dero\n\nConfirm buy", scid, name, coll, amt_str)
	default:

	}

	Market.Confirming = true

	label := widget.NewLabel(text)
	label.Wrapping = fyne.TextWrapWord
	label.Alignment = fyne.TextAlignCenter

	confirm := widget.NewButton("Confirm", func() {
		switch b {
		case 0:
			rpc.BidBuyNFA(scid, "Bid", amt)
		case 1:
			rpc.BidBuyNFA(scid, "BuyItNow", amt)
		default:

		}

		obj.Trailing.(*fyne.Container).Objects[1] = reset
		obj.Trailing.(*fyne.Container).Objects[1].Refresh()
		Market.Confirming = false
	})

	cancel := widget.NewButton("Cancel", func() {
		obj.Trailing.(*fyne.Container).Objects[1] = reset
		obj.Trailing.(*fyne.Container).Objects[1].Refresh()
		Market.Confirming = false
	})

	left := container.NewVBox(confirm)
	right := container.NewVBox(cancel)
	buttons := container.NewAdaptiveGrid(2, left, right)

	content := container.NewVBox(layout.NewSpacer(), label, layout.NewSpacer(), buttons)

	go func() {
		for rpc.Wallet.Connect && rpc.Daemon.Connect {
			time.Sleep(time.Second)
		}

		obj.Trailing.(*fyne.Container).Objects[1] = reset
		obj.Trailing.(*fyne.Container).Objects[1].Refresh()
		Market.Confirming = false
	}()

	return container.NewMax(content)
}

// Confirm a cancel or close action of listed NFA
//   - c defines close or cancel
//   - Confirmation string from Market.Tab
//   - Pass main window obj to reset to
func ConfirmCancelClose(scid string, c int, obj *container.Split, reset fyne.CanvasObject) fyne.CanvasObject {
	var text, coll, name string

	if Market.Tab == "Buy" {
		listing := checkNfaBuyListing(scid)
		split := strings.Split(listing, "   ")
		if len(split) == 4 {
			coll = split[0]
			name = split[1]
		}
	} else if Market.Tab == "Auction" {
		listing := checkNfaAuctionListing(scid)
		split := strings.Split(listing, "   ")
		if len(split) == 4 {
			coll = split[0]
			name = split[1]
		}
	}
	switch c {
	case 0:
		text = fmt.Sprintf("Close listing for SCID:\n\n%s\n\nAsset: %s\n\nCollection: %s", scid, name, coll)
	case 1:
		text = fmt.Sprintf("Cancel listing for SCID:\n\n%s\n\nAsset: %s\n\nCollection: %s", scid, name, coll)
	default:

	}

	Market.Confirming = true

	label := widget.NewLabel(text)
	label.Wrapping = fyne.TextWrapWord
	label.Alignment = fyne.TextAlignCenter

	confirm := widget.NewButton("Confirm", func() {
		switch c {
		case 0:
			rpc.CancelCloseNFA(scid, "CloseListing")
		case 1:
			rpc.CancelCloseNFA(scid, "CancelListing")

		default:

		}
		Market.Cancel_button.Hide()
		Market.Viewing = ""
		Market.Viewing_coll = ""
		obj.Trailing.(*fyne.Container).Objects[1] = reset
		obj.Trailing.(*fyne.Container).Objects[1].Refresh()
		Market.Confirming = false
	})

	cancel := widget.NewButton("Cancel", func() {
		obj.Trailing.(*fyne.Container).Objects[1] = reset
		obj.Trailing.(*fyne.Container).Objects[1].Refresh()
		Market.Confirming = false
	})

	left := container.NewVBox(confirm)
	right := container.NewVBox(cancel)
	buttons := container.NewAdaptiveGrid(2, left, right)

	content := container.NewVBox(layout.NewSpacer(), label, layout.NewSpacer(), buttons)

	return container.NewMax(content)
}

// NFA auction listings object
//   - Gets images and details for Market objects on selected
func AuctionListings() fyne.Widget {
	Market.Auction_list = widget.NewList(
		func() int {
			return len(Market.Auctions)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(Market.Auctions[i])
		})

	Market.Auction_list.OnSelected = func(id widget.ListItemID) {
		if id != 0 {
			split := strings.Split(Market.Auctions[id], "   ")
			if split[3] != Market.Viewing {
				Market.Entry.SetText("")
				clearNfaImages()
				Market.Viewing = split[3]
				go GetNfaImages(split[3])
				go GetAuctionDetails(split[3])
			}
		}
	}

	return Market.Auction_list
}

// NFA buy now listings object
//   - Gets images and details for Market objects on selected
func BuyNowListings() fyne.Widget {
	Market.Buy_list = widget.NewList(
		func() int {
			return len(Market.Buy_now)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(Market.Buy_now[i])
		})

	Market.Buy_list.OnSelected = func(id widget.ListItemID) {
		if id != 0 {
			split := strings.Split(Market.Buy_now[id], "   ")
			if split[3] != Market.Viewing {
				clearNfaImages()
				Market.Viewing = split[3]
				go GetNfaImages(split[3])
				go GetBuyNowDetails(split[3])
			}
		}
	}

	return Market.Buy_list
}

// NFA market icon image with frame
//   - Pass res for frame resource
func NfaIcon(res fyne.Resource) fyne.CanvasObject {
	Market.Icon.Resize(fyne.NewSize(94, 94))
	Market.Icon.Move(fyne.NewPos(8, 3))

	frame := canvas.NewImageFromResource(res)
	frame.SetMinSize(fyne.NewSize(100, 100))
	frame.Resize(fyne.NewSize(100, 100))
	frame.Move(fyne.NewPos(5, 0))

	cont := *container.NewWithoutLayout(&Market.Icon, frame)

	return &cont
}

// Badge for dReam Tools enabled assets
//   - Pass res for frame resource
func ToolsBadge(res fyne.Resource) fyne.CanvasObject {
	badge := *canvas.NewImageFromResource(bundle.ResourceDReamToolsPng)
	badge.Resize(fyne.NewSize(94, 94))
	badge.Move(fyne.NewPos(8, 3))

	frame := canvas.NewImageFromResource(res)
	frame.SetMinSize(fyne.NewSize(100, 100))
	frame.Resize(fyne.NewSize(100, 100))
	frame.Move(fyne.NewPos(5, 0))

	cont := *container.NewWithoutLayout(&badge, frame)

	return &cont
}

// NFA cover image for market display
func NfaImg(img canvas.Image) *fyne.Container {
	Market.Cover.Resize(fyne.NewSize(266, 400))
	Market.Cover.Move(fyne.NewPos(425, -213))

	cont := container.NewWithoutLayout(&img)

	return cont
}

// Show text while loading market image
func loadingText() *fyne.Container {
	Market.Loading = canvas.NewText("Loading", bundle.TextColor)
	Market.Loading.TextSize = 18
	Market.Loading.Move(fyne.NewPos(400, 0))

	cont := container.NewWithoutLayout(Market.Loading)

	return cont
}

// Do while market loading text is showing
func loadingTextLoop() {
	if len(Market.Loading.Text) < 21 {
		for i := 0; i < 3; i++ {
			Market.Loading.Text = Market.Loading.Text + "."
			Market.Loading.Refresh()
			Market.Details_box.Objects[0].Refresh()
			time.Sleep(300 * time.Millisecond)
		}
	}
}

// Clears all market NFA images
func clearNfaImages() {
	Market.Details_box.Objects[1].(*fyne.Container).Objects[1] = layout.NewSpacer()
	Market.Icon = *canvas.NewImageFromImage(nil)
	Market.Details_box.Objects[1].Refresh()

	Market.Cover = *canvas.NewImageFromImage(nil)
	Market.Details_box.Objects[0] = loadingText()
	Market.Details_box.Objects[0].Refresh()
	Market.Details_box.Refresh()
}

// Set up market info objects
func NfaMarketInfo() fyne.CanvasObject {
	Market.Name = canvas.NewText(" Name: ", bundle.TextColor)
	Market.Type = canvas.NewText(" Asset Type: ", bundle.TextColor)
	Market.Collection = canvas.NewText(" Collection: ", bundle.TextColor)
	Market.Description = canvas.NewText(" Description: ", bundle.TextColor)
	Market.Creator = canvas.NewText(" Creator: ", bundle.TextColor)
	Market.Art_fee = canvas.NewText(" Artificer Fee: ", bundle.TextColor)
	Market.Royalty = canvas.NewText(" Royalty: ", bundle.TextColor)
	Market.Start_price = canvas.NewText(" Start Price: ", bundle.TextColor)
	Market.Owner = canvas.NewText(" Owner: ", bundle.TextColor)
	Market.Owner_update = canvas.NewText(" Owner can update: ", bundle.TextColor)
	Market.Current_bid = canvas.NewText(" Current Bid: ", bundle.TextColor)
	Market.Bid_price = canvas.NewText(" Minimum Bid: ", bundle.TextColor)
	Market.Bid_count = canvas.NewText(" Bids: ", bundle.TextColor)
	Market.End_time = canvas.NewText(" Ends At: ", bundle.TextColor)

	Market.Name.TextSize = 18
	Market.Type.TextSize = 18
	Market.Collection.TextSize = 18
	Market.Description.TextSize = 18
	Market.Creator.TextSize = 18
	Market.Art_fee.TextSize = 18
	Market.Royalty.TextSize = 18
	Market.Start_price.TextSize = 18
	Market.Owner.TextSize = 18
	Market.Owner_update.TextSize = 18
	Market.Bid_price.TextSize = 18
	Market.Current_bid.TextSize = 18
	Market.Bid_count.TextSize = 18
	Market.End_time.TextSize = 18

	Market.Icon.SetMinSize(fyne.NewSize(94, 94))
	Market.Cover.SetMinSize(fyne.NewSize(133, 200))

	return AuctionInfo()
}

// Container for auction info objects
func AuctionInfo() fyne.CanvasObject {
	Market.Details_box = *container.NewVBox(
		NfaImg(Market.Cover),
		container.NewHBox(NfaIcon(bundle.ResourceAvatarFramePng), layout.NewSpacer()),
		Market.Name,
		Market.Type,
		Market.Collection,
		Market.Description,
		Market.Creator,
		Market.Owner,
		Market.Art_fee,
		Market.Royalty,
		Market.Start_price,
		Market.Current_bid,
		Market.Bid_price,
		Market.Bid_count,
		Market.End_time)

	Market.Details_box.Refresh()

	return &Market.Details_box
}

// Refresh Market images
func RefreshNfaImages() {
	if Market.Cover.Resource != nil {
		Market.Details_box.Objects[0] = NfaImg(Market.Cover)
		Market.Details_box.Objects[0].Refresh()
	} else {
		go loadingTextLoop()
	}

	if Market.Icon.Resource != nil {
		Market.Details_box.Objects[1].(*fyne.Container).Objects[0] = NfaIcon(bundle.ResourceAvatarFramePng)
		Market.Details_box.Objects[1].Refresh()
	}
	view := Market.Viewing_coll
	if view == "AZYPC" || view == "SIXPC" || view == "AZYPCB" || view == "SIXPCB" {
		Market.Details_box.Objects[1].(*fyne.Container).Objects[1] = ToolsBadge(bundle.ResourceAvatarFramePng)
	} else {
		Market.Details_box.Objects[1].(*fyne.Container).Objects[1] = layout.NewSpacer()
	}

	Market.Details_box.Objects[1].Refresh()
}

// Set auction display content to default values
func ResetAuctionInfo() {
	Market.Bid_amt = 0
	clearNfaImages()
	Market.Name.Text = (" Name: ")
	Market.Name.Refresh()
	Market.Type.Text = (" Asset Type: ")
	Market.Type.Refresh()
	Market.Collection.Text = (" Collection: ")
	Market.Collection.Refresh()
	Market.Description.Text = (" Description: ")
	Market.Description.Refresh()
	Market.Creator.Text = (" Creator: ")
	Market.Creator.Refresh()
	Market.Art_fee.Text = (" Artificer Fee: ")
	Market.Art_fee.Refresh()
	Market.Royalty.Text = (" Royalty: ")
	Market.Royalty.Refresh()
	Market.Start_price.Text = (" Start Price: ")
	Market.Start_price.Refresh()
	Market.Owner.Text = (" Owner: ")
	Market.Owner.Refresh()
	Market.Owner_update.Text = (" Owner can update: ")
	Market.Owner_update.Refresh()
	Market.Current_bid.Text = (" Current Bid: ")
	Market.Current_bid.Refresh()
	Market.Bid_price.Text = (" Minimum Bid: ")
	Market.Bid_price.Refresh()
	Market.Bid_count.Text = (" Bids: ")
	Market.Bid_count.Refresh()
	Market.End_time.Text = (" Ends At: ")
	Market.End_time.Refresh()
	Market.Details_box.Refresh()
}

// Container for buy now info objects
func BuyNowInfo() fyne.CanvasObject {
	Market.Details_box = *container.NewVBox(
		NfaImg(Market.Cover),
		container.NewHBox(NfaIcon(bundle.ResourceAvatarFramePng), layout.NewSpacer()),
		Market.Name,
		Market.Type,
		Market.Collection,
		Market.Description,
		Market.Creator,
		Market.Owner,
		Market.Art_fee,
		Market.Royalty,
		Market.Start_price,
		Market.End_time)

	Market.Details_box.Refresh()

	return &Market.Details_box
}

// Set buy now display content to default values
func ResetBuyInfo() {
	Market.Buy_amt = 0
	clearNfaImages()
	Market.Name.Text = (" Name: ")
	Market.Name.Refresh()
	Market.Type.Text = (" Asset Type: ")
	Market.Type.Refresh()
	Market.Collection.Text = (" Collection: ")
	Market.Collection.Refresh()
	Market.Description.Text = (" Description: ")
	Market.Description.Refresh()
	Market.Creator.Text = (" Creator: ")
	Market.Creator.Refresh()
	Market.Art_fee.Text = (" Artificer Fee: ")
	Market.Art_fee.Refresh()
	Market.Royalty.Text = (" Royalty: ")
	Market.Royalty.Refresh()
	Market.Start_price.Text = (" Buy now for: ")
	Market.Start_price.Refresh()
	Market.Owner.Text = (" Owner: ")
	Market.Owner.Refresh()
	Market.Owner_update.Text = (" Owner can update: ")
	Market.Owner_update.Refresh()
	Market.End_time.Text = (" Ends At: ")
	Market.End_time.Refresh()
	Market.Details_box.Refresh()
}

// Switch triggered when market tab changes
func MarketTab(ti *container.TabItem) {
	switch ti.Text {
	case "Auctions":
		go FindNfaListings(nil)
		Market.Tab = "Auction"
		Market.Auction_list.UnselectAll()
		Market.Viewing = ""
		Market.Viewing_coll = ""
		Market.Market_button.Text = "Bid"
		Market.Market_button.Refresh()
		Market.Entry.SetText("0.0")
		Market.Entry.Enable()
		ResetAuctionInfo()
		AuctionInfo()
	case "Buy Now":
		go FindNfaListings(nil)
		Market.Tab = "Buy"
		Market.Buy_list.UnselectAll()
		Market.Viewing = ""
		Market.Viewing_coll = ""
		Market.Market_button.Text = "Buy"
		Market.Entry.Disable()
		Market.Market_button.Refresh()
		Market.Details_box.Refresh()
		ResetBuyInfo()
		BuyNowInfo()
	}
}

// NFA market layout
func PlaceMarket() *container.Split {
	details := container.NewMax(NfaMarketInfo())

	tabs := container.NewAppTabs(
		container.NewTabItem("Auctions", AuctionListings()),
		container.NewTabItem("Buy Now", BuyNowListings()))

	tabs.SetTabLocation(container.TabLocationTop)
	tabs.OnSelected = func(ti *container.TabItem) {
		MarketTab(ti)
	}

	Market.Tab = "Auction"

	scroll_top := widget.NewButtonWithIcon("", fyne.Theme.Icon(fyne.CurrentApp().Settings().Theme(), "arrowUp"), func() {
		switch Market.Tab {
		case "Buy":
			Market.Buy_list.ScrollToTop()
		case "Auction":
			Market.Auction_list.ScrollToTop()
		default:

		}
	})

	scroll_bottom := widget.NewButtonWithIcon("", fyne.Theme.Icon(fyne.CurrentApp().Settings().Theme(), "arrowDown"), func() {
		switch Market.Tab {
		case "Buy":
			Market.Buy_list.ScrollToBottom()
		case "Auction":
			Market.Auction_list.ScrollToBottom()
		default:

		}
	})

	scroll_top.Importance = widget.LowImportance
	scroll_bottom.Importance = widget.LowImportance

	scroll_cont := container.NewVBox(container.NewHBox(layout.NewSpacer(), scroll_top, scroll_bottom))

	max := container.NewMax(bundle.Alpha120, tabs, scroll_cont)

	details_box := container.NewVBox(layout.NewSpacer(), details)

	menu_top := container.NewHSplit(details_box, max)
	menu_top.SetOffset(0)

	Market.Market_button = widget.NewButton("Bid", func() {
		scid := Market.Viewing
		if len(scid) == 64 {
			text := Market.Market_button.Text
			Market.Market_button.Hide()
			if text == "Bid" {
				amt := ToAtomicFive(Market.Entry.Text)
				menu_top.Trailing.(*fyne.Container).Objects[1] = BidBuyConfirm(scid, amt, 0, menu_top, container.NewMax(tabs, scroll_cont))
				menu_top.Trailing.(*fyne.Container).Objects[1].Refresh()
			} else if text == "Buy" {
				menu_top.Trailing.(*fyne.Container).Objects[1] = BidBuyConfirm(scid, Market.Buy_amt, 1, menu_top, container.NewMax(tabs, scroll_cont))
				menu_top.Trailing.(*fyne.Container).Objects[1].Refresh()
			}
		}
	})

	Market.Market_button.Hide()

	Market.Cancel_button = widget.NewButton("Cancel", func() {
		if len(Market.Viewing) == 64 {
			Market.Cancel_button.Hide()
			menu_top.Trailing.(*fyne.Container).Objects[1] = ConfirmCancelClose(Market.Viewing, 1, menu_top, container.NewMax(tabs, scroll_cont))
			menu_top.Trailing.(*fyne.Container).Objects[1].Refresh()
		}
	})

	Market.Close_button = widget.NewButton("Close", func() {
		if len(Market.Viewing) == 64 {
			Market.Close_button.Hide()
			menu_top.Trailing.(*fyne.Container).Objects[1] = ConfirmCancelClose(Market.Viewing, 0, menu_top, container.NewMax(tabs, scroll_cont))
			menu_top.Trailing.(*fyne.Container).Objects[1].Refresh()
		}
	})

	Market.Cancel_button.Hide()
	Market.Close_button.Hide()

	Market.Market_box = *container.NewAdaptiveGrid(6, MarketEntry(), Market.Market_button, layout.NewSpacer(), layout.NewSpacer(), Market.Close_button, Market.Cancel_button)
	Market.Market_box.Hide()

	menu_bottom := container.NewAdaptiveGrid(1, &Market.Market_box)

	menu_box := container.NewVSplit(menu_top, menu_bottom)
	menu_box.SetOffset(1)

	return menu_box
}

// Recheck owned assets button
//   - tag for log print
//   - pass recheck for desired check
func RecheckButton(tag string, recheck func()) fyne.CanvasObject {
	button := widget.NewButton("Check Assets", func() {
		if !Gnomes.Wait {
			log.Printf("[%s] Rechecking Assets\n", tag)
			go recheck()
		}
	})

	return button
}

// dReams recheck owned assets routine
func RecheckDreamsAssets() {
	Gnomes.Wait = true
	Assets.Assets = []string{}
	CheckDreamsNFAs(false, nil)
	CheckDreamsG45s(false, nil)
	if Control.Dapp_list["Holdero"] {
		if rpc.Wallet.Connect {
			Control.Names.Options = []string{rpc.Wallet.Address[0:12]}
			CheckWalletNames(rpc.Wallet.Address)
		}
	}
	sort.Strings(Assets.Assets)
	Assets.Asset_list.UnselectAll()
	Assets.Asset_list.Refresh()
	Gnomes.Wait = false
}

// Owned asset tab layout
//   - tag for log print
//   - games enables dReams asset selects
//   - recheck for RecheckButton() func
//   - menu resources for side menus
//   - w for main window dialog
func PlaceAssets(tag string, games bool, recheck func(), menu_icon, menu_background fyne.Resource, w fyne.Window) *container.Split {
	items_box := container.NewAdaptiveGrid(2)

	if games {
		games_cont := container.NewVBox(
			holdero.FaceSelect(),
			holdero.BackSelect(),
			holdero.ThemeSelect(),
			holdero.AvatarSelect(Assets.Asset_map),
			holdero.SharedDecks(),
			RecheckButton(tag, recheck),
			layout.NewSpacer())

		cont := container.NewHScroll(games_cont)
		cont.SetMinSize(fyne.NewSize(290, 35.1875))

		items_box.Add(cont)
	}

	items_box.Add(container.NewAdaptiveGrid(1, AssetStats()))

	player_input := container.NewVBox(items_box, layout.NewSpacer())

	tabs := container.NewAppTabs(
		container.NewTabItem("Owned", AssetList()))

	tabs.OnSelected = func(ti *container.TabItem) {

	}

	scroll_top := widget.NewButtonWithIcon("", fyne.Theme.Icon(fyne.CurrentApp().Settings().Theme(), "arrowUp"), func() {
		Assets.Asset_list.ScrollToTop()
	})

	scroll_bottom := widget.NewButtonWithIcon("", fyne.Theme.Icon(fyne.CurrentApp().Settings().Theme(), "arrowDown"), func() {
		Assets.Asset_list.ScrollToBottom()
	})

	scroll_top.Importance = widget.LowImportance
	scroll_bottom.Importance = widget.LowImportance

	scroll_cont := container.NewVBox(container.NewHBox(layout.NewSpacer(), scroll_top, scroll_bottom))

	max := container.NewMax(bundle.Alpha120, tabs, scroll_cont)

	header_name_entry := widget.NewEntry()
	header_name_entry.PlaceHolder = "Name:"
	header_descr_entry := widget.NewEntry()
	header_descr_entry.PlaceHolder = "Description"
	header_icon_entry := widget.NewEntry()
	header_icon_entry.PlaceHolder = "Icon:"

	header_button := widget.NewButton("Set Headers", func() {
		scid := Assets.Index_entry.Text
		if len(scid) == 64 && header_name_entry.Text != "dReam Tables" && header_name_entry.Text != "dReams" {
			if _, ok := rpc.FindStringKey(rpc.GnomonSCID, scid, rpc.Daemon.Rpc).(string); ok {
				max.Objects[1] = setHeaderConfirm(header_name_entry.Text, header_descr_entry.Text, header_icon_entry.Text, scid, max.Objects, tabs)
				max.Objects[1].Refresh()
			} else {
				dialog.NewInformation("Check back soon", "SCID not stored on the main Gnomon SC yet\n\nOnce stored, you can set your SCID headers", w).Show()
			}
		}
	})

	header_contr := container.NewVBox(header_name_entry, header_descr_entry, header_icon_entry, header_button)
	Assets.Header_box = *container.NewAdaptiveGrid(2, header_contr)
	Assets.Header_box.Hide()

	player_input.Add(&Assets.Header_box)

	player_box := container.NewHBox(player_input)

	menu_top := container.NewHSplit(player_box, max)
	menu_bottom := container.NewAdaptiveGrid(1, IndexEntry(menu_icon, menu_background))

	menu_box := container.NewVSplit(menu_top, menu_bottom)
	menu_box.SetOffset(1)

	return menu_box
}

// Set wallet and chain display content for menu
func MenuDisplay() fyne.CanvasObject {
	Assets.Gnomes_sync = canvas.NewText("", color.RGBA{31, 150, 200, 210})
	Assets.Gnomes_height = canvas.NewText(" Gnomon Height: ", bundle.TextColor)
	Assets.Daem_height = canvas.NewText(" Daemon Height: ", bundle.TextColor)
	Assets.Wall_height = canvas.NewText(" Wallet Height: ", bundle.TextColor)
	Assets.Dreams_bal = canvas.NewText(" dReams Balance: ", bundle.TextColor)
	Assets.Dero_bal = canvas.NewText(" Dero Balance: ", bundle.TextColor)
	// price := getOgre("DERO-USDT")
	Assets.Dero_price = canvas.NewText(" Dero Price: $", bundle.TextColor)

	Assets.Gnomes_sync.TextSize = 18
	Assets.Gnomes_height.TextSize = 18
	Assets.Daem_height.TextSize = 18
	Assets.Wall_height.TextSize = 18
	Assets.Dreams_bal.TextSize = 18
	Assets.Dero_bal.TextSize = 18
	Assets.Dero_price.TextSize = 18
	exLabel := canvas.NewText(" 1 Dero = 333 dReams", bundle.TextColor)
	exLabel.TextSize = 18

	box := container.NewVBox(
		Assets.Gnomes_sync,
		Assets.Gnomes_height,
		Assets.Daem_height,
		Assets.Wall_height,
		Assets.Dreams_bal,
		Assets.Dero_bal,
		Assets.Dero_price, exLabel)

	return box
}

// Icon image for Holdero tables and asset viewing
//   - Pass res as frame resource
func IconImg(res fyne.Resource) *fyne.Container {
	Assets.Icon.SetMinSize(fyne.NewSize(100, 100))
	Assets.Icon.Resize(fyne.NewSize(94, 94))
	Assets.Icon.Move(fyne.NewPos(8, 3))

	frame := canvas.NewImageFromResource(res)
	frame.Resize(fyne.NewSize(100, 100))
	frame.Move(fyne.NewPos(5, 0))

	cont := container.NewWithoutLayout(&Assets.Icon, frame)

	return cont
}

// Display for owned asset info
func AssetStats() fyne.CanvasObject {
	Assets.Collection = canvas.NewText(" Collection: ", bundle.TextColor)
	Assets.Name = canvas.NewText(" Name: ", bundle.TextColor)

	Assets.Name.TextSize = 18
	Assets.Collection.TextSize = 18

	Assets.Stats_box = *container.NewVBox(Assets.Collection, Assets.Name, IconImg(nil))

	return &Assets.Stats_box
}

// Confirmation for setting SCID headers
//   - name, desc and icon of SCID header
//   - Pass main window obj to reset to
func setHeaderConfirm(name, desc, icon, scid string, obj []fyne.CanvasObject, reset *container.AppTabs) fyne.CanvasObject {
	label := widget.NewLabel("Headers for SCID:\n\n" + scid + "\n\nName: " + name + "\n\nDescription: " + desc + "\n\nIcon: " + icon)
	label.Wrapping = fyne.TextWrapWord
	label.Alignment = fyne.TextAlignCenter

	confirm_button := widget.NewButton("Confirm", func() {
		rpc.SetHeaders(name, desc, icon, scid)
		obj[1] = reset
		obj[1].Refresh()
	})

	cancel_button := widget.NewButton("Cancel", func() {
		obj[1] = reset
		obj[1].Refresh()

	})

	alpha := container.NewMax(canvas.NewRectangle(color.RGBA{0, 0, 0, 120}))
	buttons := container.NewAdaptiveGrid(2, confirm_button, cancel_button)
	content := container.NewVBox(layout.NewSpacer(), label, layout.NewSpacer(), buttons)

	return container.NewMax(alpha, content)
}

// Full routine for NFA market and scanning wallet for NFAs, can use PlaceAssets() and PlaceMarket() layouts
//   - tag for log print
//   - quit for exit chan
//   - connected box for DeroRpcEntries
func RunNFAMarket(tag string, quit chan struct{}, connect_box *dwidget.DeroRpcEntries) {
	go func() {
		time.Sleep(6 * time.Second)
		ticker := time.NewTicker(3 * time.Second)
		offset := 0

		for {
			select {
			case <-ticker.C: // do on interval
				rpc.Ping()
				rpc.EchoWallet(tag)

				// Get all NFA listings
				if !GnomonClosing() && offset%2 == 0 {
					FindNfaListings(nil)
					if offset > 19 {
						offset = 0
					}
				}

				rpc.GetBalance()
				if !rpc.Wallet.Connect {
					rpc.Wallet.Balance = 0
				}

				// Refresh Dero balance and Gnomon endpoint
				connect_box.RefreshBalance()
				if !rpc.Signal.Startup {
					GnomonEndPoint()
				}

				// If connected daemon connected start looking for Gnomon sync with daemon
				if rpc.Daemon.Connect && Gnomes.Init && !GnomonClosing() {
					connect_box.Disconnect.SetChecked(true)
					// Get indexed SCID count
					contracts := Gnomes.Indexer.Backend.GetAllOwnersAndSCIDs()
					Gnomes.SCIDS = uint64(len(contracts))
					if Gnomes.SCIDS > 0 {
						Gnomes.Checked = true
					}

					height := rpc.DaemonHeight(tag, rpc.Daemon.Rpc)
					if Gnomes.Indexer.LastIndexedHeight >= int64(height)-3 {
						Gnomes.Sync = true
					} else {
						Gnomes.Sync = false
						Gnomes.Checked = false
					}

					// Enable index controls and check if wallet is connected
					go DisableIndexControls(false)
					if rpc.Wallet.Connect {
						Market.Market_box.Show()
						Control.Claim_button.Show()
						// Update live market info
						if len(Market.Viewing) == 64 {
							if Market.Tab == "Buy" {
								GetBuyNowDetails(Market.Viewing)
								go RefreshNfaImages()
							} else {
								GetAuctionDetails(Market.Viewing)
								go RefreshNfaImages()
							}
						}
					} else {
						Market.Market_box.Hide()
						Control.List_button.Hide()
						Control.Send_asset.Hide()
						Control.Claim_button.Hide()
					}

					// Check wallet for all owned NFAs and refresh content
					go CheckAllNFAs(false, nil)
					indexed_scids := " Indexed SCIDs: " + strconv.Itoa(int(Gnomes.SCIDS))
					Assets.Gnomes_index.Text = (indexed_scids)
					Assets.Gnomes_index.Refresh()
					Assets.Stats_box = *container.NewVBox(Assets.Collection, Assets.Name, IconImg(bundle.ResourceAvatarFramePng))
					Assets.Stats_box.Refresh()
				} else {
					connect_box.Disconnect.SetChecked(false)
					DisableIndexControls(true)
				}

				if rpc.Daemon.Connect {
					rpc.Signal.Startup = false
				}

				offset++

			case <-quit: // exit
				log.Printf("[%s] Closing\n", tag)
				if Gnomes.Icon_ind != nil {
					Gnomes.Icon_ind.Stop()
				}
				ticker.Stop()
				return
			}
		}
	}()
}

// Get search filters from on chain store
func FetchFilters() {
	check := []string{}
	if stored, ok := rpc.FindStringKey(rpc.RatingSCID, "market_filter", rpc.Daemon.Rpc).(string); ok {
		if h, err := hex.DecodeString(stored); err == nil {
			if err = json.Unmarshal(h, &check); err != nil {
				log.Println("[FetchFilters] Market Filter", err)
			} else {
				Market.Filters = check
			}
		}
	} else {
		log.Println("[FetchFilters] Could not get Market Filter")
	}
}
