package menu

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image/color"
	"sort"
	"strconv"
	"strings"
	"sync"
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
	sync.RWMutex
	Tab          string
	Entry        *dwidget.DeroAmts
	Loading      *widget.ProgressBarInfinite
	Icon         canvas.Image
	Cover        canvas.Image
	Details      fyne.Container
	Actions      fyne.Container
	Confirming   bool
	Searching    bool
	DreamsFilter bool
	downloading  bool
	Auctions     []NFAListing
	Buys         []NFAListing
	Wallets      []NFAListing
	Filters      []string
	Display      struct {
		Name        *widget.Entry
		Type        *widget.Entry
		Collection  *widget.Entry
		Description *widget.Entry
		Creator     *widget.Entry
		Owner       *widget.Entry
		Update      *widget.Entry
		Price       *widget.Entry
		Artificer   *widget.Entry
		Royalty     *widget.Entry
		Ends        *widget.Entry
		Bid         struct {
			Count   *widget.Entry
			Current *widget.Entry
			Price   *widget.Entry
		}
	}
	Button struct {
		BidBuy *widget.Button
		Cancel *widget.Button
		Close  *widget.Button
	}
	Viewing struct {
		Asset      string
		Collection string
	}
	List struct {
		Auction *widget.List
		Buy     *widget.List
		Wallet  *widget.List
	}
	amount struct {
		buy uint64
		bid uint64
	}
}

// NFA listing data
type NFAListing struct {
	Name        string `json:"name"`
	Collection  string `json:"collection"`
	Description string `json:"description"`
	SCID        string `json:"scid"`
	Icon        []byte `json:"icon"`
	IconURL     string `json:"iconURL"`
}

// Market contains widget a list objects for NFA market
var Market marketObjects

// Trim input string to specified len
func TrimStringLen(str string, l int) string {
	if len(str) > l {
		return str[0:l]
	}

	return str
}

// Sorts auction list by name
func (m *marketObjects) SortAuctions() {
	sort.Slice(m.Auctions, func(i, j int) bool {
		return m.Auctions[i].Name < m.Auctions[j].Name
	})
	m.List.Auction.Refresh()
}

// Sorts buy now list by name
func (m *marketObjects) SortBuys() {
	sort.Slice(m.Buys, func(i, j int) bool {
		return m.Buys[i].Name < m.Buys[j].Name
	})
	m.List.Buy.Refresh()
}

// Sorts wallet listings list by name
func (m *marketObjects) SortMyList() {
	sort.Slice(m.Wallets, func(i, j int) bool {
		return m.Wallets[i].Name < m.Wallets[j].Name
	})
	m.List.Wallet.Refresh()
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
		coll = listing.Collection
		name = listing.Name
		text = fmt.Sprintf("Bidding on SCID:\n\n%s\n\nAsset: %s\n\nCollection: %s\n\nBid amount: %s Dero", scid, name, coll, amt_str)
	} else {
		title = "Buy"
		listing, _, _ := checkNFABuyListing(scid)
		coll = listing.Collection
		name = listing.Name
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
func ConfirmCancelClose(scid string, close bool, d *dreams.AppObject) {
	var text, title, coll, name string
	Market.Confirming = true
	list, _ := CheckNFAListingType(scid)
	switch list {
	case 1:
		listing, _, _ := checkNFAAuctionListing(scid)
		coll = listing.Collection
		name = listing.Name
	case 2:
		listing, _, _ := checkNFABuyListing(scid)
		coll = listing.Collection
		name = listing.Name
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
			Market.Viewing.Asset = ""
			Market.Viewing.Collection = ""
			Market.Button.Cancel.Hide()
		}

		Market.Confirming = false
		done <- struct{}{}
	}, d.Window)

	go ShowConfirmDialog(done, confirm)
}

// List item object for market
func listItem() fyne.CanvasObject {
	spacer := canvas.NewImageFromImage(nil)
	spacer.SetMinSize(fyne.NewSize(70, 70))
	return container.NewStack(container.NewBorder(
		nil,
		nil,
		container.NewStack(container.NewCenter(spacer), canvas.NewImageFromImage(nil)),
		nil,
		container.NewBorder(
			nil,
			widget.NewLabel(""),
			nil,
			nil,
			widget.NewLabel(""))))
}

// Update func for listItem
func updateListItem(i widget.ListItemID, o fyne.CanvasObject, asset []NFAListing) {
	if i > len(asset)-1 {
		return
	}

	a := asset[i]
	header := fmt.Sprintf("%s   %s", a.Name, a.Collection)

	if o.(*fyne.Container).Objects[0].(*fyne.Container).Objects[0].(*fyne.Container).Objects[0].(*widget.Label).Text != header {
		o.(*fyne.Container).Objects[0].(*fyne.Container).Objects[0].(*fyne.Container).Objects[0].(*widget.Label).SetText(header)

		have, err := gnomes.StorageExists(a.Collection, a.Name)
		if err != nil {
			have = false
			logger.Errorln("[updateListItem]", err)
		}

		if have {
			var new Asset
			gnomes.GetStorage(a.Collection, a.Name, &new)
			if new.Image != nil && !bytes.Equal(new.Image, bundle.ResourceMarketCirclePng.StaticContent) {
				img := canvas.NewImageFromReader(bytes.NewReader(new.Image), a.Name)
				img.SetMinSize(fyne.NewSize(66, 66))
				o.(*fyne.Container).Objects[0].(*fyne.Container).Objects[1].(*fyne.Container).Objects[0].(*fyne.Container).Objects[0] = img
			} else {
				have = false
			}
		}

		if !have {
			if img, err := dreams.DownloadBytes(a.IconURL); err == nil {
				canv := canvas.NewImageFromReader(bytes.NewReader(img), a.Name)
				canv.SetMinSize(fyne.NewSize(66, 66))
				o.(*fyne.Container).Objects[0].(*fyne.Container).Objects[1].(*fyne.Container).Objects[0].(*fyne.Container).Objects[0] = canv
			} else {
				unknown := canvas.NewImageFromResource(bundle.ResourceMarketCirclePng)
				unknown.SetMinSize(fyne.NewSize(66, 66))
				o.(*fyne.Container).Objects[0].(*fyne.Container).Objects[1].(*fyne.Container).Objects[0].(*fyne.Container).Objects[0] = unknown
			}
		}
		o.(*fyne.Container).Objects[0].(*fyne.Container).Objects[0].(*fyne.Container).Objects[1].(*widget.Label).SetText(a.SCID)

		frame := canvas.NewImageFromResource(bundle.ResourceFramePng)
		frame.SetMinSize(fyne.NewSize(70, 70))
		o.(*fyne.Container).Objects[0].(*fyne.Container).Objects[1].(*fyne.Container).Objects[1] = frame
		o.Refresh()
	}
}

// Get NFA cover and icon images for scid and set market images
func GetNFAImages(scid string) {
	if gnomon.IsReady() && len(scid) == 64 {
		Market.Lock()
		Market.downloading = true
		defer func() {
			Market.Unlock()
			Market.downloading = false
		}()
		name, _ := gnomon.GetSCIDValuesByKey(scid, "nameHdr")
		icon, _ := gnomon.GetSCIDValuesByKey(scid, "iconURLHdr")
		cover, _ := gnomon.GetSCIDValuesByKey(scid, "coverURL")
		collection, _ := gnomon.GetSCIDValuesByKey(scid, "collection")
		if icon != nil && collection != nil {
			have, err := gnomes.StorageExists(collection[0], name[0])
			if err != nil {
				have = false
				logger.Errorln("[GetNFAImages]", err)
			}

			if have {
				var new Asset
				gnomes.GetStorage(collection[0], name[0], &new)
				if new.Image != nil && !bytes.Equal(new.Image, bundle.ResourceMarketCirclePng.StaticContent) {
					Market.Icon = *canvas.NewImageFromReader(bytes.NewReader(new.Image), name[0])
					Market.Details.Objects[0].(*fyne.Container).Objects[0].(*fyne.Container).Objects[1] = NFAIcon()
				} else {
					have = false
				}
			}

			if !have {
				if img, err := dreams.DownloadBytes(icon[0]); err == nil {
					Market.Icon = *canvas.NewImageFromReader(bytes.NewReader(img), name[0])
					Market.Details.Objects[0].(*fyne.Container).Objects[0].(*fyne.Container).Objects[1] = NFAIcon()
				} else {
					Market.Icon = *canvas.NewImageFromResource(bundle.ResourceMarketCirclePng)
					Market.Details.Objects[0].(*fyne.Container).Objects[0].(*fyne.Container).Objects[1] = NFAIcon()
				}
			}

			view := Market.Viewing.Collection
			if view == "AZY-Playing card decks" || view == "SIXPC" || view == "AZY-Playing card backs" || view == "SIXPCB" {
				Market.Details.Objects[0].(*fyne.Container).Objects[0].(*fyne.Container).Objects[2] = ToolsBadge()
			} else {
				Market.Details.Objects[0].(*fyne.Container).Objects[0].(*fyne.Container).Objects[2] = container.NewStack(layout.NewSpacer())
			}
		} else {
			Market.Icon = *canvas.NewImageFromResource(bundle.ResourceMarketCirclePng)
			Market.Details.Objects[0].(*fyne.Container).Objects[0].(*fyne.Container).Objects[1] = NFAIcon()
		}

		if cover != nil {
			Market.Cover, _ = dreams.DownloadCanvas(cover[0], name[0]+"-cover")
			if Market.Cover.Resource != nil {
				Market.Details.Objects[1].(*fyne.Container).Objects[0] = NFACoverImg()
				if Market.Loading != nil {
					Market.Loading.Stop()
				}
			} else {
				Market.Details.Objects[1].(*fyne.Container).Objects[0] = loadingBar()
			}
		} else {
			Market.Cover = *canvas.NewImageFromImage(nil)
			Market.Details.Objects[1].(*fyne.Container).Objects[0] = NFACoverImg()
		}
	}
}

// Returns NFA market icon image with frame
func NFAIcon() fyne.CanvasObject {
	Market.Icon.SetMinSize(fyne.NewSize(90, 90))
	border := container.NewBorder(layout.NewSpacer(), layout.NewSpacer(), layout.NewSpacer(), layout.NewSpacer(), &Market.Icon)

	frame := canvas.NewImageFromResource(bundle.ResourceFramePng)
	frame.SetMinSize(fyne.NewSize(100, 100))

	return container.NewStack(border, frame)
}

var toolsBadge = *canvas.NewImageFromResource(bundle.ResourceDReamToolsPng)

// Returns badge for dReam Tools enabled assets
func ToolsBadge() fyne.CanvasObject {
	toolsBadge.SetMinSize(fyne.NewSize(90, 90))
	border := container.NewBorder(layout.NewSpacer(), layout.NewSpacer(), layout.NewSpacer(), layout.NewSpacer(), &toolsBadge)

	frame := canvas.NewImageFromResource(bundle.ResourceFramePng)
	frame.SetMinSize(fyne.NewSize(100, 100))

	return container.NewStack(border, frame)
}

// Returns NFA cover image for market display
func NFACoverImg() fyne.CanvasObject {
	Market.Cover.SetMinSize(fyne.NewSize(400, 600))

	return container.NewCenter(&Market.Cover)
}

// Loading bar for NFA cover image
func loadingBar() fyne.CanvasObject {
	Market.Loading = widget.NewProgressBarInfinite()
	spacer := canvas.NewRectangle(color.Transparent)
	spacer.SetMinSize(fyne.NewSize(400, 600))
	Market.Loading.Start()

	return container.NewCenter(container.NewStack(Market.Loading, spacer), container.NewCenter(canvas.NewText("Loading...", bundle.TextColor)))
}

// Clears all market NFA images
func clearNFAImages() {
	Market.Lock()
	defer Market.Unlock()

	Market.Details.Objects[0].(*fyne.Container).Objects[0].(*fyne.Container).Objects[2] = layout.NewSpacer()
	Market.Icon = *canvas.NewImageFromImage(nil)

	Market.Details.Objects[1].(*fyne.Container).Objects[0] = canvas.NewImageFromImage(nil)
	Market.Cover = *canvas.NewImageFromImage(nil)
}

// Initialize market display objects
func NFAMarketInfo() fyne.Container {
	Market.Display.Name = widget.NewEntry()
	Market.Display.Type = widget.NewEntry()
	Market.Display.Collection = widget.NewEntry()
	Market.Display.Description = widget.NewMultiLineEntry()
	Market.Display.Creator = widget.NewEntry()
	Market.Display.Artificer = widget.NewEntry()
	Market.Display.Royalty = widget.NewEntry()
	Market.Display.Price = widget.NewEntry()
	Market.Display.Owner = widget.NewEntry()
	Market.Display.Update = widget.NewEntry()
	Market.Display.Bid.Current = widget.NewEntry()
	Market.Display.Bid.Price = widget.NewEntry()
	Market.Display.Bid.Count = widget.NewEntry()
	Market.Display.Ends = widget.NewEntry()

	Market.Display.Name.Disable()
	Market.Display.Type.Disable()
	Market.Display.Collection.Disable()
	Market.Display.Description.Disable()
	Market.Display.Creator.Disable()
	Market.Display.Artificer.Disable()
	Market.Display.Royalty.Disable()
	Market.Display.Price.Disable()
	Market.Display.Owner.Disable()
	Market.Display.Update.Disable()
	Market.Display.Bid.Current.Disable()
	Market.Display.Bid.Price.Disable()
	Market.Display.Bid.Count.Disable()
	Market.Display.Ends.Disable()

	Market.Icon.SetMinSize(fyne.NewSize(94, 94))
	Market.Cover.SetMinSize(fyne.NewSize(400, 600))

	Market.Display.Description.Wrapping = fyne.TextWrapWord

	return AuctionInfo()
}

// Returns container for auction display objects
func AuctionInfo() fyne.Container {
	auction_form := []*widget.FormItem{}
	auction_form = append(auction_form, widget.NewFormItem("Name", Market.Display.Name))
	auction_form = append(auction_form, widget.NewFormItem("Asset Type", Market.Display.Type))
	auction_form = append(auction_form, widget.NewFormItem("Collection", Market.Display.Collection))
	auction_form = append(auction_form, widget.NewFormItem("Ends", Market.Display.Ends))
	auction_form = append(auction_form, widget.NewFormItem("Bids", Market.Display.Bid.Count))
	auction_form = append(auction_form, widget.NewFormItem("Description", Market.Display.Description))
	auction_form = append(auction_form, widget.NewFormItem("Creator", Market.Display.Creator))
	auction_form = append(auction_form, widget.NewFormItem("Owner", Market.Display.Owner))
	auction_form = append(auction_form, widget.NewFormItem("Artificer %", Market.Display.Artificer))
	auction_form = append(auction_form, widget.NewFormItem("Royalty %", Market.Display.Royalty))

	auction_form = append(auction_form, widget.NewFormItem("Owner Update", Market.Display.Update))
	auction_form = append(auction_form, widget.NewFormItem("Start Price", Market.Display.Price))

	auction_form = append(auction_form, widget.NewFormItem("Current Bid", Market.Display.Bid.Current))

	form_spacer := canvas.NewRectangle(color.Transparent)
	form_spacer.SetMinSize(fyne.NewSize(330, 0))
	auction_form = append(auction_form, widget.NewFormItem("", container.NewStack(form_spacer)))

	form := widget.NewForm(auction_form...)

	padding := canvas.NewRectangle(color.Transparent)
	padding.SetMinSize(fyne.NewSize(123, 0))

	return *container.NewHBox(
		container.NewVBox(
			container.NewHBox(padding, NFAIcon(), layout.NewSpacer()),
			form),
		container.NewCenter(layout.NewSpacer()))
}

// Reset auction display content to default values
func ResetAuctionInfo() {
	Market.amount.bid = 0
	clearNFAImages()
	Market.Display.Name.SetText("")
	Market.Display.Type.SetText("")
	Market.Display.Collection.SetText("")
	Market.Display.Description.SetText("")
	Market.Display.Creator.SetText("")
	Market.Display.Artificer.SetText("")
	Market.Display.Royalty.SetText("")
	Market.Display.Price.SetText("")
	Market.Display.Owner.SetText("")
	Market.Display.Update.SetText("")
	Market.Display.Bid.Current.SetText("")
	Market.Display.Bid.Price.SetText("")
	Market.Display.Bid.Count.SetText("")
	Market.Display.Ends.SetText("")
}

// Returns container for unlisted display objects
func NotListedInfo() fyne.Container {
	unlisted_form := []*widget.FormItem{}
	unlisted_form = append(unlisted_form, widget.NewFormItem("Name", Market.Display.Name))
	unlisted_form = append(unlisted_form, widget.NewFormItem("Asset Type", Market.Display.Type))
	unlisted_form = append(unlisted_form, widget.NewFormItem("Collection", Market.Display.Collection))
	unlisted_form = append(unlisted_form, widget.NewFormItem("Description", Market.Display.Description))
	unlisted_form = append(unlisted_form, widget.NewFormItem("Creator", Market.Display.Creator))
	unlisted_form = append(unlisted_form, widget.NewFormItem("Owner", Market.Display.Owner))
	unlisted_form = append(unlisted_form, widget.NewFormItem("Artificer %", Market.Display.Artificer))
	unlisted_form = append(unlisted_form, widget.NewFormItem("Royalty %", Market.Display.Royalty))

	unlisted_form = append(unlisted_form, widget.NewFormItem("Owner Update", Market.Display.Update))

	form_spacer := canvas.NewRectangle(color.Transparent)
	form_spacer.SetMinSize(fyne.NewSize(330, 0))
	unlisted_form = append(unlisted_form, widget.NewFormItem("", container.NewStack(form_spacer)))

	form := widget.NewForm(unlisted_form...)

	padding := canvas.NewRectangle(color.Transparent)
	padding.SetMinSize(fyne.NewSize(123, 0))

	return *container.NewHBox(
		container.NewVBox(
			container.NewHBox(padding, NFAIcon(), layout.NewSpacer()),
			form),
		container.NewCenter(layout.NewSpacer()))
}

// Reset unlisted NFA display content to default values
func ResetNotListedInfo() {
	Market.amount.bid = 0
	clearNFAImages()
	Market.Display.Name.SetText("")
	Market.Display.Type.SetText("")
	Market.Display.Collection.SetText("")
	Market.Display.Description.SetText("")
	Market.Display.Creator.SetText("")
	Market.Display.Artificer.SetText("")
	Market.Display.Royalty.SetText("")
	Market.Display.Owner.SetText("")
	Market.Display.Update.SetText("")
}

// Returns container for NFA buy now display objects
func BuyNowInfo() fyne.Container {
	buy_form := []*widget.FormItem{}
	buy_form = append(buy_form, widget.NewFormItem("Name", Market.Display.Name))
	buy_form = append(buy_form, widget.NewFormItem("Asset Type", Market.Display.Type))
	buy_form = append(buy_form, widget.NewFormItem("Collection", Market.Display.Collection))
	buy_form = append(buy_form, widget.NewFormItem("Ends", Market.Display.Ends))
	buy_form = append(buy_form, widget.NewFormItem("Description", Market.Display.Description))
	buy_form = append(buy_form, widget.NewFormItem("Creator", Market.Display.Creator))
	buy_form = append(buy_form, widget.NewFormItem("Owner", Market.Display.Owner))
	buy_form = append(buy_form, widget.NewFormItem("Artificer %", Market.Display.Artificer))
	buy_form = append(buy_form, widget.NewFormItem("Royalty %", Market.Display.Royalty))

	buy_form = append(buy_form, widget.NewFormItem("Owner Update", Market.Display.Update))
	buy_form = append(buy_form, widget.NewFormItem("Price", Market.Display.Price))

	form_spacer := canvas.NewRectangle(color.Transparent)
	form_spacer.SetMinSize(fyne.NewSize(330, 0))

	buy_form = append(buy_form, widget.NewFormItem("", container.NewStack(form_spacer)))

	form := widget.NewForm(buy_form...)

	padding := canvas.NewRectangle(color.Transparent)
	padding.SetMinSize(fyne.NewSize(123, 0))

	return *container.NewHBox(
		container.NewVBox(
			container.NewHBox(padding, NFAIcon(), layout.NewSpacer()),
			form),
		container.NewCenter(layout.NewSpacer()))
}

// Reset buy now display content to default values
func ResetBuyInfo() {
	Market.amount.buy = 0
	clearNFAImages()
	Market.Display.Name.SetText("")
	Market.Display.Type.SetText("")
	Market.Display.Collection.SetText("")
	Market.Display.Description.SetText("")
	Market.Display.Creator.SetText("")
	Market.Display.Artificer.SetText("")
	Market.Display.Royalty.SetText("")
	Market.Display.Price.SetText("")
	Market.Display.Owner.SetText("")
	Market.Display.Update.SetText("")
	Market.Display.Ends.SetText("")
}

// Place NFA market layout
func PlaceMarket(d *dreams.AppObject) *container.Split {
	auction_info := NFAMarketInfo()

	buy_info := BuyNowInfo()

	not_listed_info := NotListedInfo()

	Market.Button.Cancel = widget.NewButton("Cancel", func() {
		if len(Market.Viewing.Asset) == 64 {
			Market.Button.Cancel.Hide()
			ConfirmCancelClose(Market.Viewing.Asset, false, d)
		} else {
			dialog.NewInformation("NFA Cancel", "SCID error", d.Window).Show()
		}
	})
	Market.Button.Cancel.Importance = widget.HighImportance
	Market.Button.Cancel.Hide()

	Market.Button.Close = widget.NewButton("Close", func() {
		if len(Market.Viewing.Asset) == 64 {
			Market.Button.Close.Hide()
			ConfirmCancelClose(Market.Viewing.Asset, true, d)
		} else {
			dialog.NewInformation("NFA Close", "SCID error", d.Window).Show()
		}
	})
	Market.Button.Close.Importance = widget.HighImportance
	Market.Button.Close.Hide()

	var market *container.Split

	// Search NFA object
	var dest_addr, searching string
	var message_button *widget.Button
	var scids = make(map[string]string)
	search_entry := xwidget.NewCompletionEntry([]string{})
	search_entry.Wrapping = fyne.TextWrap(fyne.TextTruncateClip)
	search_entry.OnChanged = func(s string) {
		if Market.downloading {
			search_entry.SetText(searching)
			return
		}

		scid := scids[s]
		if len(scid) == 64 {
			if scid != Market.Viewing.Asset {
				searching = s
				Market.Entry.SetText("0.0")
				Market.Viewing.Asset = scid
				list, addr := CheckNFAListingType(scid)
				dest_addr = addr
				switch list {
				case 1:
					Market.Details = auction_info
					Market.Tab = "Auction"
					Market.Button.BidBuy.Text = "Bid"
					Market.Entry.Enable()
					Market.Entry.Show()
					ResetAuctionInfo()
					go GetNFAImages(scid)
					go GetAuctionDetails(scid)
				case 2:
					Market.Details = buy_info
					Market.Tab = "Buy"
					Market.Button.BidBuy.Text = "Buy"
					Market.Entry.Disable()
					Market.Entry.Show()
					ResetBuyInfo()
					go GetNFAImages(scid)
					go GetBuyNowDetails(scid)
				default:
					Market.Details = not_listed_info
					Market.Tab = "Buy"
					Market.Entry.Disable()
					Market.Entry.Hide()
					ResetNotListedInfo()
					go GetNFAImages(scid)
					go GetUnlistedDetails(scid)
				}

				Market.Details.Objects[1].(*fyne.Container).Objects[0] = loadingBar()
				market.SetOffset(0.62)
			}
		} else {
			dest_addr = ""
		}

		if len(dest_addr) == 66 {
			message_button.Show()
		} else {
			message_button.Hide()
		}
	}

	search_by := widget.NewRadioGroup([]string{"Collection", "Name", "Description", "SCID"}, nil)
	search_by.Horizontal = true
	search_by.SetSelected("Collection")
	search_by.Required = true

	search_button := widget.NewButtonWithIcon("", dreams.FyneIcon("search"), func() {
		if search_entry.Text != "" && rpc.Wallet.IsConnected() {
			var i int
			switch search_by.Selected {
			case "Collection":
				i = 0
			case "Name":
				i = 1
			case "Description":
				i = 2
			case "SCID":
				if len(search_entry.Text) != 64 {
					dialog.NewInformation("Search", "Not a valid SCID", d.Window).Show()
					return
				}
				i = 3
			}

			if scids = SearchNFAsBy(i, search_entry.Text); scids == nil || len(scids) < 1 {
				dialog.NewInformation("No results", "Nothing found", d.Window).Show()
			} else {
				var showing []string
				for k := range scids {
					showing = append(showing, k)
				}
				sort.Strings(showing)
				search_entry.SetOptions(showing)
				search_entry.ShowCompletion()
			}
		}
	})

	clear_button := widget.NewButtonWithIcon("", dreams.FyneIcon("searchReplace"), func() {
		search_entry.SetOptions([]string{})
		search_entry.SetText("")
	})
	clear_button.Importance = widget.LowImportance

	show_results := widget.NewButtonWithIcon("", dreams.FyneIcon("arrowDropDown"), func() {
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

	// NFA List objects
	Market.List.Auction = widget.NewList(
		func() int {
			return len(Market.Auctions)
		},
		listItem,
		func(i widget.ListItemID, o fyne.CanvasObject) {
			go updateListItem(i, o, Market.Auctions)
		})

	Market.List.Buy = widget.NewList(
		func() int {
			return len(Market.Buys)
		},
		listItem,
		func(i widget.ListItemID, o fyne.CanvasObject) {
			go updateListItem(i, o, Market.Buys)
		})

	Market.List.Wallet = widget.NewList(
		func() int {
			return len(Market.Wallets)
		},
		listItem,
		func(i widget.ListItemID, o fyne.CanvasObject) {
			updateListItem(i, o, Market.Wallets)
		})

	tabs := container.NewAppTabs(
		container.NewTabItemWithIcon("", bundle.ResourceMarketCirclePng, layout.NewSpacer()),
		container.NewTabItem("Auctions", Market.List.Auction),
		container.NewTabItem("Buy Now", Market.List.Buy),
		container.NewTabItem("My Listings", container.NewBorder(
			nil,
			container.NewBorder(nil, nil, Market.Button.Close, Market.Button.Cancel),
			nil,
			nil,
			Market.List.Wallet)),

		container.NewTabItem("Search", container.NewBorder(
			search_cont,
			container.NewHBox(layout.NewSpacer(), message_button),
			nil,
			nil,
			container.NewHBox())))

	tabs.DisableIndex(0)
	tabs.SetTabLocation(container.TabLocationTop)
	tabs.OnSelected = func(ti *container.TabItem) {
		Market.Viewing.Asset = ""
		Market.Viewing.Collection = ""
		Market.Entry.SetText("0.0")
		switch ti.Text {
		case "Auctions":
			go FindNFAListings(nil, nil)
			Market.Tab = "Auction"
			Market.List.Auction.UnselectAll()
			Market.List.Wallet.UnselectAll()
			Market.Button.BidBuy.Text = "Bid"
			Market.Button.BidBuy.Refresh()
			Market.Entry.Show()
			Market.Entry.Enable()
			ResetAuctionInfo()
			Market.Details = auction_info
		case "Buy Now":
			go FindNFAListings(nil, nil)
			Market.Tab = "Buy"
			Market.List.Buy.UnselectAll()
			Market.List.Wallet.UnselectAll()
			Market.Button.BidBuy.Text = "Buy"
			Market.Button.BidBuy.Refresh()
			Market.Entry.Show()
			Market.Entry.Disable()
			ResetBuyInfo()
			Market.Details = buy_info
		case "My Listings":
			go FindNFAListings(nil, nil)
			Market.Tab = "Buy"
			Market.List.Auction.UnselectAll()
			Market.List.Wallet.UnselectAll()
			Market.List.Buy.UnselectAll()
			Market.Entry.Hide()
			Market.Entry.Disable()
			ResetBuyInfo()
			Market.Details = not_listed_info
		case "Search":
			Market.Tab = "Buy"
			Market.List.Auction.UnselectAll()
			Market.List.Wallet.UnselectAll()
			Market.List.Buy.UnselectAll()
			Market.Entry.Hide()
			Market.Entry.Disable()
			ResetBuyInfo()
			Market.Details = not_listed_info
		}

		Market.Button.Close.Hide()
		Market.Button.Cancel.Hide()
		Market.Button.BidBuy.Hide()
	}

	Market.Tab = "Auction"

	scroll_top := widget.NewButtonWithIcon("", fyne.Theme.Icon(fyne.CurrentApp().Settings().Theme(), "arrowUp"), func() {
		switch Market.Tab {
		case "Buy":
			Market.List.Buy.ScrollToTop()
		case "Auction":
			Market.List.Auction.ScrollToTop()
		default:

		}
	})

	scroll_bottom := widget.NewButtonWithIcon("", fyne.Theme.Icon(fyne.CurrentApp().Settings().Theme(), "arrowDown"), func() {
		switch Market.Tab {
		case "Buy":
			Market.List.Buy.ScrollToBottom()
		case "Auction":
			Market.List.Auction.ScrollToBottom()
		default:

		}
	})

	scroll_top.Importance = widget.LowImportance
	scroll_bottom.Importance = widget.LowImportance

	scroll_cont := container.NewVBox(container.NewHBox(layout.NewSpacer(), scroll_top, scroll_bottom))

	min_size := bundle.Alpha120
	min_size.SetMinSize(fyne.NewSize(420, 0))

	max := container.NewStack(min_size, tabs, scroll_cont)

	Market.Details = auction_info

	// Market action button
	Market.Button.BidBuy = widget.NewButton("Bid", func() {
		scid := Market.Viewing.Asset
		if len(scid) == 64 {
			text := Market.Button.BidBuy.Text
			Market.Button.BidBuy.Hide()
			if text == "Bid" {
				amt := rpc.ToAtomic(Market.Entry.Text, 5)
				BidBuyConfirm(scid, amt, true, d)
			} else if text == "Buy" {
				BidBuyConfirm(scid, Market.amount.buy, false, d)
			}
		} else {
			dialog.NewInformation("Market", "SCID error", d.Window).Show()
		}
	})
	Market.Button.BidBuy.Importance = widget.HighImportance
	Market.Button.BidBuy.Hide()

	entry_spacer := canvas.NewRectangle(color.Transparent)
	entry_spacer.SetMinSize(fyne.NewSize(210, 0))

	button_spacer := canvas.NewRectangle(color.Transparent)
	button_spacer.SetMinSize(fyne.NewSize(40, 0))

	// Market amount entry
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

	// NFA actions container
	Market.Actions = *container.NewBorder(nil, nil, nil, container.NewStack(button_spacer, Market.Button.BidBuy), Market.Entry)
	Market.Actions.Hide()

	padding := canvas.NewRectangle(color.Transparent)
	padding.SetMinSize(fyne.NewSize(123, 0))

	details := container.NewBorder(nil, container.NewHBox(padding, container.NewStack(entry_spacer, &Market.Actions)), nil, nil, &Market.Details)

	market = container.NewHSplit(details, max)
	market.SetOffset(0.66)

	// Selected indexes
	var buy_index widget.ListItemID
	var wallet_index widget.ListItemID
	var auction_index widget.ListItemID

	// Lists get images and details for Market objects on selected if not downloading already
	Market.List.Auction.OnSelected = func(id widget.ListItemID) {
		if Market.downloading && id != auction_index {
			Market.List.Buy.Select(auction_index)
			return
		}

		scid := Market.Auctions[id].SCID
		if scid != Market.Viewing.Asset {
			auction_index = id
			Market.Entry.SetText("")
			clearNFAImages()
			Market.Viewing.Asset = scid
			go GetNFAImages(scid)
			go GetAuctionDetails(scid)
			Market.Details.Objects[1].(*fyne.Container).Objects[0] = loadingBar()
			market.SetOffset(0.62)
		}
	}

	Market.List.Buy.OnSelected = func(id widget.ListItemID) {
		if Market.downloading && id != buy_index {
			Market.List.Buy.Select(buy_index)
			return
		}

		scid := Market.Buys[id].SCID
		if scid != Market.Viewing.Asset {
			buy_index = id
			clearNFAImages()
			Market.Viewing.Asset = scid
			go GetNFAImages(scid)
			go GetBuyNowDetails(scid)
			Market.Details.Objects[1].(*fyne.Container).Objects[0] = loadingBar()
			market.SetOffset(0.62)
		}
	}

	Market.List.Wallet.OnSelected = func(id widget.ListItemID) {
		if Market.downloading && id != wallet_index {
			Market.List.Buy.Select(wallet_index)
			return
		}

		scid := Market.Wallets[id].SCID
		if scid != Market.Viewing.Asset {
			wallet_index = id
			clearNFAImages()
			Market.Viewing.Asset = scid
			go GetNFAImages(scid)
			go GetUnlistedDetails(scid)
			Market.Details.Objects[1].(*fyne.Container).Objects[0] = loadingBar()
			market.SetOffset(0.62)
		}
	}

	Control.Lock()
	Control.Dapps["NFA Market"] = true
	Control.Unlock()

	go RunNFAMarket(d, max)

	return market
}

// Splash screen for when listings lists syncing
func syncScreen() (max *fyne.Container, bar *widget.ProgressBar) {
	text := canvas.NewText("Syncing...", color.White)
	text.Alignment = fyne.TextAlignCenter
	text.TextSize = 21

	img := canvas.NewImageFromResource(bundle.ResourceMarketCirclePng)
	img.SetMinSize(fyne.NewSize(150, 150))

	bar = widget.NewProgressBar()
	bar.TextFormatter = func() string {
		return ""
	}

	max = container.NewBorder(
		dwidget.LabelColor(container.NewVBox(widget.NewLabel(""))),
		nil,
		nil,
		nil,
		container.NewCenter(img, text), bar)

	return
}

// Routine for NFA market, finds listings and disables controls, see PlaceAssets() and PlaceMarket() layouts
func RunNFAMarket(d *dreams.AppObject, cont *fyne.Container) {
	offset := 0
	synced := false
	for {
		select {
		case <-d.Receive(): // do on interval
			if !rpc.Wallet.IsConnected() || !rpc.Daemon.IsConnected() {
				Market.Auctions = []NFAListing{}
				Market.Buys = []NFAListing{}
				Market.Button.BidBuy.Hide()
				Market.List.Auction.UnselectAll()
				Market.List.Wallet.UnselectAll()
				Market.List.Buy.UnselectAll()
				Market.Viewing.Collection = ""
				Market.Viewing.Asset = ""
				ResetAuctionInfo()
				synced = false
				d.WorkDone()
				continue
			}

			if !synced && gnomes.Scan(d.IsConfiguring()) {
				cont.Objects[2].(*fyne.Container).Hide()
				reset := cont.Objects[1]
				screen, bar := syncScreen()
				cont.Objects[1] = screen
				logger.Println("[NFA Market] Syncing")
				FindNFAListings(nil, bar)
				synced = true
				cont.Objects[1] = reset
				cont.Objects[2].(*fyne.Container).Show()
			}

			// If connected daemon connected start looking for Gnomon sync with daemon
			if rpc.Daemon.IsConnected() && gnomon.IsRunning() {
				// Enable index controls and check if wallet is connected
				DisableIndexControls(false)
				if rpc.Wallet.IsConnected() {
					Market.Actions.Show()
					Assets.Claim.Show()
					if d.OnSubTab("Market") {
						// Update live market info
						if len(Market.Viewing.Asset) == 64 {
							if Market.Tab == "Buy" {
								GetBuyNowDetails(Market.Viewing.Asset)
							} else {
								GetAuctionDetails(Market.Viewing.Asset)
							}
						}
					}

					if gnomon.IsSynced() {
						if offset%5 == 0 {
							if d.OnSubTab("Market") {
								FindNFAListings(nil, nil)
								if gnomon.DBStorageType() == "boltdb" {
									for _, r := range Market.Auctions {
										gnomes.StoreBolt(r.Collection, r.Name, r)
									}

									for _, r := range Market.Buys {
										gnomes.StoreBolt(r.Collection, r.Name, r)
									}
								}
							}
						}
					}
				} else {
					Market.Actions.Hide()
					Assets.Button.List.Hide()
					Assets.Button.Send.Hide()
					Assets.Button.Rescan.Hide()
					Assets.Claim.Hide()
					Assets.Asset = []Asset{}
				}
				Info.RefreshIndexed()
			} else {
				DisableIndexControls(true)
				Assets.Asset = []Asset{}
			}

			offset++
			if offset > 19 {
				offset = 0
			}

			d.WorkDone()

		case <-d.CloseDapp(): // exit
			logger.Println("[NFA Market] Done")
			return
		}
	}
}

// Get search filters from on chain store
func GetFilters(check string) (filter []string) {
	if stored, ok := rpc.GetStringKey(rpc.RatingSCID, check, rpc.Daemon.Rpc).(string); ok {
		if h, err := hex.DecodeString(stored); err == nil {
			if err = json.Unmarshal(h, &filter); err != nil {
				logger.Errorln("[GetFilters]", check, err)
			}
		}
	} else {
		logger.Errorln("[GetFilters] Could not get", check)
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
func checkNFAAuctionListing(scid string) (asset NFAListing, owned, expired bool) {
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
							asset.Name = header[0]
							asset.Collection = coll[0]
							asset.Description = desc_check
							asset.SCID = scid
							if icon, _ := gnomon.GetSCIDValuesByKey(scid, "iconURLHdr"); icon != nil {
								asset.IconURL = icon[0]
							}

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
							asset.Name = header[0]
							asset.Collection = coll[0]
							asset.Description = desc_check
							asset.SCID = scid
							if icon, _ := gnomon.GetSCIDValuesByKey(scid, "iconURLHdr"); icon != nil {
								asset.IconURL = icon[0]
							}

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
func checkNFABuyListing(scid string) (asset NFAListing, owned, expired bool) {
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
							asset.Name = header[0]
							asset.Collection = coll[0]
							asset.Description = desc_check
							asset.SCID = scid
							if icon, _ := gnomon.GetSCIDValuesByKey(scid, "iconURLHdr"); icon != nil {
								asset.IconURL = icon[0]
							}

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
							asset.Name = header[0]
							asset.Collection = coll[0]
							asset.Description = desc_check
							asset.SCID = scid
							if icon, _ := gnomon.GetSCIDValuesByKey(scid, "iconURLHdr"); icon != nil {
								asset.IconURL = icon[0]
							}

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
//   - by 0 is collection, 1 is name, 2 is description, 3 is scid
func SearchNFAsBy(by int, prefix string) (results map[string]string) {
	if gnomon.IsReady() {
		assets := gnomon.GetAllOwnersAndSCIDs()
		results = make(map[string]string)
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
									asset := coll[0] + "   " + name[0] + "   " + desc_check
									results[asset] = sc
								}
							case 1:
								if strings.HasPrefix(name[0], prefix) {
									desc_check := TrimStringLen(desc[0], 66)
									asset := coll[0] + "   " + name[0] + "   " + desc_check
									results[asset] = sc
								}
							case 2:
								if strings.Contains(desc[0], prefix) {
									desc_check := TrimStringLen(desc[0], 66)
									asset := coll[0] + "   " + name[0] + "   " + desc_check
									results[asset] = sc
								}
							case 3:
								if sc == prefix {
									desc_check := TrimStringLen(desc[0], 66)
									asset := coll[0] + "   " + name[0] + "   " + desc_check
									results[asset] = sc
									break
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

// Get auction details for current asset and set display objects
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
				Market.Viewing.Collection = collection[0]

				Market.Display.Name.SetText(name[0])

				Market.Display.Type.SetText(AssetType(collection[0], typeHdr[0]))

				Market.Display.Collection.SetText(collection[0])

				Market.Display.Description.SetText(description[0])

				if Market.Display.Creator.Text != creator[0] {
					Market.Display.Creator.SetText(creator[0])
				}

				if Market.Display.Owner.Text != owner[0] {
					Market.Display.Owner.SetText(owner[0])
				}
				if owner_update[0] == 1 {
					Market.Display.Update.SetText("Yes")
				} else {
					Market.Display.Update.SetText("No")
				}

				Market.Display.Artificer.SetText(strconv.Itoa(int(artFee[0])) + "%")

				Market.Display.Royalty.SetText(strconv.Itoa(int(royalty[0])) + "%")

				price := float64(start[0])
				str := fmt.Sprintf("%.5f", price/100000)
				Market.Display.Price.SetText(str + " Dero")

				Market.Display.Bid.Count.SetText(strconv.Itoa(int(bids[0])))

				end, _ := rpc.MsToTime(strconv.Itoa(int(endTime[0]) * 1000))
				Market.Display.Ends.SetText(end.String())

				if current != nil {
					value := float64(current[0])
					str := fmt.Sprintf("%.5f", value/100000)
					Market.Display.Bid.Current.SetText(str)
				} else {
					Market.Display.Bid.Current.SetText("")
				}

				if bid_price != nil {
					value := float64(bid_price[0])
					str := fmt.Sprintf("%.5f", value/100000)
					if bid_price[0] == 0 {
						Market.amount.bid = start[0]
					} else {
						Market.amount.bid = bid_price[0]
					}
					Market.Display.Bid.Price.SetText(str)
				} else {
					Market.amount.bid = 0
					Market.Display.Bid.Price.SetText("")
				}

				if amt, err := strconv.ParseFloat(Market.Entry.Text, 64); err == nil {
					value := float64(Market.amount.bid) / 100000
					if amt == 0 || amt < value {
						amt := fmt.Sprintf("%.5f", value)
						Market.Entry.SetText(amt)
					}
				}

				now := uint64(time.Now().Unix())
				if owner[0] == rpc.Wallet.Address {
					if now < startTime[0]+300 && startTime[0] > 0 && !Market.Confirming {
						Market.Button.Cancel.Show()
					} else {
						Market.Button.Cancel.Hide()
					}

					if now > endTime[0] && endTime[0] > 0 && !Market.Confirming {
						Market.Button.Close.Show()
					} else {
						Market.Button.Close.Hide()
					}
				} else {
					Market.Button.Close.Hide()
					Market.Button.Cancel.Hide()
				}

				Market.Button.BidBuy.Show()
				if now > endTime[0] || Market.Confirming {
					Market.Button.BidBuy.Hide()
				}
			}()
		}
	}
}

// Get buy now details for current asset and set display objects
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
				Market.Viewing.Collection = collection[0]

				Market.Display.Name.SetText(name[0])

				Market.Display.Type.SetText(AssetType(collection[0], typeHdr[0]))

				Market.Display.Collection.SetText(collection[0])

				Market.Display.Description.SetText(description[0])

				if Market.Display.Creator.Text != creator[0] {
					Market.Display.Creator.SetText(creator[0])
				}

				if Market.Display.Owner.Text != owner[0] {
					Market.Display.Owner.SetText(owner[0])
				}

				if owner_update[0] == 1 {
					Market.Display.Update.SetText("Yes")
				} else {
					Market.Display.Update.SetText("No")
				}

				Market.Display.Artificer.SetText(strconv.Itoa(int(artFee[0])) + "%")

				Market.Display.Royalty.SetText(strconv.Itoa(int(royalty[0])) + "%")

				Market.amount.buy = start[0]
				value := float64(start[0])
				str := fmt.Sprintf("%.5f", value/100000)
				Market.Display.Price.SetText(str + " Dero")

				Market.Entry.SetText(str)
				Market.Entry.Disable()
				end, _ := rpc.MsToTime(strconv.Itoa(int(endTime[0]) * 1000))
				Market.Display.Ends.SetText(end.String())

				now := uint64(time.Now().Unix())
				if owner[0] == rpc.Wallet.Address {
					if now < startTime[0]+300 && startTime[0] > 0 && !Market.Confirming {
						Market.Button.Cancel.Show()
					} else {
						Market.Button.Cancel.Hide()
					}

					if now > endTime[0] && endTime[0] > 0 && !Market.Confirming {
						Market.Button.Close.Show()
					} else {
						Market.Button.Close.Hide()
					}
				} else {
					Market.Button.Close.Hide()
					Market.Button.Cancel.Hide()
				}

				Market.Button.BidBuy.Show()
				if now > endTime[0] || Market.Confirming {
					Market.Button.BidBuy.Hide()
				}
			}()
		}
	}
}

// Get details for unlisted NFA and set display objects
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
				Market.Viewing.Collection = collection[0]

				Market.Display.Name.SetText(name[0])

				Market.Display.Type.SetText(AssetType(collection[0], typeHdr[0]))

				Market.Display.Collection.SetText(collection[0])

				Market.Display.Description.SetText(description[0])

				if Market.Display.Creator.Text != creator[0] {
					Market.Display.Creator.SetText(creator[0])
				}

				if Market.Display.Owner.Text != owner[0] {
					Market.Display.Owner.SetText(owner[0])
				}

				if owner_update[0] == 1 {
					Market.Display.Update.SetText("Yes")
				} else {
					Market.Display.Update.SetText("No")
				}

				Market.Display.Artificer.SetText(strconv.Itoa(int(artFee[0])) + "%")

				Market.Display.Royalty.SetText(strconv.Itoa(int(royalty[0])) + "%")

				Market.Entry.SetText("0")
				Market.Entry.Disable()

				now := uint64(time.Now().Unix())
				if owner[0] == rpc.Wallet.Address {
					if now < startTime[0]+300 && startTime[0] > 0 && !Market.Confirming {
						Market.Button.Cancel.Show()
					} else {
						Market.Button.Cancel.Hide()
					}

					if now > endTime[0] && endTime[0] > 0 && !Market.Confirming {
						Market.Button.Close.Show()
					} else {
						Market.Button.Close.Hide()
					}
				} else {
					Market.Button.Close.Hide()
					Market.Button.Cancel.Hide()
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

// Scan index for any active auction and buy now NFA listings
//   - Pass scids from db store, can be nil arg to check all from db
//   - progress widget to update ui
func FindNFAListings(scids map[string]string, progress *widget.ProgressBar) {
	if Market.Searching || !gnomon.HasChecked() {
		return
	}

	Market.Searching = true
	defer func() {
		Market.Searching = false
	}()

	if gnomon.IsReady() && rpc.IsReady() {
		auction := []NFAListing{}
		buy_now := []NFAListing{}
		my_list := []NFAListing{}
		if scids == nil {
			scids = gnomon.GetAllOwnersAndSCIDs()
		}

		if progress != nil {
			progress.Max = float64(len(scids))
		}

		for sc := range scids {
			if !gnomon.IsRunning() {
				return
			}

			a, owned, expired := checkNFAAuctionListing(sc)

			if a.Name != "" && !expired {
				auction = append(auction, a)
			}

			if owned {
				my_list = append(my_list, a)
			}

			b, owned, expired := checkNFABuyListing(sc)

			if b.Name != "" && !expired {
				buy_now = append(buy_now, b)
			}

			if owned {
				my_list = append(my_list, b)
			}

			if progress != nil {
				progress.SetValue(progress.Value + 1)
			}
		}

		if !gnomon.IsRunning() {
			return
		}

		Market.Auctions = auction
		Market.SortAuctions()
		Market.Buys = buy_now
		Market.SortBuys()
		Market.Wallets = my_list
		Market.SortMyList()
	}
}

// Check if wallet owns any indexed NFAs
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
						if owner[0] == rpc.Wallet.Address && owner[0] != "" {
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
