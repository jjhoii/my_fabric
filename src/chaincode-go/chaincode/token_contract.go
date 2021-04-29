package chaincode

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for transferring tokens between accounts
type SmartContract struct {
	contractapi.Contract
}

// event provides an organized struct for emitting events
type event struct {
	From  string `json:"from"`
	To    string `json:"to"`
	Value int    `json:"value"`
}

type User struct {
	ID      string `json:"ID"`
	Type    string `json:"type"`
	Balance int    `json:"balance"`
}

type Transaction struct {
	TXID  string `json:"TXID"`
	From  string `json:"from"`
	To    string `json:"to"`
	Value int    `json:"value"`
}

// InitLedger adds a base set of assets to the ledger
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	assets := []User{
		{ID: "TestUser", Type: "user", Balance: 100000},
		{ID: "TestSeller", Type: "seller", Balance: 0},
	}

	for _, asset := range assets {
		assetJSON, err := json.Marshal(asset)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(asset.ID, assetJSON)
		if err != nil {
			return fmt.Errorf("failed to put to world state. %v", err)
		}
	}

	return nil
}

// BalanceOf returns the balance of the given account
func (s *SmartContract) BalanceOf(ctx contractapi.TransactionContextInterface, id string) (int, error) {
	user, err := s.GetUser(ctx, id)
	if err != nil {
		return 0, fmt.Errorf("user id %s does not exist", id)
	}

	balance := user.Balance

	return balance, nil
}

// TransferFrom transfers the value amount from the "from" address to the "to" address
// This function triggers a Transfer event
func (s *SmartContract) TransferFrom(ctx contractapi.TransactionContextInterface, from string, to string, value int) error {

	// Initiate the transfer
	err := s.transferHelper(ctx, from, to, value)
	if err != nil {
		return fmt.Errorf("failed to transfer: %v", err)
	}

	// Emit the Transfer event
	err = s.SetEvent(ctx, "Transfer", event{from, to, value})
	if err != nil {
		return err
	}

	log.Printf("%s transfer %d balance to %s", from, value, to)

	return nil
}

// Helper Functions

// transferHelper is a helper function that transfers tokens from the "from" address to the "to" address
// Dependant functions include Transfer and TransferFrom
func (s *SmartContract) transferHelper(ctx contractapi.TransactionContextInterface, from string, to string, value int) error {

	if from == to {
		return fmt.Errorf("cannot transfer to and from same client account")
	}

	if value < 0 { // transfer of 0 is allowed in ERC-20, so just validate against negative amounts
		return fmt.Errorf("transfer amount cannot be negative")
	}

	fromUser, err := s.GetUser(ctx, from)
	if err != nil {
		return err
	}

	toUser, err := s.GetUser(ctx, to)
	if err != nil {
		return err
	}

	if fromUser.Balance < value {
		return fmt.Errorf("user balance lower than %d", value)
	}

	beforeFromUserBalance := fromUser.Balance
	beforeToUserBalance := toUser.Balance
	fromUser.Balance -= value
	toUser.Balance += value

	//update
	fromUserJSON, err := json.Marshal(fromUser)
	if err != nil {
		return err
	}

	toUserJSON, err := json.Marshal(toUser)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(from, fromUserJSON)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(to, toUserJSON)
	if err != nil {
		return err
	}

	log.Printf("client %s balance updated from %d to %d", from, beforeFromUserBalance, fromUser.Balance)
	log.Printf("recipient %s balance updated from %d to %d", to, beforeToUserBalance, toUser.Balance)

	return nil
}

func (s *SmartContract) SetEvent(ctx contractapi.TransactionContextInterface, eventName string, e event) error {
	// Emit the Transfer event
	transferEvent := e
	transferEventJSON, err := json.Marshal(transferEvent)
	if err != nil {
		return fmt.Errorf("failed to obtain JSON encoding: %v", err)
	}
	err = ctx.GetStub().SetEvent("Transfer", transferEventJSON)
	if err != nil {
		return fmt.Errorf("failed to set event: %v", err)
	}

	return nil
}

func (s *SmartContract) GetUser(ctx contractapi.TransactionContextInterface, id string) (*User, error) {
	// do something
	transactionJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}

	var user User
	err = json.Unmarshal(transactionJSON, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *SmartContract) UserExist(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	userJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return userJSON != nil, nil
}

func (s *SmartContract) GetTransaction(ctx contractapi.TransactionContextInterface, txid string) (*Transaction, error) {
	transactionJSON, err := ctx.GetStub().GetState(txid)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}

	var transaction Transaction
	err = json.Unmarshal(transactionJSON, &transaction)
	if err != nil {
		return nil, err
	}
	return &transaction, nil
}

func (s *SmartContract) SetTransaction(ctx contractapi.TransactionContextInterface, from string, to string, balance int) (*Transaction, error) {
	txid := ctx.GetStub().GetTxID()
	transaction := Transaction{TXID: txid, From: from, To: to, Value: balance}
	transactionJSON, err := json.Marshal(transaction)
	if err != nil {
		return nil, err
	}

	err = ctx.GetStub().PutState(txid, transactionJSON)

	return &transaction, err
}
