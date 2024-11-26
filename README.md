# MellowPoC

## Circuits

To calculate holder's points based on pzETH's USDC value, there are two circuits involved.

The first is [MellowRewardsCircuit](./prover/circuits/circuit_rewards.go), which uses holder's initial pzETH amount and Deposit/Withdraw receipts to calculate time-weighted pzETH amount. Each receipts will be used to add/sub holders' pzETH amount iff event emitted with holder's address.

The [MellowRewards2Circuit](./prover/circuits/circuit_rewards2.go) is the simplified implementation for the first one. It will be used for holders who <b>didn't</b> submit deposit transaction nor withdrawal request. On such condition, we assume that pzETH amount will not change during this period of time because `point accumulation rate will change upon Deposit or WithdrawalRequested event.`

For simplicity, [pzETH's USDC value](https://github.com/brevis-network/brevis-mellow/blob/2c1935b725230841d1aa13cd4d11b667da490009/prover/circuits/circuit_rewards2.go#L56) is hardcoded inside the Demo Circuit and it will be replaced with on-chain oracle's storage value.


## Contracts

The [contract](./contracts/contracts/MellowHolderReward.sol) logic is so straightforward that it receives BrevisRequest contract callback with each holder's point and emits events after decoding the callback data.