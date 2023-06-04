// derbnb property contract v1

Function InitializePrivate() Uint64
	10 IF EXISTS("metadata") THEN GOTO 100
	20 SEND_ASSET_TO_ADDRESS(SIGNER(), 1, SCID())
	30 STORE("metadata","")
	40 STORE("changed", 0)
	50 RETURN 0
	100 RETURN 1
End Function

Function StoreLocation(location String) Uint64
	10 IF ASSETVALUE(SCID()) != 1 THEN GOTO 100
	20 IF location == "" THEN GOTO 100
	30 STORE("location_"+ITOA(LOAD("changed")), location)
	40 IF LOAD("changed") < 6 THEN GOTO 60
	50 DELETE("location_"+ITOA(LOAD("changed")-5))
	60 STORE("changed", LOAD("changed")+1)
	70 SEND_ASSET_TO_ADDRESS(SIGNER(), 1, SCID())
	80 RETURN 0 
	100 RETURN 1
End Function

Function UpdateMetadata(metadata String) Uint64
	10 IF ASSETVALUE(SCID()) != 1 THEN GOTO 100
	20 STORE("metadata", metadata)
	30 SEND_ASSET_TO_ADDRESS(SIGNER(), 1, SCID())
	40 RETURN 0
	100 RETURN 1
End Function