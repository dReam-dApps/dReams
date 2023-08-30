package menu

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image/color"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/dReam-dApps/dReams/bundle"
	"github.com/dReam-dApps/dReams/dwidget"
	"github.com/dReam-dApps/dReams/rpc"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/validation"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	xwidget "fyne.io/x/fyne/widget"
)

type marketObjects struct {
	Tab           string
	Entry         *dwidget.DeroAmts
	Name          *widget.Entry
	Type          *widget.Entry
	Collection    *widget.Entry
	Description   *widget.Entry
	Creator       *widget.Entry
	Owner         *widget.Entry
	Owner_update  *widget.Entry
	Start_price   *widget.Entry
	Art_fee       *widget.Entry
	Royalty       *widget.Entry
	Bid_count     *widget.Entry
	Buy_price     *widget.Entry
	Current_bid   *widget.Entry
	Bid_price     *widget.Entry
	End_time      *widget.Entry
	Loading       *widget.ProgressBarInfinite
	Market_button *widget.Button
	Cancel_button *widget.Button
	Close_button  *widget.Button
	Auction_list  *widget.List
	Buy_list      *widget.List
	My_listings   *widget.List
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
	My_list       []string
	Filters       []string
}

type assetObjects struct {
	Swap          *fyne.Container
	Balances      *widget.List
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
	Icon          canvas.Image
	Stats_box     fyne.Container
	Header_box    fyne.Container
}

var Assets assetObjects
var Market marketObjects
var dReamsNFAs = []assetCount{
	{name: "AZYPC", count: 23},
	{name: "AZYPCB", count: 53},
	{name: "AZYDS", count: 10},
	{name: "DBC", count: 8},
	{name: "SIXPC", count: 9},
	{name: "SIXPCB", count: 10},
	{name: "SIXART", count: 17},
	{name: "HighStrangeness", count: 354},
	{name: "Dorblings NFA", count: 110},
	// // TODO correct counts
	// {name: "TestChars", count: 8},
	// {name: "TestItems", count: 8},
	// {name: "Dero Desperados", count: 5},
	// {name: "Desperado Guns", count: 5},
}

func (a *assetObjects) Add(name, scid string) {
	a.Assets = append(a.Assets, name+"   "+scid)
	a.Asset_map[name] = scid
}

// NFA market amount entry
func MarketEntry() fyne.CanvasObject {
	Market.Entry = dwidget.NewDeroEntry("", 0.1, 1)
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
		listing, _, _ := checkNfaAuctionListing(scid)
		split := strings.Split(listing, "   ")
		if len(split) == 4 {
			coll = split[0]
			name = split[1]
		}
		text = fmt.Sprintf("Bidding on SCID:\n\n%s\n\nAsset: %s\n\nCollection: %s\n\nBid amount: %s Dero\n\nConfirm bid", scid, name, coll, amt_str)
	case 1:
		listing, _, _ := checkNfaBuyListing(scid)
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
		for rpc.IsReady() {
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

	list, _ := CheckNFAListingType(scid)
	switch list {
	case 1:
		listing, _, _ := checkNfaAuctionListing(scid)
		split := strings.Split(listing, "   ")
		if len(split) == 4 {
			coll = split[0]
			name = split[1]
		}
	case 2:
		listing, _, _ := checkNfaBuyListing(scid)
		split := strings.Split(listing, "   ")
		if len(split) == 4 {
			coll = split[0]
			name = split[1]
		}
	default:

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
				Market.Details_box.Objects[1] = loadingBar()
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
				Market.Details_box.Objects[1] = loadingBar()
			}
		}
	}

	return Market.Buy_list
}

// NFA listing for connected wallet
//   - Gets images and details for Market objects on selected
func MyNFAListings() fyne.Widget {
	Market.My_listings = widget.NewList(
		func() int {
			return len(Market.My_list)
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			o.(*widget.Label).SetText(Market.My_list[i])
		})

	Market.My_listings.OnSelected = func(id widget.ListItemID) {
		if id != 0 {
			split := strings.Split(Market.My_list[id], "   ")
			if split[3] != Market.Viewing {
				clearNfaImages()
				Market.Viewing = split[3]
				go GetNfaImages(split[3])
				go GetUnlistedDetails(split[3])
				Market.Details_box.Objects[1] = loadingBar()
			}
		}
	}

	return Market.My_listings
}

// Search NFA objects
func SearchNFAs() fyne.CanvasObject {
	var dest_addr string
	search_entry := xwidget.NewCompletionEntry([]string{})
	search_entry.Wrapping = fyne.TextTruncate
	search_entry.OnCursorChanged = func() {
		split := strings.Split(search_entry.Text, "   ")
		if len(split) > 3 {
			if split[3] != Market.Viewing {
				Market.Viewing = split[3]
				list, addr := CheckNFAListingType(split[3])
				dest_addr = addr
				switch list {
				case 1:
					Market.Tab = "Auction"
					Market.Market_button.Text = "Bid"
					Market.Entry.SetText("0.0")
					Market.Entry.Enable()
					ResetAuctionInfo()
					AuctionInfo()
					clearNfaImages()
					go GetNfaImages(split[3])
					go GetAuctionDetails(split[3])
				case 2:
					Market.Tab = "Buy"
					Market.Market_button.Text = "Buy"
					Market.Entry.SetText("0.0")
					Market.Entry.Disable()
					ResetBuyInfo()
					BuyNowInfo()
					clearNfaImages()
					go GetNfaImages(split[3])
					go GetBuyNowDetails(split[3])
				default:
					Market.Tab = "Buy"
					Market.Entry.SetText("0.0")
					Market.Entry.Disable()
					ResetNotListedInfo()
					NotListedInfo()
					clearNfaImages()
					go GetNfaImages(split[3])
					go GetUnlistedDetails(split[3])
				}
			}

			Market.Details_box.Objects[1] = loadingBar()
		}
	}

	search_by := widget.NewRadioGroup([]string{"Collection   ", "Name"}, nil)
	search_by.Horizontal = true
	search_by.SetSelected("Collection   ")
	search_by.Required = true

	search_button := widget.NewButtonWithIcon("", fyne.Theme.Icon(fyne.CurrentApp().Settings().Theme(), "search"), func() {
		if search_entry.Text != "" && rpc.Wallet.Connect {
			switch search_by.Selected {
			case "Collection   ":
				if results := SearchNFAsBy(0, search_entry.Text); results != nil {
					search_entry.SetOptions(results)
					search_entry.ShowCompletion()
				}
			case "Name":
				if results := SearchNFAsBy(1, search_entry.Text); results != nil {
					search_entry.SetOptions(results)
					search_entry.ShowCompletion()
				}
			}
		}
	})

	clear_button := widget.NewButtonWithIcon("", fyne.Theme.Icon(fyne.CurrentApp().Settings().Theme(), "searchReplace"), func() {
		search_entry.SetOptions([]string{" Collection,  Name,  Description,  SCID:"})
		search_entry.SetText("")
	})

	show_results := widget.NewButtonWithIcon("", fyne.Theme.Icon(fyne.CurrentApp().Settings().Theme(), "arrowDropDown"), func() {
		search_entry.ShowCompletion()
	})

	search_cont := container.NewBorder(container.NewCenter(search_by), nil, container.NewHBox(clear_button, show_results), search_button, search_entry)

	message_button := widget.NewButton("Message Owner", func() {
		if rpc.Wallet.IsConnected() && dest_addr != "" {
			SendMessageMenu(dest_addr, bundle.ResourceDReamsIconAltPng)
		}
	})

	return container.NewBorder(search_cont, message_button, nil, nil, container.NewHBox())
}

// NFA market icon image with frame
//   - Pass res for frame resource
func NfaIcon(res fyne.Resource) fyne.CanvasObject {
	Market.Icon.SetMinSize(fyne.NewSize(90, 90))
	border := container.NewBorder(layout.NewSpacer(), layout.NewSpacer(), layout.NewSpacer(), layout.NewSpacer(), &Market.Icon)

	frame := canvas.NewImageFromResource(res)
	frame.SetMinSize(fyne.NewSize(100, 100))

	return container.NewMax(border, frame)
}

// Badge for dReam Tools enabled assets
//   - Pass res for frame resource
func ToolsBadge(res fyne.Resource) fyne.CanvasObject {
	badge := *canvas.NewImageFromResource(bundle.ResourceDReamToolsPng)
	badge.SetMinSize(fyne.NewSize(90, 90))
	border := container.NewBorder(layout.NewSpacer(), layout.NewSpacer(), layout.NewSpacer(), layout.NewSpacer(), &badge)

	frame := canvas.NewImageFromResource(res)
	frame.SetMinSize(fyne.NewSize(100, 100))

	return container.NewMax(border, frame)
}

// NFA cover image for market display
func NfaImg(img canvas.Image) fyne.CanvasObject {
	Market.Cover.SetMinSize(fyne.NewSize(400, 600))

	return container.NewCenter(&img)
}

// Loading bar for NFA cover image
func loadingBar() fyne.CanvasObject {
	Market.Loading = widget.NewProgressBarInfinite()
	spacer := canvas.NewRectangle(color.RGBA{0, 0, 0, 0})
	spacer.SetMinSize(fyne.NewSize(0, 21))
	Market.Loading.Start()

	return container.NewVBox(
		layout.NewSpacer(),
		container.NewMax(spacer, Market.Loading, container.NewCenter(canvas.NewText("Loading...", bundle.TextColor))),
		layout.NewSpacer())
}

// Clears all market NFA images
func clearNfaImages() {
	Market.Details_box.Objects[0].(*fyne.Container).Objects[0].(*fyne.Container).Objects[1] = layout.NewSpacer()
	Market.Icon = *canvas.NewImageFromImage(nil)
	Market.Details_box.Objects[0].Refresh()

	Market.Cover = *canvas.NewImageFromImage(nil)
	Market.Details_box.Objects[1] = canvas.NewImageFromImage(nil)
	Market.Details_box.Objects[1].Refresh()
	Market.Details_box.Refresh()
}

// Set up market info objects
func NfaMarketInfo() fyne.CanvasObject {
	Market.Name = widget.NewEntry()
	Market.Type = widget.NewEntry()
	Market.Collection = widget.NewEntry()
	Market.Description = widget.NewMultiLineEntry()
	Market.Creator = widget.NewEntry()
	Market.Art_fee = widget.NewEntry()
	Market.Royalty = widget.NewEntry()
	Market.Start_price = widget.NewEntry()
	Market.Owner = widget.NewEntry()
	Market.Owner_update = widget.NewEntry()
	Market.Current_bid = widget.NewEntry()
	Market.Bid_price = widget.NewEntry()
	Market.Bid_count = widget.NewEntry()
	Market.End_time = widget.NewEntry()

	Market.Name.Disable()
	Market.Type.Disable()
	Market.Collection.Disable()
	Market.Description.Disable()
	Market.Creator.Disable()
	Market.Art_fee.Disable()
	Market.Royalty.Disable()
	Market.Start_price.Disable()
	Market.Owner.Disable()
	Market.Owner_update.Disable()
	Market.Current_bid.Disable()
	Market.Bid_price.Disable()
	Market.Bid_count.Disable()
	Market.End_time.Disable()

	Market.Icon.SetMinSize(fyne.NewSize(94, 94))
	Market.Cover.SetMinSize(fyne.NewSize(400, 600))

	return AuctionInfo()
}

// Container for auction info objects
func AuctionInfo() fyne.CanvasObject {
	auction_form := []*widget.FormItem{}
	auction_form = append(auction_form, widget.NewFormItem("Name", Market.Name))
	auction_form = append(auction_form, widget.NewFormItem("Asset Type", Market.Type))
	auction_form = append(auction_form, widget.NewFormItem("Collection", Market.Collection))
	auction_form = append(auction_form, widget.NewFormItem("Ends", Market.End_time))
	auction_form = append(auction_form, widget.NewFormItem("Bids", Market.Bid_count))
	auction_form = append(auction_form, widget.NewFormItem("Description", Market.Description))
	auction_form = append(auction_form, widget.NewFormItem("Creator", Market.Creator))
	auction_form = append(auction_form, widget.NewFormItem("Owner", Market.Owner))
	auction_form = append(auction_form, widget.NewFormItem("Artificer %", Market.Art_fee))
	auction_form = append(auction_form, widget.NewFormItem("Royalty %", Market.Royalty))

	auction_form = append(auction_form, widget.NewFormItem("Owner Update", Market.Owner_update))
	auction_form = append(auction_form, widget.NewFormItem("Start Price", Market.Start_price))

	auction_form = append(auction_form, widget.NewFormItem("Current Bid", Market.Current_bid))

	Market.Details_box = *container.NewAdaptiveGrid(2,
		container.NewVBox(container.NewHBox(NfaIcon(bundle.ResourceAvatarFramePng), layout.NewSpacer()), widget.NewForm(auction_form...)),
		NfaImg(Market.Cover))

	Market.Description.Wrapping = fyne.TextWrapWord
	Market.Details_box.Refresh()

	return &Market.Details_box
}

// Container for unlisted info objects
func NotListedInfo() fyne.CanvasObject {
	unlisted_form := []*widget.FormItem{}
	unlisted_form = append(unlisted_form, widget.NewFormItem("Name", Market.Name))
	unlisted_form = append(unlisted_form, widget.NewFormItem("Asset Type", Market.Type))
	unlisted_form = append(unlisted_form, widget.NewFormItem("Collection", Market.Collection))
	unlisted_form = append(unlisted_form, widget.NewFormItem("Description", Market.Description))
	unlisted_form = append(unlisted_form, widget.NewFormItem("Creator", Market.Creator))
	unlisted_form = append(unlisted_form, widget.NewFormItem("Owner", Market.Owner))
	unlisted_form = append(unlisted_form, widget.NewFormItem("Artificer %", Market.Art_fee))
	unlisted_form = append(unlisted_form, widget.NewFormItem("Royalty %", Market.Royalty))

	unlisted_form = append(unlisted_form, widget.NewFormItem("Owner Update", Market.Owner_update))

	Market.Details_box = *container.NewAdaptiveGrid(2,
		container.NewVBox(container.NewHBox(NfaIcon(bundle.ResourceAvatarFramePng), layout.NewSpacer()), widget.NewForm(unlisted_form...)),
		NfaImg(Market.Cover))

	Market.Description.Wrapping = fyne.TextWrapWord
	Market.Details_box.Refresh()

	return &Market.Details_box
}

// Set unlisted display content to default values
func ResetNotListedInfo() {
	Market.Bid_amt = 0
	clearNfaImages()
	Market.Name.SetText("Name:")
	Market.Type.SetText("Asset Type:")
	Market.Collection.SetText("Collection:")
	Market.Description.SetText("Description:")
	Market.Creator.SetText("Creator:")
	Market.Art_fee.SetText("Artificer:")
	Market.Royalty.SetText("Royalty:")
	Market.Owner.SetText("Owner:")
	Market.Owner_update.SetText("Owner can update:")
	Market.Details_box.Refresh()
}

// Refresh Market images
func RefreshNfaImages() {
	if Market.Cover.Resource != nil {
		Market.Details_box.Objects[1] = NfaImg(Market.Cover)
		if Market.Loading != nil {
			Market.Loading.Stop()
		}
	} else {
		Market.Details_box.Objects[1] = loadingBar()
	}

	if Market.Icon.Resource != nil {
		Market.Details_box.Objects[0].(*fyne.Container).Objects[0].(*fyne.Container).Objects[0] = NfaIcon(bundle.ResourceAvatarFramePng)
	}
	view := Market.Viewing_coll
	if view == "AZYPC" || view == "SIXPC" || view == "AZYPCB" || view == "SIXPCB" {
		Market.Details_box.Objects[0].(*fyne.Container).Objects[0].(*fyne.Container).Objects[1] = ToolsBadge(bundle.ResourceAvatarFramePng)
	} else {
		Market.Details_box.Objects[0].(*fyne.Container).Objects[0].(*fyne.Container).Objects[1] = layout.NewSpacer()
	}

	Market.Details_box.Refresh()
}

// Set auction display content to default values
func ResetAuctionInfo() {
	Market.Bid_amt = 0
	clearNfaImages()
	Market.Name.SetText("Name:")
	Market.Type.SetText("Asset Type:")
	Market.Collection.SetText("Collection:")
	Market.Description.SetText("Description:")
	Market.Creator.SetText("Creator:")
	Market.Art_fee.SetText("Artificer:")
	Market.Royalty.SetText("Royalty:")
	Market.Start_price.SetText("Start Price:")
	Market.Owner.SetText("Owner:")
	Market.Owner_update.SetText("Owner can update:")
	Market.Current_bid.SetText("Current Bid:")
	Market.Bid_price.SetText("Minimum Bid:")
	Market.Bid_count.SetText("Bids:")
	Market.End_time.SetText("Ends At:")
	Market.Details_box.Refresh()
}

// Container for buy now info objects
func BuyNowInfo() fyne.CanvasObject {
	buy_form := []*widget.FormItem{}
	buy_form = append(buy_form, widget.NewFormItem("Name", Market.Name))
	buy_form = append(buy_form, widget.NewFormItem("Asset Type", Market.Type))
	buy_form = append(buy_form, widget.NewFormItem("Collection", Market.Collection))
	buy_form = append(buy_form, widget.NewFormItem("Ends", Market.End_time))
	buy_form = append(buy_form, widget.NewFormItem("Description", Market.Description))
	buy_form = append(buy_form, widget.NewFormItem("Creator", Market.Creator))
	buy_form = append(buy_form, widget.NewFormItem("Owner", Market.Owner))
	buy_form = append(buy_form, widget.NewFormItem("Artificer %", Market.Art_fee))
	buy_form = append(buy_form, widget.NewFormItem("Royalty %", Market.Royalty))

	buy_form = append(buy_form, widget.NewFormItem("Owner Update", Market.Owner_update))
	buy_form = append(buy_form, widget.NewFormItem("Price", Market.Start_price))

	Market.Details_box = *container.NewAdaptiveGrid(2,
		container.NewVBox(container.NewHBox(NfaIcon(bundle.ResourceAvatarFramePng), layout.NewSpacer()), widget.NewForm(buy_form...)),
		NfaImg(Market.Cover))

	Market.Description.Wrapping = fyne.TextWrapWord
	Market.Details_box.Refresh()

	return &Market.Details_box
}

// Set buy now display content to default values
func ResetBuyInfo() {
	Market.Buy_amt = 0
	clearNfaImages()
	Market.Name.SetText("Name:")
	Market.Type.SetText("Asset Type:")
	Market.Collection.SetText("Collection:")
	Market.Description.SetText("Description:")
	Market.Creator.SetText("Creator:")
	Market.Art_fee.SetText("Artificer:")
	Market.Royalty.SetText("Royalty:")
	Market.Start_price.SetText("Buy now for:")
	Market.Owner.SetText("Owner:")
	Market.Owner_update.SetText("Owner can update:")
	Market.End_time.SetText("Ends At:")
	Market.Details_box.Refresh()
}

// Switch triggered when market tab changes
func MarketTab(ti *container.TabItem) {
	switch ti.Text {
	case "Auctions":
		go FindNfaListings(nil)
		Market.Tab = "Auction"
		Market.Auction_list.UnselectAll()
		Market.My_listings.UnselectAll()
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
		Market.Entry.SetText("0.0")
		Market.Entry.Disable()
		Market.Market_button.Refresh()
		ResetBuyInfo()
		BuyNowInfo()
	case "My Listings":
		go FindNfaListings(nil)
		Market.Tab = "Buy"
		Market.My_listings.UnselectAll()
		Market.Viewing = ""
		Market.Viewing_coll = ""
		Market.Entry.SetText("0.0")
		Market.Entry.Disable()
		ResetBuyInfo()
		BuyNowInfo()
	case "Search":
		Market.Tab = "Buy"
		Market.Auction_list.UnselectAll()
		Market.Buy_list.UnselectAll()
		Market.Viewing = ""
		Market.Viewing_coll = ""
		Market.Entry.SetText("0.0")
		Market.Entry.Disable()
		ResetBuyInfo()
		BuyNowInfo()
	}

	Market.Close_button.Hide()
	Market.Cancel_button.Hide()
	Market.Market_button.Hide()
}

// NFA market layout
func PlaceMarket() *container.Split {
	details := container.NewMax(NfaMarketInfo())

	tabs := container.NewAppTabs(
		container.NewTabItem("Auctions", AuctionListings()),
		container.NewTabItem("Buy Now", BuyNowListings()),
		container.NewTabItem("My Listings", MyNFAListings()),
		container.NewTabItem("Search", SearchNFAs()))

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

	details_box := container.NewVBox(layout.NewSpacer(), details, layout.NewSpacer())

	menu_top := container.NewHSplit(details_box, max)
	menu_top.SetOffset(0.66)

	Market.Market_button = widget.NewButton("Bid", func() {
		scid := Market.Viewing
		if len(scid) == 64 {
			text := Market.Market_button.Text
			Market.Market_button.Hide()
			if text == "Bid" {
				amt := rpc.ToAtomic(Market.Entry.Text, 5)
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

// Returns search filter with all enabled NFAs
func ReturnEnabledNFAs(assets map[string]bool) (filters []string) {
	for name, enabled := range assets {
		if enabled {
			if isDreamsNfaName(name) {
				filters = append(filters, fmt.Sprintf(`330 STORE("nameHdr", "%s`, name))
			} else if isDreamsNfaCollection(name) {
				filters = append(filters, fmt.Sprintf(`450 STORE("collection", "%s`, name))
			}
		}
	}

	return
}

func ReturnAssetCount() (count int) {
	count = Control.NFA_count + Control.G45_count - 10
	if count < 2 {
		count = 2
	}

	return
}

// Options for enabling NFA collection
func enableNFAOpts(asset assetCount) (opts *widget.RadioGroup) {
	onChanged := func(s string) {
		if s == "Yes" {
			Control.Lock()
			Control.Enabled_assets[asset.name] = true
			Control.NFA_count += asset.count
			Control.Unlock()
			return
		}

		Control.Lock()
		defer Control.Unlock()
		Control.Enabled_assets[asset.name] = false
		if Control.NFA_count >= asset.count {
			Control.NFA_count -= asset.count
		}
	}

	if !Control.once {
		opts = widget.NewRadioGroup([]string{"Yes", "No"}, nil)
		opts.Required = true
		opts.Horizontal = true
		if Control.Enabled_assets[asset.name] {
			opts.OnChanged = onChanged
			opts.SetSelected("Yes")
		} else {
			opts.SetSelected("No")
			opts.OnChanged = onChanged
		}

		return
	}

	opts = widget.NewRadioGroup([]string{"Yes", "No"}, nil)
	opts.Required = true
	opts.Horizontal = true
	if Control.Enabled_assets[asset.name] {
		opts.SetSelected("Yes")
	} else {
		opts.SetSelected("No")
	}
	opts.OnChanged = onChanged

	return
}

// Options for enabling G45 collection
func enableG45Opts(asset assetCount) (opts *widget.RadioGroup) {
	onChanged := func(s string) {
		if s == "Yes" {
			Control.Lock()
			Control.Enabled_assets[asset.name] = true
			Control.G45_count += asset.count
			Control.Unlock()
			return
		}

		Control.Lock()
		defer Control.Unlock()
		Control.Enabled_assets[asset.name] = false
		if Control.G45_count >= asset.count {
			Control.G45_count -= asset.count
		}
	}

	if !Control.once {
		opts = widget.NewRadioGroup([]string{"Yes", "No"}, nil)
		opts.Required = true
		opts.Horizontal = true
		if Control.Enabled_assets[asset.name] {
			opts.OnChanged = onChanged
			opts.SetSelected("Yes")
		} else {
			opts.SetSelected("No")
			opts.OnChanged = onChanged
		}

		return
	}

	opts = widget.NewRadioGroup([]string{"Yes", "No"}, nil)
	opts.Required = true
	opts.Horizontal = true
	if Control.Enabled_assets[asset.name] {
		opts.SetSelected("Yes")
	} else {
		opts.SetSelected("No")
	}
	opts.OnChanged = onChanged

	return
}

// Enable asset collection objects
// intro used to set label if initial boot screen
func EnabledCollections(intro bool) (obj fyne.CanvasObject) {
	collection_form := []*widget.FormItem{}
	enable_all := widget.NewButton("Enable All", func() {
		for _, item := range collection_form {
			item.Widget.(*widget.RadioGroup).SetSelected("Yes")

		}
	})

	disable_all := widget.NewButton("Disable All", func() {
		for _, item := range collection_form {
			item.Widget.(*widget.RadioGroup).SetSelected("No")
		}
	})

	for _, asset := range dReamsNFAs {
		collection_form = append(collection_form, widget.NewFormItem(asset.name, enableNFAOpts(asset)))
	}

	for _, asset := range dReamsG45s {
		collection_form = append(collection_form, widget.NewFormItem(asset.name, enableG45Opts(asset)))
	}

	Control.once = true
	if Control.NFA_count < 3 {
		Control.NFA_count = 3
	}

	label := canvas.NewText("You will need to delete Gnomon DB and resync for changes to take effect ", bundle.TextColor)
	label.Alignment = fyne.TextAlignCenter
	if intro {
		label.Text = "Enable Asset Collections"
	}

	return container.NewBorder(
		nil,
		container.NewBorder(nil, nil, enable_all, disable_all, label),
		nil,
		nil,
		container.NewVScroll(container.NewCenter(widget.NewForm(collection_form...))))

}

// Returns string with all enabled asset names formatted for a label
func returnEnabledNames(assets map[string]bool) (text string) {
	var names []string
	for name, enabled := range assets {
		if enabled {
			if isDreamsNfaName(name) {
				names = append(names, name)
			} else if isDreamsNfaCollection(name) {
				names = append(names, name)
			}
		}
	}

	for name, enabled := range assets {
		if enabled && IsDreamsG45(name) {
			names = append(names, name)
		}
	}

	sort.Strings(names)

	for _, n := range names {
		text = text + n + "\n\n"
	}

	return
}

// Owned asset tab layout
//   - tag for log print
//   - assets is array of widgets used for asset selections
//   - menu_icon resources for side menus
//   - w for main window dialog
func PlaceAssets(tag string, assets []fyne.Widget, menu_icon fyne.Resource, w fyne.Window) *container.Split {
	items_box := container.NewAdaptiveGrid(2)

	asset_selects := container.NewVBox()
	for _, sel := range assets {
		asset_selects.Add(sel)
	}
	asset_selects.Add(layout.NewSpacer())

	cont := container.NewHScroll(asset_selects)
	cont.SetMinSize(fyne.NewSize(290, 35.1875))

	items_box.Add(cont)

	items_box.Add(container.NewAdaptiveGrid(1, AssetStats()))

	player_input := container.NewVBox(items_box, layout.NewSpacer())

	enable_opts := EnabledCollections(false)

	tabs := container.NewAppTabs(
		container.NewTabItem("Owned", AssetList()))

	if len(asset_selects.Objects) > 1 {
		tabs.Append(container.NewTabItem("Enabled", enable_opts))
	}

	tabs.OnSelected = func(ti *container.TabItem) {
		if ti.Text == "Enabled" {
			if rpc.Daemon.IsConnected() {
				dialog.NewInformation("Assets", "Shut down Gnomon to make changes to asset index", w).Show()
				tabs.Selected().Content = container.NewVScroll(container.NewVBox(dwidget.NewCenterLabel("Currently Enabled:"), dwidget.NewCenterLabel(returnEnabledNames(Control.Enabled_assets))))

				return
			}
			tabs.Selected().Content = enable_opts
		}
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
	menu_bottom := container.NewAdaptiveGrid(1, IndexEntry(menu_icon, w))

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
	Assets.Dero_price = canvas.NewText(" Dero Price: $", bundle.TextColor)

	Assets.Gnomes_sync.TextSize = 18
	Assets.Gnomes_height.TextSize = 18
	Assets.Daem_height.TextSize = 18
	Assets.Wall_height.TextSize = 18
	Assets.Dero_price.TextSize = 18

	Assets.Gnomes_sync.Alignment = fyne.TextAlignCenter
	Assets.Gnomes_height.Alignment = fyne.TextAlignCenter
	Assets.Daem_height.Alignment = fyne.TextAlignCenter
	Assets.Wall_height.Alignment = fyne.TextAlignCenter
	Assets.Dero_price.Alignment = fyne.TextAlignCenter

	return container.NewVBox(
		Assets.Gnomes_sync,
		Assets.Gnomes_height,
		Assets.Daem_height,
		Assets.Wall_height,
		Assets.Dero_price)
}

// Icon image for Holdero tables and asset viewing
//   - Pass res as frame resource
func IconImg(res fyne.Resource) *fyne.Container {
	Assets.Icon.SetMinSize(fyne.NewSize(100, 100))
	Assets.Icon.Resize(fyne.NewSize(94, 94))
	Assets.Icon.Move(fyne.NewPos(7, 3))

	frame := canvas.NewImageFromResource(res)
	frame.Resize(fyne.NewSize(100, 100))
	frame.Move(fyne.NewPos(4, 0))

	return container.NewWithoutLayout(&Assets.Icon, frame)
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
func RunNFAMarket(tag string, quit, done chan struct{}, connect_box *dwidget.DeroRpcEntries) {
	logger.Printf("[%s] %s %s %s\n", tag, rpc.DREAMSv, runtime.GOOS, runtime.GOARCH)
	time.Sleep(6 * time.Second)
	ticker := time.NewTicker(3 * time.Second)
	offset := 0

	for {
		select {
		case <-ticker.C: // do on interval
			rpc.Ping()
			rpc.EchoWallet(tag)

			// Get all NFA listings
			if Gnomes.IsRunning() && offset%2 == 0 {
				FindNfaListings(nil)
				if offset > 19 {
					offset = 0
				}
			}

			rpc.Wallet.GetBalance()

			// Refresh Dero balance and Gnomon endpoint
			connect_box.RefreshBalance()
			if !rpc.Startup {
				GnomonEndPoint()
			}

			// If connected daemon connected start looking for Gnomon sync with daemon
			if rpc.Daemon.IsConnected() && Gnomes.IsRunning() {
				connect_box.Disconnect.SetChecked(true)
				// Get indexed SCID count
				Gnomes.IndexContains()
				if Gnomes.HasIndex(1) {
					Gnomes.Checked(true)
				}

				if Gnomes.Indexer.LastIndexedHeight >= Gnomes.Indexer.ChainHeight-3 {
					Gnomes.Synced(true)
				} else {
					Gnomes.Synced(false)
					Gnomes.Checked(false)
				}

				// Enable index controls and check if wallet is connected
				go DisableIndexControls(false)
				if rpc.Wallet.IsConnected() {
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

			if rpc.Daemon.IsConnected() {
				rpc.Startup = false
			}

			offset++

		case <-quit: // exit
			logger.Printf("[%s] Closing...\n", tag)
			if Gnomes.Icon_ind != nil {
				Gnomes.Icon_ind.Stop()
			}
			ticker.Stop()
			time.Sleep(time.Second)
			done <- struct{}{}
			return
		}
	}
}

// Get search filters from on chain store
func FetchFilters(check string) (filter []string) {
	if stored, ok := rpc.FindStringKey(rpc.RatingSCID, check, rpc.Daemon.Rpc).(string); ok {
		if h, err := hex.DecodeString(stored); err == nil {
			if err = json.Unmarshal(h, &filter); err != nil {
				logger.Errorln("[FetchFilters]", check, err)
			}
		}
	} else {
		logger.Errorln("[FetchFilters] Could not get", check)
	}

	return
}
