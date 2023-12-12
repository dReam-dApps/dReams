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
	"github.com/dReam-dApps/dReams/gnomes"
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
//   - bid true for auction, false for sale
func BidBuyConfirm(scid string, amt uint64, bid bool, d *dreams.AppObject) {
	var text, title, coll, name string
	Market.Confirming = true
	f := float64(amt)
	amt_str := fmt.Sprintf("%.5f", f/100000)
	if bid {
		title = "Bid"
		listing, _, _ := checkNFAAuctionListing(scid)
		split := strings.Split(listing, "   ")
		if len(split) == 4 {
			coll = split[0]
			name = split[1]
		}
		text = fmt.Sprintf("Bidding on SCID:\n\n%s\n\nAsset: %s\n\nCollection: %s\n\nBid amount: %s Dero", scid, name, coll, amt_str)
	} else {
		title = "Buy"
		listing, _, _ := checkNFABuyListing(scid)
		split := strings.Split(listing, "   ")
		if len(split) == 4 {
			coll = split[0]
			name = split[1]
		}
		text = fmt.Sprintf("Buying SCID:\n\n%s\n\nAsset: %s\n\nCollection: %s\n\nAmount: %s Dero", scid, name, coll, amt_str)

	}

	label := widget.NewLabel(text)
	label.Wrapping = fyne.TextWrapWord
	label.Alignment = fyne.TextAlignCenter

	done := make(chan struct{})
	confirm := dialog.NewConfirm(title, text, func(b bool) {
		if b {
			if bid {
				if tx := rpc.BidBuyNFA(scid, "Bid", amt); tx != "" {
					go ShowTxDialog("NFA Bid", fmt.Sprintf("TXID: %s", tx), tx, 3*time.Second, d.Window)
				} else {
					go ShowTxDialog("NFA Bid", "TX error, check logs", tx, 3*time.Second, d.Window)
				}
			} else {
				if tx := rpc.BidBuyNFA(scid, "BuyItNow", amt); tx != "" {
					go ShowTxDialog("NFA Buy", fmt.Sprintf("TXID: %s", tx), tx, 3*time.Second, d.Window)
				} else {
					go ShowTxDialog("NFA Buy", "TX error, check logs", tx, 3*time.Second, d.Window)
				}
			}
		}
		Market.Confirming = false
		done <- struct{}{}
	}, d.Window)

	go ShowConfirmDialog(done, confirm)
}

// Confirm a cancel or close action of listed NFA
//   - close true to close listing, false to cancel
//   - Confirmation string from Market.Tab
//   - Pass main window obj to reset to
func ConfirmCancelClose(scid string, close bool, d *dreams.AppObject) {
	var text, title, coll, name string
	Market.Confirming = true
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
	if close {
		text = fmt.Sprintf("Close listing for SCID:\n\n%s\n\nAsset: %s\n\nCollection: %s", scid, name, coll)
	} else {
		text = fmt.Sprintf("Cancel listing for SCID:\n\n%s\n\nAsset: %s\n\nCollection: %s", scid, name, coll)
	}

	label := widget.NewLabel(text)
	label.Wrapping = fyne.TextWrapWord
	label.Alignment = fyne.TextAlignCenter

	done := make(chan struct{})
	confirm := dialog.NewConfirm(title, text, func(b bool) {
		if b {
			if close {
				if tx := rpc.CancelCloseNFA(scid, true); tx != "" {
					go ShowTxDialog("NFA Close", fmt.Sprintf("TXID: %s", tx), tx, 3*time.Second, d.Window)
				} else {
					go ShowTxDialog("NFA Close", "TX error, check logs", tx, 3*time.Second, d.Window)
				}
			} else {
				if tx := rpc.CancelCloseNFA(scid, false); tx != "" {
					go ShowTxDialog("NFA Cancel", fmt.Sprintf("TXID: %s", tx), tx, 3*time.Second, d.Window)
				} else {
					go ShowTxDialog("NFA Cancel", "TX error, check logs", tx, 3*time.Second, d.Window)
				}
			}
			Market.Viewing = ""
			Market.Viewing_coll = ""
			Market.Cancel_button.Hide()
		}

		Market.Confirming = false
		done <- struct{}{}
	}, d.Window)

	go ShowConfirmDialog(done, confirm)
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
func SearchNFAs(d *dreams.AppObject) fyne.CanvasObject {
	var dest_addr string
	var message_button *widget.Button
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
		} else {
			dest_addr = ""
		}

		if len(dest_addr) == 66 {
			message_button.Show()
		} else {
			message_button.Hide()
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
	clear_button.Importance = widget.LowImportance

	show_results := widget.NewButtonWithIcon("", fyne.Theme.Icon(fyne.CurrentApp().Settings().Theme(), "arrowDropDown"), func() {
		search_entry.ShowCompletion()
	})

	search_cont := container.NewBorder(container.NewCenter(search_by), nil, container.NewHBox(clear_button, show_results), search_button, search_entry)

	message_button = widget.NewButton("Message Owner", func() {
		if dest_addr != "" {
			SendMessageMenu(dest_addr, bundle.ResourceDReamsIconAltPng)
		} else {
			dialog.NewInformation("No Address", "Could not get owner address", d.Window).Show()
		}
	})
	message_button.Importance = widget.HighImportance
	message_button.Hide()

	return container.NewBorder(search_cont, container.NewHBox(layout.NewSpacer(), message_button), nil, nil, container.NewHBox())
}

// Get NFA image files
func GetNFAImages(scid string) {
	if gnomon.IsReady() && len(scid) == 64 {
		name, _ := gnomon.GetSCIDValuesByKey(scid, "nameHdr")
		icon, _ := gnomon.GetSCIDValuesByKey(scid, "iconURLHdr")
		cover, _ := gnomon.GetSCIDValuesByKey(scid, "coverURL")
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

	frame := canvas.NewImageFromResource(bundle.ResourceFramePng)
	frame.SetMinSize(fyne.NewSize(100, 100))

	return container.NewStack(border, frame)
}

var toolsBadge = *canvas.NewImageFromResource(bundle.ResourceDReamToolsPng)

// Badge for dReam Tools enabled assets
func ToolsBadge() fyne.CanvasObject {
	toolsBadge.SetMinSize(fyne.NewSize(90, 90))
	border := container.NewBorder(layout.NewSpacer(), layout.NewSpacer(), layout.NewSpacer(), layout.NewSpacer(), &toolsBadge)

	frame := canvas.NewImageFromResource(bundle.ResourceFramePng)
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

	// Market.Details_box.Refresh()
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
func PlaceMarket(d *dreams.AppObject) *container.Split {
	details := container.NewStack(NFAMarketInfo())

	tabs := container.NewAppTabs(
		container.NewTabItem("Auctions", AuctionListings()),
		container.NewTabItem("Buy Now", BuyNowListings()),
		container.NewTabItem("My Listings", MyNFAListings()),
		container.NewTabItem("Search", SearchNFAs(d)))

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
				BidBuyConfirm(scid, amt, true, d)
			} else if text == "Buy" {
				BidBuyConfirm(scid, Market.Buy_amt, false, d)
			}
		}
	})
	Market.Market_button.Importance = widget.HighImportance
	Market.Market_button.Hide()

	Market.Cancel_button = widget.NewButton("Cancel", func() {
		if len(Market.Viewing) == 64 {
			Market.Cancel_button.Hide()
			ConfirmCancelClose(Market.Viewing, false, d)
		}
	})

	Market.Close_button = widget.NewButton("Close", func() {
		if len(Market.Viewing) == 64 {
			Market.Close_button.Hide()
			ConfirmCancelClose(Market.Viewing, true, d)
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
	synced := false

	for {
		select {
		case <-ticker.C: // do on interval
			rpc.Ping()
			rpc.EchoWallet(tag)

			// Get all NFA listings
			if gnomon.IsRunning() && offset%2 == 0 {
				FindNFAListings(nil)
				if offset > 19 {
					offset = 0
				}
			}

			rpc.Wallet.GetBalance()

			// Refresh Dero balance and Gnomon endpoint
			connect_box.RefreshBalance()
			if !rpc.Startup {
				gnomes.GnomonEndPoint()
			}

			// If connected daemon connected start looking for Gnomon sync with daemon
			if rpc.Daemon.IsConnected() && gnomon.IsRunning() {
				connect_box.Disconnect.SetChecked(true)
				// Get indexed SCID count
				gnomon.IndexContains()
				if gnomon.HasIndex(1) {
					gnomon.Checked(true)
				}

				if gnomon.GetLastHeight() >= gnomon.GetChainHeight()-3 {
					gnomon.Synced(true)
				} else {
					gnomon.Synced(false)
					gnomon.Checked(false)
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
					Assets.Button.Rescan.Hide()
					Assets.Claim.Hide()
					Assets.Asset = []Asset{}
				}

				// Check wallet for all owned NFAs and store icons in boltdb
				if gnomon.IsSynced() && !synced {
					CheckAllNFAs(nil)
					Assets.List.Refresh()
					if gnomon.DBStorageType() == "boltdb" {
						for _, r := range Assets.Asset {
							gnomes.StoreBolt(r.Collection, r.Name, r)
						}
					}
					synced = true
				}
				Info.RefreshIndexed()
			} else {
				connect_box.Disconnect.SetChecked(false)
				DisableIndexControls(true)
				Assets.Asset = []Asset{}
			}

			if rpc.Daemon.IsConnected() {
				rpc.Startup = false
			}

			offset++

		case <-quit: // exit
			logger.Printf("[%s] Closing...\n", tag)
			if gnomes.Indicator.Icon != nil {
				gnomes.Indicator.Icon.Stop()
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
	if gnomon.IsReady() {
		if owner, _ := gnomon.GetSCIDValuesByKey(scid, "owner"); owner != nil {
			if listType, _ := gnomon.GetSCIDValuesByKey(scid, "listType"); listType != nil {
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
	if gnomon.IsReady() {
		if creator, _ := gnomon.GetSCIDValuesByKey(scid, "creatorAddr"); creator != nil {
			listType, _ := gnomon.GetSCIDValuesByKey(scid, "listType")
			header, _ := gnomon.GetSCIDValuesByKey(scid, "nameHdr")
			coll, _ := gnomon.GetSCIDValuesByKey(scid, "collection")
			desc, _ := gnomon.GetSCIDValuesByKey(scid, "descrHdr")
			if listType != nil && header != nil && coll != nil && desc != nil {
				if Market.DreamsFilter {
					if IsDreamsNFACollection(coll[0]) {
						if listType[0] == "auction" {
							desc_check := TrimStringLen(desc[0], 66)
							asset = coll[0] + "   " + header[0] + "   " + desc_check + "   " + scid
							if owner, _ := gnomon.GetSCIDValuesByKey(scid, "owner"); owner != nil {
								if owner[0] == rpc.Wallet.Address {
									owned = true
								}
							}

							if _, endTime := gnomon.GetSCIDValuesByKey(scid, "endBlockTime"); endTime != nil {
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
							if owner, _ := gnomon.GetSCIDValuesByKey(scid, "owner"); owner != nil {
								if owner[0] == rpc.Wallet.Address {
									owned = true
								}
							}

							if _, endTime := gnomon.GetSCIDValuesByKey(scid, "endBlockTime"); endTime != nil {
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
	if gnomon.IsReady() {
		if creator, _ := gnomon.GetSCIDValuesByKey(scid, "creatorAddr"); creator != nil {
			listType, _ := gnomon.GetSCIDValuesByKey(scid, "listType")
			header, _ := gnomon.GetSCIDValuesByKey(scid, "nameHdr")
			coll, _ := gnomon.GetSCIDValuesByKey(scid, "collection")
			desc, _ := gnomon.GetSCIDValuesByKey(scid, "descrHdr")
			if listType != nil && header != nil && coll != nil && desc != nil {
				if Market.DreamsFilter {
					if IsDreamsNFACollection(coll[0]) {
						if listType[0] == "sale" {
							desc_check := TrimStringLen(desc[0], 66)
							asset = coll[0] + "   " + header[0] + "   " + desc_check + "   " + scid
							if owner, _ := gnomon.GetSCIDValuesByKey(scid, "owner"); owner != nil {
								if owner[0] == rpc.Wallet.Address {
									owned = true
								}
							}

							if _, endTime := gnomon.GetSCIDValuesByKey(scid, "endBlockTime"); endTime != nil {
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
							if owner, _ := gnomon.GetSCIDValuesByKey(scid, "owner"); owner != nil {
								if owner[0] == rpc.Wallet.Address {
									owned = true
								}
							}

							if _, endTime := gnomon.GetSCIDValuesByKey(scid, "endBlockTime"); endTime != nil {
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
	if gnomon.IsReady() {
		results = []string{" Collection,  Name,  Description,  SCID:"}
		assets := gnomon.GetAllOwnersAndSCIDs()

		for sc := range assets {
			if !gnomon.IsReady() {
				return
			}

			if file, _ := gnomon.GetSCIDValuesByKey(sc, "fileURL"); file != nil {
				if ValidNFA(file[0]) {
					if name, _ := gnomon.GetSCIDValuesByKey(sc, "nameHdr"); name != nil {
						coll, _ := gnomon.GetSCIDValuesByKey(sc, "collection")
						desc, _ := gnomon.GetSCIDValuesByKey(sc, "descrHdr")
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
	if gnomon.IsReady() && len(scid) == 64 {
		name, _ := gnomon.GetSCIDValuesByKey(scid, "nameHdr")
		collection, _ := gnomon.GetSCIDValuesByKey(scid, "collection")
		description, _ := gnomon.GetSCIDValuesByKey(scid, "descrHdr")
		creator, _ := gnomon.GetSCIDValuesByKey(scid, "creatorAddr")
		owner, _ := gnomon.GetSCIDValuesByKey(scid, "owner")
		typeHdr, _ := gnomon.GetSCIDValuesByKey(scid, "typeHdr")
		_, owner_update := gnomon.GetSCIDValuesByKey(scid, "ownerCanUpdate")
		_, start := gnomon.GetSCIDValuesByKey(scid, "startPrice")
		_, current := gnomon.GetSCIDValuesByKey(scid, "currBidAmt")
		_, bid_price := gnomon.GetSCIDValuesByKey(scid, "currBidPrice")
		_, royalty := gnomon.GetSCIDValuesByKey(scid, "royalty")
		_, bids := gnomon.GetSCIDValuesByKey(scid, "bidCount")
		_, endTime := gnomon.GetSCIDValuesByKey(scid, "endBlockTime")
		_, startTime := gnomon.GetSCIDValuesByKey(scid, "startBlockTime")
		_, artFee := gnomon.GetSCIDValuesByKey(scid, "artificerFee")

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
	if gnomon.IsReady() && len(scid) == 64 {
		name, _ := gnomon.GetSCIDValuesByKey(scid, "nameHdr")
		collection, _ := gnomon.GetSCIDValuesByKey(scid, "collection")
		description, _ := gnomon.GetSCIDValuesByKey(scid, "descrHdr")
		creator, _ := gnomon.GetSCIDValuesByKey(scid, "creatorAddr")
		owner, _ := gnomon.GetSCIDValuesByKey(scid, "owner")
		typeHdr, _ := gnomon.GetSCIDValuesByKey(scid, "typeHdr")
		_, owner_update := gnomon.GetSCIDValuesByKey(scid, "ownerCanUpdate")
		_, start := gnomon.GetSCIDValuesByKey(scid, "startPrice")
		_, royalty := gnomon.GetSCIDValuesByKey(scid, "royalty")
		_, endTime := gnomon.GetSCIDValuesByKey(scid, "endBlockTime")
		_, startTime := gnomon.GetSCIDValuesByKey(scid, "startBlockTime")
		_, artFee := gnomon.GetSCIDValuesByKey(scid, "artificerFee")

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
	if gnomon.IsReady() && len(scid) == 64 {
		name, _ := gnomon.GetSCIDValuesByKey(scid, "nameHdr")
		collection, _ := gnomon.GetSCIDValuesByKey(scid, "collection")
		description, _ := gnomon.GetSCIDValuesByKey(scid, "descrHdr")
		creator, _ := gnomon.GetSCIDValuesByKey(scid, "creatorAddr")
		owner, _ := gnomon.GetSCIDValuesByKey(scid, "owner")
		typeHdr, _ := gnomon.GetSCIDValuesByKey(scid, "typeHdr")
		_, owner_update := gnomon.GetSCIDValuesByKey(scid, "ownerCanUpdate")
		_, start := gnomon.GetSCIDValuesByKey(scid, "startPrice")
		_, royalty := gnomon.GetSCIDValuesByKey(scid, "royalty")
		_, endTime := gnomon.GetSCIDValuesByKey(scid, "endBlockTime")
		_, startTime := gnomon.GetSCIDValuesByKey(scid, "startBlockTime")
		_, artFee := gnomon.GetSCIDValuesByKey(scid, "artificerFee")

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
	if gnomon.IsReady() {
		_, artFee := gnomon.GetSCIDValuesByKey(scid, "artificerFee")
		_, royalty := gnomon.GetSCIDValuesByKey(scid, "royalty")

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
	if gnomon.IsReady() && rpc.IsReady() {
		auction := []string{" Collection,  Name,  Description,  SCID:"}
		buy_now := []string{" Collection,  Name,  Description,  SCID:"}
		my_list := []string{" Collection,  Name,  Description,  SCID:"}
		if assets == nil {
			assets = gnomon.GetAllOwnersAndSCIDs()
		}

		for sc := range assets {
			if !gnomon.IsRunning() {
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

		if !gnomon.IsRunning() {
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
//   - Pass scids from db store, can be nil arg to check all from db
func CheckAllNFAs(scids map[string]string) {
	if gnomon.IsReady() {
		if scids == nil {
			scids = gnomon.GetAllOwnersAndSCIDs()
		}

		Assets.Asset = []Asset{}
		Theme.Select.Options = []string{}

		for sc := range scids {
			if !rpc.Wallet.IsConnected() || !gnomon.IsRunning() {
				break
			}

			if header, _ := gnomon.GetSCIDValuesByKey(sc, "nameHdr"); header != nil {
				owner, _ := gnomon.GetSCIDValuesByKey(sc, "owner")
				file, _ := gnomon.GetSCIDValuesByKey(sc, "fileURL")
				collection, _ := gnomon.GetSCIDValuesByKey(sc, "collection")
				icon, _ := gnomon.GetSCIDValuesByKey(sc, "iconURLHdr")
				if owner != nil && file != nil && collection != nil && icon != nil {
					if owner[0] == rpc.Wallet.Address && ValidNFA(file[0]) {
						var add Asset
						add.Name = header[0]
						add.Collection = collection[0]
						add.SCID = sc
						if typeHdr, _ := gnomon.GetSCIDValuesByKey(sc, "typeHdr"); typeHdr != nil {
							add.Type = AssetType(collection[0], typeHdr[0])
						}

						if collection[0] == "AZY-Deroscapes" || collection[0] == "SIXART" {
							Theme.Add(header[0], owner[0])
						}
						Assets.Add(add, icon[0])
					}
				}
			} else {
				if data, _ := gnomon.GetSCIDValuesByKey(sc, "metadata"); data != nil {
					icon, _ := gnomon.GetSCIDValuesByKey(sc, "iconURLHdr")
					owner, _ := gnomon.GetSCIDValuesByKey(sc, "owner")
					minter, _ := gnomon.GetSCIDValuesByKey(sc, "minter")
					collection, _ := gnomon.GetSCIDValuesByKey(sc, "collection")

					if data != nil && minter != nil && collection != nil && owner != nil && icon != nil {
						if owner[0] == rpc.Wallet.Address {
							var add Asset
							if minter[0] == Seals_mint && collection[0] == Seals_coll {
								var seal Seal
								if err := json.Unmarshal([]byte(data[0]), &seal); err == nil {
									check := strings.Trim(seal.Name, " #0123456789")
									if check == "Dero Seals" {
										add.Name = seal.Name
										add.Collection = check
										add.Type = "Avatar"
										if img, err := dreams.DownloadBytes(ParseURL(seal.Image)); err == nil {
											add.Image = img
										} else {
											logger.Errorln("[CheckAllNFAs]", err)
										}

										Assets.Add(add, icon[0])
									}
								}
							} else if minter[0] == ATeam_mint && collection[0] == ATeam_coll {
								var agent Agent
								if err := json.Unmarshal([]byte(data[0]), &agent); err == nil {
									add.Name = agent.Name
									add.Collection = "Dero A-Team"
									add.Type = "Avatar"
									if img, err := dreams.DownloadBytes(ParseURL(agent.Image)); err == nil {
										add.Image = img
									} else {
										logger.Errorln("[CheckAllNFAs]", err)
									}

									Assets.Add(add, icon[0])
								}
							} else if minter[0] == Degen_mint && collection[0] == Degen_coll {
								var degen Degen
								if err := json.Unmarshal([]byte(data[0]), &degen); err == nil {
									add.Name = degen.Name
									add.Collection = "Dero Degens"
									add.Type = "Avatar"
									if img, err := dreams.DownloadBytes(ParseURL(degen.Image)); err == nil {
										add.Image = img
									} else {
										logger.Errorln("[CheckAllNFAs]", err)
									}

									Assets.Add(add, icon[0])
								}
							}
						}
					}
				}
			}
		}

		Theme.Sort()
		Theme.Select.Options = append(Control.Themes, Theme.Select.Options...)
		Assets.SortList()
	}
}
