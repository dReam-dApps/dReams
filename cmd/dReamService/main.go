package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/SixofClubsss/dReams/menu"
	"github.com/SixofClubsss/dReams/prediction"
	"github.com/SixofClubsss/dReams/rpc"
	"github.com/docopt/docopt-go"
)

// Run dReamsService process from dReams prediction package

var enable_transfers bool
var command_line string = `dReamService
App to run dReamService as a single process from dReams package, powered by Gnomon.

Usage:
  dReamService [options]
  dReamService -h | --help

Options:
  -h --help                      Show this screen.
  --daemon=<127.0.0.1:10102>     Set daemon rpc address to connect.
  --wallet=<127.0.0.1:10103>     Set wallet rpc address to connect.
  --login=<user:pass>     	 Wallet rpc user:pass for auth.
  --transfers=<false>        True/false value for enabling processing transfers to integrated address.
  --debug=<true>     		 True/false value for enabling terminal debug.
  --fastsync=<true>	         Gnomon option,  true/false value to define loading at chain height on start up.
  --num-parallel-blocks=<5>      Gnomon option,  defines the number of parallel blocks to index.`

// Set opts when starting dReamService
func flags() (version string) {
	version = rpc.DREAMSv
	arguments, err := docopt.ParseArgs(command_line, nil, version)

	if err != nil {
		log.Fatalf("Error while parsing arguments: %s\n", err)
	}

	fastsync := true
	if arguments["--fastsync"] != nil {
		if arguments["--fastsync"].(string) == "false" {
			fastsync = false
		}
	}

	parallel := 1
	if arguments["--num-parallel-blocks"] != nil {
		s := arguments["--num-parallel-blocks"].(string)
		switch s {
		case "2":
			parallel = 2
		case "3":
			parallel = 3
		case "4":
			parallel = 4
		case "5":
			parallel = 5
		default:
			parallel = 1
		}
	}

	// Set default rpc params
	rpc.Daemon.Rpc = "127.0.0.1:10102"
	rpc.Wallet.Rpc = "127.0.0.1:10103"

	if arguments["--daemon"] != nil {
		if arguments["--daemon"].(string) != "" {
			rpc.Daemon.Rpc = arguments["--daemon"].(string)
		}
	}

	if arguments["--wallet"] != nil {
		if arguments["--wallet"].(string) != "" {
			rpc.Wallet.Rpc = arguments["--wallet"].(string)
		}
	}

	if arguments["--login"] != nil {
		if arguments["--login"].(string) != "" {
			rpc.Wallet.UserPass = arguments["--login"].(string)
		}
	}

	// Default false, integrated addresses generated through dReams
	transfers := false
	if arguments["--transfers"] != nil {
		if arguments["--transfers"].(string) == "true" {
			transfers = true
		}
	}

	debug := true
	if arguments["--debug"] != nil {
		if arguments["--debug"].(string) == "false" {
			debug = false
		}
	}

	rpc.Wallet.Service = true
	menu.Gnomes.Trim = true
	enable_transfers = transfers
	prediction.Service.Debug = debug
	menu.Gnomes.Fast = fastsync
	menu.Gnomes.Para = parallel
	menu.Gnomes.Import = true

	return
}

func init() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println()
		menu.Gnomes.Stop("dReamService")
		rpc.Wallet.Connected(false)
		rpc.Wallet.Service = false
		for prediction.Service.Processing {
			log.Println("[dReamService] Waiting for service to close")
			time.Sleep(3 * time.Second)
		}
		log.Println("[dReamService] Closing")
		os.Exit(0)
	}()
}

func main() {
	n := runtime.NumCPU()
	runtime.GOMAXPROCS(n)

	v := flags()
	log.Println("[dReamService]", v, runtime.GOOS, runtime.GOARCH)

	// Check for daemon connection
	rpc.Ping()
	if !rpc.Daemon.Connect {
		log.Fatalf("[dReamService] Daemon %s not connected\n", rpc.Daemon.Rpc)
	}

	// Check for wallet connection
	rpc.GetAddress("dReamService")
	if !rpc.Wallet.Connect {
		os.Exit(1)
	}

	// Start dReamService from last payload format height at minimum
	height := prediction.PAYLOAD_FORMAT

	// Set up Gnomon search filters
	filter := []string{}
	predict := rpc.GetPredictCode(0)
	if predict != "" {
		filter = append(filter, predict)
	}

	sports := rpc.GetSportsCode(0)
	if sports != "" {
		filter = append(filter, sports)
	}

	// Set up SCID rating map
	menu.Control.Contract_rating = make(map[string]uint64)

	// Start Gnomon with search filters
	go menu.StartGnomon("dReamService", "gravdb", filter, 0, 0, nil)

	// Routine for checking daemon, wallet connection and Gnomon sync
	go func() {
		for !menu.Gnomes.IsInitialized() {
			time.Sleep(time.Second)
		}

		log.Println("[dReamService] Starting when Gnomon is synced")
		height = uint64(menu.Gnomes.Indexer.ChainHeight)
		for menu.Gnomes.IsRunning() && rpc.IsReady() {
			rpc.Ping()
			rpc.EchoWallet("dReamService")
			menu.Gnomes.IndexContains()
			if menu.Gnomes.Indexer.LastIndexedHeight >= menu.Gnomes.Indexer.ChainHeight-3 && menu.Gnomes.HasIndex(9) {
				menu.Gnomes.Synced(true)
			} else {
				menu.Gnomes.Synced(false)
				if !menu.Gnomes.Start && menu.Gnomes.IsInitialized() {
					diff := menu.Gnomes.Indexer.ChainHeight - menu.Gnomes.Indexer.LastIndexedHeight
					if diff > 3 && prediction.Service.Debug {
						log.Printf("[dReamService] Gnomon has %d blocks to go\n", diff)
					}
				}
			}
			time.Sleep(3 * time.Second)
		}
	}()

	// Wait for Gnomon to sync
	for !menu.Gnomes.IsSynced() && !menu.Gnomes.HasIndex(100) {
		time.Sleep(time.Second)
	}

	time.Sleep(time.Second)

	// Populate SCID of connected wallet
	menu.PopulatePredictions(nil)
	menu.PopulateSports(nil)

	// Set added print text
	add := ""
	if enable_transfers {
		add = "and transactions"
	}

	// Start dReamService
	log.Printf("[dReamService] Processing payouts %s\n", add)
	prediction.DreamService(height, true, enable_transfers)
}
