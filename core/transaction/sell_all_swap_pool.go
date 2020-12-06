package transaction

import (
	"encoding/hex"
	"fmt"
	"github.com/MinterTeam/minter-go-node/core/code"
	"github.com/MinterTeam/minter-go-node/core/commissions"
	"github.com/MinterTeam/minter-go-node/core/state"
	"github.com/MinterTeam/minter-go-node/core/types"
	"github.com/MinterTeam/minter-go-node/formula"
	"github.com/tendermint/tendermint/libs/kv"
	"math/big"
)

type SellAllSwapPool struct {
	CoinToSell        types.CoinID
	CoinToBuy         types.CoinID
	MinimumValueToBuy *big.Int
}

func (data SellAllSwapPool) basicCheck(tx *Transaction, context *state.CheckState) *Response {
	if data.CoinToBuy == data.CoinToSell {
		return &Response{
			Code: 999,
			Log:  "identical coin",
			// Info: EncodeError(),
		}
	}

	if !context.Swap().SwapPoolExist(data.CoinToSell, data.CoinToBuy) {
		return &Response{
			Code: 999,
			Log:  "swap pool not found",
			// Info: EncodeError(),
		}
	}
	// todo
	// if err := context.Swap().CheckSwap(data.CoinToSell, data.CoinToBuy, data.ValueToSell, data.MinimumValueToBuy); err != nil {
	// 	return &Response{
	// 		Code: 999,
	// 		Log:  err.Error(),
	// 		// Info: EncodeError(),
	// 	}
	// }
	return nil
}

func (data SellAllSwapPool) String() string {
	return fmt.Sprintf("EXCHANGE SWAP POOL")
}

func (data SellAllSwapPool) Gas() int64 {
	return commissions.ConvertTx
}

func (data SellAllSwapPool) Run(tx *Transaction, context state.Interface, rewardPool *big.Int, currentBlock uint64) Response {
	sender, _ := tx.Sender()

	var checkState *state.CheckState
	var isCheck bool
	if checkState, isCheck = context.(*state.CheckState); !isCheck {
		checkState = state.NewCheckState(context.(*state.State))
	}

	response := data.basicCheck(tx, checkState)
	if response != nil {
		return *response
	}

	commissionInBaseCoin := tx.CommissionInBaseCoin()
	commission := big.NewInt(0).Set(commissionInBaseCoin)

	gasCoin := checkState.Coins().GetCoin(tx.GasCoin)

	if !tx.GasCoin.IsBaseCoin() {
		errResp := CheckReserveUnderflow(gasCoin, commissionInBaseCoin)
		if errResp != nil {
			return *errResp
		}

		commission = formula.CalculateSaleAmount(gasCoin.Volume(), gasCoin.Reserve(), gasCoin.Crr(), commissionInBaseCoin)
	}

	// amount0 := new(big.Int).Set(data.ValueToSell)
	// if tx.GasCoin == data.CoinToSell {
	// 	amount0.Add(amount0, commission)
	// }
	// if checkState.Accounts().GetBalance(sender, data.CoinToSell).Cmp(amount0) == -1 {
	// 	return Response{Code: code.InsufficientFunds} // todo
	// }

	if checkState.Accounts().GetBalance(sender, tx.GasCoin).Cmp(commission) < 0 {
		return Response{
			Code: code.InsufficientFunds,
			Log:  fmt.Sprintf("Insufficient funds for sender account: %s. Wanted %s %s", sender.String(), commission.String(), gasCoin.GetFullSymbol()),
			Info: EncodeError(code.NewInsufficientFunds(sender.String(), commission.String(), gasCoin.GetFullSymbol(), gasCoin.ID().String())),
		}
	}

	if deliverState, ok := context.(*state.State); ok {
		// amountIn, amountOut := deliverState.Swap.PairSell(data.CoinToSell, data.CoinToBuy, data.ValueToSell, data.MinimumValueToBuy)
		// deliverState.Accounts.SubBalance(sender, data.CoinToSell, amountIn)
		// deliverState.Accounts.AddBalance(sender, data.CoinToBuy, amountOut)
		// todo
		rewardPool.Add(rewardPool, commissionInBaseCoin)

		deliverState.Coins.SubVolume(tx.GasCoin, commission)
		deliverState.Coins.SubReserve(tx.GasCoin, commissionInBaseCoin)

		deliverState.Accounts.SubBalance(sender, tx.GasCoin, commission)

		deliverState.Accounts.SetNonce(sender, tx.Nonce)
	}

	tags := kv.Pairs{
		kv.Pair{Key: []byte("tx.type"), Value: []byte(hex.EncodeToString([]byte{byte(TypeSellSwapPool)}))},
		kv.Pair{Key: []byte("tx.from"), Value: []byte(hex.EncodeToString(sender[:]))},
	}

	return Response{
		Code:      code.OK,
		GasUsed:   tx.Gas(),
		GasWanted: tx.Gas(),
		Tags:      tags,
	}
}
