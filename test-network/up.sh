./network.sh down
./network.sh up createChannel -c pmchannel -ca -s couchdb
#./monitordocker.sh fabric_test &> /var/log/watch.log &

./network.sh deployCC -c pmchannel -ccn pgacc -ccp ../chaincode/pgacc -ccl go
./network.sh deployCC -c pmchannel -ccn amcc -ccp ../chaincode/amcc -ccl go
./network.sh deployCC -c pmchannel -ccn ttcc -ccp ../chaincode/ttcc -ccl go

export PATH=${PWD}/../bin:$PATH
export FABRIC_CFG_PATH=$PWD/../config/
export FABRIC_CA_CLIENT_HOME=${PWD}/organizations/peerOrganizations/org1.example.com/


./enroll-Admin.sh

./enroll-John.sh

./enroll-Amy.sh

peer chaincode invoke "${TARGET_TLS_OPTIONS[@]}" -C pmchannel -n pgacc -c '{"Args":["init","gREITAccess"]}'

: '
peer chaincode invoke "${TARGET_TLS_OPTIONS[@]}" -C pmchannel -n pgacc -c '{"Args":["setTokenizedAsset", "HamHol", "67000", "21000", "USD-T", "1", "1DPT", "33000","London"]}'

peer chaincode invoke "${TARGET_TLS_OPTIONS[@]}" -C pmchannel -n amcc -c '{"Args":["createTokenizeAsset", "KZMall", "67000", "21000", "USD-T", "1", "1DPT", "33000","Ireland"]}'

peer chaincode invoke "${TARGET_TLS_OPTIONS[@]}" -C pmchannel -n ttcc  -c '{"Args":["createAccount", "John", "0", "2000", "invest"]}'

peer chaincode invoke "${TARGET_TLS_OPTIONS[@]}" -C pmchannel -n ttcc  -c '{"Args":["buyGRET", "John", "HamHol", "500"]}'


peer chaincode query -C pmchannel -n pgacc -c '{"Args":["ReadAsset","OU"]}'

peer chaincode invoke "${TARGET_TLS_OPTIONS[@]}" -C pmchannel -n pgacc -c '{"function":"createUAinUA","Args":["OU","Prj3","Prj2"]}'

peer chaincode query -C pmchannel -n pgacc -c '{"Args":["ReadAsset","OU"]}'
'
