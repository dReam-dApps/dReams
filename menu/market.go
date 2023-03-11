package menu

import (
	"fmt"
	"image/color"
	"strconv"
	"strings"
	"time"

	"github.com/SixofClubsss/dReams/rpc"
	"github.com/SixofClubsss/dReams/table"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/validation"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type marketItems struct {
	Tab           string
	Entry         *marketAmt
	Name          *canvas.Text
	Type          *canvas.Text
	Collection    *canvas.Text
	Description   *canvas.Text
	Creator       *canvas.Text
	Owner         *canvas.Text
	Owner_update  *canvas.Text
	Start_price   *canvas.Text
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
	Buy_amt       uint64
	Bid_amt       uint64
	Viewing       string
	Viewing_coll  string
	Auctions      []string
	Buy_now       []string
}

var Market marketItems

type marketAmt struct {
	table.NumericalEntry
}

func (e *marketAmt) TypedKey(k *fyne.KeyEvent) {
	switch k.Name {
	case fyne.KeyUp:
		if f, err := strconv.ParseFloat(e.Entry.Text, 64); err == nil {
			e.Entry.SetText(strconv.FormatFloat(float64(f+0.1), 'f', 1, 64))
		}
	case fyne.KeyDown:
		if f, err := strconv.ParseFloat(e.Entry.Text, 64); err == nil {
			if f >= 0.1 {
				e.Entry.SetText(strconv.FormatFloat(float64(f-0.1), 'f', 1, 64))
			}
		}
	}
	e.Entry.TypedKey(k)
}

func MarketItems() fyne.CanvasObject {
	Market.Entry = &marketAmt{}
	Market.Entry.ExtendBaseWidget(Market.Entry)
	Market.Entry.SetText("0.0")
	Market.Entry.PlaceHolder = "Dero:"
	Market.Entry.Validator = validation.NewRegexp(`\d{1,}\.\d{1,5}$`, "Format Not Valid")
	Market.Entry.OnChanged = func(s string) {
		if Market.Entry.Validate() != nil {
			Market.Entry.SetText("0.0")
		}
	}

	Market.Market_button = widget.NewButton("Bid", func() {
		scid := Market.Viewing
		if len(scid) == 64 {
			text := Market.Market_button.Text
			Market.Market_button.Hide()
			if text == "Bid" {
				amt := ToAtomicFive(Market.Entry.Text)
				bidBuyConfirm(scid, amt, 0)
			} else if text == "Buy" {
				bidBuyConfirm(scid, Market.Buy_amt, 1)
			}
		}
	})

	Market.Market_button.Hide()

	Market.Cancel_button = widget.NewButton("Cancel", func() {
		confirmCancelClose(Market.Viewing, 1)
	})

	Market.Close_button = widget.NewButton("Close", func() {
		confirmCancelClose(Market.Viewing, 0)
	})

	Market.Market_box = *container.NewAdaptiveGrid(6, Market.Entry, Market.Market_button, layout.NewSpacer(), layout.NewSpacer(), Market.Close_button, Market.Cancel_button)
	Market.Market_box.Hide()

	return &Market.Market_box
}

func bidBuyConfirm(scid string, amt uint64, b int) {
	var text string
	f := float64(amt)
	str := fmt.Sprintf("%.5f", f/100000)
	switch b {
	case 0:
		text = "Bidding on SCID:\n\n" + scid + "\n\nBid amount: " + str + " Dero\n\nConfirm bid"
	case 1:
		text = "Buying SCID:\n\n" + scid + "\n\nBuy amount: " + str + " Dero\n\nConfirm buy"
	default:
	}

	cw := fyne.CurrentApp().NewWindow("Confirm")
	cw.Resize(fyne.NewSize(350, 350))
	cw.SetFixedSize(true)
	cw.SetIcon(Resource.SmallIcon)
	cw.SetCloseIntercept(func() {
		Market.Market_button.Show()
		cw.Close()
	})

	label := widget.NewLabel(text)
	label.Wrapping = fyne.TextWrapWord

	confirm := widget.NewButton("Confirm", func() {
		switch b {
		case 0:
			rpc.NfaBidBuy(scid, "Bid", amt)
		case 1:
			rpc.NfaBidBuy(scid, "BuyItNow", amt)

		default:

		}
		Market.Market_button.Show()
		cw.Close()
	})

	cancel := widget.NewButton("Cancel", func() {
		Market.Market_button.Show()
		cw.Close()
	})

	left := container.NewVBox(confirm)
	right := container.NewVBox(cancel)
	buttons := container.NewAdaptiveGrid(2, left, right)

	content := container.NewVBox(label, layout.NewSpacer(), buttons)

	img := *canvas.NewImageFromResource(Resource.Back2)
	cw.SetContent(
		container.New(layout.NewMaxLayout(),
			&img,
			content))
	cw.Show()
}

func confirmCancelClose(scid string, c int) {
	if len(scid) == 64 {
		var text string
		switch c {
		case 0:
			text = "Close listing for SCID:\n" + scid
		case 1:
			text = "Cancel listing for SCID:\n" + scid
		default:
		}

		cw := fyne.CurrentApp().NewWindow("Confirm")
		cw.Resize(fyne.NewSize(350, 350))
		cw.SetFixedSize(true)
		cw.SetIcon(Resource.SmallIcon)
		label := widget.NewLabel(text)
		label.Wrapping = fyne.TextWrapWord

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
			cw.Close()
		})

		cancel := widget.NewButton("Cancel", func() {
			cw.Close()
		})

		left := container.NewVBox(confirm)
		right := container.NewVBox(cancel)
		buttons := container.NewAdaptiveGrid(2, left, right)

		content := container.NewVBox(label, layout.NewSpacer(), buttons)

		img := *canvas.NewImageFromResource(Resource.Back2)
		cw.SetContent(
			container.New(layout.NewMaxLayout(),
				&img,
				content))
		cw.Show()
	}
}

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

func ToolsBadge(half bool, res fyne.Resource) fyne.CanvasObject {
	var badge canvas.Image
	if !half {
		badge = *canvas.NewImageFromResource(Resource.Tools)
		badge.Resize(fyne.NewSize(94, 94))
		badge.Move(fyne.NewPos(8, 3))
	} else {
		badge = *canvas.NewImageFromResource(Resource.ToolsH)
		badge.Resize(fyne.NewSize(94, 94))
		badge.Move(fyne.NewPos(8, 3))
	}

	frame := canvas.NewImageFromResource(res)
	frame.SetMinSize(fyne.NewSize(100, 100))
	frame.Resize(fyne.NewSize(100, 100))
	frame.Move(fyne.NewPos(5, 0))

	cont := *container.NewWithoutLayout(&badge, frame)

	return &cont
}

func NfaImg(img canvas.Image) *fyne.Container {
	Market.Cover.Resize(fyne.NewSize(266, 400))
	Market.Cover.Move(fyne.NewPos(400, -230))

	cont := container.NewWithoutLayout(&img)

	return cont
}

func loadingText() *fyne.Container {
	Market.Loading = canvas.NewText("Loading", color.White)
	Market.Loading.TextSize = 18
	Market.Loading.Move(fyne.NewPos(400, 0))

	cont := container.NewWithoutLayout(Market.Loading)

	return cont
}

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

func clearNfaImages() {
	Market.Icon = *canvas.NewImageFromImage(nil)
	Market.Details_box.Objects[1].Refresh()

	Market.Cover = *canvas.NewImageFromImage(nil)
	Market.Details_box.Objects[0] = loadingText()
	Market.Details_box.Objects[0].Refresh()
	Market.Details_box.Refresh()

}

func NfaMarketInfo() fyne.CanvasObject {
	Market.Name = canvas.NewText(" Name: ", color.White)
	Market.Type = canvas.NewText(" Asset Type: ", color.White)
	Market.Collection = canvas.NewText(" Collection: ", color.White)
	Market.Description = canvas.NewText(" Description: ", color.White)
	Market.Creator = canvas.NewText(" Creator: ", color.White)
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

func AuctionInfo() fyne.CanvasObject {
	Market.Details_box = *container.NewVBox(
		NfaImg(Market.Cover),
		container.NewHBox(NfaIcon(Resource.Frame), layout.NewSpacer()),
		Market.Name,
		Market.Type,
		Market.Collection,
		Market.Description,
		Market.Creator,
		Market.Owner,
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
		Market.Details_box.Objects[1].(*fyne.Container).Objects[0] = NfaIcon(Resource.Frame)
		Market.Details_box.Objects[1].Refresh()
	}
	view := Market.Viewing_coll
	if view == "AZYPC" || view == "SIXPC" {
		Market.Details_box.Objects[1].(*fyne.Container).Objects[1] = ToolsBadge(false, Resource.Frame)
	} else if view == "AZYPCB" || view == "SIXPCB" {
		Market.Details_box.Objects[1].(*fyne.Container).Objects[1] = ToolsBadge(true, Resource.Frame)
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

func BuyNowInfo() fyne.CanvasObject {
	Market.Details_box = *container.NewVBox(
		NfaImg(Market.Cover),
		container.NewHBox(NfaIcon(Resource.Frame), layout.NewSpacer()),
		Market.Name,
		Market.Type,
		Market.Collection,
		Market.Description,
		Market.Creator,
		Market.Owner,
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
