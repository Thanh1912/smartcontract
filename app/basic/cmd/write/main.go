package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/ardanlabs/smartcontract/app/basic/contracts/store"
	"github.com/ardanlabs/smartcontract/business/smart"
)

func main() {
	if err := run(); err != nil {
		log.Fatalln(err)
	}
}

func run() error {
	ctx := context.Background()

	client, privateKey, err := smart.Connect(smart.NetworkLocalhost)
	if err != nil {
		return err
	}

	fromAddress := crypto.PubkeyToAddress(privateKey.PublicKey)
	fmt.Println("address:", fromAddress.String())

	// =========================================================================

	contract, err := newContract(ctx, client)
	if err != nil {
		return err
	}

	version, err := contract.Version(nil)
	if err != nil {
		return err
	}
	fmt.Println("version:", version)

	// =========================================================================

	startingBalance, err := client.BalanceAt(ctx, fromAddress, nil)
	if err != nil {
		return err
	}
	defer smart.PrintBalanceDiff(ctx, startingBalance, fromAddress, client)

	// =========================================================================

	const gasLimit = 250000
	tran, err := smart.NewTransaction(ctx, gasLimit, privateKey, client)
	if err != nil {
		return err
	}

	// =========================================================================

	var key [32]byte
	var value [32]byte
	copy(key[:], []byte("name"))
	copy(value[:], []byte("brianna"))

	tx, err := contract.SetItem(tran, key, value)
	if err != nil {
		log.Fatal("SetItem ERROR:", err)
	}
	smart.PrintTransaction(tx)

	receipt, err := smart.WaitMined(ctx, tx, fromAddress, client)
	if err != nil {
		return err
	}
	smart.PrintTransactionReceipt(receipt, tx)

	return nil
}

// newContract constructs a SimpleCoin contract.
func newContract(ctx context.Context, client *ethclient.Client) (*store.Store, error) {
	data, err := os.ReadFile("contract.env")
	if err != nil {
		return nil, fmt.Errorf("readfile: %w", err)
	}
	contractID := string(data)
	fmt.Println("contractID:", contractID)

	contract, err := store.NewStore(common.HexToAddress(contractID), client)
	if err != nil {
		return nil, fmt.Errorf("NewStore: %w", err)
	}

	return contract, nil
}
