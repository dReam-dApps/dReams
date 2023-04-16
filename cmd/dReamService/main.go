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
	tranfsers := false
	if arguments["--transfers"] != nil {
		if arguments["--transfers"].(string) == "true" {
			tranfsers = true
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
	enable_transfers = tranfsers
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
		menu.StopGnomon("dReamService")
		rpc.Wallet.Connect = false
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
	go menu.StartGnomon("dReamService", filter, 0, 0, nil)

	// Routine for checking daemon, wallet connection and Gnomon sync
	go func() {
		for !menu.Gnomes.Init {
			time.Sleep(time.Second)
		}

		log.Println("[dReamService] Starting when Gnomon is synced")
		height = rpc.DaemonHeight(rpc.Daemon.Rpc)
		for menu.Gnomes.Init && !menu.GnomonClosing() && rpc.Wallet.Connect && rpc.Daemon.Connect {
			rpc.Ping()
			rpc.EchoWallet("dReamService")
			contracts := menu.Gnomes.Indexer.Backend.GetAllOwnersAndSCIDs()
			menu.Gnomes.SCIDS = uint64(len(contracts))
			if menu.Gnomes.Indexer.ChainHeight >= int64(height)-3 && menu.Gnomes.SCIDS >= 9 {
				menu.Gnomes.Sync = true
			}
			time.Sleep(3 * time.Second)
		}
	}()

	// Wait for Gnomon to sync
	for !menu.Gnomes.Sync && menu.Gnomes.SCIDS < 100 {
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
