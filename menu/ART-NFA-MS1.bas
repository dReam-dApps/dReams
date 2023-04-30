//    Copyright 2022. Civilware. All rights reserved.
//    Artificer NFA Market Standard (ART-NFA-MS1)

Function InitializePrivate() Uint64
    10  IF EXISTS("owner") == 0 THEN GOTO 300 ELSE GOTO 999
    300 STORE("artificerFee", 0)
    310 STORE("royalty", 0)
    320 STORE("ownerCanUpdate", 0)
    330 STORE("nameHdr", "<nameHdr>")
    340 STORE("descrHdr", "<descrHdr>")
    350 STORE("typeHdr", "<typeHdr>")
    360 STORE("iconURLHdr", "<iconURLHdr>")
    370 STORE("tagsHdr", "<tagsHdr>")
    400 STORE("fileCheckC", "<fileCheckC>")
    410 STORE("fileCheckS", "<fileCheckS>")
    420 STORE("fileURL", "<fileURL>")
    430 STORE("fileSignURL", "<fileSignURL>")
    440 STORE("coverURL", "<coverURL>")
    450 STORE("collection", "<collection>")
    500 IF init() == 0 THEN GOTO 600 ELSE GOTO 999
    600 RETURN 0
    999 RETURN 1
End Function

Function init() Uint64
    10  IF EXISTS("owner") == 0 THEN GOTO 20 ELSE GOTO 999
    20  STORE("owner", SIGNER())
    30  STORE("creatorAddr", SIGNER())
    40  STORE("artificerAddr", ADDRESS_RAW("dero1qy0khp9s9yw2h0eu20xmy9lth3zp5cacmx3rwt6k45l568d2mmcf6qgcsevzx"))
    50  IF IS_ADDRESS_VALID(LOAD("artificerAddr")) == 1 THEN GOTO 60 ELSE GOTO 999
    60  STORE("active", 0)
    70  STORE("scBalance", 0)
    80  STORE("cancelBuffer", 300)
    90  STORE("startBlockTime", 0)
    100 STORE("endBlockTime", 0)
    110 STORE("bidCount", 0)
    120 STORE("staticBidIncr", 10000)
    130 STORE("percentBidIncr", 1000)
    140 STORE("listType", "")
    150 STORE("charityDonatePerc", 0)
    160 STORE("startPrice", 0)
    170 STORE("currBidPrice", 0)
    180 STORE("version", "1.1.1")
    500 IF LOAD("charityDonatePerc") + LOAD("artificerFee") + LOAD("royalty") > 100 THEN GOTO 999
    600 SEND_ASSET_TO_ADDRESS(SIGNER(), 1, SCID())
    610 RETURN 0
    999 RETURN 1
End Function

Function ClaimOwnership() Uint64
    10  IF ASSETVALUE(SCID()) == 1 THEN GOTO 20 ELSE GOTO 999
    20  IF ADDRESS_STRING(SIGNER()) == "" THEN GOTO 500
    30  transferOwnership(SIGNER())
    40  SEND_ASSET_TO_ADDRESS(SIGNER(), 1, SCID())
    50  RETURN 0
    500 SEND_ASSET_TO_ADDRESS(LOAD("owner"), 1, SCID())
    510 RETURN 0
    999 RETURN 1
End Function

Function Update(iconURL String, coverURL String, fileURL String, fileSignURL String, tags String) Uint64
    10  IF LOAD("creatorAddr") == SIGNER() THEN GOTO 40 ELSE GOTO 20
    20  IF LOAD("ownerCanUpdate") == 1 THEN GOTO 30 ELSE GOTO 999
    30  IF LOAD("owner") == SIGNER() THEN GOTO 40 ELSE GOTO 999
    40  IF iconURL != "" THEN GOTO 50 ELSE GOTO 60
    50  STORE("iconURLHdr", iconURL)
    60  IF coverURL != "" THEN GOTO 70 ELSE GOTO 80
    70  STORE("coverURL", coverURL)
    80  IF fileURL != "" THEN GOTO 90 ELSE GOTO 100
    90  STORE("fileURL", fileURL)
    100 IF fileSignURL != "" THEN GOTO 110 ELSE GOTO 120
    110 STORE("fileSignURL", fileSignURL)
    120 IF tags != "" THEN GOTO 130 ELSE GOTO 140
    130 STORE("tagsHdr", tags)
    140 RETURN 0
    999 RETURN 1
End Function

Function Start(listType String, duration Uint64, startPrice Uint64, charityDonateAddr String, charityDonatePerc Uint64) Uint64
    10  dim tempPercCount as Uint64
    20  dim err as String
    30  IF ADDRESS_STRING(SIGNER()) == "" THEN GOTO 600
    40  IF ASSETVALUE(SCID()) == 1 THEN GOTO 70 ELSE GOTO 400
    70  IF listType == "auction" THEN GOTO 100 ELSE GOTO 80
    80  IF listType == "sale" THEN GOTO 100 ELSE GOTO 400
    100 IF LOAD("owner") == SIGNER() THEN GOTO 110 ELSE GOTO 400
    110 IF checkActive(LOAD("listType")) == 999 THEN GOTO 150 ELSE GOTO 400
    150 IF charityDonatePerc + LOAD("artificerFee") + LOAD("royalty") > 100 THEN GOTO 160 ELSE GOTO 190
    160 LET tempPercCount = 100 - LOAD("artificerFee") - LOAD("royalty")
    165 LET charityDonatePerc = tempPercCount
    170 STORE("charityDonatePerc", charityDonatePerc)
    175 IF IS_ADDRESS_VALID(ADDRESS_RAW(charityDonateAddr)) == 1 THEN GOTO 180 ELSE GOTO 400
    180 STORE("charityDonateAddr", ADDRESS_RAW(charityDonateAddr))
    185 GOTO 210
    190 IF charityDonatePerc > 0 THEN GOTO 195 ELSE GOTO 210
    195 IF IS_ADDRESS_VALID(ADDRESS_RAW(charityDonateAddr)) == 1 THEN GOTO 200 ELSE GOTO 400
    200 STORE("charityDonatePerc", charityDonatePerc)
    205 STORE("charityDonateAddr", ADDRESS_RAW(charityDonateAddr))
    210 STORE("listType", listType)
    220 STORE("scBalance", 1)
    230 STORE("startBlockTime", BLOCK_TIMESTAMP())
    240 STORE("endBlockTime", generateEndBlock(duration, BLOCK_TIMESTAMP()))
    250 STORE("startPrice", startPrice)
    270 STORE("active", 1)
    300 RETURN 0
    400 IF ASSETVALUE(SCID()) > 0 THEN GOTO 410 ELSE GOTO 999
    410 SEND_ASSET_TO_ADDRESS(SIGNER(), ASSETVALUE(SCID()), SCID())
    420 RETURN 0
    600 IF ASSETVALUE(SCID()) > 0 THEN GOTO 610 ELSE GOTO 999
    610 SEND_ASSET_TO_ADDRESS(LOAD("owner"), ASSETVALUE(SCID()), SCID())
    620 RETURN 0
    999 RETURN 1
End Function

Function BuyItNow() Uint64
    10  dim activeFlag as Uint64
    15  IF ADDRESS_STRING(SIGNER()) == "" THEN GOTO 999
    20  IF LOAD("owner") == SIGNER() THEN GOTO 920 ELSE GOTO 30
    30  LET activeFlag = checkActive("sale")
    40  IF activeFlag == 0 THEN GOTO 50 ELSE GOTO 500
    50  IF LOAD("scBalance") == 1 THEN GOTO 60 ELSE GOTO 920
    60  IF DEROVALUE() >= LOAD("startPrice") THEN GOTO 70 ELSE GOTO 920
    70  SEND_ASSET_TO_ADDRESS(SIGNER(), LOAD("scBalance"), SCID())
    80  STORE("scBalance", 0)
    90  processDEROFinalPayment(DEROVALUE())
    95  transferOwnership(SIGNER())
    96  resetVars(1)
    97  STORE("previousSalePrice", DEROVALUE())
    100 RETURN 0
    500 IF activeFlag == 999 THEN GOTO 920 ELSE GOTO 510
    510 IF activeFlag == 111 THEN GOTO 520 ELSE GOTO 920
    520 SEND_ASSET_TO_ADDRESS(LOAD("owner"), LOAD("scBalance"), SCID())
    530 STORE("scBalance", 0)
    540 resetVars(1)
    920 IF DEROVALUE() > 0 THEN GOTO 925 ELSE GOTO 930
    925 SEND_DERO_TO_ADDRESS(SIGNER(), DEROVALUE())
    930 RETURN 0
    999 RETURN 1
End Function

Function Bid() Uint64
    10  dim activeFlag, bidAmt as Uint64
    15  IF ADDRESS_STRING(SIGNER()) == "" THEN GOTO 999
    25  LET bidAmt = DEROVALUE()
    30  IF LOAD("owner") == SIGNER() THEN GOTO 920 ELSE GOTO 35
    35  LET activeFlag = checkActive("auction")
    40  IF activeFlag == 0 THEN GOTO 50 ELSE GOTO 500
    50  IF LOAD("scBalance") == 1 THEN GOTO 51 ELSE GOTO 920
    51  IF EXISTS(SIGNER() + "-bidDate") == 1 THEN GOTO 60 ELSE GOTO 70
    60  IF LOAD(SIGNER() + "-bidDate") < BLOCK_TIMESTAMP() THEN GOTO 70 ELSE GOTO 920
    70  IF bidAmt >= LOAD("startPrice") THEN GOTO 80 ELSE GOTO 920
    80  IF bidAmt >= LOAD("currBidPrice") THEN GOTO 90 ELSE GOTO 920
    90  STORE("currBidPrice", findLesserIncrease(bidAmt))
    100 outbidReturns()
    120 STORE("currBidAddr", SIGNER())
    130 STORE("currBidAmt", bidAmt)
    140 STORE(SIGNER() + "-bidDate", BLOCK_TIMESTAMP())
    150 STORE("bidCount", LOAD("bidCount") + 1)
    170 IF LOAD("endBlockTime") - 900 <= BLOCK_TIMESTAMP() THEN GOTO 180 ELSE GOTO 190
    180 STORE("endBlockTime", BLOCK_TIMESTAMP() + 900)
    190 RETURN 0
    500 IF activeFlag == 999 THEN GOTO 920 ELSE GOTO 510
    510 IF activeFlag == 111 THEN GOTO 520 ELSE GOTO 920
    520 IF bidAmt > 0 THEN GOTO 530 ELSE GOTO 540
    530 SEND_DERO_TO_ADDRESS(SIGNER(), bidAmt)
    540 processHighestBidder()
    550 RETURN 0
    920 IF bidAmt > 0 THEN GOTO 925 ELSE GOTO 930
    925 SEND_DERO_TO_ADDRESS(SIGNER(), bidAmt)
    930 RETURN 0
    999 RETURN 1
End Function

Function CloseListing() Uint64
    10  IF LOAD("owner") == SIGNER() THEN GOTO 20 ELSE GOTO 999
    20  IF checkActive(LOAD("listType")) == 111 THEN GOTO 30 ELSE GOTO 999
    30  IF LOAD("listType") == "auction" THEN GOTO 40 ELSE GOTO 200
    40  IF LOAD("bidCount") > 0 THEN GOTO 50 ELSE GOTO 210
    50  processHighestBidder()
    60  RETURN 0
    200 IF LOAD("listType") == "sale" THEN GOTO 210 ELSE GOTO 999
    210 SEND_ASSET_TO_ADDRESS(LOAD("owner"), LOAD("scBalance"), SCID())
    220 STORE("scBalance", 0)
    230 resetVars(1)
    240 RETURN 0
    999 RETURN 1
End Function

Function CancelListing() Uint64
    10  dim tempCounter as Uint64
    30  IF LOAD("owner") == SIGNER() THEN GOTO 50 ELSE GOTO 999
    50  IF checkActive(LOAD("listType")) == 0 THEN GOTO 60 ELSE GOTO 999
    60  IF (LOAD("startBlockTime") + LOAD("cancelBuffer")) >= BLOCK_TIMESTAMP() THEN GOTO 460 ELSE GOTO 999
    460 outbidReturns()
    600 SEND_ASSET_TO_ADDRESS(LOAD("owner"), LOAD("scBalance"), SCID())
    610 STORE("scBalance", 0)
    620 resetVars(1)
    630 RETURN 0
    999 RETURN 1
End Function

Function transferOwnership(newOwner String) Uint64
    10  IF LOAD("owner") == newOwner THEN GOTO 40 ELSE GOTO 20
    20  STORE("previousOwner", LOAD("owner"))
    30  STORE("owner", newOwner)
    40  RETURN 0
End Function

Function generateEndBlock(duration Uint64, startBlockTime Uint64) Uint64
    10  dim timeinseconds, endBlockTime as Uint64
    20  LET timeinseconds = 3600 * duration
    30  IF timeinseconds == 0 THEN GOTO 40 ELSE GOTO 50
    40  LET timeinseconds = 3600
    50  IF timeinseconds > 604800 THEN GOTO 60 ELSE GOTO 70
    60  LET timeinseconds = 604800
    70  LET endBlockTime = startBlockTime + timeinseconds
    80  RETURN endBlockTime
End Function

Function checkActive(listType String) Uint64
    10  IF LOAD("startBlockTime") <= BLOCK_TIMESTAMP() THEN GOTO 30 ELSE GOTO 900
    30  IF LOAD("scBalance") == 1 THEN GOTO 40 ELSE GOTO 900
    40  IF LOAD("endBlockTime") > BLOCK_TIMESTAMP() THEN GOTO 50 ELSE GOTO 500
    50  IF LOAD("listType") == listType THEN GOTO 200 ELSE GOTO 910
    200 STORE("active", 1)
    210 RETURN 0
    500 STORE("active", 0)
    520 RETURN 111
    900 STORE("active", 0)
    910 RETURN 999
End Function

Function processHighestBidder() Uint64
    10  dim bidAmt as Uint64
    20  dim bidAddr as String
    30  LET bidAddr = LOAD("owner")
    100 IF EXISTS("currBidAmt") == 1 THEN GOTO 110 ELSE GOTO 310
    110 IF EXISTS("currBidAddr") == 1 THEN GOTO 120 ELSE GOTO 310
    120 IF LOAD("currBidAddr") != "" THEN GOTO 130 ELSE GOTO 310
    130 IF LOAD("currBidAmt") > 0 THEN GOTO 140 ELSE GOTO 310
    140 LET bidAddr = LOAD("currBidAddr")
    150 LET bidAmt = LOAD("currBidAmt")
    310 SEND_ASSET_TO_ADDRESS(bidAddr, LOAD("scBalance"), SCID())
    320 processDEROFinalPayment(bidAmt)
    350 transferOwnership(bidAddr)
    360 STORE("scBalance", 0)
    370 IF EXISTS(bidAddr + "-bidDate") == 1 THEN GOTO 380 ELSE GOTO 390
    380 DELETE(bidAddr + "-bidDate")
    390 DELETE("currBidAddr")
    400 DELETE("currBidAmt")
    410 IF bidAmt > 0 THEN GOTO 420 ELSE GOTO 600
    420 STORE("previousAuctionPrice", bidAmt)
    600 resetVars(1)
    610 RETURN 0
End Function

Function outbidReturns() Uint64
    20  IF EXISTS("currBidAddr") == 1 THEN GOTO 30 ELSE GOTO 900
    30  IF EXISTS("currBidAmt") == 1 THEN GOTO 40 ELSE GOTO 900
    40  IF LOAD("currBidAmt") > 0 THEN GOTO 50 ELSE GOTO 900
    50  IF LOAD("currBidAddr") != "" THEN GOTO 60 ELSE GOTO 900
    60  SEND_DERO_TO_ADDRESS(LOAD("currBidAddr"), LOAD("currBidAmt"))
    800 DELETE(LOAD("currBidAddr") + "-bidDate")
    810 DELETE("currBidAddr")
    820 DELETE("currBidAmt")
    900 RETURN 0
End Function

Function processDEROFinalPayment(saleAmt Uint64) Uint64
    10  dim payoutAmt, royaltyPaymt, artificerPaymt, charityPaymt as Uint64
    20  IF saleAmt == 0 THEN GOTO 200 ELSE GOTO 60
    60  IF LOAD("royalty") > 0 THEN GOTO 65 ELSE GOTO 80
    65  LET royaltyPaymt = LOAD("royalty") * saleAmt / 100
    66  IF royaltyPaymt > 0 THEN GOTO 70 ELSE GOTO 80
    70  SEND_DERO_TO_ADDRESS(LOAD("creatorAddr"), royaltyPaymt)
    80  IF LOAD("artificerFee") > 0 THEN GOTO 85 ELSE GOTO 100
    85  LET artificerPaymt = LOAD("artificerFee") * saleAmt / 100
    86  IF artificerPaymt > 0 THEN GOTO 90 ELSE GOTO 100
    90  SEND_DERO_TO_ADDRESS(LOAD("artificerAddr"), artificerPaymt)
    100 IF LOAD("charityDonatePerc") > 0 THEN GOTO 105 ELSE GOTO 120
    105 LET charityPaymt = LOAD("charityDonatePerc") * saleAmt / 100
    106 IF charityPaymt > 0 THEN GOTO 110 ELSE GOTO 120
    110 SEND_DERO_TO_ADDRESS(LOAD("charityDonateAddr"), charityPaymt)
    120 LET payoutAmt = saleAmt - royaltyPaymt - artificerPaymt - charityPaymt
    125 IF payoutAmt > 0 THEN GOTO 130 ELSE GOTO 200
    130 SEND_DERO_TO_ADDRESS(LOAD("owner"), payoutAmt)
    200 RETURN 0
End Function

Function resetVars(forceReset Uint64) Uint64
    10  IF forceReset == 0 THEN GOTO 20 ELSE GOTO 30
    20  IF checkActive(LOAD("listType")) == 999 THEN GOTO 20 ELSE GOTO 900
    30  STORE("startBlockTime", 0)
    40  STORE("endBlockTime", 0)
    50  STORE("bidCount", 0)
    60  STORE("active", 0)
    80  STORE("startPrice", 0)
    90  STORE("currBidPrice", 0)
    100 STORE("listType", "")
    110 STORE("charityDonateAddr", "")
    120 STORE("charityDonatePerc", 0)
    200 RETURN 0
    900 RETURN 999
End Function

Function findLesserIncrease(bidAmt Uint64) Uint64
    10  dim percentCalc, staticCalc as Uint64
    20  LET percentCalc = bidAmt + (bidAmt * LOAD("percentBidIncr") / 10000)
    30  LET staticCalc = bidAmt + LOAD("staticBidIncr")
    50  IF percentCalc < staticCalc THEN GOTO 100 ELSE GOTO 200
    100 RETURN percentCalc
    200 RETURN staticCalc
End Function