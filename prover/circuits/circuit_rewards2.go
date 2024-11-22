package circuits

import (
	"github.com/brevis-network/brevis-sdk/sdk"
)

type MellowRewards2Circuit struct {
	Accounts    [NumHolders]sdk.Uint248 // Holders' addresses
	StartBlkNum sdk.Uint32              // Start block number
	EndBlkNum   sdk.Uint32              // End block number
}

var _ sdk.AppCircuit = &MellowRewards2Circuit{}

func (c *MellowRewards2Circuit) Allocate() (maxReceipts, maxStorage, maxTransactions int) {
	// Our app is only ever going to use one storage data at a time so
	// we can simply limit the max number of data for storage to 1 and
	// 0 for all others
	return 0, MaxStorage, 0
}

func (c *MellowRewards2Circuit) Define(api *sdk.CircuitAPI, in sdk.DataInput) error {
	api.AssertInputsAreUnique()
	uint32 := api.Uint32
	uint248 := api.Uint248
	zeroB32 := api.ToBytes32(sdk.ConstUint248(0))
	uint32.IsEqual(uint32.Add(c.StartBlkNum, sdk.ConstUint32(BlockRange-1)), c.EndBlkNum)

	blks := make([]sdk.Uint32, NumHolders)
	accumulatedResult := make([]sdk.Uint248, NumHolders)
	for i := range accumulatedResult {
		accumulatedResult[i] = sdk.ConstUint248(0)
	}

	for i, account := range c.Accounts {
		invalidAccount := uint248.IsEqual(account, sdk.ConstUint248(0))
		accountSlot := in.StorageSlots.Raw[i]
		blks[i] = accountSlot.BlockNum
		slot := api.Keccak256([]sdk.Bytes32{api.ToBytes32(account), zeroB32}, []int32{256, 256})
		validData := uint248.And(
			api.ToUint248(uint32.IsEqual(c.StartBlkNum, accountSlot.BlockNum)), // correct block number
			api.Bytes32.IsEqual(accountSlot.Slot, slot),
			api.Uint248.IsEqual(accountSlot.Contract, PzETHToken),
		)
		valid := uint248.Or(invalidAccount, validData)
		uint248.AssertIsEqual(valid, sdk.ConstUint248(1))

		accumulatedResult[i] = api.ToUint248(accountSlot.Value)
	}

	api.OutputUint(64, api.ToUint248(c.StartBlkNum))
	api.OutputUint(64, api.ToUint248(c.EndBlkNum))

	for i, account := range c.Accounts {
		api.OutputAddress(account)
		usdValue := uint248.Mul(accumulatedResult[i], sdk.ConstUint248(369038))
		q, _ := uint248.Div(usdValue, sdk.ConstUint248(4000*100))
		// Use fixed usd price 3690.38 for pzETH around c.EndBlkNum
		api.OutputUint(248, q)
	}
	return nil
}
