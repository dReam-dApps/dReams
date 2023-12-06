package menu

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image/color"
	"strings"
	"time"

	dreams "github.com/dReam-dApps/dReams"
	"github.com/dReam-dApps/dReams/bundle"
	"github.com/dReam-dApps/dReams/dwidget"
	"github.com/dReam-dApps/dReams/rpc"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/validation"
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

var Market marketObjects

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
		listing, _, _ := checkNFAAuctionListing(scid)
		split := strings.Split(listing, "   ")
		if len(split) == 4 {
			coll = split[0]
			name = split[1]
		}
		text = fmt.Sprintf("Bidding on SCID:\n\n%s\n\nAsset: %s\n\nCollection: %s\n\nBid amount: %s Dero\n\nConfirm bid", scid, name, coll, amt_str)
	case 1:
		listing, _, _ := checkNFABuyListing(scid)
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

	confirm := widget.NewButtonWithIcon("Confirm", dreams.FyneIcon("confirm"), func() {
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
	confirm.Importance = widget.HighImportance

	cancel := widget.NewButtonWithIcon("Cancel", dreams.FyneIcon("cancel"), func() {
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

	return container.NewStack(content)
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
		listing, _, _ := checkNFAAuctionListing(scid)
		split := strings.Split(listing, "   ")
		if len(split) == 4 {
			coll = split[0]
			name = split[1]
		}
	case 2:
		listing, _, _ := checkNFABuyListing(scid)
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

	confirm := widget.NewButtonWithIcon("Confirm", dreams.FyneIcon("confirm"), func() {
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
	confirm.Importance = widget.HighImportance

	cancel := widget.NewButtonWithIcon("Cancel", dreams.FyneIcon("cancel"), func() {
		obj.Trailing.(*fyne.Container).Objects[1] = reset
		obj.Trailing.(*fyne.Container).Objects[1].Refresh()
		Market.Confirming = false
	})

	left := container.NewVBox(confirm)
	right := container.NewVBox(cancel)
	buttons := container.NewAdaptiveGrid(2, left, right)

	content := container.NewVBox(layout.NewSpacer(), label, layout.NewSpacer(), buttons)

	return container.NewStack(content)
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
				clearNFAImages()
				Market.Viewing = split[3]
				go GetNFAImages(split[3])
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
				clearNFAImages()
				Market.Viewing = split[3]
				go GetNFAImages(split[3])
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
				clearNFAImages()
				Market.Viewing = split[3]
				go GetNFAImages(split[3])
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
	search_entry.OnChanged = func(s string) {
		split := strings.Split(s, "   ")
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
					clearNFAImages()
					go GetNFAImages(split[3])
					go GetAuctionDetails(split[3])
				case 2:
					Market.Tab = "Buy"
					Market.Market_button.Text = "Buy"
					Market.Entry.SetText("0.0")
					Market.Entry.Disable()
					ResetBuyInfo()
					BuyNowInfo()
					clearNFAImages()
					go GetNFAImages(split[3])
					go GetBuyNowDetails(split[3])
				default:
					Market.Tab = "Buy"
					Market.Entry.SetText("0.0")
					Market.Entry.Disable()
					ResetNotListedInfo()
					NotListedInfo()
					clearNFAImages()
					go GetNFAImages(split[3])
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
		if dest_addr != "" {
			SendMessageMenu(dest_addr, bundle.ResourceDReamsIconAltPng)
		}
	})

	return container.NewBorder(search_cont, message_button, nil, nil, container.NewHBox())
}

// NFA market icon image with frame
func NFAIcon() fyne.CanvasObject {
	Market.Icon.SetMinSize(fyne.NewSize(90, 90))
	border := container.NewBorder(layout.NewSpacer(), layout.NewSpacer(), layout.NewSpacer(), layout.NewSpacer(), &Market.Icon)

	frame := canvas.NewImageFromResource(bundle.ResourceAvatarFramePng)
	frame.SetMinSize(fyne.NewSize(100, 100))

	return container.NewStack(border, frame)
}

var toolsBadge = *canvas.NewImageFromResource(bundle.ResourceDReamToolsPng)

// Badge for dReam Tools enabled assets
func ToolsBadge() fyne.CanvasObject {
	toolsBadge.SetMinSize(fyne.NewSize(90, 90))
	border := container.NewBorder(layout.NewSpacer(), layout.NewSpacer(), layout.NewSpacer(), layout.NewSpacer(), &toolsBadge)

	frame := canvas.NewImageFromResource(bundle.ResourceAvatarFramePng)
	frame.SetMinSize(fyne.NewSize(100, 100))

	return container.NewStack(border, frame)
}

// NFA cover image for market display
func NFACoverImg() fyne.CanvasObject {
	Market.Cover.SetMinSize(fyne.NewSize(400, 600))

	return container.NewCenter(&Market.Cover)
}

// Loading bar for NFA cover image
func loadingBar() fyne.CanvasObject {
	Market.Loading = widget.NewProgressBarInfinite()
	spacer := canvas.NewRectangle(color.RGBA{0, 0, 0, 0})
	spacer.SetMinSize(fyne.NewSize(0, 21))
	Market.Loading.Start()

	return container.NewVBox(
		layout.NewSpacer(),
		container.NewStack(spacer, Market.Loading, container.NewCenter(canvas.NewText("Loading...", bundle.TextColor))),
		layout.NewSpacer())
}

// Clears all market NFA images
func clearNFAImages() {
	Market.Details_box.Objects[0].(*fyne.Container).Objects[0].(*fyne.Container).Objects[1] = layout.NewSpacer()
	Market.Icon = *canvas.NewImageFromImage(nil)

	Market.Cover = *canvas.NewImageFromImage(nil)
	Market.Details_box.Objects[1] = canvas.NewImageFromImage(nil)
}

// Set up market info objects
func NFAMarketInfo() fyne.CanvasObject {
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

	Market.Description.Wrapping = fyne.TextWrapWord

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

	spacer := canvas.NewRectangle(color.Transparent)
	spacer.SetMinSize(fyne.NewSize(330, 0))
	auction_form = append(auction_form, widget.NewFormItem("", container.NewStack(spacer)))

	form := widget.NewForm(auction_form...)

	Market.Details_box = *container.NewAdaptiveGrid(2,
		container.NewVBox(
			container.NewHBox(NFAIcon(), layout.NewSpacer()),
			form),
		NFACoverImg())

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

	spacer := canvas.NewRectangle(color.Transparent)
	spacer.SetMinSize(fyne.NewSize(330, 0))
	unlisted_form = append(unlisted_form, widget.NewFormItem("", container.NewStack(spacer)))

	form := widget.NewForm(unlisted_form...)

	Market.Details_box = *container.NewAdaptiveGrid(2,
		container.NewVBox(
			container.NewHBox(NFAIcon(), layout.NewSpacer()),
			form),
		NFACoverImg())

	return &Market.Details_box
}

// Set unlisted display content to default values
func ResetNotListedInfo() {
	Market.Bid_amt = 0
	clearNFAImages()
	Market.Name.SetText("Name:")
	Market.Type.SetText("Asset Type:")
	Market.Collection.SetText("Collection:")
	Market.Description.SetText("Description:")
	Market.Creator.SetText("Creator:")
	Market.Art_fee.SetText("Artificer:")
	Market.Royalty.SetText("Royalty:")
	Market.Owner.SetText("Owner:")
	Market.Owner_update.SetText("Owner can update:")
}

// Refresh Market images
func RefreshNFAImages() {
	if Market.Cover.Resource != nil {
		Market.Details_box.Objects[1] = NFACoverImg()
		if Market.Loading != nil {
			Market.Loading.Stop()
		}
	} else {
		Market.Details_box.Objects[1] = loadingBar()
	}

	if Market.Icon.Resource != nil {
		Market.Details_box.Objects[0].(*fyne.Container).Objects[0].(*fyne.Container).Objects[0] = NFAIcon()
	}

	view := Market.Viewing_coll
	if view == "AZYPC" || view == "SIXPC" || view == "AZYPCB" || view == "SIXPCB" {
		Market.Details_box.Objects[0].(*fyne.Container).Objects[0].(*fyne.Container).Objects[1] = ToolsBadge()
	} else {
		Market.Details_box.Objects[0].(*fyne.Container).Objects[0].(*fyne.Container).Objects[1] = layout.NewSpacer()
	}

	Market.Details_box.Refresh()
}

// Set auction display content to default values
func ResetAuctionInfo() {
	Market.Bid_amt = 0
	clearNFAImages()
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

	spacer := canvas.NewRectangle(color.Transparent)
	spacer.SetMinSize(fyne.NewSize(330, 0))
	buy_form = append(buy_form, widget.NewFormItem("", container.NewStack(spacer)))

	form := widget.NewForm(buy_form...)

	Market.Details_box = *container.NewAdaptiveGrid(2,
		container.NewVBox(
			container.NewHBox(NFAIcon(), layout.NewSpacer()),
			form),
		NFACoverImg())

	return &Market.Details_box
}

// Set buy now display content to default values
func ResetBuyInfo() {
	Market.Buy_amt = 0
	clearNFAImages()
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
}

// Switch triggered when market tab changes
func MarketTab(ti *container.TabItem) {
	switch ti.Text {
	case "Auctions":
		go FindNFAListings(nil)
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
		go FindNFAListings(nil)
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
		go FindNFAListings(nil)
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
	details := container.NewStack(NFAMarketInfo())

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

	// TDOD revisit scroll here when cover image is showing
	min_size := bundle.Alpha120
	min_size.SetMinSize(fyne.NewSize(420, 0))

	max := container.NewStack(min_size, tabs, scroll_cont)

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
				menu_top.Trailing.(*fyne.Container).Objects[1] = BidBuyConfirm(scid, amt, 0, menu_top, container.NewStack(tabs, scroll_cont))
				menu_top.Trailing.(*fyne.Container).Objects[1].Refresh()
			} else if text == "Buy" {
				menu_top.Trailing.(*fyne.Container).Objects[1] = BidBuyConfirm(scid, Market.Buy_amt, 1, menu_top, container.NewStack(tabs, scroll_cont))
				menu_top.Trailing.(*fyne.Container).Objects[1].Refresh()
			}
		}
	})

	Market.Market_button.Hide()

	Market.Cancel_button = widget.NewButton("Cancel", func() {
		if len(Market.Viewing) == 64 {
			Market.Cancel_button.Hide()
			menu_top.Trailing.(*fyne.Container).Objects[1] = ConfirmCancelClose(Market.Viewing, 1, menu_top, container.NewStack(tabs, scroll_cont))
			menu_top.Trailing.(*fyne.Container).Objects[1].Refresh()
		}
	})

	Market.Close_button = widget.NewButton("Close", func() {
		if len(Market.Viewing) == 64 {
			Market.Close_button.Hide()
			menu_top.Trailing.(*fyne.Container).Objects[1] = ConfirmCancelClose(Market.Viewing, 0, menu_top, container.NewStack(tabs, scroll_cont))
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

// Full routine for NFA market and scanning wallet for NFAs, can use PlaceAssets() and PlaceMarket() layouts
//   - tag for log print
//   - quit for exit chan
//   - connected box for DeroRpcEntries
func RunNFAMarket(tag string, quit, done chan struct{}, connect_box *dwidget.DeroRpcEntries) {
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
				FindNFAListings(nil)
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
							go RefreshNFAImages()
						} else {
							GetAuctionDetails(Market.Viewing)
							go RefreshNFAImages()
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
				Info.RefreshIndexed()
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
