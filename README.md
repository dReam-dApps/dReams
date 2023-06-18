# dReams
Interact with a variety of different products and services on [Dero's](https://dero.io) decentralized application platform. 

![dReamTablesFooter](https://user-images.githubusercontent.com/84689659/170848755-d2cb4933-df2b-46f9-80e6-4349621871a3.png)

1. [Project](#project) 
2. [dApps](#dapps) 
3. [Features](#features) 
4. [Build](#build) 
5. [Packages](#packages) 
	- [rpc](#rpc)
	- [menu](#menu)
	- [dwidget](#dwidget)
	- [bundle](#bundle)
6. [Donations](#donations) 
7. [Licensing](#licensing) 

### Project
dReams is a open source platform application that houses multiple *desktop* dApps and utilities built on Dero. dReams has two facets to its use. 

As a application
>With a wide array of features from games to blockchain services, dReams is a point of entry into the world of Dero.

As a repository
> dReams serves as a source for building Dero applications. Written in [Go](https://go.dev/) and using the [Fyne toolkit](https://fyne.io/), the dReams repository is constructed into packages with function imports for many different Dero necessities. 

Download the latest [release](https://github.com/SixofClubsss/dReams/releases) to use dReams or build from source.
### dApps
All dApps are ran on chain in a decentralized manner. dReams and packages are interfaces to interact with these on chain services. With the main dReams application, users can access the dApps below from one place. 
- **Holdero**
	- Multiplayer Texas Hold'em style poker
	- In game assets 
	- Deployable contracts
	- dReam Tools
- **Baccarat**
	- Single player table game
	- In game assets
- **dSports and dPrediction**
	- P2P betting and predictions
	- Deployable contracts
	- dReam Service 
- **Iluma**
	- Tarot readings
	- Custom cards and artwork
	- Querent's companion
- **DerBnb**
	- Property rental management app
	- Mint property tokens
	- Manage rentals and bookings with Dero private messaging
- **NFA Marketplace**
	- View and manage owned assets
	- View and manage listings
	- Mint NFAs 
### Features
- [Gnomon](https://github.com/civilware/gnomon) Index with user controls
- Gnomon header controls
- Send Dero messages
- Send Dero assets
- Deployable contract rating system
- Dynamic app updates from on chain data
- Shared config files for platform wide use

### Build
- Install latest [Go version](https://go.dev/doc/install)
- Install [Fyne](https://developer.fyne.io/started/) dependencies.
- Clone repo and build with:

```
git clone https://github.com/SixofClubsss/dReams.git
cd dReams
cd cmd/dReams
go build .
./dReams
```
## Packages
dReams repo is built as packages. With imports from the Dero code base, dReams variable structures are complete with the basics needs for building Dero applications that can run alone, or ones that could be integrated into dReams.
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

	"github.com/SixofClubsss/dReams/rpc"
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
### menu
The menu package contains the base components used for Gnomon indexing. `StartGnomon()` allows apps to run a instance of Gnomon with search filter and pass optional func for any custom index requirements. NFA related items such as the dReams NFA marketplace and asset controls can be independently imported for use in other dApps, it can be used with or without dReams filters. There are menu panels and custom Dero indicators that can be imported. This example starts Gnomon with NFA search filter. 
```
package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/SixofClubsss/dReams/menu"
	"github.com/SixofClubsss/dReams/rpc"
)

// dReams menu StartGnomon() example

// Name my app
const app_tag = "My_app"

func main() {
	// Initialize Gnomon fast sync
	menu.Gnomes.Fast = true

	// Initialize rpc address to rpc.Daemon var
	rpc.Daemon.Rpc = "127.0.0.1:10102"

	rpc.Ping()
	// Check for daemon connection, if daemon is not connected we won't start Gnomon
	if rpc.Daemon.Connect {
		// Initialize NFA search filter and start Gnomon
		filter := []string{menu.NFA_SEARCH_FILTER}
		menu.StartGnomon(app_tag, "boltdb", filter, 0, 0, nil)

		// Exit with ctrl-C
		var exit bool
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		go func() {
			<-c
			exit = true
		}()

		// Gnomon will continue to run if daemon is connected
		for !exit && rpc.Daemon.Connect {
			contracts := menu.Gnomes.GetAllOwnersAndSCIDs()
			log.Printf("[%s] Index contains %d contracts\n", app_tag, len(contracts))
			time.Sleep(3 * time.Second)
			rpc.Ping()
		}

		// Stop Gnomon
		menu.Gnomes.Stop(app_tag)
	}

	log.Printf("[%s] Done\n", app_tag)
}
```
### dwidget
The dwidget package is a extension to fyne widgets that intends to make creating dApps simpler and quicker with widgets specified for use with Dero. Numerical entries have prefix, increment and decimal control and pre-configured connection boxes can be used that are tied into dReams rpc vars and have default Dero connection addresses populated. There is objects for shutdown control as well as a spot for the dReams indicators, or new ones. This example starts a Fyne gui app using `VerticalEntries()` to start Gnomon when connected.
```
package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"

	"github.com/SixofClubsss/dReams/dwidget"
	"github.com/SixofClubsss/dReams/menu"
	"github.com/SixofClubsss/dReams/rpc"
)

// dReams dwidget VerticalEntries() example

// Name my app
const app_tag = "My_app"

func main() {
	// Initialize Gnomon fast sync
	menu.Gnomes.Fast = true

	// Initialize fyne app
	a := app.New()

	// Initialize fyne window with size
	w := a.NewWindow(app_tag)
	w.Resize(fyne.NewSize(300, 100))
	w.SetMaster()

	// When window closes, stop Gnomon if running
	w.SetCloseIntercept(func() {
		if menu.Gnomes.Init {
			menu.Gnomes.Stop(app_tag)
		}
		w.Close()
	})

	// Initialize dwidget connection box
	connect_box := dwidget.VerticalEntries(app_tag, 1)

	// When connection button is pressed we will connect to wallet rpc,
	// and start Gnomon with NFA search filter if it is not running
	connect_box.Button.OnTapped = func() {
		rpc.GetAddress(app_tag)
		rpc.Ping()
		if rpc.Daemon.Connect && !menu.Gnomes.IsInitialized() && !menu.Gnomes.Start {
			go menu.StartGnomon(app_tag, "boltdb", []string{menu.NFA_SEARCH_FILTER}, 0, 0, nil)
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
	"github.com/SixofClubsss/dReams/bundle"
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
	cont.Add(container.NewAdaptiveGrid(3, widget.NewLabel("Label"), widget.NewEntry(), widget.NewButton("Button", nil)))
	cont.Add(container.NewAdaptiveGrid(3, widget.NewLabel("Label"), widget.NewCheck("Check", nil), widget.NewButton("Button", nil)))
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

![DeroDonations](https://user-images.githubusercontent.com/84689659/165414903-44164e7e-4277-44f8-b1fe-8d139f559db1.jpg)

---

### Licensing

dReams platform and packages are free and open source. 
The source code is published under the [MIT](https://github.com/SixofClubsss/dReams/blob/main/LICENSE) License. 
Copyright © 2023 dReam dApps 
