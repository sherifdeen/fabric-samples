/*
 * Copyright IBM Corp All Rights Reserved
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric-chaincode-go/pkg/cid"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-protos-go/peer"
)

// SimpleAsset implements a simple chaincode to manage an account
type TokenTransaction struct {
}

// Asset describes basic details of what makes up a simple account
type Account struct {
	ID          string `json:"ID"`
	GREIT       string `json:"GREIT"` //to int
	ERC20       string `json:"ERC20"` //to int
	AccountType string `json:"accountType"`
}

// Init is called during chaincode instantiation to initialize any
// data. Note that chaincode upgrade also calls this function to reset
// or to migrate data.
func (r *TokenTransaction) Init(stub shim.ChaincodeStubInterface) peer.Response {

	return shim.Success(nil)
}

// Invoke is called per transaction on the chaincode. Each transaction is
// either a 'get' or a 'set' on the account created by Init function. The Set
// method may create a new account by specifying a new key-value pair.
func (r *TokenTransaction) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	// Extract the function and args from the transaction proposal
	fn, args := stub.GetFunctionAndParameters()

	var result string
	var err error
	AssetStruct := &Account{}
	if fn == "get" {
		result, err = get(stub, args)
	} else if fn == "createAccount" {
		err = createAccount(stub, args[0], args[1], args[2], args[3])
	} else if fn == "readAccount" {
		//result, err = readAccount(stub, args[0])
		AssetStruct, err = readAccount(stub, args[0])
		if AssetStruct != nil {
			reqBodyBytes := new(bytes.Buffer)
			json.NewEncoder(reqBodyBytes).Encode(*AssetStruct)
			return shim.Success(reqBodyBytes.Bytes()) //converts AssetStruct to byte slice
		}
	} else if fn == "deleteAccount" {
		err = deleteAccount(stub, args[0])
	} else if fn == "buyGRET" {
		result, err = makeGREITPurchase(stub, args[0], args[1], args[2])
	}

	if err != nil {
		return shim.Error(err.Error())
	}

	// Return the result as success payload
	if result != "" {
		return shim.Success([]byte(result))
	}

	return shim.Success([]byte(nil)) //Success([]byte(nil))
}

// Set stores the account (both key and value) on the ledger. If the key exists,
// it will override the value with the new one
/*func set(stub shim.ChaincodeStubInterface, args []string) (string, error) {
	if len(args) != 2 {
		return "", fmt.Errorf("Incorrect arguments. Expecting a key and a value")
	}

	err := stub.PutState(args[0], []byte(args[1]))
	if err != nil {
		return "", fmt.Errorf("Failed to set account: %s", args[0])
	}
	return args[1], nil
}*/

func invokeAccesssDecisionChaincode(stub shim.ChaincodeStubInterface, invokeParams []string) (bool, error) {
	invokeArgs := make([][]byte, len(invokeParams))
	for i, arg := range invokeParams {
		invokeArgs[i] = []byte(arg)
	}

	invokeResponse := stub.InvokeChaincode("pgacc", invokeArgs, "pmchannel")
	if invokeResponse.Status != shim.OK {
		return false, fmt.Errorf("Failed to complete access decision chaincode invocation. Got error: %s", invokeResponse.Payload)
	}

	isInvocationAuthorized := string(invokeResponse.Payload)

	/*if err != nil {
		return false, fmt.Errorf("Failed to convert invocation response to boolean. Got error: %s", err)
	}*/

	if isInvocationAuthorized == "allow" {
		return true, nil
	}

	return false, fmt.Errorf("Chaincode invocation failed to grant access. Got response: %s", isInvocationAuthorized)
}

// Get returns the value of the specified account key
func get(stub shim.ChaincodeStubInterface, args []string) (string, error) {
	params := []string{"get", "OU"}
	queryArgs := make([][]byte, len(params))
	for i, arg := range params {
		queryArgs[i] = []byte(arg)
	}

	response := stub.InvokeChaincode("pgacc", queryArgs, "mychannel")
	if response.Status != shim.OK {
		return "", fmt.Errorf("Failed to query chaincode. Got error: %s", response.Payload)
	}
	return string(response.Payload), nil
}

func createAccount(stub shim.ChaincodeStubInterface, id, gret, erc20, accountType string) error {
	exists, err := accountExists(stub, id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the account %s already exists", id)
	}

	tokenedAsset := Account{
		ID:          id,
		GREIT:       gret,
		ERC20:       erc20,
		AccountType: accountType,
	}

	tokenedAssetJSON, err := json.Marshal(tokenedAsset)
	if err != nil {
		return err
	}

	return stub.PutState(id, tokenedAssetJSON)
}

func readAccount(stub shim.ChaincodeStubInterface, assetID string) (*Account, error) {
	//func readAccount(stub shim.ChaincodeStubInterface, assetID string) (string, error) {
	x509, _ := cid.GetX509Certificate(stub)
	source := x509.Subject.CommonName

	readAssetParams := []string{"isTokenExchangeAuthorized", source, "ReadAsset", assetID, "gREITAccess"}

	canReadAsset, err := invokeAccesssDecisionChaincode(stub, readAssetParams)

	if err != nil {
		return nil, fmt.Errorf("Quary of chaincode failed. %w", err)
	}

	if canReadAsset {
		//return "Access authorized", nil
		accountSliceByte, err := stub.GetState(assetID)
		if err != nil {
			return nil, fmt.Errorf("failed to read from world state: %v", err)
		}
		if accountSliceByte == nil {
			return nil, fmt.Errorf("the account %s does not exist", assetID)
		}

		var account Account
		err = json.Unmarshal(accountSliceByte, &account)

		if err != nil {
			return nil, err
		}

		return &account, nil
	} else {
		return nil, fmt.Errorf("%s not authorized to read %s", source, assetID)
	}

}

func deleteAccount(stub shim.ChaincodeStubInterface, id string) error {
	exists, err := accountExists(stub, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the account %s does not exist", id)
	}

	return stub.DelState(id)
}

func makeGREITPurchase(stub shim.ChaincodeStubInterface, acctID, assetID, erc20 string) (string, error) {
	x509, _ := cid.GetX509Certificate(stub)
	source := x509.Subject.CommonName

	ttParams := []string{"isTokenExchangeAuthorized", source, "exchangeToken", acctID, "gREITAccess"}

	canExchangeToken, err := invokeAccesssDecisionChaincode(stub, ttParams)

	if err != nil {
		return "", fmt.Errorf("Chaincode invocation error %w", err)
	}

	/*tExchangeArgs := make([][]byte, len(ttParams))
	for i, arg := range tExchangeArgs {
		tExchangeArgs[i] = []byte(arg)
	}

	tExchangeResponse := stub.InvokeChaincode("pgacc", tExchangeArgs, "pmchannel")
	if tExchangeResponse.Status != shim.OK {
		return "", fmt.Errorf("Failed to complete the purchase of GREIT token. Got error: %s", tExchangeResponse.Payload)
	}

	canExchangeToken, err := strconv.ParseBool(string(tExchangeResponse.Payload))*/

	cciParams := []string{"isBuyTokenizeAssetAuthorized", source, "buyAssetToken", assetID, "gREITAccess"}
	/*buyTrnxArgs := make([][]byte, len(cciParams))
	for i, arg := range cciParams {
		buyTrnxArgs[i] = []byte(arg)
	}

	buyTrnxResponse := stub.InvokeChaincode("pgacc", buyTrnxArgs, "pmchannel")
	if buyTrnxResponse.Status != shim.OK {
		return "", fmt.Errorf("Failed to complete the purchase of GREIT token. Got error: %s", buyTrnxResponse.Payload)
	}

	canBuyGRET, err := strconv.ParseBool(string(buyTrnxResponse.Payload))*/
	canBuyGRET, err := invokeAccesssDecisionChaincode(stub, cciParams)

	if err != nil {
		return "", fmt.Errorf("Chaincode invocation error %w", err)
	}

	if canBuyGRET && canExchangeToken {
		//return "Access authorized", nil
		account, err := readAccount(stub, acctID)

		if err != nil {
			return "", err
		}
		//i, _ := strconv.Atoi(v)
		wantGreit, _ := strconv.Atoi(erc20)
		erc20InAccount, _ := strconv.Atoi(account.ERC20)
		GREITInAccount, _ := strconv.Atoi(account.GREIT)

		if erc20InAccount < wantGreit {
			return "", fmt.Errorf("the erc20 in your account %d is less than the greit token %d you want", erc20InAccount, wantGreit)
		}

		params := []string{"buyTokenizeAsset", assetID, erc20}
		queryArgs := make([][]byte, len(params))
		for i, arg := range params {
			queryArgs[i] = []byte(arg)
		}

		response := stub.InvokeChaincode("amcc", queryArgs, "pmchannel")
		if response.Status != shim.OK {
			return "", fmt.Errorf("Failed to complete the purchase of GREIT token. Got error: %s", response.Payload)
		}
		GREITPurchased, _ := strconv.Atoi(string(response.Payload))

		GREITInAccount += GREITPurchased
		account.GREIT = strconv.Itoa(GREITInAccount)
		erc20InAccount -= GREITPurchased

		accountJSON, err := json.Marshal(account)
		if err != nil {
			return " ", err
		}

		err = stub.PutState(account.ID, accountJSON)
		if err != nil {
			return " ", err
		}

		return strconv.Itoa(GREITInAccount), nil
	} else {
		return "", fmt.Errorf("Buy operation using %s on %s is unauthorized for %s", acctID, assetID, source)
	}

}

// AssetExists returns true when account with given ID exists in world state
func accountExists(stub shim.ChaincodeStubInterface, id string) (bool, error) {
	accountSliceByte, err := stub.GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return accountSliceByte != nil, nil
}

// main function starts up the chaincode in the container during instantiate
func main() {
	if err := shim.Start(new(TokenTransaction)); err != nil {
		fmt.Printf("Error starting TokenTransaction chaincode: %s", err)
	}
}
