# dReams
Interact with a variety of different products and services on [Dero's](https://dero.io) decentralized application platform. 

![dReamsFooter](https://raw.githubusercontent.com/SixofClubsss/dreamdappsite/main/assets/dReamerUp.png)

1. [Project](#project) 
2. [dApps](#dapps) 
3. [Features](#features) 
4. [Build](#build) 
5. [Packages](#packages) 
	- [rpc](#rpc)
	- [gnomes](#gnomes)
	- [menu](#menu)
	- [dwidget](#dwidget)
	- [bundle](#bundle)
6. [Donations](#donations) 
7. [Licensing](#licensing) 

### Project
dReams is a open source platform application that houses multiple dApps and utilities built on Dero. dReams has two facets to its use. 

![goMod](https://img.shields.io/github/go-mod/go-version/dReam-dApps/dReams.svg)![goReport](https://goreportcard.com/badge/github.com/dReam-dApps/dReams)[![goDoc](https://img.shields.io/badge/godoc-reference-blue.svg)](https://pkg.go.dev/github.com/dReam-dApps/dReams)

As a application
> With a wide array of features from games to blockchain services, dReams app is a *desktop* point of entry into the privacy preserving world of Dero.

As a repository
> dReams serves as a source for building Dero applications. Written in [Go](https://go.dev/) and using the [Fyne toolkit](https://fyne.io/), the dReams repository is constructed into packages with function imports for many different Dero necessities. 

Download the latest [release](https://github.com/dReam-dApps/dReams/releases) or [build from source](#build) to use dReams.

![windowsOS](https://raw.githubusercontent.com/SixofClubsss/dreamdappsite/main/assets/os-windows-green.svg)![macOS](https://raw.githubusercontent.com/SixofClubsss/dreamdappsite/main/assets/os-macOS-green.svg)![linuxOS](https://raw.githubusercontent.com/SixofClubsss/dreamdappsite/main/assets/os-linux-green.svg)

dReams [Template](https://github.com/dReam-dApps/Template) can be used to help create new Dero dApps.

### dApps
All dApps are ran on chain in a decentralized manner. dReams and packages are solely interfaces to interact with these on chain services. With the main dReams application, users can access the dApps below from one place.
- **[Grokked](https://github.com/SixofClubsss/Grokked)**
	- Proof of attention game
	- A player is randomly chosen as the Grok
	- If they do not pass the Grok in the time frame they are removed from the game   
	- The time frame gets shorter every turn, last player standing wins
	- Deployable contracts
	- Leader boards
- **[Holdero](https://github.com/SixofClubsss/Holdero)**
	- Multiplayer Texas Hold'em style poker
	- In game assets 
	- Deployable contracts
	- Multiple tokens supported
	- dReam Tools
- **[Baccarat](https://github.com/SixofClubsss/Baccarat)**
	- Single player table game
	- In game assets
	- Multiple tokens supported
- **[dSports and dPrediction](https://github.com/SixofClubsss/dPrediction)**
	- P2P betting and predictions
	- Deployable contracts
	- dService 
- **[Iluma](https://github.com/SixofClubsss/Iluma)**
	- Tarot readings
	- Custom cards and artwork by Kalina Lux
	- Querent's companion
- **[Duels](https://github.com/SixofClubsss/Duels)**
	- Duel Dero assets in a over or under showdown style game
	- Three game modes, regular, death match and hardcore 
	- Asset graveyard and leader board
- **[NFA Marketplace](https://github.com/civilware/artificer-nfa-standard)**
	- View and manage owned assets
	- View and manage listings
	- Search NFAs
	- Mint NFAs 
- **More dApps to come...**
### Features
- [Gnomon](https://github.com/civilware/gnomon) with UI controls
- Create customs Gnomon indexes
- Gnomon header controls
- Send Dero messages
- Send Dero assets
- Deployable contract rating system
- Dynamic app updates from on chain data
- Import only the dApps and collections you want to use
- Shared config files for platform wide use

### Build
- Install latest [Go version](https://go.dev/doc/install)
- Install [Fyne](https://developer.fyne.io/started/) dependencies.
- Clone repo and build with:

```
git clone https://github.com/dReam-dApps/dReams.git
cd dReams
cd cmd/dReams
go build .
./dReams
```
## Packages
dReams repo is built as packages. With imports from the Dero code base, dReams variable structures are complete with the basics needs for building Dero applications that can run alone, or ones that could be integrated into dReams. 

dReams [Template](https://github.com/dReam-dApps/Template) can be used as a UI starting point and you can view our [Examples](https://github.com/dReam-dApps/Examples) repo for further references. 
### rpc
The rpc package contains all of the basic functionality needed to set up clients, check connectivity and read blockchain and wallet information. There are arbitrary rpc calls which any dApp can make use of such as the NFA calls, `SendMessage()` or `SendAsset()` with optional payload. This example checks for daemon and wallet rpc connectivity.
```
package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dReam-dApps/dReams/rpc"
)

// dReams rpc connection example

// Name my app
const app_tag = "My_app"

func main() {
	// Initialize rpc addresses to rpc.Daemon and rpc.Wallet vars
	rpc.Daemon.Rpc = "127.0.0.1:10102"
	rpc.Wallet.Rpc = "127.0.0.1:10103"
	// Initialize rpc.Wallet.UserPass for rpc user:pass

	// Check for daemon connection
	rpc.Ping()

	// Check for wallet connection and get address
	rpc.GetAddress(app_tag)

	// Exit with ctrl-C
	var exit bool
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		log.Printf("[%s] Closing\n", app_tag)
		exit = true
	}()

	// Loop will check for daemon and wallet connection and
	// print wallet height and balance. It will keep
	// running while daemon and wallet are connected or until exit
	for !exit && rpc.IsReady() {
		rpc.Wallet.GetBalance()
		rpc.GetWalletHeight(app_tag)
		log.Printf("[%s] Height: %d   Dero Balance: %s\n", app_tag, rpc.Wallet.Height, rpc.FromAtomic(rpc.Wallet.Balance, 5))
		time.Sleep(3 * time.Second)
		rpc.Ping()
		rpc.EchoWallet(app_tag)
	}

	log.Printf("[%s] Not connected\n", app_tag)
}
```
### gnomes
The gnomes package contains the base components used for Gnomon indexing. `StartGnomon()` allows apps to run a instance of Gnomon with search filter and pass optional func for any custom index requirements.  
```
package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/civilware/Gnomon/structures"
	"github.com/dReam-dApps/dReams/gnomes"
	"github.com/dReam-dApps/dReams/rpc"
	"github.com/sirupsen/logrus"
)

// dReams gnomes StartGnomon() example

// Name my app
const app_tag = "My_app"

// Log output
var logger = structures.Logger.WithFields(logrus.Fields{})

// Gnomon instance from gnomes package
var gnomon = gnomes.NewGnomes()

func main() {
	// Initialize Gnomon fast sync true to sync db immediately
	gnomon.SetFastsync(true, false, 100)

	// Initialize rpc address to rpc.Daemon var
	rpc.Daemon.Rpc = "127.0.0.1:10102"

	// Initialize logger to Stdout
	gnomes.InitLogrusLog(logrus.InfoLevel)

	rpc.Ping()
	// Check for daemon connection, if daemon is not connected we won't start Gnomon
	if rpc.Daemon.IsConnected() {
		// Initialize NFA search filter and start Gnomon
		filter := []string{gnomes.NFA_SEARCH_FILTER}
		gnomes.StartGnomon(app_tag, "boltdb", filter, 0, 0, nil)

		// Exit with ctrl-C
		var exit bool
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-c
			exit = true
		}()

		// Gnomon will continue to run if daemon is connected
		for !exit && rpc.Daemon.IsConnected() {
			contracts := gnomon.GetAllOwnersAndSCIDs()
			logger.Printf("[%s] Index contains %d contracts\n", app_tag, len(contracts))
			time.Sleep(3 * time.Second)
			rpc.Ping()
		}

		// Stop Gnomon
		gnomon.Stop(app_tag)
	}

	logger.Printf("[%s] Done\n", app_tag)
}
```
### menu 
NFA related items such as the dReams NFA marketplace and asset controls can be independently imported for use in other dApps, it can be used with or without dReams filters. There are menu panels and custom Dero indicators that can be imported. This example shows how to import them as app tabs.
```
package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	dreams "github.com/dReam-dApps/dReams"
	"github.com/dReam-dApps/dReams/bundle"
	"github.com/dReam-dApps/dReams/menu"
)

// dReams menu PlaceMarket and PlaceAsset example

// Name my app
const app_tag = "My_app"

func main() {
	// Intialize Fyne window app and window into dReams app object
	a := app.New()
	w := a.NewWindow(app_tag)
	w.Resize(fyne.NewSize(900, 700))
	d := dreams.AppObject{
		App:    a,
		Window: w,
	}

	// Simple asset profile with wallet name entry and theme select
	line := canvas.NewLine(bundle.TextColor)
	profile := []*widget.FormItem{}
	profile = append(profile, widget.NewFormItem("Name", menu.NameEntry()))
	profile = append(profile, widget.NewFormItem("", layout.NewSpacer()))
	profile = append(profile, widget.NewFormItem("", container.NewVBox(line)))
	profile = append(profile, widget.NewFormItem("Theme", menu.ThemeSelect(&d)))
	profile = append(profile, widget.NewFormItem("", container.NewVBox(line)))

	// Rescan button function in asset tab
	rescan := func() {
		// What you want to scan wallet for
	}

	// Place asset and market layouts into tabs
	tabs := container.NewAppTabs(
		container.NewTabItem("Assets", menu.PlaceAssets(app_tag, widget.NewForm(profile...), rescan, bundle.ResourceDReamsIconAltPng, &d)),
		container.NewTabItem("Market", menu.PlaceMarket(&d)))

	// Place tabs as window content and run app
	d.Window.SetContent(tabs)
	d.Window.ShowAndRun()
}
```
### dwidget
The dwidget package is a extension to fyne widgets that intends to make creating dApps simpler and quicker with widgets specified for use with Dero. Numerical entries have prefix, increment and decimal control and pre-configured connection boxes can be used that are tied into dReams rpc vars and have default Dero connection addresses populated. There is objects for shutdown control as well as a spot for the dReams indicators, or new ones. This example starts a Fyne gui app using `VerticalEntries()` to start Gnomon when connected.
```
package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"

	"github.com/dReam-dApps/dReams/dwidget"
	"github.com/dReam-dApps/dReams/gnomes"
	"github.com/dReam-dApps/dReams/rpc"
	"github.com/sirupsen/logrus"
)

// dReams dwidget NewVerticalEntries() example

// Name my app
const app_tag = "My_app"

// Gnomon instance from gnomes package
var gnomon = gnomes.NewGnomes()

func main() {
	// Initialize Gnomon fast sync true to sync db immediately
	gnomon.SetFastsync(true, false, 100)

	// Initialize logger to Stdout
	gnomes.InitLogrusLog(logrus.InfoLevel)

	// Initialize fyne app
	a := app.New()

	// Initialize fyne window with size
	w := a.NewWindow(app_tag)
	w.Resize(fyne.NewSize(300, 100))
	w.SetMaster()

	// When window closes, stop Gnomon if running
	w.SetCloseIntercept(func() {
		if gnomon.IsInitialized() {
			gnomon.Stop(app_tag)
		}
		w.Close()
	})

	// Initialize dwidget connection box
	connect_box := dwidget.NewVerticalEntries(app_tag, 1)

	// When connection button is pressed we will connect to wallet rpc,
	// and start Gnomon with NFA search filter if it is not running
	connect_box.Button.OnTapped = func() {
		rpc.GetAddress(app_tag)
		rpc.Ping()
		if rpc.Daemon.Connect && !gnomon.IsInitialized() && !gnomon.IsStarting() {
			go gnomes.StartGnomon(app_tag, "boltdb", []string{gnomes.NFA_SEARCH_FILTER}, 0, 0, nil)
		}
	}

	// Place connection box and start app
	w.SetContent(connect_box.Container)
	w.ShowAndRun()
}
```
### bundle
The bundle package contains all dReams resources. Images, gifs and fonts can be imported as well as the two Dero styled base app themes for Fyne. This example starts a Fyne gui app with various widgets to show case both Dero themes and image imports from bundle.
```
package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/dReam-dApps/dReams/bundle"
	"github.com/dReam-dApps/dReams/dwidget"
)

// Name my app
const app_tag = "My_app"

func main() {
	// Initialize app color to bundle var
	bundle.AppColor = color.Black

	// Initialize fyne app with Dero theme
	a := app.New()
	a.Settings().SetTheme(bundle.DeroTheme(bundle.AppColor))

	// Initialize fyne window with size and icon from bundle package
	w := a.NewWindow(app_tag)
	w.SetIcon(bundle.ResourceBlueBadge3Png)
	w.Resize(fyne.NewSize(300, 100))
	w.SetMaster()

	// Initialize fyne container and add some various widgets for viewing purposes
	cont := container.NewVBox()
	cont.Add(container.NewAdaptiveGrid(3, dwidget.NewCenterLabel("Label"), widget.NewEntry(), widget.NewButton("Button", nil)))
	cont.Add(container.NewAdaptiveGrid(3, widget.NewLabel("Label"), widget.NewCheck("Check", nil), dwidget.NewLine(30, 30, bundle.TextColor)))
	cont.Add(widget.NewPasswordEntry())
	cont.Add(widget.NewSlider(0, 100))

	// Widget to change theme
	change_theme := widget.NewRadioGroup([]string{"Dark", "Light"}, func(s string) {
		switch s {
		case "Dark":
			bundle.AppColor = color.Black
		case "Light":
			bundle.AppColor = color.White
		default:

		}

		a.Settings().SetTheme(bundle.DeroTheme(bundle.AppColor))
	})
	change_theme.Horizontal = true
	cont.Add(container.NewCenter(change_theme))

	// Add a image from bundle package
	gnomon_img := canvas.NewImageFromResource(bundle.ResourceGnomonIconPng)
	gnomon_img.SetMinSize(fyne.NewSize(45, 45))
	cont.Add(container.NewCenter(gnomon_img))

	// Adding last widget
	select_entry := widget.NewSelect([]string{"Choice 1", "Choice 2", "Choice 3"}, nil)
	cont.Add(select_entry)

	// Place widget container and start app
	w.SetContent(cont)
	w.ShowAndRun()
}
```

## Donations
- **Dero Address**: dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn

![DeroDonations](https://raw.githubusercontent.com/SixofClubsss/dreamdappsite/main/assets/DeroDonations.jpg)

---

### Licensing

dReams platform and packages are free and open source.    
The source code is published under the [MIT](https://github.com/dReam-dApps/dReams/blob/main/LICENSE) License.   
Copyright © 2023 dReam dApps   
