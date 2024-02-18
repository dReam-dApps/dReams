# Changelog

This file lists the changes to dReams repo with each version.

## 0.11.x - In Progress

### Added
* dSkullz Collection
* ParseSmartContract() for directory and when minting NFA
* Auction highest bidder display
* `rpc` GetNameToAddress() and send to name
* `dreams` GetMaxSize()


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