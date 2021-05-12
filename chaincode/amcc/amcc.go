/*
 * Copyright IBM Corp All Rights Reserved
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-protos-go/peer"
)

// SimpleAsset implements a simple chaincode to manage an asset
type AssetManagement struct {
}

// Asset describes basic details of what makes up a simple asset
type TokenizeAsset struct {
	ID            string `json:"ID"`
	GRET          string `json:"GRET"`   //to int
	Issued        string `json:"issued"` //to int
	Currency      string `json:"currency"`
	MinToken      string `json:"minToken"` //to int
	PricePerToken string `json:"pricePerToken"`
	Available     string `json:"available"` //to int
	Location      string `json:"location"`
}

// Init is called during chaincode instantiation to initialize any
// data. Note that chaincode upgrade also calls this function to reset
// or to migrate data.
func (r *AssetManagement) Init(stub shim.ChaincodeStubInterface) peer.Response {

	return shim.Success(nil)
}

// Invoke is called per transaction on the chaincode. Each transaction is
// either a 'get' or a 'set' on the asset created by Init function. The Set
// method may create a new asset by specifying a new key-value pair.
func (r *AssetManagement) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	// Extract the function and args from the transaction proposal
	fn, args := stub.GetFunctionAndParameters()

	var result string
	var err error
	//var tAsset *TokenizeAsset
	if fn == "get" {
		result, err = get(stub, args)
	} else if fn == "createTokenizeAsset" {
		err = createTokenizeAsset(stub, args[0], args[1], args[2], args[3], args[4], args[5], args[6], args[7])
	} else if fn == "readTokenizeAsset" {
		result, err = readTokenizeAsset(stub, args[0])
	} else if fn == "deleteTokenizeAsset" {
		err = deleteTokenizeAsset(stub, args[0])
	} else if fn == "buyTokenizeAsset" {
		result, err = buyTokenizeAsset(stub, args[0], args[1])
	}

	if err != nil {
		return shim.Error(err.Error())
	}

	// Return the result as success payload
	if result != "" {
		return shim.Success([]byte(result))
	}

	return shim.Success([]byte(nil))
}

// Set stores the asset (both key and value) on the ledger. If the key exists,
// it will override the value with the new one
/*func set(stub shim.ChaincodeStubInterface, args []string) (string, error) {
	if len(args) != 2 {
		return "", fmt.Errorf("Incorrect arguments. Expecting a key and a value")
	}

	err := stub.PutState(args[0], []byte(args[1]))
	if err != nil {
		return "", fmt.Errorf("Failed to set asset: %s", args[0])
	}
	return args[1], nil
}*/

// Get returns the value of the specified asset key
func get(stub shim.ChaincodeStubInterface, args []string) (string, error) {
	params := []string{"get", "OU"}
	queryArgs := make([][]byte, len(params))
	for i, arg := range params {
		queryArgs[i] = []byte(arg)
	}

	response := stub.InvokeChaincode("pgacc", queryArgs, "pmchannel")
	if response.Status != shim.OK {
		return "", fmt.Errorf("Failed to query chaincode. Got error: %s", response.Payload)
	}
	return string(response.Payload), nil
}

func createTokenizeAsset(stub shim.ChaincodeStubInterface, id, gret, issued, currency, min_token, ppt, available, location string) error {
	exists, err := assetExists(stub, id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the asset %s already exists", id)
	}

	tokenedAsset := TokenizeAsset{
		ID:            id,
		GRET:          gret,
		Issued:        issued,
		Currency:      currency,
		MinToken:      min_token,
		PricePerToken: ppt,
		Available:     available,
		Location:      location,
	}

	tokenedAssetJSON, err := json.Marshal(tokenedAsset)
	if err != nil {
		return err
	}

	return stub.PutState(id, tokenedAssetJSON)
}

func buyTokenizeAsset(stub shim.ChaincodeStubInterface, id, erc20 string) (string, error) {
	assetSliceByte, err := stub.GetState(id)
	if err != nil {
		return " ", fmt.Errorf("failed to read from world state: %v", err)
	}
	if assetSliceByte == nil {
		return " ", fmt.Errorf("the asset %s does not exist", id)
	}

	var asset TokenizeAsset
	err = json.Unmarshal(assetSliceByte, &asset)

	availableGREIT, _ := strconv.Atoi(asset.Available)
	issuedGREIT, _ := strconv.Atoi(asset.Issued)
	GREITToSell, _ := strconv.Atoi(erc20)
	if availableGREIT > 5000 {
		availableGREIT -= GREITToSell
		issuedGREIT += GREITToSell
		asset.Available = strconv.Itoa(availableGREIT)
		asset.Issued = strconv.Itoa(issuedGREIT)

		tokenedAssetJSON, err := json.Marshal(asset)
		if err != nil {
			return "", err
		}

		err = stub.PutState(id, tokenedAssetJSON)
		if err != nil {
			return "", err
		}

		return erc20, nil
	}
	return "", nil
}

func readTokenizeAsset(stub shim.ChaincodeStubInterface, id string) (string, error) {
	assetSliceByte, err := stub.GetState(id)
	if err != nil {
		return " ", fmt.Errorf("failed to read from world state: %v", err)
	}
	if assetSliceByte == nil {
		return " ", fmt.Errorf("the asset %s does not exist", id)
	}

	var asset TokenizeAsset
	err = json.Unmarshal(assetSliceByte, &asset)

	assetJSON, err := json.Marshal(asset)

	if err != nil {
		return " ", err
	}

	return string(assetJSON), nil
}

func deleteTokenizeAsset(stub shim.ChaincodeStubInterface, id string) error {
	exists, err := assetExists(stub, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the asset %s does not exist", id)
	}

	return stub.DelState(id)
}

// AssetExists returns true when asset with given ID exists in world state
func assetExists(stub shim.ChaincodeStubInterface, id string) (bool, error) {
	assetSliceByte, err := stub.GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return assetSliceByte != nil, nil
}

// main function starts up the chaincode in the container during instantiate
func main() {
	if err := shim.Start(new(AssetManagement)); err != nil {
		fmt.Printf("Error starting AssetManagement chaincode: %s", err)
	}
}
