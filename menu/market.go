package menu

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image/color"
	"sort"
	"strconv"
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

// Trim input string to specified len
func TrimStringLen(str string, l int) string {
	if len(str) > l {
		return str[0:l]
	}

	return str
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

// Get NFA image files
func GetNFAImages(scid string) {
	if Gnomes.IsReady() && len(scid) == 64 {
		name, _ := Gnomes.GetSCIDValuesByKey(scid, "nameHdr")
		icon, _ := Gnomes.GetSCIDValuesByKey(scid, "iconURLHdr")
		cover, _ := Gnomes.GetSCIDValuesByKey(scid, "coverURL")
		if icon != nil {
			Market.Icon, _ = dreams.DownloadCanvas(icon[0], name[0])
			Market.Cover, _ = dreams.DownloadCanvas(cover[0], name[0]+"-cover")
		} else {
			Market.Icon = *canvas.NewImageFromImage(nil)
			Market.Cover = *canvas.NewImageFromImage(nil)
		}
	}
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
	if view == "AZY-Playing card decks" || view == "SIXPC" || view == "AZY-Playing card backs" || view == "SIXPCB" {
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
					Assets.Claim.Show()
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
					Assets.Button.List.Hide()
					Assets.Button.Send.Hide()
					Assets.Claim.Hide()
				}

				// Check wallet for all owned NFAs and refresh content
				go CheckAllNFAs(false, nil)
				Info.RefreshIndexed()
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

// Check NFA listing type and return owner address
//   - Auction returns 1
//   - Sale returns 2
func CheckNFAListingType(scid string) (list int, addr string) {
	if Gnomes.IsReady() {
		if owner, _ := Gnomes.GetSCIDValuesByKey(scid, "owner"); owner != nil {
			if listType, _ := Gnomes.GetSCIDValuesByKey(scid, "listType"); listType != nil {
				addr = owner[0]
				switch listType[0] {
				case "auction":
					list = 1
				case "sale":
					list = 2
				default:

				}
			}
		}
	}
	return
}

// Check if NFA SCID is listed for auction
//   - Market.DreamsFilter false for all NFA listings
func checkNFAAuctionListing(scid string) (asset string, owned, expired bool) {
	if Gnomes.IsReady() {
		if creator, _ := Gnomes.GetSCIDValuesByKey(scid, "creatorAddr"); creator != nil {
			listType, _ := Gnomes.GetSCIDValuesByKey(scid, "listType")
			header, _ := Gnomes.GetSCIDValuesByKey(scid, "nameHdr")
			coll, _ := Gnomes.GetSCIDValuesByKey(scid, "collection")
			desc, _ := Gnomes.GetSCIDValuesByKey(scid, "descrHdr")
			if listType != nil && header != nil && coll != nil && desc != nil {
				if Market.DreamsFilter {
					if IsDreamsNFACollection(coll[0]) {
						if listType[0] == "auction" {
							desc_check := TrimStringLen(desc[0], 66)
							asset = coll[0] + "   " + header[0] + "   " + desc_check + "   " + scid
							if owner, _ := Gnomes.GetSCIDValuesByKey(scid, "owner"); owner != nil {
								if owner[0] == rpc.Wallet.Address {
									owned = true
								}
							}

							if _, endTime := Gnomes.GetSCIDValuesByKey(scid, "endBlockTime"); endTime != nil {
								now := uint64(time.Now().Unix())
								if now > endTime[0] && endTime[0] > 0 {
									expired = true
								}
							}
						}
					}
				} else {
					var hidden bool
					for _, addr := range Market.Filters {
						if creator[0] == addr {
							hidden = true
						}
					}

					if !hidden {
						if listType[0] == "auction" {
							desc_check := TrimStringLen(desc[0], 66)
							asset = coll[0] + "   " + header[0] + "   " + desc_check + "   " + scid
							if owner, _ := Gnomes.GetSCIDValuesByKey(scid, "owner"); owner != nil {
								if owner[0] == rpc.Wallet.Address {
									owned = true
								}
							}

							if _, endTime := Gnomes.GetSCIDValuesByKey(scid, "endBlockTime"); endTime != nil {
								now := uint64(time.Now().Unix())
								if now > endTime[0] && endTime[0] > 0 {
									expired = true
								}
							}
						}
					}
				}
			}
		}
	}

	return
}

// Check if NFA SCID is listed as buy now
//   - Market.DreamsFilter false for all NFA listings
func checkNFABuyListing(scid string) (asset string, owned, expired bool) {
	if Gnomes.IsReady() {
		if creator, _ := Gnomes.GetSCIDValuesByKey(scid, "creatorAddr"); creator != nil {
			listType, _ := Gnomes.GetSCIDValuesByKey(scid, "listType")
			header, _ := Gnomes.GetSCIDValuesByKey(scid, "nameHdr")
			coll, _ := Gnomes.GetSCIDValuesByKey(scid, "collection")
			desc, _ := Gnomes.GetSCIDValuesByKey(scid, "descrHdr")
			if listType != nil && header != nil && coll != nil && desc != nil {
				if Market.DreamsFilter {
					if IsDreamsNFACollection(coll[0]) {
						if listType[0] == "sale" {
							desc_check := TrimStringLen(desc[0], 66)
							asset = coll[0] + "   " + header[0] + "   " + desc_check + "   " + scid
							if owner, _ := Gnomes.GetSCIDValuesByKey(scid, "owner"); owner != nil {
								if owner[0] == rpc.Wallet.Address {
									owned = true
								}
							}

							if _, endTime := Gnomes.GetSCIDValuesByKey(scid, "endBlockTime"); endTime != nil {
								now := uint64(time.Now().Unix())
								if now > endTime[0] && endTime[0] > 0 {
									expired = true
								}
							}
						}
					}
				} else {
					var hidden bool
					for _, addr := range Market.Filters {
						if creator[0] == addr {
							hidden = true
						}
					}

					if !hidden {
						if listType[0] == "sale" {
							desc_check := TrimStringLen(desc[0], 66)
							asset = coll[0] + "   " + header[0] + "   " + desc_check + "   " + scid
							if owner, _ := Gnomes.GetSCIDValuesByKey(scid, "owner"); owner != nil {
								if owner[0] == rpc.Wallet.Address {
									owned = true
								}
							}

							if _, endTime := Gnomes.GetSCIDValuesByKey(scid, "endBlockTime"); endTime != nil {
								now := uint64(time.Now().Unix())
								if now > endTime[0] && endTime[0] > 0 {
									expired = true
								}
							}
						}
					}
				}
			}
		}
	}

	return
}

// Search NFAs in index by name or collection
func SearchNFAsBy(by int, prefix string) (results []string) {
	if Gnomes.IsReady() {
		results = []string{" Collection,  Name,  Description,  SCID:"}
		assets := Gnomes.GetAllOwnersAndSCIDs()

		for sc := range assets {
			if !Gnomes.IsReady() {
				return
			}

			if file, _ := Gnomes.GetSCIDValuesByKey(sc, "fileURL"); file != nil {
				if ValidNFA(file[0]) {
					if name, _ := Gnomes.GetSCIDValuesByKey(sc, "nameHdr"); name != nil {
						coll, _ := Gnomes.GetSCIDValuesByKey(sc, "collection")
						desc, _ := Gnomes.GetSCIDValuesByKey(sc, "descrHdr")
						if coll != nil && desc != nil {
							switch by {
							case 0:
								if strings.HasPrefix(coll[0], prefix) {
									desc_check := TrimStringLen(desc[0], 66)
									asset := coll[0] + "   " + name[0] + "   " + desc_check + "   " + sc
									results = append(results, asset)
								}
							case 1:
								if strings.HasPrefix(name[0], prefix) {
									desc_check := TrimStringLen(desc[0], 66)
									asset := coll[0] + "   " + name[0] + "   " + desc_check + "   " + sc
									results = append(results, asset)
								}
							}
						}
					}
				}
			}
		}

		sort.Strings(results)
	}

	return
}

// Create auction tab info for current asset
func GetAuctionDetails(scid string) {
	if Gnomes.IsReady() && len(scid) == 64 {
		name, _ := Gnomes.GetSCIDValuesByKey(scid, "nameHdr")
		collection, _ := Gnomes.GetSCIDValuesByKey(scid, "collection")
		description, _ := Gnomes.GetSCIDValuesByKey(scid, "descrHdr")
		creator, _ := Gnomes.GetSCIDValuesByKey(scid, "creatorAddr")
		owner, _ := Gnomes.GetSCIDValuesByKey(scid, "owner")
		typeHdr, _ := Gnomes.GetSCIDValuesByKey(scid, "typeHdr")
		_, owner_update := Gnomes.GetSCIDValuesByKey(scid, "ownerCanUpdate")
		_, start := Gnomes.GetSCIDValuesByKey(scid, "startPrice")
		_, current := Gnomes.GetSCIDValuesByKey(scid, "currBidAmt")
		_, bid_price := Gnomes.GetSCIDValuesByKey(scid, "currBidPrice")
		_, royalty := Gnomes.GetSCIDValuesByKey(scid, "royalty")
		_, bids := Gnomes.GetSCIDValuesByKey(scid, "bidCount")
		_, endTime := Gnomes.GetSCIDValuesByKey(scid, "endBlockTime")
		_, startTime := Gnomes.GetSCIDValuesByKey(scid, "startBlockTime")
		_, artFee := Gnomes.GetSCIDValuesByKey(scid, "artificerFee")

		if name != nil && collection != nil && start != nil && royalty != nil && endTime != nil && artFee != nil && typeHdr != nil {
			go func() {
				Market.Viewing_coll = collection[0]

				Market.Name.SetText(name[0])

				Market.Type.SetText(AssetType(collection[0], typeHdr[0]))

				Market.Collection.SetText(collection[0])

				Market.Description.SetText(description[0])

				if Market.Creator.Text != creator[0] {
					Market.Creator.SetText(creator[0])
				}

				if Market.Owner.Text != owner[0] {
					Market.Owner.SetText(owner[0])
				}
				if owner_update[0] == 1 {
					Market.Owner_update.SetText("Yes")
				} else {
					Market.Owner_update.SetText("No")
				}

				Market.Art_fee.SetText(strconv.Itoa(int(artFee[0])) + "%")

				Market.Royalty.SetText(strconv.Itoa(int(royalty[0])) + "%")

				price := float64(start[0])
				str := fmt.Sprintf("%.5f", price/100000)
				Market.Start_price.SetText(str + " Dero")

				Market.Bid_count.SetText(strconv.Itoa(int(bids[0])))

				end, _ := rpc.MsToTime(strconv.Itoa(int(endTime[0]) * 1000))
				Market.End_time.SetText(end.String())

				if current != nil {
					value := float64(current[0])
					str := fmt.Sprintf("%.5f", value/100000)
					Market.Current_bid.SetText(str)
				} else {
					Market.Current_bid.SetText("")
				}

				if bid_price != nil {
					value := float64(bid_price[0])
					str := fmt.Sprintf("%.5f", value/100000)
					if bid_price[0] == 0 {
						Market.Bid_amt = start[0]
					} else {
						Market.Bid_amt = bid_price[0]
					}
					Market.Bid_price.SetText(str)
				} else {
					Market.Bid_amt = 0
					Market.Bid_price.SetText("")
				}

				if amt, err := strconv.ParseFloat(Market.Entry.Text, 64); err == nil {
					value := float64(Market.Bid_amt) / 100000
					if amt == 0 || amt < value {
						amt := fmt.Sprintf("%.5f", value)
						Market.Entry.SetText(amt)
					}
				}

				now := uint64(time.Now().Unix())
				if owner[0] == rpc.Wallet.Address {
					if now < startTime[0]+300 && startTime[0] > 0 && !Market.Confirming {
						Market.Cancel_button.Show()
					} else {
						Market.Cancel_button.Hide()
					}

					if now > endTime[0] && endTime[0] > 0 && !Market.Confirming {
						Market.Close_button.Show()
					} else {
						Market.Close_button.Hide()
					}
				} else {
					Market.Close_button.Hide()
					Market.Cancel_button.Hide()
				}

				Market.Market_button.Show()
				if now > endTime[0] || Market.Confirming {
					Market.Market_button.Hide()
				}
			}()
		}
	}
}

// Create buy now tab info for current asset
func GetBuyNowDetails(scid string) {
	if Gnomes.IsReady() && len(scid) == 64 {
		name, _ := Gnomes.GetSCIDValuesByKey(scid, "nameHdr")
		collection, _ := Gnomes.GetSCIDValuesByKey(scid, "collection")
		description, _ := Gnomes.GetSCIDValuesByKey(scid, "descrHdr")
		creator, _ := Gnomes.GetSCIDValuesByKey(scid, "creatorAddr")
		owner, _ := Gnomes.GetSCIDValuesByKey(scid, "owner")
		typeHdr, _ := Gnomes.GetSCIDValuesByKey(scid, "typeHdr")
		_, owner_update := Gnomes.GetSCIDValuesByKey(scid, "ownerCanUpdate")
		_, start := Gnomes.GetSCIDValuesByKey(scid, "startPrice")
		_, royalty := Gnomes.GetSCIDValuesByKey(scid, "royalty")
		_, endTime := Gnomes.GetSCIDValuesByKey(scid, "endBlockTime")
		_, startTime := Gnomes.GetSCIDValuesByKey(scid, "startBlockTime")
		_, artFee := Gnomes.GetSCIDValuesByKey(scid, "artificerFee")

		if name != nil && collection != nil && start != nil && royalty != nil && endTime != nil && artFee != nil && typeHdr != nil {
			go func() {
				Market.Viewing_coll = collection[0]

				Market.Name.SetText(name[0])

				Market.Type.SetText(AssetType(collection[0], typeHdr[0]))

				Market.Collection.SetText(collection[0])

				Market.Description.SetText(description[0])

				if Market.Creator.Text != creator[0] {
					Market.Creator.SetText(creator[0])
				}

				if Market.Owner.Text != owner[0] {
					Market.Owner.SetText(owner[0])
				}

				if owner_update[0] == 1 {
					Market.Owner_update.SetText("Yes")
				} else {
					Market.Owner_update.SetText("No")
				}

				Market.Art_fee.SetText(strconv.Itoa(int(artFee[0])) + "%")

				Market.Royalty.SetText(strconv.Itoa(int(royalty[0])) + "%")

				Market.Buy_amt = start[0]
				value := float64(start[0])
				str := fmt.Sprintf("%.5f", value/100000)
				Market.Start_price.SetText(str + " Dero")

				Market.Entry.SetText(str)
				Market.Entry.Disable()
				end, _ := rpc.MsToTime(strconv.Itoa(int(endTime[0]) * 1000))
				Market.End_time.SetText(end.String())

				now := uint64(time.Now().Unix())
				if owner[0] == rpc.Wallet.Address {
					if now < startTime[0]+300 && startTime[0] > 0 && !Market.Confirming {
						Market.Cancel_button.Show()
					} else {
						Market.Cancel_button.Hide()
					}

					if now > endTime[0] && endTime[0] > 0 && !Market.Confirming {
						Market.Close_button.Show()
					} else {
						Market.Close_button.Hide()
					}
				} else {
					Market.Close_button.Hide()
					Market.Cancel_button.Hide()
				}

				Market.Market_button.Show()
				if now > endTime[0] || Market.Confirming {
					Market.Market_button.Hide()
				}
			}()
		}
	}
}

// Create info for unlisted NFA
func GetUnlistedDetails(scid string) {
	if Gnomes.IsReady() && len(scid) == 64 {
		name, _ := Gnomes.GetSCIDValuesByKey(scid, "nameHdr")
		collection, _ := Gnomes.GetSCIDValuesByKey(scid, "collection")
		description, _ := Gnomes.GetSCIDValuesByKey(scid, "descrHdr")
		creator, _ := Gnomes.GetSCIDValuesByKey(scid, "creatorAddr")
		owner, _ := Gnomes.GetSCIDValuesByKey(scid, "owner")
		typeHdr, _ := Gnomes.GetSCIDValuesByKey(scid, "typeHdr")
		_, owner_update := Gnomes.GetSCIDValuesByKey(scid, "ownerCanUpdate")
		_, start := Gnomes.GetSCIDValuesByKey(scid, "startPrice")
		_, royalty := Gnomes.GetSCIDValuesByKey(scid, "royalty")
		_, endTime := Gnomes.GetSCIDValuesByKey(scid, "endBlockTime")
		_, startTime := Gnomes.GetSCIDValuesByKey(scid, "startBlockTime")
		_, artFee := Gnomes.GetSCIDValuesByKey(scid, "artificerFee")

		if name != nil && collection != nil && start != nil && royalty != nil && endTime != nil && artFee != nil && typeHdr != nil {
			go func() {
				Market.Viewing_coll = collection[0]

				Market.Name.SetText(name[0])

				Market.Type.SetText(AssetType(collection[0], typeHdr[0]))

				Market.Collection.SetText(collection[0])

				Market.Description.SetText(description[0])

				if Market.Creator.Text != creator[0] {
					Market.Creator.SetText(creator[0])
				}

				if Market.Owner.Text != owner[0] {
					Market.Owner.SetText(owner[0])
				}

				if owner_update[0] == 1 {
					Market.Owner_update.SetText("Yes")
				} else {
					Market.Owner_update.SetText("No")
				}

				Market.Art_fee.SetText(strconv.Itoa(int(artFee[0])) + "%")

				Market.Royalty.SetText(strconv.Itoa(int(royalty[0])) + "%")

				Market.Entry.SetText("0")
				Market.Entry.Disable()

				now := uint64(time.Now().Unix())
				if owner[0] == rpc.Wallet.Address {
					if now < startTime[0]+300 && startTime[0] > 0 && !Market.Confirming {
						Market.Cancel_button.Show()
					} else {
						Market.Cancel_button.Hide()
					}

					if now > endTime[0] && endTime[0] > 0 && !Market.Confirming {
						Market.Close_button.Show()
					} else {
						Market.Close_button.Hide()
					}
				} else {
					Market.Close_button.Hide()
					Market.Cancel_button.Hide()
				}
			}()
		}
	}
}

// Get percentages for a NFA
func GetListingPercents(scid string) (artP float64, royaltyP float64) {
	if Gnomes.IsReady() {
		_, artFee := Gnomes.GetSCIDValuesByKey(scid, "artificerFee")
		_, royalty := Gnomes.GetSCIDValuesByKey(scid, "royalty")

		if artFee != nil && royalty != nil {
			artP = float64(artFee[0]) / 100
			royaltyP = float64(royalty[0]) / 100

			return
		}
	}

	return
}

// Scan index for any active NFA listings
//   - Pass assets from db store, can be nil arg
func FindNFAListings(assets map[string]string) {
	if Gnomes.IsReady() && rpc.IsReady() {
		auction := []string{" Collection,  Name,  Description,  SCID:"}
		buy_now := []string{" Collection,  Name,  Description,  SCID:"}
		my_list := []string{" Collection,  Name,  Description,  SCID:"}
		if assets == nil {
			assets = Gnomes.GetAllOwnersAndSCIDs()
		}

		for sc := range assets {
			if !Gnomes.IsRunning() {
				return
			}

			a, owned, expired := checkNFAAuctionListing(sc)

			if a != "" && !expired {
				auction = append(auction, a)
			}

			if owned {
				my_list = append(my_list, a)
			}

			b, owned, expired := checkNFABuyListing(sc)

			if b != "" && !expired {
				buy_now = append(buy_now, b)
			}

			if owned {
				my_list = append(my_list, b)
			}
		}

		if !Gnomes.IsRunning() {
			return
		}

		Market.Auctions = auction
		Market.Buy_now = buy_now
		Market.My_list = my_list
		sort.Strings(Market.Auctions)
		sort.Strings(Market.Buy_now)
		sort.Strings(Market.My_list)

		Market.Auction_list.Refresh()
		Market.Buy_list.Refresh()
	}
}

// Check wallet for all indexed NFAs
//   - Pass scids from db store, can be nil arg
//   - Pass false gc for rechecks
func CheckAllNFAs(gc bool, scids map[string]string) {
	if Gnomes.IsReady() && !gc {
		if scids == nil {
			scids = Gnomes.GetAllOwnersAndSCIDs()
		}

		assets := []Asset{}
		for sc := range scids {
			if !rpc.Wallet.IsConnected() || !Gnomes.IsRunning() {
				break
			}

			if header, _ := Gnomes.GetSCIDValuesByKey(sc, "nameHdr"); header != nil {
				owner, _ := Gnomes.GetSCIDValuesByKey(sc, "owner")
				file, _ := Gnomes.GetSCIDValuesByKey(sc, "fileURL")
				if owner != nil && file != nil {
					if owner[0] == rpc.Wallet.Address && ValidNFA(file[0]) {
						if asset := GetOwnedAssetInfo(sc); asset.Collection != "" {
							assets = append(assets, asset)
						}
					}
				}
			}
		}

		sort.Slice(assets, func(i, j int) bool {
			return assets[i].Name < assets[j].Name
		})
		Assets.Asset = assets
		Assets.List.Refresh()
	}
}

// Get NFA or G45 asset info
func GetOwnedAssetInfo(scid string) (asset Asset) {
	if Gnomes.IsReady() {
		header, _ := Gnomes.GetSCIDValuesByKey(scid, "nameHdr")
		if header != nil {
			asset.SCID = scid
			asset.Name = header[0]
			collection, _ := Gnomes.GetSCIDValuesByKey(scid, "collection")
			if collection != nil {
				asset.Collection = collection[0]
				typeHdr, _ := Gnomes.GetSCIDValuesByKey(scid, "typeHdr")
				if typeHdr != nil {
					asset.Type = AssetType(collection[0], typeHdr[0])
				}
			} else {
				asset.Collection = "?"
			}

			icon, _ := Gnomes.GetSCIDValuesByKey(scid, "iconURLHdr")
			if icon != nil {
				if img, err := dreams.DownloadBytes(icon[0]); err == nil {
					asset.Image = img
				} else {
					logger.Errorln("[GetOwnedAssetInfo]", err)
				}
			}
		} else {
			data, _ := Gnomes.GetSCIDValuesByKey(scid, "metadata")
			minter, _ := Gnomes.GetSCIDValuesByKey(scid, "minter")
			collection, _ := Gnomes.GetSCIDValuesByKey(scid, "collection")
			if data != nil && minter != nil && collection != nil {
				asset.SCID = scid
				if minter[0] == Seals_mint && collection[0] == Seals_coll {
					var seal Seal
					if err := json.Unmarshal([]byte(data[0]), &seal); err == nil {
						check := strings.Trim(seal.Name, " #0123456789")
						if check == "Dero Seals" {
							asset.Name = seal.Name
							asset.Collection = check
							asset.Type = "Avatar"
							if img, err := dreams.DownloadBytes(ParseURL(seal.Image)); err == nil {
								asset.Image = img
							} else {
								logger.Errorln("[GetOwnedAssetInfo]", err)
							}
						}
					}
				} else if minter[0] == ATeam_mint && collection[0] == ATeam_coll {
					var agent Agent
					if err := json.Unmarshal([]byte(data[0]), &agent); err == nil {
						asset.Name = agent.Name
						asset.Collection = "Dero A-Team"
						asset.Type = "Avatar"
						if img, err := dreams.DownloadBytes(ParseURL(agent.Image)); err == nil {
							asset.Image = img
						} else {
							logger.Errorln("[GetOwnedAssetInfo]", err)
						}
					}
				} else if minter[0] == Degen_mint && collection[0] == Degen_coll {
					var degen Degen
					if err := json.Unmarshal([]byte(data[0]), &degen); err == nil {
						asset.Name = degen.Name
						asset.Collection = "Dero Degens"
						asset.Type = "Avatar"
						if img, err := dreams.DownloadBytes(ParseURL(degen.Image)); err == nil {
							asset.Image = img
						} else {
							logger.Errorln("[GetOwnedAssetInfo]", err)
						}
					}
				}
			}
		}
	}

	return
}
