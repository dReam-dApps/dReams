package menu

import (
	"fmt"
	"image/color"
	"strings"
	"time"

	"github.com/SixofClubsss/dReams/bundle"
	"github.com/SixofClubsss/dReams/dwidget"
	"github.com/SixofClubsss/dReams/rpc"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/validation"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type marketItems struct {
	Tab           string
	Entry         *dwidget.TenthAmt
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
	Buy_amt       uint64
	Bid_amt       uint64
	Viewing       string
	Viewing_coll  string
	Auctions      []string
	Buy_now       []string
}

var Market marketItems

// NFA market amount entry
func MarketEntry() fyne.CanvasObject {
	Market.Entry = &dwidget.TenthAmt{}
	Market.Entry.ExtendBaseWidget(Market.Entry)
	Market.Entry.SetText("0.0")
	Market.Entry.PlaceHolder = "Dero:"
	Market.Entry.Validator = validation.NewRegexp(`\d{1,}\.\d{1,5}$`, "Format Not Valid")
	Market.Entry.OnChanged = func(s string) {
		if Market.Entry.Validate() != nil {
			Market.Entry.SetText("0.0")
		}
	}

	return Market.Entry
}

// Confirm a bid or buy action of listed NFA
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
			rpc.NfaBidBuy(scid, "Bid", amt)
		case 1:
			rpc.NfaBidBuy(scid, "BuyItNow", amt)
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

	return container.NewMax(Alpha120, content)
}

// Confirm a cancel or close action of listed NFA
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
			rpc.NfaCancelClose(scid, "CloseListing")
		case 1:
			rpc.NfaCancelClose(scid, "CancelListing")

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

	return container.NewMax(Alpha120, content)
}

// NFA auction listings object
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
	Market.Loading = canvas.NewText("Loading", color.White)
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
	Market.Name = canvas.NewText(" Name: ", color.White)
	Market.Type = canvas.NewText(" Asset Type: ", color.White)
	Market.Collection = canvas.NewText(" Collection: ", color.White)
	Market.Description = canvas.NewText(" Description: ", color.White)
	Market.Creator = canvas.NewText(" Creator: ", color.White)
	Market.Art_fee = canvas.NewText(" Artificer Fee: ", color.White)
	Market.Royalty = canvas.NewText(" Royalty: ", color.White)
	Market.Start_price = canvas.NewText(" Start Price: ", color.White)
	Market.Owner = canvas.NewText(" Owner: ", color.White)
	Market.Owner_update = canvas.NewText(" Owner can update: ", color.White)
	Market.Current_bid = canvas.NewText(" Current Bid: ", color.White)
	Market.Bid_price = canvas.NewText(" Minimum Bid: ", color.White)
	Market.Bid_count = canvas.NewText(" Bids: ", color.White)
	Market.End_time = canvas.NewText(" Ends At: ", color.White)

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
