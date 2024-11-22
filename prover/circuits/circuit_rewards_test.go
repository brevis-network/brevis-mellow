package circuits

import (
	"math/big"
	"testing"

	"github.com/brevis-network/brevis-sdk/sdk"
	"github.com/brevis-network/brevis-sdk/test"

	"github.com/ethereum/go-ethereum/common"
)

func TestCircuit(t *testing.T) {
	rpc := "https://mainnet.infura.io/v3/9aa3d95b3bc440fa88ea12eaa4456161"
	localDir := "$HOME/circuitOut/myBrevisApp"
	app, err := sdk.NewBrevisApp(1, rpc, localDir)
	check(err)

	// Add withdrawal
	withdrawReceipt := sdk.ReceiptData{
		TxHash: common.HexToHash("0xd8258abf40ad1a21520df1c30b88a92f219064705aa4ef6183badb8f14861450"),
		Fields: []sdk.LogFieldData{
			{
				LogPos:     1,
				IsTopic:    true,
				FieldIndex: 1,
			},
			{
				LogPos:     1,
				IsTopic:    false,
				FieldIndex: 2,
			},
		},
	}
	app.AddReceipt(withdrawReceipt)

	// Add Deposit
	depositReceipt := sdk.ReceiptData{
		TxHash: common.HexToHash("0xbcbe1dff6837ddb97ce702a195245bbb11d119539dfc198fc7c5f1200d3202c7"),
		Fields: []sdk.LogFieldData{
			{
				LogPos:     12,
				IsTopic:    true,
				FieldIndex: 1,
			},
			{
				LogPos:     12,
				IsTopic:    false,
				FieldIndex: 1,
			},
		},
	}
	app.AddReceipt(depositReceipt)

	account0Slot := sdk.StorageData{
		BlockNum: big.NewInt(21230700),
		Address:  common.HexToAddress("0x8c9532a60E0E7C6BbD2B2c1303F63aCE1c3E9811"),
		Slot:     common.HexToHash("0x57afd083d91aa1b80d9941137e5acdccef8478a196264b2a22d0c64076fa967d"),
	}
	app.AddStorage(account0Slot)

	account1Slot := sdk.StorageData{
		BlockNum: big.NewInt(21230700),
		Address:  common.HexToAddress("0x8c9532a60E0E7C6BbD2B2c1303F63aCE1c3E9811"),
		Slot:     common.HexToHash("0x9f95c9b305e461caf70860497f662097e2be9e5a28dda747b9f070021a23af13"),
	}
	app.AddStorage(account1Slot)

	accounts0 := "0x2221B43E989eBf213D19C6a3649DB38255b60419"
	accounts1 := "0xBc3a058D1c919f6b1F48E8846246D04D467902c8"
	var accounts [NumHolders]sdk.Uint248
	accounts[0] = sdk.ConstUint248(accounts0)
	accounts[1] = sdk.ConstUint248(accounts1)
	for i := 2; i < NumHolders; i++ {
		accounts[i] = sdk.ConstUint248(0)
	}
	appCircuit := &MellowRewardsCircuit{
		Accounts:    accounts,
		StartBlkNum: sdk.ConstUint32(21230700),
		EndBlkNum:   sdk.ConstUint32(21231099),
	}
	appCircuitAssignment := &MellowRewardsCircuit{
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
	test.ProverSucceeded(t, appCircuit, appCircuitAssignment, circuitInput)
}
