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
//   - Confirmation string from Market.Tab
//   - Pass main window obj to reset to
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

// NFA auction listings object
//   - Gets images and details for Market objects on selected
func AuctionListings() fyne.Widget {
	Market.List.Auction = widget.NewList(
		func() int {
			return len(Market.Auctions)
		},
		listItem,
		func(i widget.ListItemID, o fyne.CanvasObject) {
			go updateListItem(i, o, Market.Auctions)
		})

	Market.List.Auction.OnSelected = func(id widget.ListItemID) {
		scid := Market.Auctions[id].SCID
		if scid != Market.Viewing.Asset {
			Market.Entry.SetText("")
			clearNFAImages()
			Market.Viewing.Asset = scid
			go GetNFAImages(scid)
			go GetAuctionDetails(scid)
			Market.Details.Objects[1] = loadingBar()
		}
	}

	return Market.List.Auction
}

// NFA buy now listings object
//   - Gets images and details for Market objects on selected
func BuyNowListings() fyne.Widget {
	Market.List.Buy = widget.NewList(
		func() int {
			return len(Market.Buys)
		},
		listItem,
		func(i widget.ListItemID, o fyne.CanvasObject) {
			go updateListItem(i, o, Market.Buys)
		})

	Market.List.Buy.OnSelected = func(id widget.ListItemID) {
		scid := Market.Buys[id].SCID
		if scid != Market.Viewing.Asset {
			clearNFAImages()
			Market.Viewing.Asset = scid
			go GetNFAImages(scid)
			go GetBuyNowDetails(scid)
			Market.Details.Objects[1] = loadingBar()
		}
	}

	return Market.List.Buy
}

// NFA listing for connected wallet
//   - Gets images and details for Market objects on selected
func MyNFAListings() fyne.Widget {
	Market.List.Wallet = widget.NewList(
		func() int {
			return len(Market.Wallets)
		},
		listItem,
		func(i widget.ListItemID, o fyne.CanvasObject) {
			updateListItem(i, o, Market.Wallets)
		})

	Market.List.Wallet.OnSelected = func(id widget.ListItemID) {
		scid := Market.Wallets[id].SCID
		if scid != Market.Viewing.Asset {
			clearNFAImages()
			Market.Viewing.Asset = scid
			go GetNFAImages(scid)
			go GetUnlistedDetails(scid)
			Market.Details.Objects[1] = loadingBar()
		}
	}

	return Market.List.Wallet
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
			if split[3] != Market.Viewing.Asset {
				Market.Viewing.Asset = split[3]
				list, addr := CheckNFAListingType(split[3])
				dest_addr = addr
				switch list {
				case 1:
					Market.Tab = "Auction"
					Market.Button.BidBuy.Text = "Bid"
					Market.Entry.SetText("0.0")
					Market.Entry.Enable()
					ResetAuctionInfo()
					AuctionInfo()
					clearNFAImages()
					go GetNFAImages(split[3])
					go GetAuctionDetails(split[3])
				case 2:
					Market.Tab = "Buy"
					Market.Button.BidBuy.Text = "Buy"
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

			Market.Details.Objects[1] = loadingBar()
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

// Get NFA cover and icon images
func GetNFAImages(scid string) {
	if gnomon.IsReady() && len(scid) == 64 {
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
				}
			}

			view := Market.Viewing.Collection
			if view == "AZY-Playing card decks" || view == "SIXPC" || view == "AZY-Playing card backs" || view == "SIXPCB" {
				Market.Details.Objects[0].(*fyne.Container).Objects[0].(*fyne.Container).Objects[2] = ToolsBadge()
			} else {
				Market.Details.Objects[0].(*fyne.Container).Objects[0].(*fyne.Container).Objects[2] = layout.NewSpacer()
			}
		} else {
			Market.Icon = *canvas.NewImageFromResource(bundle.ResourceMarketCirclePng)
		}

		if cover != nil {
			Market.Cover, _ = dreams.DownloadCanvas(cover[0], name[0]+"-cover")
			if Market.Cover.Resource != nil {
				Market.Details.Objects[1] = NFACoverImg()
				if Market.Loading != nil {
					Market.Loading.Stop()
				}
			} else {
				Market.Details.Objects[1] = loadingBar()
			}
		} else {
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
	spacer := canvas.NewRectangle(color.Transparent)
	spacer.SetMinSize(fyne.NewSize(400, 600))
	Market.Loading.Start()

	return container.NewCenter(container.NewStack(Market.Loading, spacer), container.NewCenter(canvas.NewText("Loading...", bundle.TextColor)))
}

// Clears all market NFA images
func clearNFAImages() {
	Market.Details.Objects[0].(*fyne.Container).Objects[0].(*fyne.Container).Objects[2] = layout.NewSpacer()
	Market.Icon = *canvas.NewImageFromImage(nil)

	Market.Cover = *canvas.NewImageFromImage(nil)
	Market.Details.Objects[1] = canvas.NewImageFromImage(nil)
}

// Set up market info objects
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

// Container for auction info objects
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
		NFACoverImg())
}

// Container for unlisted info objects
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
		NFACoverImg())
}

// Set unlisted display content to default values
func ResetNotListedInfo() {
	Market.amount.bid = 0
	clearNFAImages()
	Market.Display.Name.SetText("Name:")
	Market.Display.Type.SetText("Asset Type:")
	Market.Display.Collection.SetText("Collection:")
	Market.Display.Description.SetText("Description:")
	Market.Display.Creator.SetText("Creator:")
	Market.Display.Artificer.SetText("Artificer:")
	Market.Display.Royalty.SetText("Royalty:")
	Market.Display.Owner.SetText("Owner:")
	Market.Display.Update.SetText("Owner can update:")
}

// Set auction display content to default values
func ResetAuctionInfo() {
	Market.amount.bid = 0
	clearNFAImages()
	Market.Display.Name.SetText("Name:")
	Market.Display.Type.SetText("Asset Type:")
	Market.Display.Collection.SetText("Collection:")
	Market.Display.Description.SetText("Description:")
	Market.Display.Creator.SetText("Creator:")
	Market.Display.Artificer.SetText("Artificer:")
	Market.Display.Royalty.SetText("Royalty:")
	Market.Display.Price.SetText("Start Price:")
	Market.Display.Owner.SetText("Owner:")
	Market.Display.Update.SetText("Owner can update:")
	Market.Display.Bid.Current.SetText("Current Bid:")
	Market.Display.Bid.Price.SetText("Minimum Bid:")
	Market.Display.Bid.Count.SetText("Bids:")
	Market.Display.Ends.SetText("Ends At:")
}

// Container for buy now info objects
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
		NFACoverImg())
}

// Set buy now display content to default values
func ResetBuyInfo() {
	Market.amount.buy = 0
	clearNFAImages()
	Market.Display.Name.SetText("Name:")
	Market.Display.Type.SetText("Asset Type:")
	Market.Display.Collection.SetText("Collection:")
	Market.Display.Description.SetText("Description:")
	Market.Display.Creator.SetText("Creator:")
	Market.Display.Artificer.SetText("Artificer:")
	Market.Display.Royalty.SetText("Royalty:")
	Market.Display.Price.SetText("Buy now for:")
	Market.Display.Owner.SetText("Owner:")
	Market.Display.Update.SetText("Owner can update:")
	Market.Display.Ends.SetText("Ends At:")
}

// NFA market layout
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

	tabs := container.NewAppTabs(
		container.NewTabItem("Auctions", AuctionListings()),
		container.NewTabItem("Buy Now", BuyNowListings()),
		container.NewTabItem("My Listings", container.NewBorder(
			nil,
			container.NewBorder(nil, nil, Market.Button.Close, Market.Button.Cancel),
			nil,
			nil,
			MyNFAListings())),

		container.NewTabItem("Search", SearchNFAs(d)))

	tabs.SetTabLocation(container.TabLocationTop)
	tabs.OnSelected = func(ti *container.TabItem) {
		switch ti.Text {
		case "Auctions":
			go FindNFAListings(nil)
			Market.Tab = "Auction"
			Market.List.Auction.UnselectAll()
			Market.List.Wallet.UnselectAll()
			Market.Viewing.Asset = ""
			Market.Viewing.Collection = ""
			Market.Button.BidBuy.Text = "Bid"
			Market.Button.BidBuy.Refresh()
			Market.Entry.Show()
			Market.Entry.SetText("0.0")
			Market.Entry.Enable()
			ResetAuctionInfo()
			Market.Details = auction_info
		case "Buy Now":
			go FindNFAListings(nil)
			Market.Tab = "Buy"
			Market.List.Buy.UnselectAll()
			Market.List.Wallet.UnselectAll()
			Market.Viewing.Asset = ""
			Market.Viewing.Collection = ""
			Market.Button.BidBuy.Text = "Buy"
			Market.Button.BidBuy.Refresh()
			Market.Entry.Show()
			Market.Entry.SetText("0.0")
			Market.Entry.Disable()
			ResetBuyInfo()
			Market.Details = buy_info
		case "My Listings":
			go FindNFAListings(nil)
			Market.Tab = "Buy"
			Market.List.Auction.UnselectAll()
			Market.List.Wallet.UnselectAll()
			Market.List.Buy.UnselectAll()
			Market.Viewing.Asset = ""
			Market.Viewing.Collection = ""
			Market.Entry.Hide()
			Market.Entry.SetText("0.0")
			Market.Entry.Disable()
			ResetBuyInfo()
			Market.Details = not_listed_info
		case "Search":
			Market.Tab = "Buy"
			Market.List.Auction.UnselectAll()
			Market.List.Wallet.UnselectAll()
			Market.List.Buy.UnselectAll()
			Market.Viewing.Asset = ""
			Market.Viewing.Collection = ""
			Market.Entry.Hide()
			Market.Entry.SetText("0.0")
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

	Market.Actions = *container.NewBorder(nil, nil, nil, container.NewStack(button_spacer, Market.Button.BidBuy), MarketEntry())
	Market.Actions.Hide()

	padding := canvas.NewRectangle(color.Transparent)
	padding.SetMinSize(fyne.NewSize(123, 0))

	details := container.NewBorder(nil, container.NewHBox(padding, container.NewStack(entry_spacer, &Market.Actions)), nil, nil, &Market.Details)

	market := container.NewHSplit(details, max)
	market.SetOffset(0.66)

	Control.Lock()
	Control.Dapps["NFA Market"] = true
	Control.Unlock()

	go RunNFAMarket(d)

	return market
}

// Routine for NFA market, finds listings and disables controls, see PlaceAssets() and PlaceMarket() layouts
func RunNFAMarket(d *dreams.AppObject) {
	offset := 0
	for {
		select {
		case <-d.Receive(): // do on interval
			if !rpc.Wallet.IsConnected() || !rpc.Daemon.IsConnected() {
				Market.Button.BidBuy.Hide()
				d.WorkDone()
				continue
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
								FindNFAListings(nil)
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

// Scan index for any active NFA listings
//   - Pass assets from db store, can be nil arg
func FindNFAListings(assets map[string]string) {
	if Market.Searching {
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
		if assets == nil {
			assets = gnomon.GetAllOwnersAndSCIDs()
		}

		for sc := range assets {
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
