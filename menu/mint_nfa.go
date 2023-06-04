package menu

import (
	_ "embed"
	"errors"
	"fmt"
	"image/color"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/validation"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"github.com/SixofClubsss/dReams/bundle"
	"github.com/SixofClubsss/dReams/dwidget"
	"github.com/SixofClubsss/dReams/rpc"
	"github.com/deroproject/derohe/walletapi"
)

//go:embed ART-NFA-MS1.bas
var ART_NFA_MS1 string

// Tree object containing NFA minting instructions screen
func HowToMintNFA(button *widget.Button) fyne.CanvasObject {
	list := map[string][]string{
		"":                       {"How To Mint NFAs"},
		"How To Mint NFAs":       {"Get Started", "Single Asset", "Collection"},
		"Get Started":            {"A NFA consists of four main parts: Asset file, Cover Image, Icon image, Dero file sign", "Each NFA is its own self contained marketplace", "This tool automates three areas of NFA installs: File sign, Contract creation, Contract install", "Storage is not provided at this point", "Entries with a * are mutable, meaning they can be updated by creator (or owner) after install", "Gas fees to install a NFA are ~0.20000 Dero", "There is a 0.00500 Dero dev fee for minting a NFA with this tool", "If minting a collection fees will be paid as each contract is installed, a total will be shown before hand", "For further info read the NFA documentation at github.com/civilware/artificer-nfa-standard"},
		"Single Asset":           {"Disable the Collection check", "Type the name of your asset into the collection entry and click the folder button on right to set up NFA-Creation directory for your single asset", "NFA-Creation Directory", "File sign can be imported from file by clicking the file button to right of check C entry, or follow next step if you require file sign", "Single File Sign", "Fill out the rest of the information for your NFA and when complete the Create Contracts button will show", "Click Create Contract and confirm information, this will populate your bas folder with your asset contract", "Type the name of your asset in name entry and Install Contract will show if contract exists in bas folder", "Click Install Contract and confirm the install address is same as signing address", "NFA is now installed, check your wallet for NFA balance"},
		"Single File Sign":       {"Enter minting wallet file password and open minting wallet file", "Place asset file into asset folder", "Enter the name of your asset in name entry, select extension to match file", "Click Sign File and confirm information", "Once confirmed, the file check C and file check S will population with your file signs"},
		"Collection":             {"Collection automation installs assets of same name with incrementing numbers", "Enable the Collection check", "Type the name of your collection into the collection entry and click the folder button on right to set up NFA-Creation directory for your collection", "NFA-Creation Directory", "Enter the starting number and ending number for your collection", "File signs can be done externally and placed into sign folder, or follow next step if you require file signs", "Collection File Signs", "Make sure file signs are in sign directory for contract creation", "Fill out the rest of the information for your NFA collection and when complete the Create Contracts button will show", "The Asset Number sections are where it will add the incrementing number to your input to make the collection", "The + - buttons on top right can add or remove a increment section from Url paths", "Click Create Contract and confirm information", "Contract creation loop will start and populate your bas folder with your asset contracts, takes about 1 second per contract", "Type the name of your asset in name entry and Install Contract will show if contract exists in bas folder", "Click Install Contract and confirm the install address is same as signing address", "Minting loop will now start and installs one bas contract per block", "For larger collections this could take some time, 120 installs could take around 1 hour to complete", "If 100%, NFA collection is now installed, check your wallet for NFA balances"},
		"Collection File Signs":  {"Enter minting wallet file password and open minting wallet file", "Place numbered asset files into asset folder", "Enter the name of your asset in name entry, select extension to match file", "Click Sign File and confirm information", "This starts a file sign loop of your selected range and stores all signed files in sign directory, takes about 1 second per sign"},
		"NFA-Creation Directory": {"NFA-Creation directory stores collection and single asset directories", "Inside of your asset or collection directory are five sub directories", "Your main asset files are stored in asset", "Contracts created are stored in bas", "Signed files are stored in sign", "Cover and icon are optional directories at this point and are not used in the install process"},
	}

	tree := widget.NewTreeWithStrings(list)

	tree.OnBranchClosed = func(uid widget.TreeNodeID) {
		tree.UnselectAll()
		if uid == "How To Mint NFA" {
			tree.CloseAllBranches()
		}
	}

	tree.OnBranchOpened = func(uid widget.TreeNodeID) {
		tree.Select(uid)
	}

	tree.OpenBranch("How To Mint NFAs")

	return container.NewBorder(nil, button, nil, nil, tree)
}

// Create a new NFA contract string with passed values
func CreateNFAContract(art, royalty, update, name, descrip, typeHdr, icon, tags, fileCheckC, fileCheckS, fileUrl, fileSignUrl, coverURl, collection string) (contract string) {
	contract = ART_NFA_MS1[0:231] + art + ART_NFA_MS1[232:259] + royalty + ART_NFA_MS1[260:294] + update + ART_NFA_MS1[295:323] + name + ART_NFA_MS1[332:362] + descrip + ART_NFA_MS1[372:401] + typeHdr +
		ART_NFA_MS1[410:442] + icon + ART_NFA_MS1[454:483] + tags + ART_NFA_MS1[492:524] + fileCheckC + ART_NFA_MS1[536:568] + fileCheckS + ART_NFA_MS1[580:609] + fileUrl + ART_NFA_MS1[618:651] + fileSignUrl +
		ART_NFA_MS1[664:694] + coverURl + ART_NFA_MS1[704:736] + collection + ART_NFA_MS1[748:]

	return
}

// Visual spacer for string increment breaks
func incrementSpacer() fyne.CanvasObject {
	add_text := canvas.NewText("Asset Number", bundle.TextColor)
	add_text.Alignment = fyne.TextAlignCenter

	space_color := color.RGBA{105, 90, 205, 210}
	if bundle.AppColor == color.White {
		space_color = color.RGBA{31, 150, 200, 210}
	}

	rect := canvas.NewRectangle(space_color)

	return container.NewMax(rect, add_text)
}

// Place objects for NFA minting of collections or single mint
func PlaceNFAMint(tag string, window fyne.Window) fyne.CanvasObject {
	collection_enable := widget.NewCheck("Collection", nil)
	sign_button := widget.NewButton("File Sign", nil)
	contracts_button := widget.NewButton("Create Contract", nil)
	install_button := widget.NewButton("Install Contract", nil)

	collection_high_entry := dwidget.DeroAmtEntry("", 1, 0)
	collection_high_entry.SetPlaceHolder("Ending At:")
	collection_high_entry.AllowFloat = false
	collection_high_entry.Validator = validation.NewRegexp(`^[^0]\d{0,}$`, "Uint required")

	collection_low_entry := dwidget.DeroAmtEntry("", 1, 0)
	collection_low_entry.SetPlaceHolder("Starting At:")
	collection_low_entry.AllowFloat = false
	collection_low_entry.Validator = validation.NewRegexp(`^[^0]\d{0,}$`, "Uint required")

	collection_entry := widget.NewEntry()
	collection_entry.SetPlaceHolder("Collection:")
	collection_entry.Validator = validation.NewRegexp(`^\w{2,}`, "String required")

	update_select := widget.NewSelect([]string{"No", "Yes"}, nil)
	update_select.PlaceHolder = "Owner Can Update:"

	name_entry := widget.NewEntry()
	name_entry.SetPlaceHolder("Asset Name:")
	name_entry.Validator = validation.NewRegexp(`^\w{2,}`, "String required")
	name_cont := container.NewGridWithRows(1, name_entry)

	extension_select := widget.NewSelect([]string{".jpg", ".png", ".gif", ".mp3", ".mp4", ".pdf", ".zip", ".7z", ".avi", ".mov", ".ogg"}, nil)
	extension_select.PlaceHolder = "ext"

	set_up_collec := widget.NewButtonWithIcon("", fyne.Theme.Icon(fyne.CurrentApp().Settings().Theme(), "folderNew"), func() {
		if collection_entry.Text != "" {
			if collection_entry.Validate() == nil {
				info_message := dialog.NewInformation("Collection Exists", "Check NFA-Creation directory", window)
				info_message.Resize(fyne.NewSize(300, 150))
				info_message.Show()
				return
			}

			info := fmt.Sprintf("Creating NFA directories for %s", collection_entry.Text)
			confirm := dialog.NewConfirm("Create directories", info, func(b bool) {
				if b {
					SetUpNFACreation(tag, collection_entry.Text)
					collection_entry.SetText(collection_entry.Text)
				}
			}, window)

			confirm.Resize(fyne.NewSize(600, 240))
			confirm.Show()
		}
	})

	descr_entry := widget.NewMultiLineEntry()
	descr_entry.SetPlaceHolder("Asset Description:")
	descr_entry.Validator = validation.NewRegexp(`^\w{1,}`, "String required")

	type_select := widget.NewSelect([]string{"Book", "Code", "File", "Image", "Movie", "Music", "Package", "Text"}, nil)
	type_select.PlaceHolder = "Asset Type"

	tags_entry := widget.NewEntry()
	tags_entry.SetPlaceHolder("Asset Tags:")
	tags_entry.Validator = validation.NewRegexp(`^\#\w{1,}|^\w{1,}`, "String required")

	checkC_entry := widget.NewEntry()
	checkC_entry.SetPlaceHolder("File Sign C:")
	checkC_entry.Validator = validation.NewRegexp(`^\w{61,64}$`, "Invalid")

	checkS_entry := widget.NewEntry()
	checkS_entry.SetPlaceHolder("File Sign S:")
	checkS_entry.Validator = validation.NewRegexp(`^\w{61,64}$`, "Invalid")

	import_signs := widget.NewButtonWithIcon("", fyne.Theme.Icon(fyne.CurrentApp().Settings().Theme(), "document"), func() {
		read_filesign := dialog.NewFileOpen(func(uc fyne.URIReadCloser, err error) {
			if err == nil && uc != nil {
				if sign_data, err := os.ReadFile(uc.URI().Path()); err != nil {
					log.Printf("[%s] Cannot read input file %s\n", tag, err)
				} else {
					if string(sign_data[0:35]) == "-----BEGIN DERO SIGNED MESSAGE-----" {
						split := strings.Split(string(sign_data[0:247]), "\n")
						read_checkC := strings.TrimSpace(split[2][3:])
						read_checkS := strings.TrimSpace(split[3][3:])
						checkC_entry.SetText(read_checkC)
						checkS_entry.SetText(read_checkS)
					} else {
						log.Printf("[%s] Not a valid Dero .sign file\n", tag)
					}
				}
			}
		}, window)

		read_filesign.SetConfirmText("Open Signature")
		read_filesign.SetFilter(storage.NewExtensionFileFilter([]string{".sign"}))
		read_filesign.Resize(fyne.NewSize(900, 600))
		read_filesign.Show()
	})

	var file_entries, cover_entries, icon_entries, sign_entries *fyne.Container
	file_entry_start := widget.NewEntry()
	file_entry_start.SetPlaceHolder("Asset File URL:")
	file_entry_start.Validator = validation.NewRegexp(`^\w{2,}`, "String required")
	file_entry_mid := widget.NewEntry()
	file_entry_mid.SetPlaceHolder("Asset File URL End:")
	file_entry_end := widget.NewEntry()
	file_entry_end.SetPlaceHolder("Asset File URL End:")

	cover_entry_start := widget.NewEntry()
	cover_entry_start.SetPlaceHolder("Cove URL:")
	cover_entry_start.Validator = validation.NewRegexp(`^\w{2,}`, "String required")
	cover_entry_mid := widget.NewEntry()
	cover_entry_mid.SetPlaceHolder("Cover URL End:")
	cover_entry_end := widget.NewEntry()
	cover_entry_end.SetPlaceHolder("Cover URL End:")

	icon_entry_start := widget.NewEntry()
	icon_entry_start.SetPlaceHolder("Icon URL:")
	icon_entry_start.Validator = validation.NewRegexp(`^\w{2,}`, "String required")
	icon_entry_mid := widget.NewEntry()
	icon_entry_mid.SetPlaceHolder("Icon URL End:")
	icon_entry_end := widget.NewEntry()
	icon_entry_end.SetPlaceHolder("Icon URL End:")

	sign_entry_start := widget.NewEntry()
	sign_entry_start.Validator = validation.NewRegexp(`^\w{2,}`, "String required")
	sign_entry_start.SetPlaceHolder("Sign URL:")
	sign_entry_mid := widget.NewEntry()
	sign_entry_mid.SetPlaceHolder("Sign URL End:")
	sign_entry_end := widget.NewEntry()
	sign_entry_end.SetPlaceHolder("Sign URL End:")

	file_entries = container.NewGridWithRows(1, file_entry_start)
	cover_entries = container.NewGridWithRows(1, cover_entry_start)
	icon_entries = container.NewGridWithRows(1, icon_entry_start)
	sign_entries = container.NewGridWithRows(1, sign_entry_start)

	collection_low_entry.OnChanged = func(s string) {
		start := rpc.StringToInt(s)
		if collection_high_entry.Text == "" {
			collection_high_entry.SetText(strconv.Itoa(start + 2))
		}

		end := rpc.StringToInt(collection_high_entry.Text)
		if start >= end {
			if start < 2 {
				start = 2
			}
			collection_low_entry.SetText(strconv.Itoa(start - 2))
		}
	}

	collection_high_entry.OnChanged = func(s string) {
		start := rpc.StringToInt(collection_low_entry.Text)
		end := rpc.StringToInt(s)

		if end <= start {
			collection_high_entry.SetText(strconv.Itoa(end + 2))
		}
	}

	url_add_incr := widget.NewButtonWithIcon("", fyne.Theme.Icon(fyne.CurrentApp().Settings().Theme(), "contentAdd"), func() {
		switch len(file_entries.Objects) {
		case 0:
			file_entries.Add(file_entry_start)
			cover_entries.Add(cover_entry_start)
			icon_entries.Add(icon_entry_start)
			sign_entries.Add(sign_entry_start)
			file_entry_start.SetPlaceHolder("File URL:")
			cover_entry_start.SetPlaceHolder("Cover URL:")
			icon_entry_start.SetPlaceHolder("Icon URL:")
			sign_entry_start.SetPlaceHolder("Sign URL:")
		case 1:
			file_entry_start.SetPlaceHolder("File URL Start:")
			cover_entry_start.SetPlaceHolder("Cover URL Start:")
			icon_entry_start.SetPlaceHolder("Icon URL Start:")
			sign_entry_start.SetPlaceHolder("Sign URL Start:")
			name_cont.Add(incrementSpacer())
			file_entries.Add(incrementSpacer())
			file_entries.Add(file_entry_mid)
			file_entry_mid.SetPlaceHolder("File URL End:")
			cover_entries.Add(incrementSpacer())
			cover_entries.Add(cover_entry_mid)
			cover_entry_mid.SetPlaceHolder("Cover URL End:")
			icon_entries.Add(incrementSpacer())
			icon_entries.Add(icon_entry_mid)
			icon_entry_mid.SetPlaceHolder("Icon URL End:")
			sign_entries.Add(incrementSpacer())
			sign_entries.Add(sign_entry_mid)
			sign_entry_mid.SetPlaceHolder("Sign URL End:")
		case 3:
			file_entry_start.SetPlaceHolder("File URL Start:")
			cover_entry_start.SetPlaceHolder("Cover URL Start:")
			icon_entry_start.SetPlaceHolder("Icon URL Start:")
			sign_entry_start.SetPlaceHolder("Sign URL Start:")
			file_entries.Add(incrementSpacer())
			file_entries.Add(file_entry_end)
			file_entry_mid.SetPlaceHolder("File URL Mid:")
			cover_entries.Add(incrementSpacer())
			cover_entries.Add(cover_entry_end)
			cover_entry_mid.SetPlaceHolder("Cover URL Mid:")
			icon_entries.Add(incrementSpacer())
			icon_entries.Add(icon_entry_end)
			icon_entry_mid.SetPlaceHolder("Icon URL Mid:")
			sign_entries.Add(incrementSpacer())
			sign_entries.Add(sign_entry_end)
			sign_entry_mid.SetPlaceHolder("Sign URL Mid:")
		default:

		}

		file_entries.Refresh()
	})

	url_remove_incr := widget.NewButtonWithIcon("", fyne.Theme.Icon(fyne.CurrentApp().Settings().Theme(), "contentRemove"), func() {
		l := len(file_entries.Objects)
		if l > 3 {
			file_entries.Remove(file_entries.Objects[l-1])
			file_entries.Remove(file_entries.Objects[l-2])
			cover_entries.Remove(cover_entries.Objects[l-1])
			cover_entries.Remove(cover_entries.Objects[l-2])
			icon_entries.Remove(icon_entries.Objects[l-1])
			icon_entries.Remove(icon_entries.Objects[l-2])
			sign_entries.Remove(sign_entries.Objects[l-1])
			sign_entries.Remove(sign_entries.Objects[l-2])
			file_entry_mid.SetPlaceHolder("File URL End:")
			cover_entry_mid.SetPlaceHolder("Cover URL End:")
			icon_entry_mid.SetPlaceHolder("Icon URL End:")
			sign_entry_mid.SetPlaceHolder("Sign URL End:")
		}
	})

	collection_enable.OnChanged = func(b bool) {
		contracts_button.Hide()
		install_button.Hide()
		if b {
			file_entry_start.SetPlaceHolder("File URL Start:")
			cover_entry_start.SetPlaceHolder("Cover URL Start:")
			icon_entry_start.SetPlaceHolder("Icon URL Start:")
			sign_entry_start.SetPlaceHolder("Sign URL Start:")
			file_entry_mid.SetPlaceHolder("File URL End:")
			cover_entry_mid.SetPlaceHolder("Cover URL End:")
			icon_entry_mid.SetPlaceHolder("Icon URL End:")
			sign_entry_mid.SetPlaceHolder("Sign URL End:")
			import_signs.Disable()
			checkC_entry.Validator = validation.NewRegexp(`^`, "")
			checkS_entry.Validator = validation.NewRegexp(`^`, "")
			checkC_entry.SetText("")
			checkS_entry.SetText("")
			checkC_entry.SetPlaceHolder("Loading from sign directory")
			checkS_entry.SetPlaceHolder("Loading from sign directory")
			checkC_entry.Disable()
			checkS_entry.Disable()
			url_add_incr.Show()
			url_remove_incr.Show()
			collection_low_entry.Show()
			collection_high_entry.Show()
			name_cont.Add(incrementSpacer())
			file_entries.Add(incrementSpacer())
			file_entries.Add(file_entry_mid)
			cover_entries.Add(incrementSpacer())
			cover_entries.Add(cover_entry_mid)
			icon_entries.Add(incrementSpacer())
			icon_entries.Add(icon_entry_mid)
			sign_entries.Add(incrementSpacer())
			sign_entries.Add(sign_entry_mid)
			collection_entry.SetText(collection_entry.Text)
		} else {
			file_entry_start.SetPlaceHolder("File URL:")
			cover_entry_start.SetPlaceHolder("Cover URL:")
			icon_entry_start.SetPlaceHolder("Icon URL:")
			sign_entry_start.SetPlaceHolder("Sign URL:")
			import_signs.Enable()
			checkC_entry.Validator = validation.NewRegexp(`^\w{61,64}$`, "Invalid")
			checkS_entry.Validator = validation.NewRegexp(`^\w{61,64}$`, "Invalid")
			checkC_entry.SetText("")
			checkS_entry.SetText("")
			checkC_entry.SetPlaceHolder("File Sign C:")
			checkS_entry.SetPlaceHolder("File Sign S:")
			checkC_entry.Enable()
			checkS_entry.Enable()
			url_add_incr.Hide()
			url_remove_incr.Hide()
			name_cont.Refresh()
			collection_low_entry.Hide()
			collection_high_entry.Hide()

			if len(file_entries.Objects) > 1 {
				name_cont.Remove(name_cont.Objects[1])
				file_entries.RemoveAll()
				cover_entries.RemoveAll()
				icon_entries.RemoveAll()
				sign_entries.RemoveAll()
			}

			collection_entry.SetText(collection_entry.Text)
			file_entries.Add(file_entry_start)
			cover_entries.Add(cover_entry_start)
			icon_entries.Add(icon_entry_start)
			sign_entries.Add(sign_entry_start)
		}
	}

	royalty_entry := dwidget.DeroAmtEntry("", 1, 0)
	royalty_entry.AllowFloat = false
	royalty_entry.SetPlaceHolder("% you will get from each sale:")
	royalty_entry.Validator = validation.NewRegexp(`^\d{1,2}$`, "Uint required")

	art_entry := dwidget.DeroAmtEntry("", 1, 0)
	art_entry.AllowFloat = false
	art_entry.SetPlaceHolder("% Artificer will get from each sale:")
	art_entry.Validator = validation.NewRegexp(`^\d{1,2}$`, "Uint required")
	art_entry.OnChanged = func(s string) {
		art := rpc.StringToInt(s)
		roy := rpc.StringToInt(royalty_entry.Text)
		if art+roy > 100 {
			art_entry.SetText(strconv.Itoa(art - 1))
		}
	}

	royalty_entry.OnChanged = func(s string) {
		art := rpc.StringToInt(art_entry.Text)
		roy := rpc.StringToInt(s)
		if art+roy > 100 {
			royalty_entry.SetText(strconv.Itoa(roy - 1))
		}
	}

	wallet_label := widget.NewLabel("Signing address:")
	wallet_label.Alignment = fyne.TextAlignCenter
	wallet_label.Wrapping = fyne.TextWrapWord

	rpc_label := widget.NewLabel(fmt.Sprintf("Installing address: %s", rpc.Wallet.Address))
	rpc_label.Alignment = fyne.TextAlignCenter
	rpc_label.Wrapping = fyne.TextWrapWord

	var wallet_file_path string
	wallet_pass_entry := widget.NewPasswordEntry()
	wallet_pass_entry.SetPlaceHolder("Wallet file pass")
	wallet_pass_entry.OnChanged = func(s string) {
		rpc_label.SetText(fmt.Sprintf("Installing address: %s", rpc.Wallet.Address))
		wallet_label.SetText("Signing address:")
		sign_button.Hide()
	}

	open_wallet_button := widget.NewButton("Open Wallet File", func() {
		open_wallet := dialog.NewFileOpen(func(uc fyne.URIReadCloser, err error) {
			wallet_label.SetText("Signing address:")
			if err == nil && uc != nil {
				if rpc.Wallet.File, err = walletapi.Open_Encrypted_Wallet(uc.URI().Path(), wallet_pass_entry.Text); err == nil {
					if _, err := os.ReadFile(uc.URI().Path()); err != nil {
						log.Printf("[%s] Cannot read wallet file %s\n", tag, err)
					} else {
						rpc.Wallet.File.SetNetwork(true)
						log.Printf("[%s] Wallet file found , Wallet is registered: %t\n", tag, rpc.Wallet.File.IsRegistered())
						wallet_label.SetText(fmt.Sprintf("Signing address: %s", rpc.Wallet.File.GetAddress().String()))
						rpc_label.SetText(fmt.Sprintf("Installing address: %s", rpc.Wallet.Address))
						wallet_file_path = uc.URI().Path()
						error_message := dialog.NewInformation("Wallet", fmt.Sprintf("Signing with: %s", rpc.Wallet.File.GetAddress().String()), window)
						error_message.Resize(fyne.NewSize(300, 150))
						error_message.Show()
						go rpc.Wallet.File.Close_Encrypted_Wallet()
					}
				} else {
					log.Printf("[%s] Wallet %s\n", tag, err)
					error_message := dialog.NewInformation("Wallet", "Invalid password", window)
					error_message.Resize(fyne.NewSize(300, 150))
					error_message.Show()
					sign_button.Hide()
				}
				collection_entry.SetText(collection_entry.Text)
			}
		}, window)

		open_wallet.Resize(fyne.NewSize(900, 600))
		open_wallet.Show()
	})

	sign_button.OnTapped = func() {
		_, sign_path := SetUpNFACreation(tag, collection_entry.Text)
		file_name := fmt.Sprintf("%s%s.sign", name_entry.Text, extension_select.Selected)
		if collection_enable.Checked {
			file_name = fmt.Sprintf("%s%s%s.sign", name_entry.Text, collection_low_entry.Text, extension_select.Selected)
		}

		_, err := os.Stat(sign_path + "/" + file_name)
		if !os.IsNotExist(err) {
			info := fmt.Sprintf("Check %s/%s", sign_path, file_name)
			info_message := dialog.NewInformation("File Exists", info, window)
			info_message.Resize(fyne.NewSize(300, 150))
			info_message.Show()
			return
		}

		ext := extension_select.Selected

		var input string
		if collection_enable.Checked {
			input = fmt.Sprintf("NFA-Creation/%s/asset/%s%s%s", collection_entry.Text, name_entry.Text, collection_low_entry.Text, ext)
		} else {
			input = fmt.Sprintf("NFA-Creation/%s/asset/%s%s", collection_entry.Text, name_entry.Text, ext)
		}
		_, err = os.Stat(input)
		if os.IsNotExist(err) {
			info := fmt.Sprintf("%s not found", input)
			error_message := dialog.NewInformation("File Sign", info, window)
			error_message.Resize(fyne.NewSize(300, 150))
			error_message.Show()
			return
		}

		if rpc.Wallet.File != nil && extension_select.SelectedIndex() >= 0 {
			var err error
			if rpc.Wallet.File, err = walletapi.Open_Encrypted_Wallet(wallet_file_path, wallet_pass_entry.Text); err == nil {
				if _, err := os.ReadFile(wallet_file_path); err != nil {
					log.Printf("[%s] Cannot read wallet file %s\n", tag, err)
					error_message := dialog.NewInformation("Error", "Could not read wallet file", window)
					error_message.Resize(fyne.NewSize(300, 150))
					error_message.Show()
					sign_button.Hide()
					return
				}
			} else {
				log.Printf("[%s] Wallet %s\n", tag, err)
				error_message := dialog.NewInformation("Wallet", "Invalid password", window)
				error_message.Resize(fyne.NewSize(300, 150))
				error_message.Show()
				sign_button.Hide()
				return
			}

			rpc.Wallet.File.SetNetwork(true)
			address := rpc.Wallet.File.GetAddress().String()
			if collection_enable.Checked {
				var count, ending_at int
				if count = rpc.StringToInt(collection_low_entry.Text); count < 1 {
					log.Printf("[%s] Not starting signs from 0\n", tag)
					go rpc.Wallet.File.Close_Encrypted_Wallet()
					return
				}

				if ending_at = rpc.StringToInt(collection_high_entry.Text); ending_at < count {
					log.Printf("[%s] Ending is less than starting at\n", tag)
					go rpc.Wallet.File.Close_Encrypted_Wallet()
					return
				}

				total := ending_at - count + 1

				info := fmt.Sprintf("You are about to sign %d asset files\n\nSigning address: %s\n\nSigning: %s%d%s to %s%d%s", total, address, name_entry.Text, count, ext, name_entry.Text, ending_at, ext)
				confirm := dialog.NewConfirm("File sign", info, func(b bool) {
					if b {
						go func() {
							wait := true
							progress_label := widget.NewLabel("")
							progress_label.Alignment = fyne.TextAlignCenter
							progress := widget.NewProgressBar()
							progress.Min = float64(count)
							progress.Max = float64(ending_at)
							progress_cont := container.NewBorder(nil, progress_label, nil, nil, progress)
							wait_message := dialog.NewCustom("Signing Files", "Stop", progress_cont, window)
							wait_message.SetOnClosed(func() {
								wait = false
							})
							wait_message.Resize(fyne.NewSize(450, 150))
							wait_message.Show()

							log.Printf("[%s] Starting sign loop\n", tag)

							for wait {
								i := strconv.Itoa(count)
								signing_asset := fmt.Sprintf("Signing %s%s%s", name_entry.Text, i, ext)
								input_file := fmt.Sprintf("NFA-Creation/%s/asset/%s%s%s", collection_entry.Text, name_entry.Text, i, ext)
								output_file := fmt.Sprintf("NFA-Creation/%s/sign/%s%s%s%s", collection_entry.Text, name_entry.Text, i, ext, ".sign")
								_, err := os.Stat(output_file)
								if !os.IsNotExist(err) {
									log.Printf("[%s] %s exists\n", tag, output_file)
									progress_label.SetText("Error " + output_file + " exists")
									wait_message.SetDismissText("Close")
									info := fmt.Sprintf("Check %s", output_file)
									info_message := dialog.NewInformation("File Exists", info, window)
									info_message.Resize(fyne.NewSize(300, 150))
									info_message.Show()
									break
								}

								progress_label.SetText(signing_asset)
								if data, err := os.ReadFile(input_file); err != nil {
									log.Printf("[%s] Cannot read input file %s\n", tag, err)
									progress_label.SetText("Error " + input_file + " not found")
									wait_message.SetDismissText("Close")
									error_message := dialog.NewInformation("Error", input_file+" not found", window)
									error_message.Resize(fyne.NewSize(300, 150))
									error_message.Show()
									break
								} else if err := os.WriteFile(output_file, rpc.Wallet.File.SignData(data), 0600); err != nil {
									log.Printf("[%s] Cannot write output file %s\n", tag, output_file)
									progress_label.SetText("Error writing " + output_file)
									wait_message.SetDismissText("Close")
									error_message := dialog.NewInformation("Error", "Could not write "+output_file, window)
									error_message.Resize(fyne.NewSize(300, 150))
									error_message.Show()
									break
								} else {
									log.Printf("[%s] Successfully signed file, please check %s\n", tag, output_file)
								}

								time.Sleep(time.Second)
								count++
								progress.SetValue(float64(count))
								if count > ending_at {
									progress_label.SetText(fmt.Sprintf("Signed files in NFA-Creation/%s/sign", collection_entry.Text))
									wait_message.SetDismissText("Done")
									break
								}
							}

							wait = false
							log.Printf("[%s] Sign loop complete\n", tag)
							go rpc.Wallet.File.Close_Encrypted_Wallet()
						}()
					} else {
						go rpc.Wallet.File.Close_Encrypted_Wallet()
					}
				}, window)

				confirm.Resize(fyne.NewSize(600, 240))
				confirm.Show()
			} else {
				ext := extension_select.Selected
				info := fmt.Sprintf("You are about to sign asset file %s%s\n\nSigning address: %s", name_entry.Text, ext, address)
				confirm := dialog.NewConfirm("File sign", info, func(b bool) {
					if b {
						input_file := fmt.Sprintf("NFA-Creation/%s/asset/%s%s", collection_entry.Text, name_entry.Text, ext)
						output_file := fmt.Sprintf("NFA-Creation/%s/sign/%s%s%s", collection_entry.Text, name_entry.Text, ext, ".sign")
						if data, err := os.ReadFile(input_file); err != nil {
							log.Printf("[%s] Cannot read input file %s\n", tag, err)
							error_message := dialog.NewInformation("Error", "Could not read file", window)
							error_message.Resize(fyne.NewSize(300, 150))
							error_message.Show()
						} else if err := os.WriteFile(output_file, rpc.Wallet.File.SignData(data), 0600); err != nil {
							log.Printf("[%s] Cannot write output file %s\n", tag, output_file)
						} else {
							log.Printf("[%s] Successfully signed file. please check %s\n", tag, output_file)
							if sign_data, err := os.ReadFile(output_file); err == nil {
								if string(sign_data[0:35]) == "-----BEGIN DERO SIGNED MESSAGE-----" {
									split := strings.Split(string(sign_data[0:247]), "\n")
									read_checkC := strings.TrimSpace(split[2][3:])
									read_checkS := strings.TrimSpace(split[3][3:])
									checkC_entry.SetText(read_checkC)
									checkS_entry.SetText(read_checkS)
									info_message := dialog.NewInformation("File Signed", output_file, window)
									info_message.Resize(fyne.NewSize(300, 150))
									info_message.Show()
								}
							}
						}
					}
					go rpc.Wallet.File.Close_Encrypted_Wallet()
				}, window)

				confirm.Resize(fyne.NewSize(600, 240))
				confirm.Show()
			}

			go rpc.Wallet.File.Close_Encrypted_Wallet()
		}
	}

	contracts_button.OnTapped = func() {
		collection := collection_entry.Text
		art := art_entry.Text
		royalty := royalty_entry.Text
		update := strconv.Itoa(update_select.SelectedIndex())

		description := descr_entry.Text
		typeHdr := type_select.Selected
		tags := tags_entry.Text
		ext := extension_select.Selected

		name := name_entry.Text
		if collection_enable.Checked {
			name = name + collection_low_entry.Text
		}

		save_path, sign_path := SetUpNFACreation(tag, collection)
		file_name := fmt.Sprintf("%s.bas", name)

		_, err := os.Stat(save_path + "/bas/" + file_name)
		if !os.IsNotExist(err) {
			info := fmt.Sprintf("Check %s/bas/%s", save_path, file_name)
			info_message := dialog.NewInformation("File Exists", info, window)
			info_message.Resize(fyne.NewSize(300, 150))
			info_message.Show()
			return
		}

		checkC := checkC_entry.Text
		checkS := checkS_entry.Text

		file := file_entry_start.Text
		cover := cover_entry_start.Text
		icon := icon_entry_start.Text
		sign := sign_entry_start.Text

		if collection_enable.Checked {
			var count, ending_at int
			if count = rpc.StringToInt(collection_low_entry.Text); count < 1 {
				log.Printf("[%s] Not starting collection from 0\n", tag)
				error_message := dialog.NewInformation("Contract Creation", "Not starting collection from 0", window)
				error_message.Resize(fyne.NewSize(300, 150))
				error_message.Show()
				return
			}

			if ending_at = rpc.StringToInt(collection_high_entry.Text); ending_at < count {
				log.Printf("[%s] Ending is less than starting at\n", tag)
				error_message := dialog.NewInformation("Contract Creation", "Ending number is less than starting number", window)
				error_message.Resize(fyne.NewSize(300, 150))
				error_message.Show()
				return
			}

			total := ending_at - count + 1
			info := fmt.Sprintf("You are about to create %d total .bas contract files\n\nCreating for %s%d%s to %s%d%s\n\nEnsure all sign files are in place", total, name_entry.Text, count, ext, name_entry.Text, ending_at, ext)
			confirm := dialog.NewConfirm("Contract Creation", info, func(b bool) {
				if b {
					go func() {
						wait := true
						name = name_entry.Text + collection_low_entry.Text
						progress_label := widget.NewLabel("")
						progress_label.Alignment = fyne.TextAlignCenter
						progress := widget.NewProgressBar()
						progress.Min = float64(count)
						progress.Max = float64(ending_at)
						progress_cont := container.NewBorder(nil, progress_label, nil, nil, progress)
						wait_message := dialog.NewCustom("Creating Contracts", "Stop", progress_cont, window)
						wait_message.SetOnClosed(func() {
							wait = false
						})
						wait_message.Resize(fyne.NewSize(450, 150))
						wait_message.Show()

						log.Printf("[%s] Starting contract creation loop\n", tag)

						for wait && count <= ending_at {
							incr := strconv.Itoa(count)
							name = name_entry.Text + incr
							file_name := fmt.Sprintf("%s.bas", name)
							progress_label.SetText("Creating " + file_name)
							full_save_path := save_path + "/bas/" + file_name
							_, err := os.Stat(full_save_path)
							if !os.IsNotExist(err) {
								wait = false
								progress_label.SetText("Error " + full_save_path + " exists")
								wait_message.SetDismissText("Close")
								info := fmt.Sprintf("Check %s", full_save_path)
								info_message := dialog.NewInformation("File Exists", info, window)
								info_message.Resize(fyne.NewSize(300, 150))
								info_message.Show()
								log.Printf("[%s] Contract creation loop complete\n", tag)
								return
							}

							if f, err := os.Create(full_save_path); err == nil {
								l := len(file_entries.Objects)
								switch l {
								case 3:
									file = file_entry_start.Text + incr + file_entry_mid.Text
									cover = cover_entry_start.Text + incr + cover_entry_mid.Text
									icon = icon_entry_start.Text + incr + icon_entry_mid.Text
									sign = sign_entry_start.Text + incr + sign_entry_mid.Text
								case 5:
									file = file_entry_start.Text + incr + file_entry_mid.Text + incr + file_entry_end.Text
									cover = cover_entry_start.Text + incr + cover_entry_mid.Text + incr + cover_entry_end.Text
									icon = icon_entry_start.Text + incr + icon_entry_mid.Text + incr + icon_entry_end.Text
									sign = sign_entry_start.Text + incr + sign_entry_mid.Text + incr + sign_entry_end.Text
								default:
									wait = false
									progress_label.SetText("Error could not make Urls")
									wait_message.SetDismissText("Close")
									error_message := dialog.NewInformation("Error", "Could not make Urls", window)
									error_message.Resize(fyne.NewSize(300, 150))
									error_message.Show()
									os.Remove(full_save_path)
									log.Printf("[%s] Contract creation loop complete\n", tag)
									return
								}

								ext := extension_select.Selected + ".sign"
								full_sign_path := sign_path + "/" + name + ext
								if data, err := os.ReadFile(full_sign_path); err != nil {
									wait = false
									log.Printf("[%s] Cannot read input file %s\n", tag, err)
									progress_label.SetText("Error could not read " + full_sign_path)
									wait_message.SetDismissText("Close")
									error_message := dialog.NewInformation("Error", full_sign_path+" not found", window)
									error_message.Resize(fyne.NewSize(300, 150))
									error_message.Show()
									os.Remove(full_save_path)
									log.Printf("[%s] Contract creation loop complete\n", tag)
									return
								} else {
									if string(data[0:35]) == "-----BEGIN DERO SIGNED MESSAGE-----" {
										split := strings.Split(string(data[0:247]), "\n")
										checkC = strings.TrimSpace(split[2][3:])
										checkS = strings.TrimSpace(split[3][3:])
									} else {
										log.Printf("[%s] %s not a valid Dero .sign file\n", tag, full_sign_path)
										wait = false
										progress_label.SetText(full_sign_path + " not a valid Dero .sign file")
										wait_message.SetDismissText("Close")
										error_message := dialog.NewInformation("Error", full_sign_path+" not a valid Dero .sign file", window)
										error_message.Resize(fyne.NewSize(300, 150))
										error_message.Show()
										os.Remove(full_save_path)
										log.Printf("[%s] Contract creation loop complete\n", tag)
										return
									}
								}

								new_contract := CreateNFAContract(art, royalty, update, name, description, typeHdr, icon, tags, checkC, checkS, file, sign, cover, collection)
								if _, err := f.WriteString(new_contract); err != nil {
									log.Printf("[%s] %s\n", tag, err)
									wait = false
									progress_label.SetText(fmt.Sprintf("Error writing %s.bas", name))
									wait_message.SetDismissText("Close")
									error_message := dialog.NewInformation("Error", fmt.Sprintf("Could not write %s", full_save_path), window)
									error_message.Resize(fyne.NewSize(300, 150))
									error_message.Show()
									os.Remove(full_save_path)
									log.Printf("[%s] Contract creation loop complete\n", tag)
									return
								}
							} else {
								log.Printf("[%s] %s\n", tag, err)
								wait = false
								progress_label.SetText(fmt.Sprintf("Error creating %s.bas", name))
								wait_message.SetDismissText("Close")
								error_message := dialog.NewInformation("Error", fmt.Sprintf("Error creating %s.bas", full_save_path), window)
								error_message.Resize(fyne.NewSize(300, 150))
								error_message.Show()
								log.Printf("[%s] Contract creation loop complete\n", tag)
								return
							}

							log.Printf("[%s] Saved NFA Contract, please check %s\n", tag, full_save_path)

							time.Sleep(time.Second)
							count++
							progress.SetValue(float64(count))

						}

						wait = false
						wait_message.SetDismissText("Done")
						progress_label.SetText("Contracts in " + save_path + "/bas")
						log.Printf("[%s] Contract creation loop complete\n", tag)
					}()
				}
			}, window)

			confirm.Resize(fyne.NewSize(600, 240))
			confirm.Show()
		} else {
			ext := extension_select.Selected
			info := fmt.Sprintf("You are about to create a .bas contract file for %s%s\n\nEnsure file sign is correct for this asset", name_entry.Text, ext)
			confirm := dialog.NewConfirm("Contract Creation", info, func(b bool) {
				if b {
					name = name_entry.Text
					file_name := fmt.Sprintf("%s.bas", name)
					full_save_path := save_path + "/bas/" + file_name
					if f, err := os.Create(full_save_path); err == nil {
						file = file_entry_start.Text
						cover = cover_entry_start.Text
						icon = icon_entry_start.Text
						sign = sign_entry_start.Text

						checkC = checkC_entry.Text
						checkS = checkS_entry.Text

						new_contract := CreateNFAContract(art, royalty, update, name, description, typeHdr, icon, tags, checkC, checkS, file, sign, cover, collection)
						if _, err := f.WriteString(new_contract); err != nil {
							log.Printf("[%s] %s\n", tag, err)
							return
						}
					} else {
						log.Printf("[%s] %s\n", tag, err)
						error_message := dialog.NewInformation("Error", fmt.Sprintf("Error creating %s", full_save_path), window)
						error_message.Resize(fyne.NewSize(300, 150))
						error_message.Show()
						return
					}

					create_message := dialog.NewInformation("Created Contract", "Check "+full_save_path, window)
					create_message.Resize(fyne.NewSize(240, 150))
					create_message.Show()

					log.Printf("[%s] Saved NFA Contract, please check %s\n", tag, full_save_path)
				}
			}, window)

			confirm.Resize(fyne.NewSize(600, 240))
			confirm.Show()
		}
	}

	install_button.OnTapped = func() {
		if rpc.Wallet.Connect {
			if collection_enable.Checked {
				var count, ending_at int
				if count = rpc.StringToInt(collection_low_entry.Text); count < 1 {
					log.Printf("[%s] Not starting installs from 0\n", tag)
					error_message := dialog.NewInformation("NFA Install", "Not starting installs from 0", window)
					error_message.Resize(fyne.NewSize(300, 150))
					error_message.Show()
					return
				}

				if ending_at = rpc.StringToInt(collection_high_entry.Text); ending_at < count {
					log.Printf("[%s] Ending is less than starting at\n", tag)
					error_message := dialog.NewInformation("NFA Install", "Ending number is less than starting number", window)
					error_message.Resize(fyne.NewSize(300, 150))
					error_message.Show()
					return
				}

				var input string
				if collection_enable.Checked {
					input = fmt.Sprintf("NFA-Creation/%s/bas/%s%s.bas", collection_entry.Text, name_entry.Text, collection_low_entry.Text)
				} else {
					input = fmt.Sprintf("NFA-Creation/%s/bas/%s.bas", collection_entry.Text, name_entry.Text)
				}
				_, err := os.Stat(input)
				if os.IsNotExist(err) {
					info := fmt.Sprintf("%s not found", input)
					error_message := dialog.NewInformation("Error", info, window)
					error_message.Resize(fyne.NewSize(300, 150))
					error_message.Show()
					return
				}

				total := ending_at - count + 1
				total_fees := float64(0.21) * float64(total)
				info := fmt.Sprintf("You are about to install %d asset files\n\nEnsure all immutable info is correct on bas contracts as this process is irreversible\n\nTotal fees to install this NFA collection will be ~%.5f Dero\n\nRefer how to mint guide for any questions\n\nWallet address: %s\n\nInstalling: %s%d.bas to %s%d.bas", total, total_fees, rpc.Wallet.Address, name_entry.Text, count, name_entry.Text, ending_at)
				confirm := dialog.NewConfirm("NFA Install", info, func(b bool) {
					if b {
						go func() {
							wait := true
							progress_label := widget.NewLabel("")
							progress_label.Alignment = fyne.TextAlignCenter
							progress := widget.NewProgressBar()
							progress.Min = float64(count)
							progress.Max = float64(ending_at)
							progress_cont := container.NewBorder(nil, progress_label, nil, nil, progress)
							wait_message := dialog.NewCustom("NFA Install", "Stop", progress_cont, window)
							wait_message.SetOnClosed(func() {
								if wait {
									wait = false
								}
							})
							wait_message.Resize(fyne.NewSize(450, 150))
							wait_message.Show()

							log.Printf("[%s] Starting install loop\n", tag)

							for wait && count <= ending_at {
								if !rpc.Wallet.Connect {
									progress_label.SetText("Error wallet disconnected")
									wait_message.SetDismissText("Close")
									error_message := dialog.NewInformation("Error", "Wallet rpc disconnected", window)
									error_message.Resize(fyne.NewSize(300, 150))
									error_message.Show()
									break
								}

								input_file := fmt.Sprintf("NFA-Creation/%s/bas/%s%d.bas", collection_entry.Text, name_entry.Text, count)
								if _, err := os.Stat(input_file); err == nil {
									log.Printf("[%s] Installing %s\n", tag, input_file)
									progress_label.SetText(fmt.Sprintf("Installing %s", input_file))
								} else if errors.Is(err, os.ErrNotExist) {
									log.Printf("[%s] %s not found\n", tag, input_file)
									progress_label.SetText(fmt.Sprintf("Error %s file not found", input_file))
									wait_message.SetDismissText("Close")
									error_message := dialog.NewInformation("Error", input_file+" not found", window)
									error_message.Resize(fyne.NewSize(300, 150))
									error_message.Show()
									break
								}

								file, err := os.ReadFile(input_file)
								if err != nil {
									log.Printf("[%s] %s\n", tag, err)
									progress_label.SetText(fmt.Sprintf("Error reading %s", input_file))
									wait_message.SetDismissText("Close")
									error_message := dialog.NewInformation("Error", fmt.Sprintf("Could not read %s", input_file), window)
									error_message.Resize(fyne.NewSize(300, 150))
									error_message.Show()
									break
								}

								if string(file) == "" {
									error_message := dialog.NewInformation("Error", fmt.Sprintf("%s is a empty file", input_file), window)
									error_message.Resize(fyne.NewSize(240, 150))
									error_message.Show()
									break
								}

								if tx := rpc.UploadNFAContract(string(file)); tx == "" {
									progress_label.SetText("Error installing " + input_file)
									wait_message.SetDismissText("Close")
									error_message := dialog.NewInformation("Error", fmt.Sprintf("Could not install %s", input_file), window)
									error_message.Resize(fyne.NewSize(300, 150))
									error_message.Show()
									break
								} else {
									log.Printf("[%s] Confirming install TX\n", tag)
									rpc.ConfirmTx(tx, tag, 45)
								}

								count++
								progress.SetValue(float64(count))
								if count > ending_at {
									progress_label.SetText("All contracts installed, check wallet for NFA balance")
									wait_message.SetDismissText("Done")
									break
								}

								time.Sleep(6 * time.Second)
							}

							wait = false
							log.Printf("[%s] Install loop complete\n", tag)
						}()
					}
				}, window)
				confirm.Resize(fyne.NewSize(600, 240))
				confirm.Show()
			} else {
				info := fmt.Sprintf("You are about to install asset %s.bas\n\nEnsure all immutable info is correct on bas contract as this process is irreversible\n\nFees to install a NFA are ~0.21000 Dero\n\nRefer how to mint guide for any questions\n\nWallet address: %s", name_entry.Text, rpc.Wallet.Address)
				confirm := dialog.NewConfirm("NFA Install", info, func(b bool) {
					if b {
						input_file := fmt.Sprintf("NFA-Creation/%s/bas/%s.bas", collection_entry.Text, name_entry.Text)
						if _, err := os.Stat(input_file); err == nil {
							log.Printf("[%s] Installing %s\n", tag, input_file)
						} else if errors.Is(err, os.ErrNotExist) {
							log.Printf("[%s] %s not found\n", tag, input_file)
							return
						}

						file, err := os.ReadFile(input_file)
						if err != nil {
							log.Printf("[%s] %s\n", tag, err)
							return
						}

						if string(file) == "" {
							error_message := dialog.NewInformation("Error", fmt.Sprintf("%s is a empty file", input_file), window)
							error_message.Resize(fyne.NewSize(240, 150))
							error_message.Show()
							return
						}

						if tx := rpc.UploadNFAContract(string(file)); tx == "" {
							error_message := dialog.NewInformation("Error", "Could not install NFA", window)
							error_message.Resize(fyne.NewSize(240, 150))
							error_message.Show()
						} else {
							install_message := dialog.NewInformation("Installed NFA", tx, window)
							install_message.SetDismissText("Copy")
							install_message.SetOnClosed(func() {
								window.Clipboard().SetContent(tx)
							})
							install_message.Resize(fyne.NewSize(240, 150))
							install_message.Show()
						}

					}
				}, window)
				confirm.Resize(fyne.NewSize(600, 240))
				confirm.Show()
			}
		}
	}

	collection_cont := container.NewBorder(nil, nil, collection_enable, container.NewAdaptiveGrid(2, url_add_incr, url_remove_incr), container.NewAdaptiveGrid(2, collection_low_entry, collection_high_entry))

	wallet_cont := container.NewBorder(nil, container.NewAdaptiveGrid(2, rpc_label, wallet_label), nil, nil, container.NewAdaptiveGrid(2, wallet_pass_entry, open_wallet_button))

	instructions_button := widget.NewButton("How To Mint", nil)

	instructions_button.Importance = widget.LowImportance
	set_up_collec.Importance = widget.LowImportance
	import_signs.Importance = widget.LowImportance
	url_add_incr.Importance = widget.LowImportance
	url_remove_incr.Importance = widget.LowImportance

	collection_low_entry.Hide()
	collection_high_entry.Hide()
	url_add_incr.Hide()
	url_remove_incr.Hide()
	contracts_button.Hide()
	install_button.Hide()
	sign_button.Hide()

	mint_form := []*widget.FormItem{}
	mint_form = append(mint_form, widget.NewFormItem("", instructions_button))
	mint_form = append(mint_form, widget.NewFormItem("", collection_cont))
	mint_form = append(mint_form, widget.NewFormItem("", layout.NewSpacer()))
	mint_form = append(mint_form, widget.NewFormItem("Collection", container.NewBorder(nil, nil, nil, set_up_collec, collection_entry)))
	mint_form = append(mint_form, widget.NewFormItem("Owner Can Update", update_select))
	mint_form = append(mint_form, widget.NewFormItem("Name", container.NewBorder(nil, nil, nil, extension_select, name_cont)))
	mint_form = append(mint_form, widget.NewFormItem("Description", descr_entry))
	mint_form = append(mint_form, widget.NewFormItem("Type", type_select))
	mint_form = append(mint_form, widget.NewFormItem("Tags *", tags_entry))
	mint_form = append(mint_form, widget.NewFormItem("File Check C", container.NewBorder(nil, nil, nil, import_signs, checkC_entry)))
	mint_form = append(mint_form, widget.NewFormItem("File Check S", checkS_entry))
	mint_form = append(mint_form, widget.NewFormItem("File URL *", file_entries))
	mint_form = append(mint_form, widget.NewFormItem("Cover URL *", cover_entries))
	mint_form = append(mint_form, widget.NewFormItem("Icon URL *", icon_entries))
	mint_form = append(mint_form, widget.NewFormItem("File Sign URL *", sign_entries))
	mint_form = append(mint_form, widget.NewFormItem("", layout.NewSpacer()))
	mint_form = append(mint_form, widget.NewFormItem("Royalty %", royalty_entry))
	mint_form = append(mint_form, widget.NewFormItem("Artificer %", art_entry))
	mint_form = append(mint_form, widget.NewFormItem("", layout.NewSpacer()))
	mint_form = append(mint_form, widget.NewFormItem("Wallet File", wallet_cont))
	mint_form = append(mint_form, widget.NewFormItem("", layout.NewSpacer()))
	mint_form = append(mint_form, widget.NewFormItem("", container.NewAdaptiveGrid(3, container.NewMax(install_button), container.NewMax(contracts_button), sign_button)))

	scroll := container.NewVScroll(widget.NewForm(mint_form...))
	max := container.NewMax(scroll)
	instructions_back_button := widget.NewButton("Back", func() {
		max.Objects[0] = container.NewMax(scroll)
	})

	instructions_button.OnTapped = func() {
		max.Objects[0] = HowToMintNFA(instructions_back_button)
	}

	go func() {
		for {
			if collection_entry.Validate() == nil && update_select.SelectedIndex() >= 0 && name_entry.Validate() == nil && extension_select.SelectedIndex() >= 0 && descr_entry.Validate() == nil &&
				type_select.SelectedIndex() >= 0 && tags_entry.Validate() == nil && checkC_entry.Validate() == nil && checkS_entry.Validate() == nil &&
				file_entry_start.Validate() == nil && cover_entry_start.Validate() == nil && icon_entry_start.Validate() == nil &&
				sign_entry_start.Validate() == nil && royalty_entry.Validate() == nil && art_entry.Validate() == nil {
				if collection_enable.Checked {
					if collection_low_entry.Validate() == nil && collection_high_entry.Validate() == nil {
						asset_file_start := fmt.Sprintf("NFA-Creation/%s/asset/%s%s%s", collection_entry.Text, name_entry.Text, collection_low_entry.Text, extension_select.Selected)
						if _, err := os.ReadFile(asset_file_start); err == nil {
							contracts_button.Show()
						} else {
							contracts_button.Hide()
						}
					} else {
						contracts_button.Hide()
					}
				} else {
					asset_file := fmt.Sprintf("NFA-Creation/%s/asset/%s%s", collection_entry.Text, name_entry.Text, extension_select.Selected)
					if _, err := os.ReadFile(asset_file); err == nil {
						contracts_button.Show()
					} else {
						contracts_button.Hide()
					}
				}
			} else {
				contracts_button.Hide()
			}

			if NFACreationExists(collection_entry.Text) {
				if rpc.Wallet.File != nil && rpc.Wallet.File.Check_Password(wallet_pass_entry.Text) {
					wallet_label.SetText(fmt.Sprintf("Signing address: %s", rpc.Wallet.File.GetAddress().String()))
					if collection_enable.Checked {
						if collection_low_entry.Text != "" && collection_high_entry.Text != "" {
							sign_file_start := fmt.Sprintf("NFA-Creation/%s/asset/%s%s%s", collection_entry.Text, name_entry.Text, collection_low_entry.Text, extension_select.Selected)
							if _, err := os.ReadFile(sign_file_start); err == nil {
								sign_button.Show()
							} else {
								sign_button.Hide()
							}
						} else {
							sign_button.Hide()
						}
					} else {
						sign_file := fmt.Sprintf("NFA-Creation/%s/asset/%s%s", collection_entry.Text, name_entry.Text, extension_select.Selected)
						if _, err := os.ReadFile(sign_file); err == nil {
							sign_button.Show()
						} else {
							sign_button.Hide()
						}
					}
				} else {
					wallet_label.SetText("Signing address:")
				}
				collection_entry.Validator = validation.NewRegexp(`^\w{2,}`, "String required")
			} else {
				sign_button.Hide()
				collection_entry.Validator = validation.NewRegexp(`^\W\D\S$`, "Invliad collection directory")
			}

			if rpc.Wallet.Connect && rpc.Daemon.Connect {
				rpc_label.SetText(fmt.Sprintf("Installing address: %s", rpc.Wallet.Address))
				if collection_enable.Checked {
					if collection_low_entry.Text != "" && collection_high_entry.Text != "" {
						input_file_start := fmt.Sprintf("NFA-Creation/%s/bas/%s%s.bas", collection_entry.Text, name_entry.Text, collection_low_entry.Text)
						if _, err := os.ReadFile(input_file_start); err == nil {
							install_button.Show()
						} else {
							install_button.Hide()
						}
					} else {
						install_button.Hide()
					}
				} else {
					input_file := fmt.Sprintf("NFA-Creation/%s/bas/%s.bas", collection_entry.Text, name_entry.Text)
					if _, err := os.ReadFile(input_file); err == nil {
						install_button.Show()
					} else {
						install_button.Hide()
					}
				}
			} else {
				rpc_label.SetText("Installing address:")
				install_button.Hide()
			}

			time.Sleep(1 * time.Second)
		}
	}()

	return max
}

// Set up NFA-Creation directory with sub directory for collection or single asset,
// which contains sub directories for asset, bas, icon, cover and sign files
func SetUpNFACreation(tag, collection string) (save_path string, sign_path string) {
	main_path := "NFA-Creation"
	_, main := os.Stat(main_path)
	if os.IsNotExist(main) {
		log.Printf("[%s] Creating NFA-Creation Dir\n", tag)
		if err := os.Mkdir(main_path, 0755); err != nil {
			log.Printf("[%s] %s\n", tag, err)
			return
		}
	}

	save_path = fmt.Sprintf("%s/%s", main_path, collection)
	_, coll := os.Stat(save_path)
	if os.IsNotExist(coll) {
		err := os.Mkdir(save_path, 0755)
		if err != nil {
			log.Printf("[%s] %s\n", tag, err)
			return
		}
	}

	asset_path := save_path + "/asset"
	_, asset := os.Stat(asset_path)
	if os.IsNotExist(asset) {
		err := os.Mkdir(asset_path, 0755)
		if err != nil {
			log.Printf("[%s] %s\n", tag, err)
			return
		}
	}

	bas_path := save_path + "/bas"
	_, bas := os.Stat(bas_path)
	if os.IsNotExist(bas) {
		log.Printf("[%s] Creating bas Dir\n", tag)
		err := os.Mkdir(bas_path, 0755)
		if err != nil {
			log.Printf("[%s] %s\n", tag, err)
			return
		}
	}

	cover_path := save_path + "/cover"
	_, cover := os.Stat(cover_path)
	if os.IsNotExist(cover) {
		err := os.Mkdir(cover_path, 0755)
		if err != nil {
			log.Printf("[%s] %s\n", tag, err)
			return
		}
	}

	icon_path := save_path + "/icon"
	_, icon := os.Stat(icon_path)
	if os.IsNotExist(icon) {
		log.Printf("[%s] Creating icons Dir\n", tag)
		err := os.Mkdir(icon_path, 0755)
		if err != nil {
			log.Printf("[%s] %s\n", tag, err)
			return
		}
	}

	sign_path = save_path + "/sign"
	_, sign := os.Stat(sign_path)
	if os.IsNotExist(sign) {
		log.Printf("[%s] Creating sign Dir\n", tag)
		err := os.Mkdir(sign_path, 0755)
		if err != nil {
			log.Printf("[%s] %s\n", tag, err)
			return
		}
	}

	return
}

// Check that all creation directories exists
func NFACreationExists(collection string) bool {
	main_path := "NFA-Creation"
	_, main := os.Stat(main_path)
	if os.IsNotExist(main) {
		return false
	}

	save_path := fmt.Sprintf("%s/%s", main_path, collection)
	_, coll := os.Stat(save_path)
	if os.IsNotExist(coll) {
		return false
	}

	asset_path := save_path + "/asset"
	_, asset := os.Stat(asset_path)
	if os.IsNotExist(asset) {
		return false
	}

	bas_path := save_path + "/bas"
	_, bas := os.Stat(bas_path)
	if os.IsNotExist(bas) {
		return false
	}

	cover_path := save_path + "/cover"
	_, cover := os.Stat(cover_path)
	if os.IsNotExist(cover) {
		return false
	}

	icon_path := save_path + "/icon"
	_, icon := os.Stat(icon_path)
	if os.IsNotExist(icon) {
		return false
	}

	sign_path := save_path + "/sign"
	_, sign := os.Stat(sign_path)

	return !os.IsNotExist(sign)
}
