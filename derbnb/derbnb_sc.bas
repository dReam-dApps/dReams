Function Deposit() Uint64
10 IF ASSETVALUE(HEXDECODE(LOAD("TOKEN"))) % 10000 != 0 THEN GOTO 100
13 STORE("TREASURY",LOAD("TREASURY")+DEROVALUE())
15 IF EXISTS(ADDRESS_STRING(SIGNER())+"_SHARES") THEN GOTO 30
20 STORE(ADDRESS_STRING(SIGNER())+"_SHARES",ASSETVALUE(HEXDECODE(LOAD("TOKEN")))/10000)
21 STORE(ADDRESS_STRING(SIGNER())+"_EPOCH",(BLOCK_TIMESTAMP() - LOAD("EPOCH-INIT"))/2629743)
25 RETURN 0
30 STORE(ADDRESS_STRING(SIGNER())+"_SHARES",LOAD(ADDRESS_STRING(SIGNER())+"_SHARES")+ASSETVALUE(HEXDECODE(LOAD("TOKEN")))/10000)
31 STORE(ADDRESS_STRING(SIGNER())+"_EPOCH",(BLOCK_TIMESTAMP() - LOAD("EPOCH-INIT"))/2629743)
35 RETURN 0
100 RETURN 1
End Function

Function Withdraw(allowance Uint64, seat Uint64) Uint64
5 IF ASSETVALUE(HEXDECODE(LOAD("CEO"))) == 1 THEN GOTO 30
10 DIM EPOCH as Uint64
11 LET EPOCH = (BLOCK_TIMESTAMP()-LOAD("EPOCH-INIT"))/2629743
12 IF EXISTS("SEAT_"+seat) == 0 THEN GOTO 14
13 IF ASSETVALUE(HEXDECODE(LOAD("SEAT_"+seat))) == 1 THEN GOTO 80
14 IF EXISTS(ADDRESS_STRING(SIGNER())+"_SHARES") ==0 THEN GOTO 100
15 IF LOAD(ADDRESS_STRING(SIGNER())+"_EPOCH") >= EPOCH THEN GOTO 100
16 DIM SHARE as Uint64
17 LET SHARE = LOAD(ADDRESS_STRING(SIGNER())+"_SHARES")*LOAD("TREASURY")/100000
18 SEND_DERO_TO_ADDRESS(SIGNER(),SHARE)
19 STORE(ADDRESS_STRING(SIGNER())+"_EPOCH",EPOCH)
20 STORE("TREASURY",LOAD("TREASURY")-SHARE)
25 RETURN 0
30 IF allowance > LOAD("ALLOWANCE") THEN GOTO 100
40 SEND_DERO_TO_ADDRESS(SIGNER(),allowance)
50 STORE("ALLOWANCE",LOAD("ALLOWANCE") - allowance)
60 SEND_ASSET_TO_ADDRESS(SIGNER(),1,HEXDECODE(LOAD("CEO")))
65 STORE("TREASURY",LOAD("TREASURY")-allowance)
70 RETURN 0
80 IF LOAD("SEAT_"+seat+"_EPOCH") >= EPOCH THEN GOTO 100
81 DIM SAL as Uint64
82 LET SAL = 5*LOAD("TREASURY")/100
83 SEND_DERO_TO_ADDRESS(SIGNER(),SAL)
84 STORE("SEAT_"+seat+"_EPOCH",EPOCH)
85 STORE("TREASURY",LOAD("TREASURY")-SAL)
86 SEND_ASSET_TO_ADDRESS(SIGNER(),1,HEXDECODE(LOAD("SEAT_"+seat)))
90 RETURN 0
100 RETURN 1
End Function

Function SellShares(shares Uint64) Uint64
10 IF EXISTS(ADDRESS_STRING(SIGNER())+"_SHARES") == 0 THEN GOTO 100
20 IF LOAD(ADDRESS_STRING(SIGNER())+"_SHARES") < shares THEN GOTO 100
30 STORE(ADDRESS_STRING(SIGNER())+"_SHARES",LOAD(ADDRESS_STRING(SIGNER())+"_SHARES")-shares)
40 SEND_ASSET_TO_ADDRESS(SIGNER(),shares*10000,HEXDECODE(LOAD("TOKEN")))
99 RETURN 0
100 RETURN 1
End Function

Function ListProperty(scid String, price Uint64, damage_deposit Uint64) Uint64
10 IF EXISTS(scid+"_owner")==0 THEN GOTO 30
20 IF LOAD(scid+"_owner")==ADDRESS_STRING(SIGNER()) THEN GOTO 70 ELSE GOTO 100
30 IF ASSETVALUE(HEXDECODE(scid))!=1 THEN GOTO 100
40 STORE(scid+"_owner",ADDRESS_STRING(SIGNER()))
50 IF EXISTS(scid+"_bk_last") THEN GOTO 70
60 STORE(scid+"_bk_last",0)
70 STORE(scid+"_price", price)
80 STORE(scid+"_damage_deposit", damage_deposit)
99 RETURN 0
100 RETURN 1
End Function 

Function RemoveProperty(scid String) Uint64
10 IF LOAD(scid+"_owner") != ADDRESS_STRING(SIGNER()) THEN GOTO 100
20 DELETE(scid+"_owner")
30 SEND_ASSET_TO_ADDRESS(SIGNER(),1,HEXDECODE(scid))
99 RETURN 0
100 RETURN 1
End Function

Function ChangeAvailability(scid String, calendar_url String) Uint64
10 IF LOAD(scid+"_owner") != ADDRESS_STRING(SIGNER()) THEN GOTO 100
20 STORE(scid + "_bk_avail", calendar_url)
99 RETURN 0
100 RETURN 1
End Function

Function ConfirmBooking(scid String, stamp Uint64) Uint64
10 IF LOAD(scid+"_owner") != ADDRESS_STRING(SIGNER()) THEN GOTO 100
11 IF BLOCK_TIMESTAMP() > LOAD(scid+"_request_bk_start_"+stamp) THEN GOTO 100
15 DIM id, count as Uint64
20 LET id = LOAD(scid + "_bk_last") + 1
24 LET count = id
25 IF count == 1 THEN GOTO 30
26 LET count = count -1
27 IF LOAD(scid + "_bk_start_"+count) > LOAD(scid + "_request_bk_end_"+ stamp) THEN GOTO 25
28 IF LOAD(scid + "_bk_end_"+count) < LOAD(scid + "_request_bk_start_"+ stamp) THEN GOTO 25 ELSE GOTO 100
30 STORE(scid + "_bk_last", id)
31 STORE(scid + "_booker_" + id, LOAD(scid + "_request_booker_"+ stamp))
32 STORE(scid + "_bk_start_" + id, LOAD(scid + "_request_bk_start_"+ stamp))
33 STORE(scid + "_bk_end_" + id, LOAD(scid + "_request_bk_end_"+ stamp))
34 STORE(scid + "_payment_" + id, LOAD(scid + "_request_payment_"+ stamp))
35 SEND_DERO_TO_ADDRESS(SIGNER(),90*(LOAD(scid + "_request_payment_"+ stamp)-LOAD(scid + "_deposit_"+stamp))/100)
40 STORE("TREASURY", LOAD("TREASURY") + 10*(LOAD(scid + "_request_payment_"+ stamp)-LOAD(scid + "_deposit_"+stamp))/100)
69 STORE(scid+"_deposit_"+id, LOAD(scid+"_deposit_"+stamp))
70 DELETE(scid + "_request_booker_"+ stamp)
71 DELETE(scid + "_request_bk_start_"+ stamp)
72 DELETE(scid + "_request_bk_end_"+ stamp)
73 DELETE(scid + "_request_payment_"+ stamp)
74 DELETE(scid+"_deposit_"+stamp)
99 RETURN 0
100 RETURN 1
End Function

Function RateExperience(scid String, id Uint64, Renter Uint64, Owner Uint64, Property Uint64, Location Uint64, Overall Uint64) Uint64
10 IF ADDRESS_STRING(SIGNER()) == LOAD(scid+"_booker_"+id) THEN GOTO 40
20 IF ADDRESS_STRING(SIGNER()) == LOAD(scid+"_owner") THEN GOTO 90
30 RETURN 1
40 STORE(scid+"_"+id+"_rating_property",Property)
50 STORE(scid+"_"+id+"_rating_location",Location)
60 STORE(scid+"_"+id+"_rating_owner",Owner)
70 STORE(scid+"_"+id+"_rating_overall",Overall)
80 RETURN 0
90 STORE(scid+"_"+id+"_rating_renter",Renter)
100 RETURN 0
End Function

Function RequestBooking(scid String,stamp Uint64,start Uint64,end Uint64) Uint64
15 IF DEROVALUE() < LOAD(scid+"_price") * (end-start)/86400 + LOAD(scid+"_damage_deposit") THEN GOTO 100
20 IF ADDRESS_STRING(SIGNER()) == "" THEN GOTO 100
30 IF EXISTS(scid + "_request_bk_start_" + stamp) != 0 THEN GOTO 100
40 STORE(scid + "_request_booker_" + stamp, ADDRESS_STRING(SIGNER()))
50 STORE(scid + "_request_bk_start_" + stamp, start)
60 STORE(scid + "_request_bk_end_" + stamp, end)
70 STORE(scid + "_request_payment_"+ stamp,DEROVALUE())
75 STORE(scid+"_deposit_"+stamp, LOAD(scid+"_damage_deposit"))
99 RETURN 0
100 RETURN 1
End Function

Function CancelBooking(scid String, key Uint64) Uint64
10 IF EXISTS(scid+"_request_booker_"+key) == 0 THEN GOTO 100
15 IF SIGNER()!=ADDRESS_RAW(LOAD(scid+"_request_booker_"+key)) && SIGNER()!=ADDRESS_RAW(LOAD(scid+"_owner")) THEN GOTO 100
20 SEND_DERO_TO_ADDRESS(ADDRESS_RAW(LOAD(scid+"_request_booker_"+key)),LOAD(scid+"_request_payment_"+key))
30 DELETE(scid+"_request_booker_"+key)
40 DELETE(scid+"_request_bk_start_"+key)
50 DELETE(scid+"_request_bk_end_"+key)
60 DELETE(scid+"_request_payment_"+key)
75 DELETE(scid+"_deposit_"+key)
99 RETURN 0
100 RETURN 1
End Function

Function ReleaseDamageDeposit(scid String,id Uint64,damage Uint64,description String) Uint64
10 IF LOAD(scid+"_owner") != ADDRESS_STRING(SIGNER()) THEN GOTO 1000
30 DIM renter as String
40 DIM deposit,release as Uint64
70 LET renter=LOAD(scid+"_booker_"+id)
80 LET deposit=LOAD(scid+"_deposit_"+id)
90 IF damage>deposit THEN GOTO 1000
110 IF damage>0&&description=="" THEN GOTO 1000
120 LET release=deposit-damage
130 STORE(scid+"_"+id+"_damage_amount",damage)
140 STORE(scid+"_"+id+"_damage_description",description)
150 STORE(scid+"_"+id+"_damage_renter",renter)
170 IF damage==0 THEN GOTO 190
180 SEND_DERO_TO_ADDRESS(SIGNER(),damage)
190 IF release==0 THEN GOTO 999
200 SEND_DERO_TO_ADDRESS(ADDRESS_RAW(renter),release)
999 RETURN 0
1000 RETURN 1
End Function

Function Propose(hash String, k String, u Uint64, s String, t Uint64, seat Uint64) Uint64
10 IF ASSETVALUE(HEXDECODE(LOAD("CEO"))) != 1 THEN GOTO 13
11 SEND_ASSET_TO_ADDRESS(SIGNER(),1,HEXDECODE(LOAD("CEO")))
12 GOTO 15
13 IF ASSETVALUE(HEXDECODE(LOAD("SEAT_"+seat))) !=1 THEN GOTO 100
14 SEND_ASSET_TO_ADDRESS(SIGNER(),1,HEXDECODE(LOAD("SEAT_"+seat)))
15 STORE("APPROVE", 0)
20 IF hash =="" THEN GOTO 40
25 STORE("HASH",hash)
30 STORE("k","")
35 RETURN 0
40 STORE("k",k)
45 STORE("HASH","")
49 STORE("t",t)
50 IF t == 1 THEN GOTO 80
60 STORE("s", s)
70 RETURN 0
80 STORE("u",u)
90 RETURN 0
100 RETURN 1
End Function

Function Approve(seat Uint64) Uint64
10 IF ASSETVALUE(HEXDECODE(LOAD("SEAT_"+seat)))!=1 THEN GOTO 100
20 STORE("APPROVE",LOAD("APPROVE")+1)
30 STORE("SEAT_"+seat+"_OWNER",SIGNER())
99 RETURN 0
100 RETURN 1
End Function

Function ClaimSeat(seat Uint64) Uint64
10 IF SIGNER()!= LOAD("SEAT_"+seat+"_OWNER") THEN GOTO 100
20 SEND_ASSET_TO_ADDRESS(SIGNER(),1,HEXDECODE(LOAD("SEAT_"+seat)))
30 IF LOAD("APPROVE") == 0 THEN GOTO 99
40 STORE("APPROVE",LOAD("APPROVE")-1)
99 RETURN 0
100 RETURN 1
End Function

Function Update(code String) Uint64
10 IF ASSETVALUE(HEXDECODE(LOAD("CEO")))!=1 THEN GOTO 100
15 SEND_ASSET_TO_ADDRESS(SIGNER(),1,HEXDECODE(LOAD("CEO")))
20 IF SHA256(code) != HEXDECODE(LOAD("HASH")) THEN GOTO 100
30 IF LOAD("APPROVE") < LOAD("QUORUM") THEN GOTO 100
40 UPDATE_SC_CODE(code)
99 RETURN 0
100 RETURN 1
End Function

Function Store(k String, u Uint64, s String) Uint64
20 IF k != LOAD("k") THEN GOTO 999
40 IF LOAD("APPROVE") < LOAD("QUORUM") THEN GOTO 999
50 dim t as Uint64
60 let t = LOAD("t")
110 IF t == 0 THEN GOTO 150
120 IF t == 1 THEN GOTO 170
130 IF s!=LOAD("s") THEN GOTO 999
135 STORE(k, HEX(s))
140 RETURN 0
150 IF s!=LOAD("s") THEN GOTO 999
155 STORE(k, s)
160 RETURN 0
170 IF u!=LOAD("u") THEN GOTO 999
175 STORE(k,u)
180 RETURN 0
999 RETURN 1
End Function