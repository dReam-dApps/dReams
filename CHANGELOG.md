# Changelog

This file lists the changes to dReams repo with each version.

## 0.12.0 - In Progress

### Added

* Implement platform wide encrypted account stores for local data
* XSWD and DERO connections implemented into existing APIs and `dwidgets`
* dDice 0.x.x
* New asset type for creators, dice.
* Implement custom token support for balances
* TX confirmation indicator
* dSkullz Collection
* ParseSmartContract() for directory and when minting NFA
* Auction highest bidder display
* Show SCID in market
* View image assets in market
* `dreams` NewFyneApp() to easily create new Fyne/dReams apps
* `dreams` SetBalanceLabelText() for standardizing dApp labels
* `dreams` DownloadFile() and UnzipFile()
* `dreams` GetMaxSize() and GetImageSizeFromMemory()
* `rpc` GetNameToAddress() and sending messages/assets to name
* `rpc` HashToHexSHA256()
* `rpc` GetDaemonInfo()
* `rpc` GasEstimateInstall()
* `rpc` IsAddress(address) checks equal
* `menu` utility var to assetObjects, IsDreamsNFACreator() also returns utility
* `gnomes` GetAllSCIDInvokeDetailsByEntrypoint()
* `gnomes` GetAssetInfo()
* `gnomes` GetLiveSCVariables() and StoreLiveSCIDVariableDetails()
* `dwidget` Float64() and Uint64() methods for AmountEntry
* `dwidget` dstack file with UpdateText() and SetUpdate()
* `bundle` astrolyte font 

### Changed

* Fyne 2.5.0
* Gnomon 2.0.3-alpha.x
* Baccarat 0.x.x
* Holdero 0.x.x
* dPrediction 0.x.x
* Iluma 0.x.x
* Duels 0.x.x
* Grokked 0.x.x
* dReams standard import function for dApps is now LayoutAll()
* Removed DERO file buttons from NFA minter, is now integrated into its connection widget
* NFA-Creation directory renamed to creation
* All local storage locations contained within datashards directory
* Split type and utility display
* `rpc` removed unnecessary exported vars from wallet and created methods for File.disk
* `rpc` balance map to map[string]*Balance
* `rpc` SetDaemonClient() allows https endpoints
* `rpc` Rename Daemon.Rpc to Daemon.Endpoint and maintain similar package naming
* `rpc` Rename DaemonHeight => GetDaemonHeight
* `rpc` Ping(), GetHeight(), GetVersion(), GetInfo(), GetTx(), GetTxPool() to daemon methods
* `rpc` Remove CheckForIndex()
* `rpc` Use Wallet.Address() method instead of var
* `menu` StartDreamsIndicators() removed in favor of StartIndicators for all apps
* `menu` Theme moved to `dreams` package
* `menu` Rename DefaultThemeResource() to DefaultBackgroundResource()
* `menu` ReadDreamsConfig() and WriteDreamsConfig() deprecated, use StoreSettings(), GetSettings()

### Fixed

* Padding on scroll to buttons covered by scroll bar
* Hide claim button when disconnected
* Store local theme image file after downloading
* Save collection info when storing minting config
* `rpc` catch int cases when using convert funcs


## 0.11.1 - January 19 2024

### Added

* Xeggex price feed
* `gnomes` Version var to SC
* `rpc` GetUintKey()

### Changed

* Go 1.21.5
* Fyne 2.4.3
* Gnomon 2.0.3-alpha.5
* Baccarat 0.3.1
* Holdero 0.3.1
* dPrediction 0.3.1
* Iluma 0.3.1
* Duels 0.1.1
* Grokked 0.1.1
* Clean up client var names in `rpc`

### Fixed

* High memory use (#13)
* False TX fail prints (#11)
* Template typo


## 0.11.0 - December 23 2023

### Added

* CHANGELOG
* Pull request and issue templates
* Check for icon DB storage before downloading
* HS gold cards
* Grokked dApp
* Asset profile
* Sync screens for asset and market
* `semver` versioning 
* `gnomes` Gnomon upstream updates including forced fast sync and status 
* `gnomes` bolt storage funcs
* `gnomes` SCHeaders and SC structs
* `gnomes` tool tip
* `menu` NFAListing struct
* `menu` DefaultThemeResource, AssetIcon, ParseURL, SwitchProfileIcon, ShowTxDialog and ShowConfirmDialog funcs
* `menu` ClaimAll NFA funcs
* `dreams` DownloadBytes func
* `dwidget` NewSpacer, NewLine and AddIndicator funcs
* `rpc` PrintError and PrintLog funcs print to UI
* `rpc` IsConfirmingTx funcs

### Changed

* Fyne 2.4.1
* Icon resources 
* Duel assets enabled
* Move Theme var and funcs to `menu` package
* Removed terminal app
* Removed unneeded start flags
* Removed derbnb 
* Removed system tray and moved funcs to menu/wallet layout
* Entries OnCursorChanged to OnChanged
* dApp tab layout updated with versions
* Gnomes removed from `menu` and is now a package
* Connect objects and info layout updated
* Confirmations to dialogs  
* Asset tab layout updated and broken down into sub tabs (owned, profile, index and headers)
* Market layout updated
* NFA minter layout updated
* Import balances and swap from `holdero`
* `gnomes` Gnomes var to Gnomes interface with added methods
* `gnomes` control panel UI updated
* `menu` funcs split into smaller files
* `menu` refactor assetObjects, marketObjects, menuObjects structs
* `dreams` rename DownloadFile to DownloadCanvas
* `rpc` SessionLog versioning for dApps
* `rpc` rename FetchFees to GetFees
* `rpc` rename FetchDapps GetDapps
* `rpc` increase cancel time to eight sec

### Fixed

* Deprecated container.NewMax
* Deprecated fyne.TextTruncate
* Can't close when initializing
* Channel short cycling
* Market high resource use
* Validator hangs