package circuits

import (
	"github.com/brevis-network/brevis-sdk/sdk"
)

const (
	NumHolders      = 32
	MaxReceipts     = 32
	MaxStorage      = NumHolders
	MaxTransactions = 0
	BlockRange      = 300
)

var PzETHToken = sdk.ConstUint248("0x8c9532a60E0E7C6BbD2B2c1303F63aCE1c3E9811")

type AppCircuit struct {
	Accounts    [NumHolders]sdk.Uint248 // Holders' addresses
	StartBlkNum sdk.Uint32              // Start block number
	EndBlkNum   sdk.Uint32              // End block number
}

var _ sdk.AppCircuit = &AppCircuit{}

func (c *AppCircuit) Allocate() (maxReceipts, maxStorage, maxTransactions int) {
	// Our app is only ever going to use one storage data at a time so
	// we can simply limit the max number of data for storage to 1 and
	// 0 for all others
	return 32, MaxStorage, 0
}

func (c *AppCircuit) Define(api *sdk.CircuitAPI, in sdk.DataInput) error {
	api.AssertInputsAreUnique()
	uint32 := api.Uint32
	uint248 := api.Uint248
	zeroB32 := api.ToBytes32(sdk.ConstUint248(0))
	uint32.IsEqual(uint32.Add(c.StartBlkNum, sdk.ConstUint32(BlockRange-1)), c.EndBlkNum)

	receipts := sdk.NewDataStream(api, in.Receipts)

	amounts := make([]sdk.Uint248, NumHolders)
	blks := make([]sdk.Uint32, NumHolders)
	accumulatedResult := make([]sdk.Uint248, NumHolders)
	for i := range accumulatedResult {
		accumulatedResult[i] = sdk.ConstUint248(0)
	}

	for i, account := range c.Accounts {
		invalidAccount := uint248.IsEqual(account, sdk.ConstUint248(0))
		accountSlot := in.StorageSlots.Raw[i]
		amounts[i] = api.ToUint248(accountSlot.Value)
		blks[i] = accountSlot.BlockNum
		slot := api.Keccak256([]sdk.Bytes32{api.ToBytes32(account), zeroB32}, []int32{256, 256})
		validData := uint248.And(
			api.ToUint248(uint32.IsEqual(c.StartBlkNum, accountSlot.BlockNum)), // correct block number
			api.Bytes32.IsEqual(accountSlot.Slot, slot),
			api.Uint248.IsEqual(accountSlot.Contract, PzETHToken),
		)
		valid := uint248.Or(invalidAccount, validData)
		uint248.AssertIsEqual(valid, sdk.ConstUint248(1))
	}

	sdk.AssertSorted(receipts, func(a, b sdk.Receipt) sdk.Uint248 {
		return api.ToUint248(uint32.Or(uint32.IsLessThan(a.BlockNum, b.BlockNum), uint32.IsEqual(a.BlockNum, b.BlockNum)))
	})

	for _, receipt := range in.Receipts.Raw {
		invalidReceipt := uint32.IsZero(receipt.BlockNum)
		deposit := uint248.And(
			uint248.IsEqual(receipt.Fields[0].Contract, PzETHToken),
			uint248.IsEqual(receipt.Fields[0].IsTopic, sdk.ConstUint248(1)),
			uint248.IsEqual(receipt.Fields[0].Index, sdk.ConstUint248(1)),
			api.ToUint248(uint32.IsEqual(receipt.Fields[0].LogPos, receipt.Fields[1].LogPos)),
			uint248.IsEqual(receipt.Fields[1].IsTopic, sdk.ConstUint248(0)),
			uint248.IsEqual(receipt.Fields[1].Index, sdk.ConstUint248(1)),
		)
		withdrawal := uint248.And(
			uint248.IsEqual(receipt.Fields[0].Contract, PzETHToken),
			uint248.IsEqual(receipt.Fields[0].IsTopic, sdk.ConstUint248(1)),
			uint248.IsEqual(receipt.Fields[0].Index, sdk.ConstUint248(1)),
			api.ToUint248(uint32.IsEqual(receipt.Fields[0].LogPos, receipt.Fields[1].LogPos)),
			uint248.IsEqual(receipt.Fields[1].IsTopic, sdk.ConstUint248(0)),
			uint248.IsEqual(receipt.Fields[1].Index, sdk.ConstUint248(2)),
		)
		uint248.AssertIsEqual(uint248.Or(deposit, withdrawal, api.ToUint248(invalidReceipt)), sdk.ConstUint248(1))

		holderAddress := api.ToUint248(receipt.Fields[0].Value)
		changingAmount := api.ToUint248(receipt.Fields[1].Value)
		for j, account := range c.Accounts {
			diff := uint248.Mul(uint248.IsEqual(holderAddress, account), changingAmount)
			blockRange := uint32.Sub(receipt.BlockNum, blks[j])
			currentAmount := amounts[j]
			newAmount := uint248.Select(deposit, uint248.Add(amounts[j], diff), uint248.Select(withdrawal, uint248.Select(uint248.IsGreaterThan(amounts[j], diff), uint248.Sub(amounts[j], diff), sdk.ConstUint248(0)), sdk.ConstUint248(0)))
			accumulatedResult[j] = uint248.Add(accumulatedResult[j], uint248.Mul(api.ToUint248(blockRange), currentAmount))
			blks[j] = receipt.BlockNum
			amounts[j] = newAmount
		}
	}

	api.OutputUint(64, api.ToUint248(c.StartBlkNum))
	api.OutputUint(64, api.ToUint248(c.EndBlkNum))

	for i, account := range c.Accounts {
		api.OutputAddress(account)
		q, _ := uint248.Div(accumulatedResult[i], sdk.ConstUint248(BlockRange*4000))
		api.OutputUint(248, q)
	}
	return nil
}
