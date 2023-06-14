package derbnb

import (
	"encoding/hex"
	"encoding/json"
	"log"
	"strconv"

	"github.com/SixofClubsss/dReams/rpc"
	"github.com/deroproject/derohe/cryptography/crypto"
	dero "github.com/deroproject/derohe/rpc"
)

type property_data struct {
	Photos        []string `json:"photos"`
	Squarefootage int      `json:"squarefootage"`
	// Driveway              bool     `json:"driveway"`
	Style       string `json:"style"`
	Description string `json:"description"`
	// DistanceToCasino      string   `json:"distance-to-casino"`
	// DistanceToRestaurants string   `json:"distance-to-restaurants"`
	// FibreBroadband        bool     `json:"fibre-broadband"`
	// AcceptsDero           bool     `json:"accepts-dero"`
	// AcceptsCrypto         bool     `json:"accepts-crypto"`
	// HasPool               bool     `json:"has-pool"`
	//ParkingGarage bool `json:"parking-garage"`
	// DistanceToTrain       string   `json:"distance-to-train"`
	// DistanceToAirport     string   `json:"distance-to-airport"`
	// DistanceToSubway      string   `json:"distance-to-subway"`
	// DistanceToBus         string   `json:"distance-to-bus"`
	// DistanceToFerry       string   `json:"distance-to-ferry"`
	NumberOfBedrooms  int `json:"number-of-bedrooms"`
	MaxNumberOfGuests int `json:"max-number-of-guests"`
	//HasWasherAndDryer bool `json:"has-washer-and-dryer"`
	// HasOceanViews         bool     `json:"has-ocean-views"`
	// HasBalcony            bool     `json:"has-balcony"`
	// HasPrivatePool        bool     `json:"has-private-pool"`
	// HasAirConditioner     bool     `json:"has-air-conditioner"`
	// HasHeating            bool     `json:"has-heating"`
	// HasTV                 bool     `json:"has-TV"`
	// HasFridge             bool     `json:"has-fridge"`
	// HasStovetop           bool     `json:"has-stovetop"`
	// HasOven               bool     `json:"has-oven"`
	// HasCoffeeMaker        bool     `json:"has-coffee-maker"`
	// HasBlender            bool     `json:"has-blender"`
	// HasPrivateBeachAccess bool     `json:"has-private-beach-access"`
	// DistanceToCapital     string   `json:"distance-to-capital"`
	// DistanceToShop        string   `json:"distance-to-shop"`
	// DistanceToClubs       string   `json:"distance-to-clubs"`
	// DistanceToBeach       string   `json:"distance-to-beach"`
	// HasSmokeAlarm         bool     `json:"has-smoke-alarm"`
	// HasCO2Detector        bool     `json:"has-CO2-detector"`
	// HasFireExtinguisher   bool     `json:"has-fire-extinguisher"`
	// Basement              bool     `json:"basement"`
	// Fireplaces            int      `json:"fireplaces"`
	// Flooring              string   `json:"flooring"`
	// Dishwasher            bool     `json:"dishwasher"`
	// Refrigerator          bool     `json:"refrigerator"`
	// Stove                 string   `json:"stove"`
	// Heating               string   `json:"heating"`
	// RoadFrontage          int      `json:"RoadFrontage"`
	// Water                 string   `json:"water"`
	// Lotsize               string   `json:"lotsize"`
	// BuildingExterior      string   `json:"building-exterior"`
	// Foundation            string   `json:"foundation"`
	// Levels                int      `json:"levels"`
	// Homestyle             string   `json:"homestyle"`
	// YearBuilt             int      `json:"year-built"`
	// ConstructionMaterials string   `json:"construction-materials"`
	// Roof                  string   `json:"roof"`
	// Sewer                 string   `json:"sewer"`
	// Electric              string   `json:"electric"`
}

// Get DerBnb SC code
func BnbSearchFilter() (filter []string) {
	if rpc.Daemon.Connect {
		rpcClientD, ctx, cancel := rpc.SetDaemonClient(rpc.Daemon.Rpc)
		defer cancel()

		var result *dero.GetSC_Result
		params := dero.GetSC_Params{
			SCID:      rpc.DerBnbSCID,
			Code:      true,
			Variables: false,
		}

		err := rpcClientD.CallFor(ctx, &result, "DERO.GetSC", params)
		if err != nil {
			log.Println("[searchFilters]", err)
			return nil
		}

		filter = append(filter, result.Code)

		return
	}

	return
}

// Get image urls from DerBnb property SCID
func getImages(scid string) {
	if rpc.Daemon.Connect {
		rpcClientD, ctx, cancel := rpc.SetDaemonClient(rpc.Daemon.Rpc)
		defer cancel()

		var result *dero.GetSC_Result
		params := dero.GetSC_Params{
			SCID:      scid,
			Code:      false,
			Variables: true,
		}

		err := rpcClientD.CallFor(ctx, &result, "DERO.GetSC", params)
		if err != nil {
			log.Println("[getImages]", err)
			return
		}

		if metadata, ok := result.VariableStringKeys["metadata"].(string); ok {
			if h, err := hex.DecodeString(metadata); err == nil {
				data := property_data{}
				if err = json.Unmarshal(h, &data); err == nil {
					property_photos.Lock()
					property_photos.data[scid] = data.Photos
					property_photos.Unlock()
					return
				}
				log.Println("[getImages]", err)
			} else {
				log.Println("[getImages]", err)
			}
		}
	}
}

// Get location data from DerBnb property SCID
func getLocation(scid string) (city string, country string) {
	if rpc.Daemon.Connect {
		rpcClientD, ctx, cancel := rpc.SetDaemonClient(rpc.Daemon.Rpc)
		defer cancel()

		var result *dero.GetSC_Result
		params := dero.GetSC_Params{
			SCID:      scid,
			Code:      false,
			Variables: true,
		}

		err := rpcClientD.CallFor(ctx, &result, "DERO.GetSC", params)
		if err != nil {
			log.Println("[getLocation]", err)
			return
		}

		if changed, ok := result.VariableStringKeys["changed"].(float64); ok {
			i := int(changed) - 1
			if last, ok := result.VariableStringKeys["location_"+strconv.Itoa(i)].(string); ok {
				if h, err := hex.DecodeString(last); err == nil {
					data := location_data{}
					if err = json.Unmarshal(h, &data); err == nil {
						return data.City, data.Country
					}
					log.Println("[getLocation]", err)
				} else {
					log.Println("[getLocation]", err)
				}
			}
		}
	}

	return
}

// Get metadata from DerBnb property SCID
func getMetadata(scid string) *property_data {
	if rpc.Daemon.Connect {
		rpcClientD, ctx, cancel := rpc.SetDaemonClient(rpc.Daemon.Rpc)
		defer cancel()

		var result *dero.GetSC_Result
		params := dero.GetSC_Params{
			SCID:      scid,
			Code:      false,
			Variables: true,
		}

		err := rpcClientD.CallFor(ctx, &result, "DERO.GetSC", params)
		if err != nil {
			log.Println("[getMetadata]", err)
			return nil
		}

		if metadata, ok := result.VariableStringKeys["metadata"].(string); ok {
			if h, err := hex.DecodeString(metadata); err == nil {
				data := property_data{}
				if err = json.Unmarshal(h, &data); err == nil {
					return &data
				}
				log.Println("[getMetadata]", err)
			} else {
				log.Println("[getMetadata]", err)
			}
		}
	}

	return nil
}

// Check that SC code of asset matches DerBnb standard
func checkAssetContract(scid string) string {
	if rpc.Daemon.Connect {
		rpcClientD, ctx, cancel := rpc.SetDaemonClient(rpc.Daemon.Rpc)
		defer cancel()

		var result *dero.GetSC_Result
		params := dero.GetSC_Params{
			SCID:      scid,
			Code:      true,
			Variables: false,
		}

		err := rpcClientD.CallFor(ctx, &result, "DERO.GetSC", params)
		if err != nil {
			log.Println("[checkAssetContract]", err)
			return ""
		}

		return result.Code
	}

	return ""
}

// Request booking call to DerBnb SCID
//   - stamp is current unix timestamp
//   - s_key and e_key define start and end dates
//   - amt is Dero atomic value to send
func RequestBooking(scid string, stamp, s_key, e_key, amt uint64) {
	rpcClientW, ctx, cancel := rpc.SetWalletClient(rpc.Wallet.Rpc, rpc.Wallet.UserPass)
	defer cancel()

	arg1 := dero.Argument{Name: "entrypoint", DataType: "S", Value: "RequestBooking"}
	arg2 := dero.Argument{Name: "scid", DataType: "S", Value: scid}
	arg3 := dero.Argument{Name: "stamp", DataType: "U", Value: stamp}
	arg4 := dero.Argument{Name: "start", DataType: "U", Value: s_key}
	arg5 := dero.Argument{Name: "end", DataType: "U", Value: e_key}
	args := dero.Arguments{arg1, arg2, arg3, arg4, arg5}
	txid := dero.Transfer_Result{}

	t1 := dero.Transfer{
		Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
		Amount:      0,
		Burn:        amt,
	}

	t := []dero.Transfer{t1}
	fee := rpc.GasEstimate(rpc.DerBnbSCID, "[RequestBooking]", args, t, rpc.LowLimitFee)
	params := &dero.Transfer_Params{
		Transfers: t,
		SC_ID:     rpc.DerBnbSCID,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	if err := rpcClientW.CallFor(ctx, &txid, "transfer", params); err != nil {
		log.Println("[RequestBooking]", err)
		return
	}

	log.Println("[RequestBooking] Request TX:", txid)
	rpc.AddLog("Request Booking TX: " + txid.TXID)
}

// List a DerBnb SCID for bookings
//   - amt is price per night in Dero atomic value
//   - dd is damage deposit amount in Dero atomic value
//   - burn true if token deposit is required
func ListProperty(scid string, amt, dd uint64, burn bool) {
	rpcClientW, ctx, cancel := rpc.SetWalletClient(rpc.Wallet.Rpc, rpc.Wallet.UserPass)
	defer cancel()

	arg1 := dero.Argument{Name: "entrypoint", DataType: "S", Value: "ListProperty"}
	arg2 := dero.Argument{Name: "scid", DataType: "S", Value: scid}
	arg3 := dero.Argument{Name: "price", DataType: "U", Value: amt}
	arg4 := dero.Argument{Name: "damage_deposit", DataType: "U", Value: dd}
	args := dero.Arguments{arg1, arg2, arg3, arg4}
	txid := dero.Transfer_Result{}

	tag := "[UpdatePrices]"
	bal := uint64(0)
	if burn {
		tag = "[ListProperty]"
		bal = 1
	}

	t1 := dero.Transfer{
		SCID:        crypto.HashHexToHash(scid),
		Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
		Amount:      0,
		Burn:        bal,
		Payload_RPC: []dero.Argument{},
	}

	t := []dero.Transfer{t1}
	fee := rpc.GasEstimate(rpc.DerBnbSCID, tag, args, t, rpc.LowLimitFee)
	params := &dero.Transfer_Params{
		Transfers: t,
		SC_ID:     rpc.DerBnbSCID,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	if err := rpcClientW.CallFor(ctx, &txid, "transfer", params); err != nil {
		log.Println("[ListProperty]", err)
		return
	}

	if !burn {
		log.Println("[UpdatePrices] Update TX:", txid)
		rpc.AddLog("Update Prices TX: " + txid.TXID)
	} else {
		log.Println("[ListProperty] List TX:", txid)
		rpc.AddLog("List Property TX: " + txid.TXID)
	}
}

// Remove a DerBnb SCID from listings
func RemoveProperty(scid string) {
	rpcClientW, ctx, cancel := rpc.SetWalletClient(rpc.Wallet.Rpc, rpc.Wallet.UserPass)
	defer cancel()

	arg1 := dero.Argument{Name: "entrypoint", DataType: "S", Value: "RemoveProperty"}
	arg2 := dero.Argument{Name: "scid", DataType: "S", Value: scid}
	args := dero.Arguments{arg1, arg2}
	txid := dero.Transfer_Result{}

	t1 := dero.Transfer{
		Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
		Amount:      0,
		Burn:        0,
	}

	t := []dero.Transfer{t1}
	fee := rpc.GasEstimate(rpc.DerBnbSCID, "[RemoveProperty]", args, t, rpc.LowLimitFee)
	params := &dero.Transfer_Params{
		Transfers: t,
		SC_ID:     rpc.DerBnbSCID,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	if err := rpcClientW.CallFor(ctx, &txid, "transfer", params); err != nil {
		log.Println("[RemoveProperty]", err)
		return
	}

	log.Println("[RemoveProperty] Remove property TX:", txid)
	rpc.AddLog("Remove Property TX: " + txid.TXID)
}

// Confirm booking request on DerBnb SCID
//   - stamp is current unix timestamp
func ConfirmBooking(scid string, stamp uint64) {
	rpcClientW, ctx, cancel := rpc.SetWalletClient(rpc.Wallet.Rpc, rpc.Wallet.UserPass)
	defer cancel()

	arg1 := dero.Argument{Name: "entrypoint", DataType: "S", Value: "ConfirmBooking"}
	arg2 := dero.Argument{Name: "scid", DataType: "S", Value: scid}
	arg3 := dero.Argument{Name: "stamp", DataType: "U", Value: stamp}
	args := dero.Arguments{arg1, arg2, arg3}
	txid := dero.Transfer_Result{}

	t1 := dero.Transfer{
		Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
		Amount:      0,
		Burn:        0,
	}

	t := []dero.Transfer{t1}
	fee := rpc.GasEstimate(rpc.DerBnbSCID, "[ConfirmBooking]", args, t, rpc.LowLimitFee)
	params := &dero.Transfer_Params{
		Transfers: t,
		SC_ID:     rpc.DerBnbSCID,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	if err := rpcClientW.CallFor(ctx, &txid, "transfer", params); err != nil {
		log.Println("[ConfirmBooking]", err)
		return
	}

	log.Println("[ConfirmBooking] Confirm Booking TX:", txid)
	rpc.AddLog("Confirm Booking TX: " + txid.TXID)
}

// Release specified deposit amount of DerBnb booking
//   - desc is string comment from owner when releasing deposit
//   - id is booking id to be released
//   - amt is amount of damage to be withheld by owner in Dero atomic value
func ReleaseDamageDeposit(scid, desc string, id, amt uint64) {
	rpcClientW, ctx, cancel := rpc.SetWalletClient(rpc.Wallet.Rpc, rpc.Wallet.UserPass)
	defer cancel()

	arg1 := dero.Argument{Name: "entrypoint", DataType: "S", Value: "ReleaseDamageDeposit"}
	arg2 := dero.Argument{Name: "scid", DataType: "S", Value: scid}
	arg3 := dero.Argument{Name: "id", DataType: "U", Value: id}
	arg4 := dero.Argument{Name: "damage", DataType: "U", Value: amt}
	arg5 := dero.Argument{Name: "description", DataType: "S", Value: desc}
	args := dero.Arguments{arg1, arg2, arg3, arg4, arg5}
	txid := dero.Transfer_Result{}

	t1 := dero.Transfer{
		Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
		Amount:      0,
		Burn:        0,
	}

	t := []dero.Transfer{t1}
	fee := rpc.GasEstimate(rpc.DerBnbSCID, "[ReleaseDamageDeposit]", args, t, rpc.LowLimitFee)
	params := &dero.Transfer_Params{
		Transfers: t,
		SC_ID:     rpc.DerBnbSCID,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	if err := rpcClientW.CallFor(ctx, &txid, "transfer", params); err != nil {
		log.Println("[ReleaseDamageDeposit]", err)
		return
	}

	log.Println("[ReleaseDamageDeposit] Release Deposit TX:", txid)
	rpc.AddLog("Release Damage Deposit TX: " + txid.TXID)
}

// Cancel your requested booking
//   - id is timestamp_key of booking
func CancelBooking(scid string, id uint64) {
	rpcClientW, ctx, cancel := rpc.SetWalletClient(rpc.Wallet.Rpc, rpc.Wallet.UserPass)
	defer cancel()

	arg1 := dero.Argument{Name: "entrypoint", DataType: "S", Value: "CancelBooking"}
	arg2 := dero.Argument{Name: "scid", DataType: "S", Value: scid}
	arg3 := dero.Argument{Name: "key", DataType: "U", Value: id}
	args := dero.Arguments{arg1, arg2, arg3}
	txid := dero.Transfer_Result{}

	t1 := dero.Transfer{
		Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
		Amount:      0,
		Burn:        0,
	}

	t := []dero.Transfer{t1}
	fee := rpc.GasEstimate(rpc.DerBnbSCID, "[CancelBooking]", args, t, rpc.LowLimitFee)
	params := &dero.Transfer_Params{
		Transfers: t,
		SC_ID:     rpc.DerBnbSCID,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	if err := rpcClientW.CallFor(ctx, &txid, "transfer", params); err != nil {
		log.Println("[CancelBooking]", err)
		return
	}

	log.Println("[CancelBooking] Cancel Booking TX:", txid)
	rpc.AddLog("Cancel Booking TX: " + txid.TXID)
}

// Rate your booking experience
//   - id is booking_id to be rated
//   - owner, prop, loc, overall are the rating categories a renter can store
//   - renter is the rating category a owner can store
func RateExperience(scid string, id, renter, owner, prop, loc, overall uint64) {
	rpcClientW, ctx, cancel := rpc.SetWalletClient(rpc.Wallet.Rpc, rpc.Wallet.UserPass)
	defer cancel()

	arg1 := dero.Argument{Name: "entrypoint", DataType: "S", Value: "RateExperience"}
	arg2 := dero.Argument{Name: "scid", DataType: "S", Value: scid}
	arg3 := dero.Argument{Name: "id", DataType: "U", Value: id}
	arg4 := dero.Argument{Name: "Renter", DataType: "U", Value: renter}
	arg5 := dero.Argument{Name: "Owner", DataType: "U", Value: owner}
	arg6 := dero.Argument{Name: "Property", DataType: "U", Value: prop}
	arg7 := dero.Argument{Name: "Location", DataType: "U", Value: loc}
	arg8 := dero.Argument{Name: "Overall", DataType: "U", Value: overall}
	args := dero.Arguments{arg1, arg2, arg3, arg3, arg4, arg5, arg6, arg7, arg8}
	txid := dero.Transfer_Result{}

	t1 := dero.Transfer{
		Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
		Amount:      0,
		Burn:        0,
	}

	t := []dero.Transfer{t1}
	fee := rpc.GasEstimate(rpc.DerBnbSCID, "[RateExperience]", args, t, rpc.LowLimitFee)
	params := &dero.Transfer_Params{
		Transfers: t,
		SC_ID:     rpc.DerBnbSCID,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	if err := rpcClientW.CallFor(ctx, &txid, "transfer", params); err != nil {
		log.Println("[RateExperience]", err)
		return
	}

	log.Println("[RateExperience] Rate Experience TX:", txid)
	rpc.AddLog("Rate Experience TX: " + txid.TXID)
}

// Change availability days for booking requests
//   - cal is available_dates json object
func ChangeAvailability(scid, cal string) {
	rpcClientW, ctx, cancel := rpc.SetWalletClient(rpc.Wallet.Rpc, rpc.Wallet.UserPass)
	defer cancel()

	arg1 := dero.Argument{Name: "entrypoint", DataType: "S", Value: "ChangeAvailability"}
	arg2 := dero.Argument{Name: "scid", DataType: "S", Value: scid}
	arg3 := dero.Argument{Name: "calendar_url", DataType: "S", Value: cal}
	args := dero.Arguments{arg1, arg2, arg3}
	txid := dero.Transfer_Result{}

	t1 := dero.Transfer{
		Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
		Amount:      0,
		Burn:        0,
	}

	t := []dero.Transfer{t1}
	fee := rpc.GasEstimate(rpc.DerBnbSCID, "[ChangeAvailability]", args, t, rpc.LowLimitFee)
	params := &dero.Transfer_Params{
		Transfers: t,
		SC_ID:     rpc.DerBnbSCID,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	if err := rpcClientW.CallFor(ctx, &txid, "transfer", params); err != nil {
		log.Println("[ChangeAvailability]", err)
		return
	}

	log.Println("[ChangeAvailability] Change Availability TX:", txid)
	rpc.AddLog("Change Availability TX: " + txid.TXID)
}

// Upload a new DerBnb property token contract
func UploadBnbTokenContract() (new_scid string) {
	if rpc.Daemon.Connect && rpc.Wallet.IsConnected() {
		rpcClientW, ctx, cancel := rpc.SetWalletClient(rpc.Wallet.Rpc, rpc.Wallet.UserPass)
		defer cancel()

		args := dero.Arguments{}
		txid := dero.Transfer_Result{}

		t := dero.Transfer{
			Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
			Amount:      rpc.ListingFee,
			Payload_RPC: dero.Arguments{
				{Name: dero.RPC_DESTINATION_PORT, DataType: dero.DataUint64, Value: uint64(0x6233566812245578)},
				{Name: dero.RPC_SOURCE_PORT, DataType: dero.DataUint64, Value: uint64(0)},
				{Name: dero.RPC_COMMENT, DataType: dero.DataString, Value: "Bnb Property Minted"},
			},
		}

		params := &dero.Transfer_Params{
			Transfers: []dero.Transfer{t},
			SC_Code:   TOKEN_CONTRACT,
			SC_Value:  0,
			SC_RPC:    args,
			Ringsize:  2,
		}

		if err := rpcClientW.CallFor(ctx, &txid, "transfer", params); err != nil {
			log.Println("[UploadTokenContract]", err)
			return ""
		}

		log.Println("[UploadTokenContract] Upload TX:", txid)
		rpc.AddLog("Token Upload TX: " + txid.TXID)

		return txid.TXID
	}

	return ""
}

// Store location call for DerBnb property SCID
//   - location is location_data json object
func StoreLocation(scid, location string) {
	rpcClientW, ctx, cancel := rpc.SetWalletClient(rpc.Wallet.Rpc, rpc.Wallet.UserPass)
	defer cancel()

	arg1 := dero.Argument{Name: "entrypoint", DataType: "S", Value: "StoreLocation"}
	arg2 := dero.Argument{Name: "location", DataType: "S", Value: location}
	args := dero.Arguments{arg1, arg2}
	txid := dero.Transfer_Result{}

	t1 := dero.Transfer{
		SCID:        crypto.HashHexToHash(scid),
		Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
		Amount:      0,
		Burn:        1,
	}

	t := []dero.Transfer{t1}
	fee := rpc.GasEstimate(scid, "[StoreLocation]", args, t, rpc.LowLimitFee)
	params := &dero.Transfer_Params{
		Transfers: t,
		SC_ID:     scid,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	if err := rpcClientW.CallFor(ctx, &txid, "transfer", params); err != nil {
		log.Println("[StoreLocation]", err)
		return
	}

	log.Println("[StoreLocation] Store Location TX:", txid)
	rpc.AddLog("Store Location TX: " + txid.TXID)
}

// Update metadata call for DerBnb property SCID
//   - metadata is property_data json object
func UpdateMetadata(scid, metadata string) {
	rpcClientW, ctx, cancel := rpc.SetWalletClient(rpc.Wallet.Rpc, rpc.Wallet.UserPass)
	defer cancel()

	arg1 := dero.Argument{Name: "entrypoint", DataType: "S", Value: "UpdateMetadata"}
	arg2 := dero.Argument{Name: "metadata", DataType: "S", Value: metadata}
	args := dero.Arguments{arg1, arg2}
	txid := dero.Transfer_Result{}

	t1 := dero.Transfer{
		SCID:        crypto.HashHexToHash(scid),
		Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
		Amount:      0,
		Burn:        1,
	}

	t := []dero.Transfer{t1}
	fee := rpc.GasEstimate(scid, "[UpdateMetadata]", args, t, rpc.HighLimitFee)
	params := &dero.Transfer_Params{
		Transfers: t,
		SC_ID:     scid,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	if err := rpcClientW.CallFor(ctx, &txid, "transfer", params); err != nil {
		log.Println("[UpdateMetadata]", err)
		return
	}

	log.Println("[UpdateMetadata] Update Metadata TX:", txid)
	rpc.AddLog("Update Metadata TX: " + txid.TXID)
}

// Deposit Dero or TRVL into Derbnb SC
//   - token true for TRVL
func DepositToDerBnb(token bool, amt uint64) {
	rpcClientW, ctx, cancel := rpc.SetWalletClient(rpc.Wallet.Rpc, rpc.Wallet.UserPass)
	defer cancel()

	arg1 := dero.Argument{Name: "entrypoint", DataType: "S", Value: "Deposit"}
	args := dero.Arguments{arg1}
	txid := dero.Transfer_Result{}

	var t1 dero.Transfer
	if token {
		t1 = dero.Transfer{
			SCID:        crypto.HashHexToHash(rpc.TrvlSCID),
			Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
			Amount:      0,
			Burn:        amt,
		}
	} else {
		t1 = dero.Transfer{
			Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
			Amount:      0,
			Burn:        amt,
		}
	}

	t := []dero.Transfer{t1}
	fee := rpc.GasEstimate(rpc.DerBnbSCID, "[DepositToDerBnb]", args, t, rpc.LowLimitFee)
	params := &dero.Transfer_Params{
		Transfers: t,
		SC_ID:     rpc.DerBnbSCID,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	if err := rpcClientW.CallFor(ctx, &txid, "transfer", params); err != nil {
		log.Println("[DepositToDerBnb]", err)
		return
	}

	log.Println("[DepositToDerBnb] Deposit TX:", txid)
	rpc.AddLog("DerBnb Deposit TX: " + txid.TXID)
}

// Withdraw Dero from Derbnb SC
func WithdrawFromDerBnb() {
	rpcClientW, ctx, cancel := rpc.SetWalletClient(rpc.Wallet.Rpc, rpc.Wallet.UserPass)
	defer cancel()

	arg1 := dero.Argument{Name: "entrypoint", DataType: "S", Value: "Withdraw"}
	arg2 := dero.Argument{Name: "allowance", DataType: "U", Value: 0}
	arg3 := dero.Argument{Name: "seat", DataType: "U", Value: 99}
	args := dero.Arguments{arg1, arg2, arg3}
	txid := dero.Transfer_Result{}

	t1 := dero.Transfer{
		Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
		Amount:      0,
		Burn:        0,
	}

	t := []dero.Transfer{t1}
	fee := rpc.GasEstimate(rpc.DerBnbSCID, "[WithdrawFromDerBnb]", args, t, rpc.LowLimitFee)
	params := &dero.Transfer_Params{
		Transfers: t,
		SC_ID:     rpc.DerBnbSCID,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	if err := rpcClientW.CallFor(ctx, &txid, "transfer", params); err != nil {
		log.Println("[WithdrawFromDerBnb]", err)
		return
	}

	log.Println("[WithdrawFromDerBnb] Withdraw TX:", txid)
	rpc.AddLog("DerBnb Withdraw TX: " + txid.TXID)
}

// Sell shares stored in DerBnb SC
func SellDerBnbShares(shares uint64) {
	rpcClientW, ctx, cancel := rpc.SetWalletClient(rpc.Wallet.Rpc, rpc.Wallet.UserPass)
	defer cancel()

	arg1 := dero.Argument{Name: "entrypoint", DataType: "S", Value: "SellShares"}
	arg2 := dero.Argument{Name: "shares", DataType: "U", Value: shares}
	args := dero.Arguments{arg1, arg2}
	txid := dero.Transfer_Result{}

	t1 := dero.Transfer{
		Destination: "dero1qyr8yjnu6cl2c5yqkls0hmxe6rry77kn24nmc5fje6hm9jltyvdd5qq4hn5pn",
		Amount:      0,
		Burn:        0,
	}

	t := []dero.Transfer{t1}
	fee := rpc.GasEstimate(rpc.DerBnbSCID, "[SellDerBnbShares]", args, t, rpc.LowLimitFee)
	params := &dero.Transfer_Params{
		Transfers: t,
		SC_ID:     rpc.DerBnbSCID,
		SC_RPC:    args,
		Ringsize:  2,
		Fees:      fee,
	}

	if err := rpcClientW.CallFor(ctx, &txid, "transfer", params); err != nil {
		log.Println("[SellDerBnbShares]", err)
		return
	}

	log.Println("[SellDerBnbShares] Sell Shares TX:", txid)
	rpc.AddLog("DerBnb Sell Shares TX: " + txid.TXID)
}
