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
	//"strconv"

	"github.com/hyperledger/fabric-chaincode-go/pkg/cid"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-protos-go/peer"
)

type PolicyGraphAsset struct {
}

type Node struct {
	Name  string            `json:"name"`
	Value int               `json:"value"`
	Props map[string]string `json:"props"`
}

type Graph struct {
	ID           string                `json:"id"`
	PE           map[string]Node       `json:"pe"`
	Assignments  []map[string][]Node   `json:"assignments"`
	Associations []map[string][]string `json:"associations"`
}

type assocIndexAndHead struct {
	assocIdx    int
	assocMapKey string
}

func (g *Graph) new() *Graph {
	init := 0
	gg := Graph{
		PE:           make(map[string]Node, init),
		Assignments:  make([]map[string][]Node, init),
		Associations: make([]map[string][]string, init),
	}
	return &gg
}

func (g *Graph) createPolicyClass(pc string) {
	pcNode := Node{Name: pc, Value: len(g.PE), Props: map[string]string{"type": "pc"}}
	g.PE[pcNode.Name] = pcNode
	g.ID = pc
}

func (g *Graph) createAssignment(v, w Node) {
	/*currentLength := len(g.Assignments)
	expectedLength := v.Value + 1
	if expectedLength > currentLength {
		n := expectedLength - currentLength
		for i := 0; i <= n; i++ {
			g.Assignments = append(g.Assignments, make(map[string][]Node))
		}
	}*/

	if _, ok := g.PE[v.Name]; !ok {
		g.Assignments = append(g.Assignments, make(map[string][]Node))
		v.Value = len(g.PE) - 1
		g.PE[v.Name] = v
	}

	if g.Assignments[v.Value] == nil {
		g.Assignments[v.Value] = map[string][]Node{v.Name: {w}}
	} else {
		g.Assignments[v.Value][v.Name] = append(g.Assignments[v.Value][v.Name], w)
	}

	/*if _, ok := g.PE[v.Name]; !ok {
		g.PE[v.Name] = v
	}*/
}

func (g *Graph) createAssociation(pre, suc Node, ars []string) {
	_, preOK := g.PE[pre.Name]
	_, sucOK := g.PE[suc.Name]
	if preOK && sucOK {
		currentLength := len(g.Associations)
		expectedLength := pre.Value + 1
		if expectedLength > currentLength {
			n := expectedLength - currentLength
			for i := 0; i <= n; i++ {
				g.Associations = append(g.Associations, make(map[string][]string))
			}
		}
	}

	if g.Associations[pre.Value] == nil {
		g.Associations[pre.Value] = map[string][]string{suc.Name: ars}
	} else {
		g.Associations[pre.Value][suc.Name] = append(g.Associations[pre.Value][suc.Name], ars...)
	}
}

func (g *Graph) isReachable(s, d Node) bool {
	if s.Value == d.Value {
		return true
	}
	length := len(g.PE)
	visited := make([]bool, length)
	queue := []int{}
	var n int

	queue = append(queue, s.Value)
	visited[s.Value] = true

	for len(queue) > 0 {

		if len(queue) > 1 {
			n, queue = queue[0], queue[1:]
			//fmt.Printf("if: n = %d, queue = %v, visited = %v\n",n, queue, visited)
		} else {
			n, queue = queue[0], []int{}
			//fmt.Printf("2**%d = %d\n", i, &v)
			//fmt.Printf("else: n = %d, queue = %v, visited = %v\n",n, queue, visited)
		}

		if n == d.Value {
			return true
		}

		//fmt.Printf("node %d: edge = %v\n", n, g.Assignments[n])

		var nodeName string

		for k, _ := range g.Assignments[n] {
			nodeName = k
		}

		for i := range g.Assignments[n][nodeName] {
			nodeValue := g.Assignments[n][nodeName][i].Value
			if visited[nodeValue] == false {
				queue = append(queue, nodeValue)
				visited[nodeValue] = true
				//fmt.Println("from loop:", g.Assignments[s][k]["assign"])
			}
		}

	}
	return false
}

func (g *Graph) getAssociationsToTarget(target Node) []assocIndexAndHead {

	assocRecord := []assocIndexAndHead{}

	for i := 0; i < len(g.Associations); i++ {

		for k, _ := range g.Associations[i] {
			//fmt.Printf("From getAssociationsToTarget checking path from %s to %s\n", target.Name, g.PE[k].Name)
			if g.isReachable(target, g.PE[k]) {
				fmt.Printf("From getAssociationsToTarget path from %s to %s found\n", target.Name, g.PE[k].Name)
				assocRecord = append(assocRecord, assocIndexAndHead{assocIdx: i, assocMapKey: k})
				//return i, k
			}
		}
	}
	//fmt.Printf("From getAssociationsToTarget no path from %s to any association head node\n",target.Name)
	//return -1, " "
	return assocRecord
}

func (g *Graph) isOperationInAccessRights(assocIndx int, assocMapKey string, opRequest string) bool {
	for _, ar := range g.Associations[assocIndx][assocMapKey] {
		fmt.Printf("From isOperationInAccessRights access right = %s\n", ar)
		if ar == opRequest {
			fmt.Printf("From isOperationInAccessRights access right %s matches %s\n", opRequest, ar)
			return true
		}
	}
	fmt.Printf("From isOperationInAccessRights no matching access right for %s\n", opRequest)
	return false
}

func (g *Graph) isSourcePermitted(source Node, assocIndx int) bool {
	for _, v := range g.PE {
		fmt.Printf("From isSourcePermitted, check path from %s to %s\n", source.Name, v.Name)
		if v.Value == assocIndx {
			if g.isReachable(source, v) {
				fmt.Printf("From isSourcePermitted, path from %s to %s found\n", source.Name, v.Name)
				return true
			}
		}
	}
	fmt.Printf("From isSourcePermitted, path not found to %s\n", source.Name)
	return false
}

func operationToPermissionMapping(operation string) string {
	opToPerm := map[string]string{
		"CreateUAinUA":        "cuaua",
		"ReadAsset":           "read",
		"buyAssetToken":       "buy",
		"exchangeToken":       "exchange",
		"AssignAssetToAUM":    "launch",
		"createTokenizeAsset": "create",
	}
	fmt.Printf("From operationToPermissionMapping, returning %s for %s\n", opToPerm[operation], operation)
	return opToPerm[operation]
}

func (g *Graph) isAuthorize(op string, source, assignTail, assignHead Node) string {
	associationsToTarget := g.getAssociationsToTarget(assignHead)
	//fmt.Printf("From isAuthorize association head = %s", assocMapKey)
	if associationsToTarget == nil {
		return "deny"
	}

	opToArs := operationToPermissionMapping(op)

	for _, anAssociation := range associationsToTarget {
		if g.isOperationInAccessRights(anAssociation.assocIdx, anAssociation.assocMapKey,
			opToArs) &&
			g.isSourcePermitted(source, anAssociation.assocIdx) {
			return "allow"
		}
	}

	return "deny"
}

// main function starts up the chaincode in the container during instantiate

func (pg *PolicyGraphAsset) Init(stub shim.ChaincodeStubInterface) peer.Response {
	return shim.Success(nil)
}

func initPolicyGraph(stub shim.ChaincodeStubInterface, pcName string) error {
	var gg Graph
	pmGraph := gg.new()
	pmGraph.createPolicyClass(pcName)
	pmGraph.createAssignment(Node{Name: "Subject", Props: map[string]string{"type": "ua"}}, pmGraph.PE[pcName])
	pmGraph.createAssignment(Node{Name: "Records", Props: map[string]string{"type": "oa"}}, pmGraph.PE[pcName])
	pmGraph.createAssignment(Node{Name: "GFM", Props: map[string]string{"type": "ua"}}, pmGraph.PE["Subject"])
	pmGraph.createAssignment(Node{Name: "GFMAdmin", Props: map[string]string{"type": "ua"}}, pmGraph.PE["GFM"])
	pmGraph.createAssignment(Node{Name: "AssetOwners", Props: map[string]string{"type": "ua"}}, pmGraph.PE["Subject"])
	pmGraph.createAssignment(Node{Name: "Amy", Props: map[string]string{"type": "ua"}}, pmGraph.PE["AssetOwners"])
	pmGraph.createAssignment(Node{Name: "Investors", Props: map[string]string{"type": "ua"}}, pmGraph.PE["Subject"])
	pmGraph.createAssignment(Node{Name: "UserID", Props: map[string]string{"type": "u"}}, pmGraph.PE["Investors"])
	pmGraph.createAssignment(Node{Name: "John", Props: map[string]string{"type": "u"}}, pmGraph.PE["UserID"])
	pmGraph.createAssignment(Node{Name: "TokenTrxn", Props: map[string]string{"type": "oa"}}, pmGraph.PE["Records"])
	pmGraph.createAssignment(Node{Name: "WalletID", Props: map[string]string{"type": "oa"}}, pmGraph.PE["TokenTrxn"])
	pmGraph.createAssignment(Node{Name: "AUM", Props: map[string]string{"type": "oa"}}, pmGraph.PE["Records"])
	pmGraph.createAssignment(Node{Name: "KZMall", Props: map[string]string{"type": "oa"}}, pmGraph.PE["AUM"])
	pmGraph.createAssociation(pmGraph.PE["Investors"], pmGraph.PE["AUM"], []string{"buy", "read"})
	pmGraph.createAssociation(pmGraph.PE["AssetOwners"], pmGraph.PE["AUM"], []string{"launch"})
	pmGraph.createAssociation(pmGraph.PE["GFM"], pmGraph.PE["Records"], []string{"create"})
	pmGraph.createAssociation(pmGraph.PE["UserID"], pmGraph.PE["WalletID"], []string{"exchange", "read"})
	/*pmGraph.createPolicyClass("OU")
	pmGraph.createAssignment(Node{Name: "Div", Value: 1, Props: map[s[
	.[['''''''''''[[tring]string{"type": "ua"}}, pmGraph.PE["OU"])
	pmGraph.createAssignment(Node{Name: "Prj", Value: 2, Props: map[string]string{"type": "ua"}}, pmGraph.PE["OU"])
	pmGraph.createAssignment(Node{Name: "Assets", Value: 8, Props: map[string]string{"type": "oa"}}, pmGraph.PE["OU"])
	pmGraph.createAssignment(Node{Name: "Grp2", Value: 3, Props: map[string]string{"type": "ua"}}, pmGraph.PE["Div"])
	pmGraph.createAssignment(Node{Name: "Grp1", Value: 4, Props: map[string]string{"type": "ua"}}, pmGraph.PE["Div"])
	pmGraph.createAssignment(Node{Name: "Prj1", Value: 5, Props: map[string]string{"type": "ua"}}, pmGraph.PE["Prj"])
	pmGraph.createAssignment(Node{Name: "Prj2", Value: 6, Props: map[string]string{"type": "ua"}}, pmGraph.PE["Prj"])
	pmGraph.createAssignment(Node{Name: "AB1", Value: 9, Props: map[string]string{"type": "oa"}}, pmGraph.PE["Assets"])
	pmGraph.createAssignment(Node{Name: "asset1", Value: 10, Props: map[string]string{"type": "o"}}, pmGraph.PE["AB1"])
	pmGraph.createAssociation(pmGraph.PE["Prj2"], pmGraph.PE["AB1"], []string{"read"})
	pmGraph.createAssignment(Node{Name: "John", Value: 7, Props: map[string]string{"type": "u"}}, pmGraph.PE["Prj2"])*/

	//g := graph.Sort(gm)

	assetJSON, err := json.Marshal(pmGraph)

	if err != nil { //shim.Error(fmt.Sprintf("Failed to serialize policy graph asset: %s", assetJSON))
		return fmt.Errorf("Failed to serialize policy graph asset")
	}

	err = stub.PutState(pmGraph.ID, assetJSON)
	if err != nil {
		return fmt.Errorf("Failed to create policy graph asset") //shim.Error(fmt.Sprintf("Failed to create policy graph asset: %s", assetJSON))
	}
	return nil
}

func (pg *PolicyGraphAsset) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	// Extract the function and args from the transaction proposal
	fn, args := stub.GetFunctionAndParameters()
	//fmt.Printf("fn = %v\n args = %v\n", fn, args)

	var pmGraph *Graph

	var err error
	//var response bool
	if fn == "get" {
		pmGraph, err = ReadAsset(stub, args[0]) // Return the result as success payload
		reqBodyBytes := new(bytes.Buffer)
		json.NewEncoder(reqBodyBytes).Encode(pmGraph)
		return shim.Success(reqBodyBytes.Bytes())
	} else if fn == "init" {
		err = initPolicyGraph(stub, args[0])
	} else if fn == "setTokenizedAsset" {
		return setTokenizedAsset(stub, args)
	} else if fn == "isBuyTokenizeAssetAuthorized" {
		return isTokenTransactionAuthorized(stub, args)
	} else if fn == "launchTokenizedAsset" {
		return launchTokenizedAsset(stub, args)
	} else if fn == "isTokenExchangeAuthorized" {
		return isTokenTransactionAuthorized(stub, args)
	} else if fn == "deleteAsset" {
		err = deleteAsset(stub, args[0])
	}

	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil) //Error(err.Error())
}

func launchTokenizedAsset(stub shim.ChaincodeStubInterface, funcArgs []string) peer.Response {
	x509, _ := cid.GetX509Certificate(stub)
	source := x509.Subject.CommonName
	op := "AssignAssetToAUM"
	assignTail := funcArgs[0]
	assignHead := funcArgs[1]
	policyID := funcArgs[2]

	g, err := ReadAsset(stub, policyID)

	if err != nil {
		return shim.Error("failed to read from world state") //, fmt.Errorf("failed to read from world state: %v", err)
	}
	response := g.isAuthorize(op, g.PE[source], Node{}, g.PE[assignHead])


	fmt.Printf("From launchTokenizedAsset isAuthorize function returns %s\n", response)

	if response == "allow" {
		//return shim.Success([]byte("Access granted"))

		tailNode := Node{Name: assignTail, Props: map[string]string{"type": "oa"}}

		g.createAssignment(tailNode, g.PE[assignHead])

		policyAssetJSON, err := json.Marshal(g)

		if err != nil { //shim.Error(fmt.Sprintf("Failed to serialize policy graph asset: %s", assetJSON))
			return shim.Error(fmt.Sprintf("Failed to serialize policy graph asset: %s", policyAssetJSON))
		}

		err = stub.PutState(g.ID, policyAssetJSON)
		if err != nil {
			return shim.Error(fmt.Sprintf("Failed to write policy graph asset to ledger: %s", policyAssetJSON))
		}
	} else {
		return shim.Error(fmt.Sprintf("%s is unauthorized to launch token asset: %s", source, assignTail))
	}
	return shim.Success(nil)	
}

func setTokenizedAsset(stub shim.ChaincodeStubInterface, funcArgs []string) peer.Response {
	x509, _ := cid.GetX509Certificate(stub)
	source := x509.Subject.CommonName
	op := "createTokenizeAsset"
	policyID := funcArgs[0]
	assignHead := funcArgs[1]
	assignTail := funcArgs[2]
	

	g, err := ReadAsset(stub, policyID)

	if err != nil {
		return shim.Error("failed to read from world state") //, fmt.Errorf("failed to read from world state: %v", err)
	}
	response := g.isAuthorize(op, g.PE[source], Node{}, g.PE[assignHead])
	fmt.Printf("From setTokenizedAsset isAuthorize function returns %s\n", response)

	if response == "allow" {
		//return shim.Success([]byte("Access granted"))
		cciArgs := funcArgs[2:]
		args := []string{"createTokenizeAsset"}
		args = append(args, cciArgs...)
		setArgs := make([][]byte, len(args))
		for i, arg := range args {
			setArgs[i] = []byte(arg)
		}

		response := stub.InvokeChaincode("amcc", setArgs, "pmchannel")
		if response.Status != shim.OK {
			//return fmt.Errorf("Failed to query chaincode. Got error: %s", response.Payload)
			return shim.Error(fmt.Sprintf("Failed to query chaincode. Got error: %s", response.Payload))
		}
		return shim.Success(nil)

	}

	return shim.Error(fmt.Sprintf("%s is unauthorized to create token asset: %s", source, assignTail))

}

/*func isTokenExchangeAuthorized(stub shim.ChaincodeStubInterface, funcArgs []string) peer.Response {

}*/

func isTokenTransactionAuthorized(stub shim.ChaincodeStubInterface, funcArgs []string) peer.Response {
	source := funcArgs[0]
	op := funcArgs[1]
	target := funcArgs[2]
	policyID := funcArgs[3]

	g, err := ReadAsset(stub, policyID)

	if err != nil {
		return shim.Error("failed to read from world state") //, fmt.Errorf("failed to read from world state: %v", err)
	}
	response := g.isAuthorize(op, g.PE[source], Node{}, g.PE[target])
	fmt.Printf("From isTokenTransactionAuthorized isAuthorize function returns %s\n", response)

	//stringREsponse := strconv.FormatBool(response)
	return shim.Success([]byte(response))

	/*reqBodyBytes := new(bytes.Buffer)
	json.NewEncoder(reqBodyBytes).Encode(response)
	return shim.Success(reqBodyBytes.Bytes()) */

}

func ReadAsset(stub shim.ChaincodeStubInterface, id string) (*Graph, error) {
	graphJSON, err := stub.GetState(id)

	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if graphJSON == nil {
		return nil, fmt.Errorf("the asset %s does not exist", graphJSON)
	}

	var pmGraph Graph
	err = json.Unmarshal(graphJSON, &pmGraph)
	if err != nil {
		return nil, err
	}

	/*
		out, _ := json.Marshal(&pmg)
	*/

	return &pmGraph, nil

}

func assetExists(stub shim.ChaincodeStubInterface, id string) (bool, error) {
	assetSliceByte, err := stub.GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return assetSliceByte != nil, nil
}

func deleteAsset(stub shim.ChaincodeStubInterface, id string) error {
	exists, err := assetExists(stub, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the asset %s does not exist", id)
	}

	return stub.DelState(id)
}

func main() {
	if err := shim.Start(new(PolicyGraphAsset)); err != nil {
		fmt.Printf("Error starting SimpleAsset chaincode: %s", err)
	}
}

// SimpleAsset implements a simple chaincode to manage an asset
/*type SimpleAsset struct {
}

// Init is called during chaincode instantiation to initialize any
// data. Note that chaincode upgrade also calls this function to reset
// or to migrate data.
func (t *SimpleAsset) Init(stub shim.ChaincodeStubInterface) peer.Response {
	// Get the args from the transaction proposal
	args := stub.GetStringArgs()
	if len(args) != 2 {
		return shim.Error("Incorrect arguments. Expecting a key and a value")
	}

	// Set up any variables or assets here by calling stub.PutState()

	// We store the key and the value on the ledger
	err := stub.PutState(args[0], []byte(args[1]))
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to create asset: %s", args[0]))
	}
	return shim.Success(nil)
}

// Invoke is called per transaction on the chaincode. Each transaction is
// either a 'get' or a 'set' on the asset created by Init function. The Set
// method may create a new asset by specifying a new key-value pair.
func (t *SimpleAsset) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	// Extract the function and args from the transaction proposal
	fn, args := stub.GetFunctionAndParameters()

	var result string
	var err error
	if fn == "set" {
		result, err = set(stub, args)
	} else { // assume 'get' even if fn is nil
		result, err = get(stub, args)
	}
	if err != nil {
		return shim.Error(err.Error())
	}

	// Return the result as success payload
	return shim.Success([]byte(result))
}

// Set stores the asset (both key and value) on the ledger. If the key exists,
// it will override the value with the new one
func set(stub shim.ChaincodeStubInterface, args []string) (string, error) {
	if len(args) != 2 {
		return "", fmt.Errorf("Incorrect arguments. Expecting a key and a value")
	}

	err := stub.PutState(args[0], []byte(args[1]))
	if err != nil {
		return "", fmt.Errorf("Failed to set asset: %s", args[0])
	}
	return args[1], nil
}

// Get returns the value of the specified asset key
func get(stub shim.ChaincodeStubInterface, args []string) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("Incorrect arguments. Expecting a key")
	}

	value, err := stub.GetState(args[0])
	if err != nil {
		return "", fmt.Errorf("Failed to get asset: %s with error: %s", args[0], err)
	}
	if value == nil {
		return "", fmt.Errorf("Asset not found: %s", args[0])
	}
	return string(value), nil
}

// main function starts up the chaincode in the container during instantiate
func main() {
	if err := shim.Start(new(SimpleAsset)); err != nil {
		fmt.Printf("Error starting SimpleAsset chaincode: %s", err)
	}
}*/
