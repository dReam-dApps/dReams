package menu

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	dreams "github.com/dReam-dApps/dReams"
	"github.com/dReam-dApps/dReams/bundle"
	"github.com/dReam-dApps/dReams/dwidget"
	"github.com/dReam-dApps/dReams/rpc"
	"github.com/deroproject/derohe/dvm"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/validation"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
)

const (
	SIX_mint       = "dero1qy4ascka9rtspjvcyj6t8maazaay8t9udtt5nper3mukqkx2qtvyxqgflkpwp"
	AZY_mint       = "dero1qyfk5w2rvqpl9kzfd7fpteyp2k362y6audydcu2qrgcmj6vtasfkgqq9704gn"
	DCB_mint       = "dero1qy02stluwgh5aaawkmugqh47krtfzcq6f88jv2ydf6dkfupjca4gzqqwsmdzf"
	HS_mint        = "dero1qy8p8cw8hr8dlxyg9xjxzlxf2zznwshcxk98ad688csgmk44y9kzyqqt9g2m0"
	Dorbling_mint  = "dero1qy2tkfgpapjsgev8m9rs25209ajc6kcm42vzt2fapgk3ngmtc2guzqq52eq2r"
	Desperado_mint = "dero1qy0ydkcwuf7nvh6938jpalt5snsj2atgmdyms67rd05whuy8a2hvzqg46gh5f"
	DSkull_mint    = "dero1qyfn6esxwqk6mzx3whqtdfrz5aytu4nsfwk9hw02dp9p2nvf360r5qgzp2qre"
)

type mintConfig struct {
	Collection  string   `json:"collection"`
	Update      string   `json:"update"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	TypeHdr     string   `json:"typeHdr"`
	Tags        string   `json:"tags"`
	Start       string   `json:"start"`
	End         string   `json:"end"`
	Extension   string   `json:"extension"`
	File        []string `json:"file"`
	Cover       []string `json:"cover"`
	Icon        []string `json:"icon"`
	Sign        []string `json:"sign"`
	Multi       int      `json:"multi"`
	Royalty     string   `json:"royalty"`
	Artificer   string   `json:"artificer"`
}

//go:embed ART-NFA-MS1.bas
var ART_NFA_MS1 string

// Tree object containing NFA minting instructions screen
func HowToMintNFA(button *widget.Button) fyne.CanvasObject {
	list := map[string][]string{
		"":                      {"How To Mint NFAs"},
		"How To Mint NFAs":      {"Get Started", "Single Asset", "Collection"},
		"Get Started":           {"A NFA consists of four main parts: Asset file, Cover Image, Icon image, Dero file sign", "Each NFA is its own self contained marketplace", "This tool automates three areas of NFA installs: File sign, Contract creation, Contract install", "Storage is not provided at this point", "Entries with a * are mutable, meaning they can be updated by creator (or owner) after install", "Gas fees to install a NFA are ~0.20000 Dero", "There is a 0.00500 Dero dev fee for minting a NFA with this tool", "If minting a collection fees will be paid as each contract is installed, a total will be shown before hand", "For further info read the NFA documentation at github.com/civilware/artificer-nfa-standard"},
		"Single Asset":          {"Disable the Collection check", "Type the name of your asset into the collection entry and click the folder button on right to set up creation directory for your single asset", "Creation Directory", "File sign can be imported from file by clicking the file button to right of check C entry, or follow next step if you require file sign", "Single File Sign", "Fill out the rest of the information for your NFA and when complete the Create Contracts button will show", "Click Create Contract and confirm information, this will populate your bas folder with your asset contract", "Type the name of your asset in name entry and Install Contract will show if contract exists in bas folder", "Click Install Contract and confirm the install address is same as signing address", "NFA is now installed, check your wallet for NFA balance"},
		"Single File Sign":      {"Enter minting wallet file password and open minting wallet file", "Place asset file into asset folder", "Enter the name of your asset in name entry, select extension to match file", "Click Sign File and confirm information", "Once confirmed, the file check C and file check S will population with your file signs"},
		"Collection":            {"Collection automation installs assets of same name with incrementing numbers", "Enable the Collection check", "Type the name of your collection into the collection entry and click the folder button on right to set up creation directory for your collection", "Creation Directory", "Enter the starting number and ending number for your collection", "File signs can be done externally and placed into sign folder, or follow next step if you require file signs", "Collection File Signs", "Make sure file signs are in sign directory for contract creation", "Fill out the rest of the information for your NFA collection and when complete the Create Contracts button will show", "The Asset Number sections are where it will add the incrementing number to your input to make the collection", "The + - buttons on top right can add or remove a increment section from Url paths", "Click Create Contract and confirm information", "Contract creation loop will start and populate your bas folder with your asset contracts, takes about 1 second per contract", "Type the name of your asset in name entry and Install Contract will show if contract exists in bas folder", "Click Install Contract and confirm the install address is same as signing address", "Minting loop will now start and installs one bas contract per block", "For larger collections this could take some time, 120 installs could take around 1 hour to complete", "If 100%, NFA collection is now installed, check your wallet for NFA balances"},
		"Collection File Signs": {"Enter minting wallet file password and open minting wallet file", "Place numbered asset files into asset folder", "Enter the name of your asset in name entry, select extension to match file", "Click Sign File and confirm information", "This starts a file sign loop of your selected range and stores all signed files in sign directory, takes about 1 second per sign"},
		"Creation Directory":    {"Creation directory stores collection and single asset directories", "Inside of your asset or collection directory are five sub directories", "Your main asset files are stored in asset", "Contracts created are stored in bas", "Signed files are stored in sign", "Cover and icon are optional directories at this point and are not used in the install process"},
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

	return container.NewBorder(nil, container.NewCenter(button), nil, nil, tree)
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

	return container.NewStack(rect, add_text)
}

// Place objects for NFA minting of collections or single mint
func PlaceNFAMint(tag string, window fyne.Window) fyne.CanvasObject {
	collection_enable := widget.NewCheck("Collection", nil)

	sign_button := widget.NewButton("File Sign", nil)
	sign_button.Importance = widget.HighImportance

	contracts_button := widget.NewButton("Create Contract", nil)
	contracts_button.Importance = widget.HighImportance

	install_button := widget.NewButton("Install Contract", nil)
	install_button.Importance = widget.HighImportance

	collection_high_entry := dwidget.NewAmountEntry("", 1, 0)
	collection_high_entry.SetPlaceHolder("Ending At:")
	collection_high_entry.AllowFloat = false
	collection_high_entry.Validator = validation.NewRegexp(`^[^0]\d{0,}$`, "Uint required")

	collection_low_entry := dwidget.NewAmountEntry("", 1, 0)
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

	set_up_collec := widget.NewButtonWithIcon("", dreams.FyneIcon("folderNew"), func() {
		if collection_entry.Text != "" {
			if collection_entry.Validate() == nil {
				info_message := dialog.NewInformation("Collection Exists", "Check creation directory", window)
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

			return
		}

		info := dialog.NewInformation("Create", "Enter a collection name to create", window)
		info.SetOnClosed(collection_entry.FocusLost)
		collection_entry.FocusGained()
		info.Show()

	})

	descr_entry := widget.NewEntry()
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

	import_signs := widget.NewButtonWithIcon("", dreams.FyneIcon("document"), func() {
		read_filesign := dialog.NewFileOpen(func(uc fyne.URIReadCloser, err error) {
			if err == nil && uc != nil {
				readC, readS, _ := ReadDeroSignFile(tag, uc.URI().Path())
				checkC_entry.SetText(readC)
				checkS_entry.SetText(readS)
			}
		}, window)
		if uri, err := createURI(); err == nil {
			read_filesign.SetLocation(uri)
		}
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

	url_add_incr := widget.NewButtonWithIcon("", dreams.FyneIcon("contentAdd"), func() {
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

	url_remove_incr := widget.NewButtonWithIcon("", dreams.FyneIcon("contentRemove"), func() {
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

	royalty_entry := dwidget.NewAmountEntry("", 1, 0)
	royalty_entry.AllowFloat = false
	royalty_entry.SetPlaceHolder("% you will get from each sale:")
	royalty_entry.Validator = validation.NewRegexp(`^\d{1,2}$`, "Uint required")

	art_entry := dwidget.NewAmountEntry("", 1, 0)
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

	sign_button.OnTapped = func() {
		save_path, sign_path := SetUpNFACreation(tag, collection_entry.Text)
		file_name := fmt.Sprintf("%s%s.sign", name_entry.Text, extension_select.Selected)
		if collection_enable.Checked {
			file_name = fmt.Sprintf("%s%s%s.sign", name_entry.Text, collection_low_entry.Text, extension_select.Selected)
		}

		full_sign_path := filepath.Join(sign_path, file_name)
		_, err := os.Stat(full_sign_path)
		if !os.IsNotExist(err) {
			info := fmt.Sprintf("Check %s", full_sign_path)
			info_message := dialog.NewInformation("File Exists", info, window)
			info_message.Resize(fyne.NewSize(300, 150))
			info_message.Show()
			readC, readS, _ := ReadDeroSignFile(tag, full_sign_path)
			checkC_entry.SetText(readC)
			checkS_entry.SetText(readS)
			return
		}

		ext := extension_select.Selected

		var input string
		if collection_enable.Checked {
			input = filepath.Join(save_path, "asset", fmt.Sprintf("%s%s%s", name_entry.Text, collection_low_entry.Text, ext))
		} else {
			input = filepath.Join(save_path, "asset", fmt.Sprintf("%s%s", name_entry.Text, ext))
		}
		_, err = os.Stat(input)
		if os.IsNotExist(err) {
			info := fmt.Sprintf("%s not found", input)
			error_message := dialog.NewInformation("File Sign", info, window)
			error_message.Resize(fyne.NewSize(300, 150))
			error_message.Show()
			return
		}

		if !rpc.Wallet.File.IsNil() && extension_select.SelectedIndex() >= 0 {
			address := rpc.Wallet.Address
			if collection_enable.Checked {
				var count, ending_at int
				if count = rpc.StringToInt(collection_low_entry.Text); count < 1 {
					logger.Warnf("[%s] Not starting signs from 0\n", tag)

					return
				}

				if ending_at = rpc.StringToInt(collection_high_entry.Text); ending_at < count {
					logger.Warnf("[%s] Ending is less than starting at\n", tag)

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

							logger.Printf("[%s] Starting sign loop\n", tag)

							for wait {
								i := strconv.Itoa(count)
								signing_asset := fmt.Sprintf("Signing %s%s%s", name_entry.Text, i, ext)
								input_file := filepath.Join(save_path, "asset", fmt.Sprintf("%s%s%s", name_entry.Text, i, ext))
								output_file := filepath.Join(sign_path, fmt.Sprintf("%s%s%s%s", name_entry.Text, i, ext, ".sign"))
								_, err := os.Stat(output_file)
								if !os.IsNotExist(err) {
									logger.Warnf("[%s] %s exists\n", tag, output_file)
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
									logger.Errorf("[%s] Cannot read input file %s\n", tag, err)
									progress_label.SetText("Error " + input_file + " not found")
									wait_message.SetDismissText("Close")
									error_message := dialog.NewInformation("Error", input_file+" not found", window)
									error_message.Resize(fyne.NewSize(300, 150))
									error_message.Show()
									break
								} else if err := os.WriteFile(output_file, rpc.Wallet.File.SignData(data), 0600); err != nil {
									logger.Errorf("[%s] Cannot write output file %s\n", tag, output_file)
									progress_label.SetText("Error writing " + output_file)
									wait_message.SetDismissText("Close")
									error_message := dialog.NewInformation("Error", "Could not write "+output_file, window)
									error_message.Resize(fyne.NewSize(300, 150))
									error_message.Show()
									break
								} else {
									logger.Printf("[%s] Successfully signed file, please check %s\n", tag, output_file)
								}

								time.Sleep(time.Second)
								count++
								progress.SetValue(float64(count))
								if count > ending_at {
									progress_label.SetText(fmt.Sprintf("Signed files in creation/%s/sign", collection_entry.Text))
									wait_message.SetDismissText("Done")
									break
								}
							}

							wait = false
							logger.Printf("[%s] Sign loop complete\n", tag)
						}()
					}
				}, window)

				confirm.Resize(fyne.NewSize(600, 240))
				confirm.Show()
			} else {
				ext := extension_select.Selected
				info := fmt.Sprintf("You are about to sign asset file %s%s\n\nSigning address: %s", name_entry.Text, ext, address)
				confirm := dialog.NewConfirm("File sign", info, func(b bool) {
					if b {
						input_file := filepath.Join(save_path, "asset", fmt.Sprintf("%s%s", name_entry.Text, ext))
						output_file := filepath.Join(sign_path, fmt.Sprintf("%s%s%s", name_entry.Text, ext, ".sign"))
						if data, err := os.ReadFile(input_file); err != nil {
							logger.Errorf("[%s] Cannot read input file %s\n", tag, err)
							error_message := dialog.NewInformation("Error", "Could not read file", window)
							error_message.Resize(fyne.NewSize(300, 150))
							error_message.Show()
						} else if err := os.WriteFile(output_file, rpc.Wallet.File.SignData(data), 0600); err != nil {
							logger.Errorf("[%s] Cannot write output file %s\n", tag, output_file)
						} else {
							logger.Printf("[%s] Successfully signed file. please check %s\n", tag, output_file)
							readC, readS, _ := ReadDeroSignFile(tag, output_file)
							checkC_entry.SetText(readC)
							checkS_entry.SetText(readS)
							info_message := dialog.NewInformation("File Signed", output_file, window)
							info_message.Resize(fyne.NewSize(300, 150))
							info_message.Show()
						}
					}
				}, window)

				confirm.Resize(fyne.NewSize(600, 240))
				confirm.Show()
			}
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

		_, err := os.Stat(filepath.Join(save_path, "bas", file_name))
		if !os.IsNotExist(err) {
			info := fmt.Sprintf("Check %s", filepath.Join(save_path, "bas", file_name))
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
				logger.Warnf("[%s] Not starting collection from 0\n", tag)
				error_message := dialog.NewInformation("Contract Creation", "Not starting collection from 0", window)
				error_message.Resize(fyne.NewSize(300, 150))
				error_message.Show()
				return
			}

			if ending_at = rpc.StringToInt(collection_high_entry.Text); ending_at < count {
				logger.Warnf("[%s] Ending is less than starting at\n", tag)
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

						logger.Printf("[%s] Starting contract creation loop\n", tag)

						for wait && count <= ending_at {
							incr := strconv.Itoa(count)
							name = name_entry.Text + incr
							file_name := fmt.Sprintf("%s.bas", name)
							progress_label.SetText("Creating " + file_name)
							full_save_path := filepath.Join(save_path, "bas", file_name)
							_, err := os.Stat(full_save_path)
							if !os.IsNotExist(err) {
								wait = false
								progress_label.SetText("Error " + full_save_path + " exists")
								wait_message.SetDismissText("Close")
								info := fmt.Sprintf("Check %s", full_save_path)
								info_message := dialog.NewInformation("File Exists", info, window)
								info_message.Resize(fyne.NewSize(300, 150))
								info_message.Show()
								logger.Printf("[%s] Contract creation loop complete\n", tag)
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
									logger.Printf("[%s] Contract creation loop complete\n", tag)
									return
								}

								ext := extension_select.Selected + ".sign"
								full_sign_path := filepath.Join(sign_path, name+ext)
								if data, err := os.ReadFile(full_sign_path); err != nil {
									wait = false
									logger.Errorf("[%s] Cannot read input file %s\n", tag, err)
									progress_label.SetText("Error could not read " + full_sign_path)
									wait_message.SetDismissText("Close")
									error_message := dialog.NewInformation("Error", full_sign_path+" not found", window)
									error_message.Resize(fyne.NewSize(300, 150))
									error_message.Show()
									os.Remove(full_save_path)
									logger.Printf("[%s] Contract creation loop complete\n", tag)
									return
								} else {
									if string(data[0:35]) == "-----BEGIN DERO SIGNED MESSAGE-----" {
										split := strings.Split(string(data[0:247]), "\n")
										checkC = strings.TrimSpace(split[2][3:])
										checkS = strings.TrimSpace(split[3][3:])
									} else {
										logger.Errorf("[%s] %s not a valid Dero .sign file\n", tag, full_sign_path)
										wait = false
										progress_label.SetText(full_sign_path + " not a valid Dero .sign file")
										wait_message.SetDismissText("Close")
										error_message := dialog.NewInformation("Error", full_sign_path+" not a valid Dero .sign file", window)
										error_message.Resize(fyne.NewSize(300, 150))
										error_message.Show()
										os.Remove(full_save_path)
										logger.Printf("[%s] Contract creation loop complete\n", tag)
										return
									}
								}

								new_contract := CreateNFAContract(art, royalty, update, name, description, typeHdr, icon, tags, checkC, checkS, file, sign, cover, collection)
								if _, err := f.WriteString(new_contract); err != nil {
									logger.Errorf("[%s] %s\n", tag, err)
									wait = false
									progress_label.SetText(fmt.Sprintf("Error writing %s.bas", name))
									wait_message.SetDismissText("Close")
									error_message := dialog.NewInformation("Error", fmt.Sprintf("Could not write %s", full_save_path), window)
									error_message.Resize(fyne.NewSize(300, 150))
									error_message.Show()
									os.Remove(full_save_path)
									logger.Printf("[%s] Contract creation loop complete\n", tag)
									return
								}
							} else {
								logger.Errorf("[%s] %s\n", tag, err)
								wait = false
								progress_label.SetText(fmt.Sprintf("Error creating %s.bas", name))
								wait_message.SetDismissText("Close")
								error_message := dialog.NewInformation("Error", fmt.Sprintf("Error creating %s.bas", full_save_path), window)
								error_message.Resize(fyne.NewSize(300, 150))
								error_message.Show()
								logger.Printf("[%s] Contract creation loop complete\n", tag)
								return
							}

							logger.Printf("[%s] Saved NFA Contract, please check %s\n", tag, full_save_path)

							time.Sleep(time.Second)
							count++
							progress.SetValue(float64(count))

						}

						wait = false
						wait_message.SetDismissText("Done")
						progress_label.SetText("Contracts in " + save_path + "/bas")
						logger.Printf("[%s] Contract creation loop complete\n", tag)
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
					full_save_path := filepath.Join(save_path, "bas", file_name)
					if f, err := os.Create(full_save_path); err == nil {
						file = file_entry_start.Text
						cover = cover_entry_start.Text
						icon = icon_entry_start.Text
						sign = sign_entry_start.Text

						checkC = checkC_entry.Text
						checkS = checkS_entry.Text

						new_contract := CreateNFAContract(art, royalty, update, name, description, typeHdr, icon, tags, checkC, checkS, file, sign, cover, collection)
						if _, err := f.WriteString(new_contract); err != nil {
							logger.Errorf("[%s] %s\n", tag, err)
							return
						}
					} else {
						logger.Errorf("[%s] %s\n", tag, err)
						error_message := dialog.NewInformation("Error", fmt.Sprintf("Error creating %s", full_save_path), window)
						error_message.Resize(fyne.NewSize(300, 150))
						error_message.Show()
						return
					}

					create_message := dialog.NewInformation("Created Contract", "Check "+full_save_path, window)
					create_message.Resize(fyne.NewSize(240, 150))
					create_message.Show()

					logger.Printf("[%s] Saved NFA Contract, please check %s\n", tag, full_save_path)
				}
			}, window)

			confirm.Resize(fyne.NewSize(600, 240))
			confirm.Show()
		}
	}

	install_button.OnTapped = func() {
		if rpc.Wallet.IsConnected() {
			save_path, _ := SetUpNFACreation(tag, collection_entry.Text)
			if collection_enable.Checked {
				var count, ending_at int
				if count = rpc.StringToInt(collection_low_entry.Text); count < 1 {
					logger.Warnf("[%s] Not starting installs from 0\n", tag)
					error_message := dialog.NewInformation("NFA Install", "Not starting installs from 0", window)
					error_message.Resize(fyne.NewSize(300, 150))
					error_message.Show()
					return
				}

				if ending_at = rpc.StringToInt(collection_high_entry.Text); ending_at < count {
					logger.Warnf("[%s] Ending is less than starting at\n", tag)
					error_message := dialog.NewInformation("NFA Install", "Ending number is less than starting number", window)
					error_message.Resize(fyne.NewSize(300, 150))
					error_message.Show()
					return
				}

				var input string
				if collection_enable.Checked {
					input = filepath.Join(save_path, "bas", fmt.Sprintf("%s%s.bas", name_entry.Text, collection_low_entry.Text))
				} else {
					input = filepath.Join(save_path, "bas", fmt.Sprintf("%s.bas", name_entry.Text))
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

							logger.Printf("[%s] Starting install loop\n", tag)

							for wait && count <= ending_at {
								if !rpc.Wallet.IsConnected() {
									progress_label.SetText("Error wallet disconnected")
									wait_message.SetDismissText("Close")
									error_message := dialog.NewInformation("Error", "Wallet rpc disconnected", window)
									error_message.Resize(fyne.NewSize(300, 150))
									error_message.Show()
									break
								}

								input_file := filepath.Join(save_path, "bas", fmt.Sprintf("%s%d.bas", name_entry.Text, count))
								if _, err := os.Stat(input_file); err == nil {
									logger.Printf("[%s] Installing %s\n", tag, input_file)
									progress_label.SetText(fmt.Sprintf("Installing %s", input_file))
								} else if errors.Is(err, os.ErrNotExist) {
									logger.Errorf("[%s] %s not found\n", tag, input_file)
									progress_label.SetText(fmt.Sprintf("Error %s file not found", input_file))
									wait_message.SetDismissText("Close")
									error_message := dialog.NewInformation("Error", input_file+" not found", window)
									error_message.Resize(fyne.NewSize(300, 150))
									error_message.Show()
									break
								}

								file, err := os.ReadFile(input_file)
								if err != nil {
									logger.Errorf("[%s] %s\n", tag, err)
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

								if _, _, err := dvm.ParseSmartContract(string(file)); err != nil {
									error_message := dialog.NewInformation("Error", fmt.Sprintf("%s is not a valid SC", input_file), window)
									error_message.Resize(fyne.NewSize(240, 150))
									error_message.Show()
									return
								}

								if tx := rpc.UploadNFAContract(string(file)); tx == "" {
									progress_label.SetText("Error installing " + input_file)
									wait_message.SetDismissText("Close")
									error_message := dialog.NewInformation("Error", fmt.Sprintf("Could not install %s", input_file), window)
									error_message.Resize(fyne.NewSize(300, 150))
									error_message.Show()
									break
								} else {
									logger.Printf("[%s] Confirming install TX\n", tag)
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
							logger.Printf("[%s] Install loop complete\n", tag)
						}()
					}
				}, window)
				confirm.Resize(fyne.NewSize(600, 240))
				confirm.Show()
			} else {
				info := fmt.Sprintf("You are about to install asset %s.bas\n\nEnsure all immutable info is correct on bas contract as this process is irreversible\n\nFees to install a NFA are ~0.21000 Dero\n\nRefer how to mint guide for any questions\n\nWallet address: %s", name_entry.Text, rpc.Wallet.Address)
				confirm := dialog.NewConfirm("NFA Install", info, func(b bool) {
					if b {
						input_file := filepath.Join(save_path, "bas", fmt.Sprintf("%s.bas", name_entry.Text))
						if _, err := os.Stat(input_file); err == nil {
							logger.Printf("[%s] Installing %s\n", tag, input_file)
						} else if errors.Is(err, os.ErrNotExist) {
							logger.Errorf("[%s] %s not found\n", tag, input_file)
							return
						}

						file, err := os.ReadFile(input_file)
						if err != nil {
							logger.Errorf("[%s] %s\n", tag, err)
							return
						}

						if string(file) == "" {
							error_message := dialog.NewInformation("Error", fmt.Sprintf("%s is a empty file", input_file), window)
							error_message.Resize(fyne.NewSize(240, 150))
							error_message.Show()
							return
						}

						if _, _, err := dvm.ParseSmartContract(string(file)); err != nil {
							error_message := dialog.NewInformation("Error", fmt.Sprintf("%s is not a valid SC", input_file), window)
							error_message.Resize(fyne.NewSize(240, 150))
							error_message.Show()
							return
						}

						if tx := rpc.UploadNFAContract(string(file)); tx == "" {
							error_message := dialog.NewInformation("Error", "Could not install NFA, check log", window)
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

	// Save and load json config files with collection/asset data
	config_select := widget.NewSelect([]string{}, nil)
	config_select.PlaceHolder = "Load config"
	if dir, err := os.Open("creation"); err == nil {
		defer dir.Close()
		if files, err := dir.Readdirnames(0); err == nil {
			opts := []string{}
			for _, name := range files {
				if strings.HasSuffix(name, ".json") {
					opts = append(opts, name)
				}
			}
			sort.Strings(opts)
			config_select.Options = opts
		}
	}

	clear_all_button := widget.NewButtonWithIcon("", dreams.FyneIcon("contentUndo"), nil)
	clear_all_button.Importance = widget.LowImportance
	clear_all_button.OnTapped = func() {
		dialog.NewConfirm("Clear All", "Would you like to clear all current entries?", func(b bool) {
			if b {
				config_select.SetSelectedIndex(-1)
				collection_entry.SetText("")
				update_select.SetSelected("")
				name_entry.SetText("")
				descr_entry.SetText("")
				type_select.SetSelected("")
				tags_entry.SetText("")

				file_entry_start.SetText("")
				cover_entry_start.SetText("")
				icon_entry_start.SetText("")
				sign_entry_start.SetText("")

				file_entry_mid.SetText("")
				cover_entry_mid.SetText("")
				icon_entry_mid.SetText("")
				sign_entry_mid.SetText("")

				file_entry_end.SetText("")
				cover_entry_end.SetText("")
				icon_entry_end.SetText("")
				sign_entry_end.SetText("")

				royalty_entry.SetText("")
				art_entry.SetText("")
			}
		}, window).Show()
	}

	save_config_button := widget.NewButtonWithIcon("Save", dreams.FyneIcon("documentSave"), nil)
	save_config_button.Importance = widget.HighImportance
	save_config_button.OnTapped = func() {
		name := collection_entry.Text + ".json"
		file, err := os.Create(filepath.Join("creation", name))
		if err != nil {
			logger.Errorf("[%s] %s", tag, err)
			return
		}
		defer file.Close()

		data := &mintConfig{
			Collection:  collection_entry.Text,
			Update:      update_select.Selected,
			Name:        name_entry.Text,
			Description: descr_entry.Text,
			TypeHdr:     type_select.Selected,
			Tags:        tags_entry.Text,
			Royalty:     royalty_entry.Text,
			Artificer:   art_entry.Text,
			Start:       collection_low_entry.Text,
			End:         collection_high_entry.Text,
			Extension:   extension_select.Selected,
		}

		if !collection_enable.Checked {
			data.Multi = 1
		} else {
			data.Multi = len(file_entries.Objects)
		}

		data.File = append(data.File, file_entry_start.Text, file_entry_mid.Text, file_entry_end.Text)
		data.Cover = append(data.Cover, cover_entry_start.Text, cover_entry_mid.Text, cover_entry_end.Text)
		data.Icon = append(data.Icon, icon_entry_start.Text, icon_entry_mid.Text, icon_entry_end.Text)
		data.Sign = append(data.Sign, sign_entry_start.Text, sign_entry_mid.Text, sign_entry_end.Text)

		json, _ := json.MarshalIndent(data, "", " ")
		if _, err = file.Write(json); err != nil {
			logger.Errorf("[%s] %s", tag, err)
		}

		var have bool
		opts := config_select.Options
		for _, o := range opts {
			if o == name {
				have = true
				break
			}
		}

		if !have {
			opts = append(opts, name)
			sort.Strings(opts)
			config_select.Options = opts
		}
	}

	config_select.OnChanged = func(s string) {
		file, err := os.ReadFile(filepath.Join("creation", s))
		if err != nil {
			logger.Errorf("[%s] %s", tag, err)
			return
		}

		var data mintConfig
		if err = json.Unmarshal(file, &data); err != nil {
			logger.Errorf("[%s] %s", tag, err)
			return
		}

		collection_entry.SetText(data.Collection)
		update_select.SetSelected(data.Update)
		name_entry.SetText(data.Name)
		descr_entry.SetText(data.Description)
		type_select.SetSelected(data.TypeHdr)
		tags_entry.SetText(data.Tags)
		extension_select.SetSelected(data.Extension)

		art_entry.SetText("")
		royalty_entry.SetText(data.Royalty)
		art_entry.SetText(data.Artificer)

		collection_low_entry.SetText("")
		collection_high_entry.SetText(data.End)
		collection_low_entry.SetText(data.Start)

		for i := range data.File {
			switch i {
			case 0:
				file_entry_start.SetText(data.File[i])
				cover_entry_start.SetText(data.Cover[i])
				icon_entry_start.SetText(data.Icon[i])
				sign_entry_start.SetText(data.Sign[i])
			case 1:
				file_entry_mid.SetText(data.File[i])
				cover_entry_mid.SetText(data.Cover[i])
				icon_entry_mid.SetText(data.Icon[i])
				sign_entry_mid.SetText(data.Sign[i])
			case 2:
				file_entry_end.SetText(data.File[i])
				cover_entry_end.SetText(data.Cover[i])
				icon_entry_end.SetText(data.Icon[i])
				sign_entry_end.SetText(data.Sign[i])
			}
		}

		if data.Multi == 5 {
			collection_enable.SetChecked(true)
			switch len(file_entries.Objects) {
			case 3:
				url_add_incr.OnTapped()
			case 1:
				url_add_incr.OnTapped()
				url_add_incr.OnTapped()
			}
		} else if data.Multi == 3 {
			collection_enable.SetChecked(true)
			switch len(file_entries.Objects) {
			case 5:
				url_remove_incr.OnTapped()
			case 1:
				url_add_incr.OnTapped()
			}
		} else {
			collection_enable.SetChecked(false)
			switch len(file_entries.Objects) {
			case 5:
				url_remove_incr.OnTapped()
				url_remove_incr.OnTapped()
			case 3:
				url_remove_incr.OnTapped()
			}
		}
	}

	// Parse all contracts in bas folder
	scan_button := widget.NewButtonWithIcon("Scan", dreams.FyneIcon("broken-image"), nil)
	scan_button.Importance = widget.LowImportance
	scan_button.OnTapped = func() {
		if collection_entry.Validate() != nil {
			dialog.NewInformation("No collection", "Enter a collection to scan", window).Show()
			return
		}

		path := filepath.Join("creation", collection_entry.Text, "bas")

		files, err := filepath.Glob(fmt.Sprintf("%s%s*.bas", path, string(filepath.Separator)))
		if err != nil {
			error_message := dialog.NewInformation("Error", fmt.Sprintf("Could not read %s", path), window)
			error_message.Resize(fyne.NewSize(240, 150))
			error_message.Show()
			return
		}

		if len(files) < 1 {
			dialog.NewInformation("Scan", fmt.Sprintf("No contracts to scan in %s", path), window).Show()
			return
		}

		for _, f := range files {
			file, err := os.ReadFile(f)
			if err != nil {
				error_message := dialog.NewInformation("Error", fmt.Sprintf("Could not read %s", f), window)
				error_message.Resize(fyne.NewSize(240, 150))
				error_message.Show()
				return
			}

			if _, _, err := dvm.ParseSmartContract(string(file)); err != nil {
				error_message := dialog.NewInformation("Error", fmt.Sprintf("%s is not a valid SC", f), window)
				error_message.Resize(fyne.NewSize(240, 150))
				error_message.Show()
				return
			}
		}

		dialog.NewInformation("Scan Complete", fmt.Sprintf("No errors found in %s", path), window).Show()
	}

	mint_form := []*widget.FormItem{}
	mint_form = append(mint_form, widget.NewFormItem("", container.NewCenter(instructions_button)))

	mint_form = append(mint_form, widget.NewFormItem("", container.NewAdaptiveGrid(2, collection_cont,
		widget.NewForm(widget.NewFormItem("Config", container.NewBorder(nil, nil, clear_all_button, save_config_button, config_select))))))

	mint_form = append(mint_form, widget.NewFormItem("", layout.NewSpacer()))
	mint_form = append(mint_form, widget.NewFormItem("Collection", container.NewBorder(nil, nil, nil, container.NewHBox(set_up_collec, scan_button), collection_entry)))

	mint_form = append(mint_form, widget.NewFormItem("Owner Can Update", container.NewAdaptiveGrid(2, update_select,
		widget.NewForm(widget.NewFormItem("Name", container.NewBorder(nil, nil, nil, extension_select, name_cont))))))

	mint_form = append(mint_form, widget.NewFormItem("Tags *", container.NewAdaptiveGrid(2, tags_entry,
		widget.NewForm(widget.NewFormItem("Type   ", type_select)))))

	mint_form = append(mint_form, widget.NewFormItem("Description", descr_entry))

	mint_form = append(mint_form, widget.NewFormItem("File Sign C", container.NewAdaptiveGrid(2, checkC_entry,
		widget.NewForm(widget.NewFormItem("File Sign S", container.NewBorder(nil, nil, nil, import_signs, checkS_entry))))))

	mint_form = append(mint_form, widget.NewFormItem("File URL *", file_entries))
	mint_form = append(mint_form, widget.NewFormItem("Cover URL *", cover_entries))
	mint_form = append(mint_form, widget.NewFormItem("Icon URL *", icon_entries))
	mint_form = append(mint_form, widget.NewFormItem("File Sign URL *", sign_entries))
	mint_form = append(mint_form, widget.NewFormItem("", layout.NewSpacer()))

	mint_form = append(mint_form, widget.NewFormItem("Royalty %", container.NewHBox(container.NewVBox(container.NewStack(dwidget.NewSpacer(280, 0), royalty_entry)),
		widget.NewForm(widget.NewFormItem("Artificer %", container.NewStack(dwidget.NewSpacer(280, 0), art_entry))),
		layout.NewSpacer())))

	mint_form = append(mint_form, widget.NewFormItem("", layout.NewSpacer()))
	mint_form = append(mint_form, widget.NewFormItem("", container.NewAdaptiveGrid(3, container.NewStack(install_button), container.NewStack(contracts_button), sign_button)))

	scroll := container.NewVScroll(widget.NewForm(mint_form...))

	alpha120 := canvas.NewRectangle(color.RGBA{0, 0, 0, 120})
	if bundle.AppColor == color.White {
		alpha120 = canvas.NewRectangle(color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0x55})
	}

	max := container.NewStack(alpha120, scroll)

	instructions_back_button := widget.NewButton("Back", nil)
	instructions_back_button.Importance = widget.LowImportance
	instructions_back_button.OnTapped = func() {
		max.Objects[1] = container.NewStack(alpha120, scroll)
	}

	instructions_button.OnTapped = func() {
		max.Objects[1] = container.NewStack(alpha120, HowToMintNFA(instructions_back_button))
	}

	go func() {
		for {
			if collection_entry.Validate() == nil && update_select.SelectedIndex() >= 0 && name_entry.Validate() == nil && extension_select.SelectedIndex() >= 0 && descr_entry.Validate() == nil &&
				type_select.SelectedIndex() >= 0 && tags_entry.Validate() == nil && checkC_entry.Validate() == nil && checkS_entry.Validate() == nil &&
				file_entry_start.Validate() == nil && cover_entry_start.Validate() == nil && icon_entry_start.Validate() == nil &&
				sign_entry_start.Validate() == nil && royalty_entry.Validate() == nil && art_entry.Validate() == nil {
				if collection_enable.Checked {
					if collection_low_entry.Validate() == nil && collection_high_entry.Validate() == nil {
						asset_file_start := filepath.Join("creation", collection_entry.Text, "asset", fmt.Sprintf("%s%s%s", name_entry.Text, collection_low_entry.Text, extension_select.Selected))
						if _, err := os.ReadFile(asset_file_start); err == nil {
							contracts_button.Show()
						} else {
							contracts_button.Hide()
						}
					} else {
						contracts_button.Hide()
					}
				} else {
					asset_file := filepath.Join("creation", collection_entry.Text, "asset", fmt.Sprintf("%s%s", name_entry.Text, extension_select.Selected))
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
				if !rpc.Wallet.File.IsNil() {
					if collection_enable.Checked {
						if collection_low_entry.Text != "" && collection_high_entry.Text != "" {
							sign_file_start := filepath.Join("creation", collection_entry.Text, "asset", fmt.Sprintf("%s%s%s", name_entry.Text, collection_low_entry.Text, extension_select.Selected))
							if _, err := os.ReadFile(sign_file_start); err == nil {
								sign_button.Show()
							} else {
								sign_button.Hide()
							}
						} else {
							sign_button.Hide()
						}
					} else {
						sign_file := filepath.Join("creation", collection_entry.Text, "asset", fmt.Sprintf("%s%s", name_entry.Text, extension_select.Selected))
						if _, err := os.ReadFile(sign_file); err == nil {
							sign_button.Show()
						} else {
							sign_button.Hide()
						}
					}
				} else {
					sign_button.Hide()
				}
				collection_entry.Validator = validation.NewRegexp(`^\w{2,}`, "String required")
			} else {
				sign_button.Hide()
				collection_entry.Validator = validation.NewRegexp(`^\W\D\S$`, "Invalid collection directory")
			}

			if rpc.IsReady() {
				if collection_enable.Checked {
					if collection_low_entry.Text != "" && collection_high_entry.Text != "" {
						input_file_start := filepath.Join("creation", collection_entry.Text, "bas", fmt.Sprintf("%s%s.bas", name_entry.Text, collection_low_entry.Text))
						if _, err := os.ReadFile(input_file_start); err == nil {
							install_button.Show()
						} else {
							install_button.Hide()
						}
					} else {
						install_button.Hide()
					}
				} else {
					input_file := filepath.Join("creation", collection_entry.Text, "bas", fmt.Sprintf("%s.bas", name_entry.Text))
					if _, err := os.ReadFile(input_file); err == nil {
						install_button.Show()
					} else {
						install_button.Hide()
					}
				}
			} else {
				install_button.Hide()
			}

			time.Sleep(1 * time.Second)
		}
	}()

	return max
}

// Set up creation directory with sub directory for collection or single asset,
// which contains sub directories for asset, bas, icon, cover and sign files
func SetUpNFACreation(tag, collection string) (save_path string, sign_path string) {
	main_path := "creation"
	_, main := os.Stat(main_path)
	if os.IsNotExist(main) {
		logger.Printf("[%s] Setting up creation directory\n", tag)
		if err := os.Mkdir(main_path, 0755); err != nil {
			logger.Errorf("[%s] %s\n", tag, err)
			return
		}
	}

	save_path = filepath.Join(main_path, collection)
	_, coll := os.Stat(save_path)
	if os.IsNotExist(coll) {
		err := os.Mkdir(save_path, 0755)
		if err != nil {
			logger.Errorf("[%s] %s\n", tag, err)
			return
		}
	}

	asset_path := filepath.Join(save_path, "asset")
	_, asset := os.Stat(asset_path)
	if os.IsNotExist(asset) {
		logger.Printf("[%s] Creating assets directory\n", tag)
		err := os.Mkdir(asset_path, 0755)
		if err != nil {
			logger.Errorf("[%s] %s\n", tag, err)
			return
		}
	}

	bas_path := filepath.Join(save_path, "bas")
	_, bas := os.Stat(bas_path)
	if os.IsNotExist(bas) {
		logger.Printf("[%s] Creating bas Dir\n", tag)
		err := os.Mkdir(bas_path, 0755)
		if err != nil {
			logger.Errorf("[%s] %s\n", tag, err)
			return
		}
	}

	cover_path := filepath.Join(save_path, "cover")
	_, cover := os.Stat(cover_path)
	if os.IsNotExist(cover) {
		logger.Printf("[%s] Creating covers directory\n", tag)
		err := os.Mkdir(cover_path, 0755)
		if err != nil {
			logger.Errorf("[%s] %s\n", tag, err)
			return
		}
	}

	icon_path := filepath.Join(save_path, "icon")
	_, icon := os.Stat(icon_path)
	if os.IsNotExist(icon) {
		logger.Printf("[%s] Creating icons directory\n", tag)
		err := os.Mkdir(icon_path, 0755)
		if err != nil {
			logger.Errorf("[%s] %s\n", tag, err)
			return
		}
	}

	sign_path = filepath.Join(save_path, "sign")
	_, sign := os.Stat(sign_path)
	if os.IsNotExist(sign) {
		logger.Printf("[%s] Creating sign directory\n", tag)
		err := os.Mkdir(sign_path, 0755)
		if err != nil {
			logger.Errorf("[%s] %s\n", tag, err)
			return
		}
	}

	return
}

// Check that all creation directories exists
func NFACreationExists(collection string) bool {
	main_path := "creation"
	_, main := os.Stat(main_path)
	if os.IsNotExist(main) {
		return false
	}

	save_path := filepath.Join(main_path, collection)
	_, coll := os.Stat(save_path)
	if os.IsNotExist(coll) {
		return false
	}

	asset_path := filepath.Join(save_path, "asset")
	_, asset := os.Stat(asset_path)
	if os.IsNotExist(asset) {
		return false
	}

	bas_path := filepath.Join(save_path, "bas")
	_, bas := os.Stat(bas_path)
	if os.IsNotExist(bas) {
		return false
	}

	cover_path := filepath.Join(save_path, "cover")
	_, cover := os.Stat(cover_path)
	if os.IsNotExist(cover) {
		return false
	}

	icon_path := filepath.Join(save_path, "icon")
	_, icon := os.Stat(icon_path)
	if os.IsNotExist(icon) {
		return false
	}

	sign_path := filepath.Join(save_path, "sign")
	_, sign := os.Stat(sign_path)

	return !os.IsNotExist(sign)
}

// Read a Dero .sign file and return signer, C and S signatures
func ReadDeroSignFile(tag, sign_path string) (checkC string, checkS string, signer string) {
	if sign_data, err := os.ReadFile(sign_path); err != nil {
		logger.Errorf("[%s] Cannot read input file %s\n", tag, err)
	} else {
		if string(sign_data[0:35]) == "-----BEGIN DERO SIGNED MESSAGE-----" {
			split := strings.Split(string(sign_data[0:247]), "\n")
			signer = strings.TrimSpace(split[1][9:])
			checkC = strings.TrimSpace(split[2][3:])
			checkS = strings.TrimSpace(split[3][3:])
		} else {
			logger.Errorf("[%s] Not a valid Dero .sign file\n", tag)
		}
	}
	return
}

// Create fyne.ListableURI for current directory
func createURI() (uri fyne.ListableURI, err error) {
	var dir string
	dir, err = os.Getwd()
	if err != nil {
		logger.Println("[createURI] Failed to get current directory:", err)
		return
	}

	return storage.ListerForURI(storage.NewFileURI(dir))
}
