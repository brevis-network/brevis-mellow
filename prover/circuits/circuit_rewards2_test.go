package circuits

import (
	"encoding/json"
	"math/big"
	"os"
	"testing"

	"github.com/brevis-network/brevis-sdk/sdk"
	"github.com/brevis-network/brevis-sdk/test"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

type Holder struct {
	H HolderInfo `json:"Holder"`
}
type HolderInfo struct {
	Address string `json:"Address"`
}

func TestRewards2Circuit(t *testing.T) {
	rpc := "https://mainnet.infura.io/v3/fe9161bb028d474f908af91b81296eba"
	localDir := "$HOME/circuitOut/myBrevisApp"
	app, err := sdk.NewBrevisApp(1, rpc, localDir)
	check(err)

	data, err := os.ReadFile("holder_address.json")
	check(err)
	var s []*Holder
	err = json.Unmarshal(data, &s)
	check(err)

	num := 32
	accountsSlot := make([]sdk.StorageData, num)
	var accounts [NumHolders]sdk.Uint248
	for i, account := range s[0:32] {
		slotPreImage := make([]byte, 64)
		address := hexutil.MustDecode(account.H.Address)
		copy(slotPreImage[12:32], address)
		accounts[i] = sdk.ConstUint248(account.H.Address)
		accountsSlot[i] = sdk.StorageData{
			BlockNum: big.NewInt(21230700),
			Address:  common.HexToAddress("0x8c9532a60E0E7C6BbD2B2c1303F63aCE1c3E9811"),
			Slot:     crypto.Keccak256Hash(slotPreImage),
		}
		app.AddStorage(accountsSlot[i])
	}

	appCircuit := &MellowRewards2Circuit{
		Accounts:    accounts,
		StartBlkNum: sdk.ConstUint32(21230700),
		EndBlkNum:   sdk.ConstUint32(21231099),
	}
	appCircuitAssignment := &MellowRewards2Circuit{
		Accounts:    accounts,
		StartBlkNum: sdk.ConstUint32(21230700),
		EndBlkNum:   sdk.ConstUint32(21231099),
	}

	circuitInput, err := app.BuildCircuitInput(appCircuit)
	check(err)

	// ///////////////////////////////////////////////////////////////////////////////
	// // Testing
	// ///////////////////////////////////////////////////////////////////////////////
	test.IsSolved(t, appCircuit, appCircuitAssignment, circuitInput)
	// test.ProverSucceeded(t, appCircuit, appCircuitAssignment, circuitInput)
}
