package menu

import (
	"SixofClubsss/dReams/rpc"
	"SixofClubsss/dReams/table"
	"fmt"
	"image/color"
	"strconv"
	"strings"

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
	Market.Entry.PlaceHolder = "Dero:"
	Market.Entry.Validator = validation.NewRegexp(`\d{1,}\.\d{1,5}$`, "Format Not Valid")
	Market.Entry.OnChanged = func(s string) {
		if Market.Entry.Validate() != nil {
			Market.Entry.SetText("0.0")
		}
	}

	Market.Market_button = widget.NewButton("Bid", func() {
		if Market.Market_button.Text == "Bid" {
			rpc.NfaBidBuy(Market.Viewing, "Bid", ToAtomicFive(Market.Entry.Text))
		} else if Market.Market_button.Text == "Buy" {
			rpc.NfaBidBuy(Market.Viewing, "BuyItNow", Market.Buy_amt)
		}
	})

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

func confirmCancelClose(scid string, c int) {
	if len(scid) == 64 {
		var text string
		switch c {
		case 0:
			text = "Close listing for SCID: " + scid
		case 1:
			text = "Cancel listing for SCID: " + scid
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
				Market.Close_button.Hide()
			case 1:
				rpc.NfaCancelClose(scid, "CancelListing")
				Market.Cancel_button.Hide()
			default:
			}
			Market.Viewing = ""
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
				Market.Viewing = split[3]
				ResetAuctionInfo()
				go GetAuctionImages(split[3])
				GetAuctionDetails(split[3])
				value := float64(Market.Bid_amt)
				str := fmt.Sprintf("%.5f", value/100000)
				Market.Entry.SetText(str)
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
			Market.Viewing = split[3]
			ResetBuyInfo()
			go GetBuyNowImages(split[3])
			GetBuyNowDetails(split[3])
		}
	}

	return Market.Buy_list
}

func NfaIcon(res fyne.Resource) fyne.CanvasObject {
	Market.Icon.SetMinSize(fyne.NewSize(94, 94))
	Market.Icon.Resize(fyne.NewSize(94, 94))
	Market.Icon.Move(fyne.NewPos(8, 3))

	frame := canvas.NewImageFromResource(res)
	frame.SetMinSize(fyne.NewSize(100, 100))
	frame.Resize(fyne.NewSize(100, 100))
	frame.Move(fyne.NewPos(5, 0))

	cont := *container.NewWithoutLayout(&Market.Icon, frame)

	return &cont
}

func NfaImg(img canvas.Image) *fyne.Container {
	Market.Cover.SetMinSize(fyne.NewSize(133, 200))
	Market.Cover.Resize(fyne.NewSize(266, 400))
	Market.Cover.Move(fyne.NewPos(400, -50))

	cont := container.NewWithoutLayout(&img)

	return cont
}

func NfaMarketInfo() fyne.CanvasObject {
	Market.Name = canvas.NewText(" Name: ", color.White)
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

	return AuctionInfo()
}

func AuctionInfo() fyne.CanvasObject {
	Market.Details_box = *container.NewVBox(
		NfaImg(Market.Cover),
		NfaIcon(Resource.Frame),
		Market.Name,
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

func ResetAuctionInfo() {
	Market.Bid_amt = 0
	Market.Icon = *canvas.NewImageFromImage(nil)
	Market.Cover = *canvas.NewImageFromImage(nil)
	Market.Name.Text = (" Name: ")
	Market.Name.Refresh()
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
}

func BuyNowInfo() fyne.CanvasObject {
	Market.Details_box = *container.NewVBox(
		NfaImg(Market.Cover),
		NfaIcon(Resource.Frame),
		Market.Name,
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
	Market.Icon = *canvas.NewImageFromImage(nil)
	Market.Cover = *canvas.NewImageFromImage(nil)
	Market.Name.Text = (" Name: ")
	Market.Name.Refresh()
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
}
