package circuits

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"testing"

	"github.com/brevis-network/brevis-sdk/sdk"
	"github.com/brevis-network/brevis-sdk/sdk/proto/gwproto"
	"github.com/brevis-network/brevis-sdk/sdk/proto/sdkproto"
	"github.com/brevis-network/brevis-sdk/test"
	"google.golang.org/grpc"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

func TestE2EWholeSetupRewards2Circuit(t *testing.T) {
	rpc := "https://mainnet.infura.io/v3/9aa3d95b3bc440fa88ea12eaa4456161"
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

	outDir := "$HOME/circuitOut/myBrevisApp"
	srsDir := "$HOME/kzgsrs"

	compiledCircuit, pk, vk, _, err := sdk.Compile(&MellowRewards2Circuit{}, outDir, srsDir)
	check(err)

	check(err)
	witness, publicWitness, err := sdk.NewFullWitness(appCircuitAssignment, circuitInput)
	check(err)
	proof, err := sdk.Prove(compiledCircuit, pk, witness)
	check(err)

	// Test verifying the proof we just generated
	err = sdk.Verify(vk, publicWitness, proof)
	check(err)

	fmt.Println(">> Initiating Brevis request")
	appContract := common.HexToAddress("0x9fc16c4918a4d69d885f2ea792048f13782a522d")
	refundee := common.HexToAddress("0x1bF81EA1F2F6Afde216cD3210070936401A14Bd4")

	calldata, requestId, _, feeValue, err := app.PrepareRequest(vk, witness, 1, 11155111, refundee, appContract, 400000, gwproto.QueryOption_ZK_MODE.Enum(), "")
	fmt.Printf("calldata %x\n", calldata)
	fmt.Printf("feeValue %d\n", feeValue)
	fmt.Printf("requestId %s\n", requestId)
	fmt.Println("Don't forget to make the transaction that pays the fee by calling Brevis.sendRequest")
	check(err)

	// Submit proof to Brevis
	fmt.Println(">> Submitting my proof to Brevis")
	err = app.SubmitProof(proof)
	check(err)

	// Poll Brevis gateway for query status till the final proof is submitted
	// on-chain by Brevis and your contract is called
	fmt.Println(">> Waiting for final proof generation and submission")
	submitTx, err := app.WaitFinalProofSubmitted(context.Background())
	check(err)
	fmt.Printf(">> Final proof submitted: tx hash %s\n", submitTx)
}

func TestE2EWithProverRewards2Circuit(t *testing.T) {
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

	opts := []grpc.DialOption{grpc.WithInsecure()}
	conn, err := grpc.Dial("localhost:33247", opts...)
	check(err)
	proverClient := sdkproto.NewProverClient(conn)
	req := sdkproto.ProveRequest{}
	req.Receipts = []*sdkproto.IndexedReceipt{}
	req.Receipts = append(req.Receipts, &sdkproto.IndexedReceipt{
		Index: 0,
		Data:  convertSDKReceiptToProtoReceipt(withdrawReceipt),
	})
	req.Receipts = append(req.Receipts, &sdkproto.IndexedReceipt{
		Index: 1,
		Data:  convertSDKReceiptToProtoReceipt(depositReceipt),
	})

	req.Storages = append(req.Storages, &sdkproto.IndexedStorage{
		Index: 0,
		Data:  convertSDKStorageToProtoStorage(account0Slot),
	})

	req.Storages = append(req.Storages, &sdkproto.IndexedStorage{
		Index: 1,
		Data:  convertSDKStorageToProtoStorage(account1Slot),
	})

	req.CustomInput = &sdkproto.CustomInput{
		JsonBytes: "{\"Accounts\":[\"0x2221B43E989eBf213D19C6a3649DB38255b60419\",\"0xBc3a058D1c919f6b1F48E8846246D04D467902c8\",\"0x00\",\"0x00\",\"0x00\",\"0x00\",\"0x00\",\"0x00\",\"0x00\",\"0x00\",\"0x00\",\"0x00\",\"0x00\",\"0x00\",\"0x00\",\"0x00\",\"0x00\",\"0x00\",\"0x00\",\"0x00\",\"0x00\",\"0x00\",\"0x00\",\"0x00\",\"0x00\",\"0x00\",\"0x00\",\"0x00\",\"0x00\",\"0x00\",\"0x00\",\"0x00\"],\"StartBlkNum\":21230700,\"EndBlkNum\":21231099}",
	}
	response, err := proverClient.Prove(context.Background(), &req)
	check(err)

	gc, err := sdk.NewGatewayClient()
	check(err)

	preq := &gwproto.PrepareQueryRequest{
		ChainId:       1,
		TargetChainId: 11155111,
		ReceiptInfos: []*gwproto.ReceiptInfo{
			convertSDKReceiptToGWReceipt(withdrawReceipt),
			convertSDKReceiptToGWReceipt(depositReceipt),
		},
		StorageQueryInfos: []*gwproto.StorageQueryInfo{
			convertSDKStorageToGWStorage(account0Slot),
			convertSDKStorageToGWStorage(account1Slot),
		},
		AppCircuitInfo: response.CircuitInfo,
	}

	pres, err := gc.PrepareQuery(preq)
	check(err)

	_, err = gc.SubmitProof(&gwproto.SubmitAppCircuitProofRequest{
		QueryKey:      pres.QueryKey,
		TargetChainId: 11155111,
		Proof:         response.Proof,
	})
	check(err)
}
