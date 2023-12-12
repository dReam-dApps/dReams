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
	Searching     bool
	DreamsFilter  bool
	Buy_amt       uint64
	Bid_amt       uint64
	Viewing       string
	Viewing_coll  string
	Auctions      []NFAListing
	Buy_now       []NFAListing
	My_list       []NFAListing
	Filters       []string
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
	m.Auction_list.Refresh()
}

// Sorts buy now list by name
func (m *marketObjects) SortBuys() {
	sort.Slice(m.Buy_now, func(i, j int) bool {
		return m.Buy_now[i].Name < m.Buy_now[j].Name
	})
	m.Buy_list.Refresh()
}

// Sorts wallet listings list by name
func (m *marketObjects) SortMyList() {
	sort.Slice(m.My_list, func(i, j int) bool {
		return m.My_list[i].Name < m.My_list[j].Name
	})
	m.My_listings.Refresh()
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
			Market.Viewing = ""
			Market.Viewing_coll = ""
			Market.Cancel_button.Hide()
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
	Market.Auction_list = widget.NewList(
		func() int {
			return len(Market.Auctions)
		},
		listItem,
		func(i widget.ListItemID, o fyne.CanvasObject) {
			go updateListItem(i, o, Market.Auctions)
		})

	Market.Auction_list.OnSelected = func(id widget.ListItemID) {
		scid := Market.Auctions[id].SCID
		if scid != Market.Viewing {
			Market.Entry.SetText("")
			clearNFAImages()
			Market.Viewing = scid
			go GetNFAImages(scid)
			go GetAuctionDetails(scid)
			Market.Details_box.Objects[1] = loadingBar()
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
		listItem,
		func(i widget.ListItemID, o fyne.CanvasObject) {
			go updateListItem(i, o, Market.Buy_now)
		})

	Market.Buy_list.OnSelected = func(id widget.ListItemID) {
		scid := Market.Buy_now[id].SCID
		if scid != Market.Viewing {
			clearNFAImages()
			Market.Viewing = scid
			go GetNFAImages(scid)
			go GetBuyNowDetails(scid)
			Market.Details_box.Objects[1] = loadingBar()
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
		listItem,
		func(i widget.ListItemID, o fyne.CanvasObject) {
			updateListItem(i, o, Market.My_list)
		})

	Market.My_listings.OnSelected = func(id widget.ListItemID) {
		scid := Market.My_list[id].SCID
		if scid != Market.Viewing {
			clearNFAImages()
			Market.Viewing = scid
			go GetNFAImages(scid)
			go GetUnlistedDetails(scid)
			Market.Details_box.Objects[1] = loadingBar()
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
					Market.Details_box.Objects[0].(*fyne.Container).Objects[0].(*fyne.Container).Objects[1] = NFAIcon()
				} else {
					Market.Icon = *canvas.NewImageFromResource(bundle.ResourceMarketCirclePng)
				}
			}

			view := Market.Viewing_coll
			if view == "AZY-Playing card decks" || view == "SIXPC" || view == "AZY-Playing card backs" || view == "SIXPCB" {
				Market.Details_box.Objects[0].(*fyne.Container).Objects[0].(*fyne.Container).Objects[2] = ToolsBadge()
			} else {
				Market.Details_box.Objects[0].(*fyne.Container).Objects[0].(*fyne.Container).Objects[2] = layout.NewSpacer()
			}
		} else {
			Market.Icon = *canvas.NewImageFromResource(bundle.ResourceMarketCirclePng)
		}

		if cover != nil {
			Market.Cover, _ = dreams.DownloadCanvas(cover[0], name[0]+"-cover")
			if Market.Cover.Resource != nil {
				Market.Details_box.Objects[1] = NFACoverImg()
				if Market.Loading != nil {
					Market.Loading.Stop()
				}
			} else {
				Market.Details_box.Objects[1] = loadingBar()
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
	Market.Details_box.Objects[0].(*fyne.Container).Objects[0].(*fyne.Container).Objects[2] = layout.NewSpacer()
	Market.Icon = *canvas.NewImageFromImage(nil)

	Market.Cover = *canvas.NewImageFromImage(nil)
	Market.Details_box.Objects[1] = canvas.NewImageFromImage(nil)
}

// Set up market info objects
func NFAMarketInfo() fyne.Container {
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
func AuctionInfo() fyne.Container {
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
	unlisted_form = append(unlisted_form, widget.NewFormItem("Name", Market.Name))
	unlisted_form = append(unlisted_form, widget.NewFormItem("Asset Type", Market.Type))
	unlisted_form = append(unlisted_form, widget.NewFormItem("Collection", Market.Collection))
	unlisted_form = append(unlisted_form, widget.NewFormItem("Description", Market.Description))
	unlisted_form = append(unlisted_form, widget.NewFormItem("Creator", Market.Creator))
	unlisted_form = append(unlisted_form, widget.NewFormItem("Owner", Market.Owner))
	unlisted_form = append(unlisted_form, widget.NewFormItem("Artificer %", Market.Art_fee))
	unlisted_form = append(unlisted_form, widget.NewFormItem("Royalty %", Market.Royalty))

	unlisted_form = append(unlisted_form, widget.NewFormItem("Owner Update", Market.Owner_update))

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
func BuyNowInfo() fyne.Container {
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

// NFA market layout
func PlaceMarket(d *dreams.AppObject) *container.Split {
	auction_info := NFAMarketInfo()

	buy_info := BuyNowInfo()

	not_listed_info := NotListedInfo()

	Market.Cancel_button = widget.NewButton("Cancel", func() {
		if len(Market.Viewing) == 64 {
			Market.Cancel_button.Hide()
			ConfirmCancelClose(Market.Viewing, false, d)
		} else {
			dialog.NewInformation("NFA Cancel", "SCID error", d.Window).Show()
		}
	})
	Market.Cancel_button.Importance = widget.HighImportance
	Market.Cancel_button.Hide()

	Market.Close_button = widget.NewButton("Close", func() {
		if len(Market.Viewing) == 64 {
			Market.Close_button.Hide()
			ConfirmCancelClose(Market.Viewing, true, d)
		} else {
			dialog.NewInformation("NFA Close", "SCID error", d.Window).Show()
		}
	})
	Market.Close_button.Importance = widget.HighImportance
	Market.Close_button.Hide()

	tabs := container.NewAppTabs(
		container.NewTabItem("Auctions", AuctionListings()),
		container.NewTabItem("Buy Now", BuyNowListings()),
		container.NewTabItem("My Listings", container.NewBorder(
			nil,
			container.NewBorder(nil, nil, Market.Close_button, Market.Cancel_button),
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
			Market.Auction_list.UnselectAll()
			Market.My_listings.UnselectAll()
			Market.Viewing = ""
			Market.Viewing_coll = ""
			Market.Market_button.Text = "Bid"
			Market.Market_button.Refresh()
			Market.Entry.Show()
			Market.Entry.SetText("0.0")
			Market.Entry.Enable()
			ResetAuctionInfo()
			Market.Details_box = auction_info
		case "Buy Now":
			go FindNFAListings(nil)
			Market.Tab = "Buy"
			Market.Buy_list.UnselectAll()
			Market.Viewing = ""
			Market.Viewing_coll = ""
			Market.Market_button.Text = "Buy"
			Market.Entry.Show()
			Market.Entry.SetText("0.0")
			Market.Entry.Disable()
			Market.Market_button.Refresh()
			ResetBuyInfo()
			Market.Details_box = buy_info
		case "My Listings":
			go FindNFAListings(nil)
			Market.Tab = "Buy"
			Market.My_listings.UnselectAll()
			Market.Viewing = ""
			Market.Viewing_coll = ""
			Market.Entry.Hide()
			Market.Entry.SetText("0.0")
			Market.Entry.Disable()
			ResetBuyInfo()
			Market.Details_box = not_listed_info
		case "Search":
			Market.Tab = "Buy"
			Market.Auction_list.UnselectAll()
			Market.Buy_list.UnselectAll()
			Market.Viewing = ""
			Market.Viewing_coll = ""
			Market.Entry.Hide()
			Market.Entry.SetText("0.0")
			Market.Entry.Disable()
			ResetBuyInfo()
			Market.Details_box = not_listed_info
		}

		Market.Close_button.Hide()
		Market.Cancel_button.Hide()
		Market.Market_button.Hide()
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

	min_size := bundle.Alpha120
	min_size.SetMinSize(fyne.NewSize(420, 0))

	max := container.NewStack(min_size, tabs, scroll_cont)

	Market.Details_box = auction_info

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
		} else {
			dialog.NewInformation("Market", "SCID error", d.Window).Show()
		}
	})
	Market.Market_button.Importance = widget.HighImportance
	Market.Market_button.Hide()

	entry_spacer := canvas.NewRectangle(color.Transparent)
	entry_spacer.SetMinSize(fyne.NewSize(210, 0))

	button_spacer := canvas.NewRectangle(color.Transparent)
	button_spacer.SetMinSize(fyne.NewSize(40, 0))

	Market.Market_box = *container.NewBorder(nil, nil, nil, container.NewStack(button_spacer, Market.Market_button), MarketEntry())
	Market.Market_box.Hide()

	padding := canvas.NewRectangle(color.Transparent)
	padding.SetMinSize(fyne.NewSize(123, 0))

	details := container.NewBorder(nil, container.NewHBox(padding, container.NewStack(entry_spacer, &Market.Market_box)), nil, nil, &Market.Details_box)

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
				Market.Market_button.Hide()
				d.WorkDone()
				continue
			}

			// If connected daemon connected start looking for Gnomon sync with daemon
			if rpc.Daemon.IsConnected() && gnomon.IsRunning() {
				// Enable index controls and check if wallet is connected
				DisableIndexControls(false)
				if rpc.Wallet.IsConnected() {
					Market.Market_box.Show()
					Assets.Claim.Show()
					if d.OnSubTab("Market") {
						// Update live market info
						if len(Market.Viewing) == 64 {
							if Market.Tab == "Buy" {
								GetBuyNowDetails(Market.Viewing)
							} else {
								GetAuctionDetails(Market.Viewing)
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

									for _, r := range Market.Buy_now {
										gnomes.StoreBolt(r.Collection, r.Name, r)
									}
								}
							}
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
		Market.Buy_now = buy_now
		Market.SortBuys()
		Market.My_list = my_list
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
