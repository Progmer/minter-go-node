package transaction

import (
	"encoding/hex"
	"fmt"
	"github.com/MinterTeam/minter-go-node/core/code"
	"github.com/MinterTeam/minter-go-node/core/state"
	"github.com/MinterTeam/minter-go-node/core/types"
	"github.com/MinterTeam/minter-go-node/formula"
	"github.com/MinterTeam/minter-go-node/helpers"
	"github.com/tendermint/tendermint/libs/kv"
	"math/big"
	"regexp"
	"strconv"
)

const maxCoinNameBytes = 64
const allowedCoinSymbols = "^[A-Z0-9]{3,10}$"

var (
	minCoinSupply  = helpers.BipToPip(big.NewInt(1))
	minCoinReserve = helpers.BipToPip(big.NewInt(10000))
	maxCoinSupply  = big.NewInt(0).Exp(big.NewInt(10), big.NewInt(15+18), nil)
)

type CreateCoinData struct {
	Name                 string
	Symbol               types.CoinSymbol
	InitialAmount        *big.Int
	InitialReserve       *big.Int
	ConstantReserveRatio uint
	MaxSupply            *big.Int
}

func (data CreateCoinData) BasicCheck(tx *Transaction, context *state.CheckState) *Response {
	if data.InitialReserve == nil || data.InitialAmount == nil || data.MaxSupply == nil {
		return &Response{
			Code: code.DecodeError,
			Log:  "Incorrect tx data"}
	}

	if len(data.Name) > maxCoinNameBytes {
		return &Response{
			Code: code.InvalidCoinName,
			Log:  fmt.Sprintf("Coin name is invalid. Allowed up to %d bytes.", maxCoinNameBytes)}
	}

	if match, _ := regexp.MatchString(allowedCoinSymbols, data.Symbol.String()); !match {
		return &Response{
			Code: code.InvalidCoinSymbol,
			Log:  fmt.Sprintf("Invalid coin symbol. Should be %s", allowedCoinSymbols)}
	}

	if context.Coins().ExistsBySymbol(data.Symbol) {
		return &Response{
			Code: code.CoinAlreadyExists,
			Log:  fmt.Sprintf("Coin already exists"),
			Info: EncodeError(map[string]string{
				"code": strconv.Itoa(int(code.CoinAlreadyExists)),
			}),
		}

	}

	if data.ConstantReserveRatio < 10 || data.ConstantReserveRatio > 100 {
		return &Response{
			Code: code.WrongCrr,
			Log:  fmt.Sprintf("Constant Reserve Ratio should be between 10 and 100"),
			Info: EncodeError(map[string]string{
				"code": strconv.Itoa(int(code.WrongCrr)),
			})}
	}

	if data.InitialAmount.Cmp(minCoinSupply) == -1 || data.InitialAmount.Cmp(data.MaxSupply) == 1 {
		return &Response{
			Code: code.WrongCoinSupply,
			Log:  fmt.Sprintf("Coin supply should be between %s and %s", minCoinSupply.String(), data.MaxSupply.String()),
			Info: EncodeError(map[string]string{
				"code": strconv.Itoa(int(code.WrongCoinSupply)),
			})}
	}

	if data.MaxSupply.Cmp(maxCoinSupply) == 1 {
		return &Response{
			Code: code.WrongCoinSupply,
			Log:  fmt.Sprintf("Max coin supply should be less than %s", maxCoinSupply),
			Info: EncodeError(map[string]string{
				"code": strconv.Itoa(int(code.WrongCoinSupply)),
			})}
	}

	if data.InitialReserve.Cmp(minCoinReserve) == -1 {
		return &Response{
			Code: code.WrongCoinSupply,
			Log:  fmt.Sprintf("Coin reserve should be greater than or equal to %s", minCoinReserve.String()),
			Info: EncodeError(map[string]string{
				"code": strconv.Itoa(int(code.WrongCoinSupply)),
			})}
	}

	return nil
}

func (data CreateCoinData) String() string {
	return fmt.Sprintf("CREATE COIN symbol:%s reserve:%s amount:%s crr:%d",
		data.Symbol.String(), data.InitialReserve, data.InitialAmount, data.ConstantReserveRatio)
}

func (data CreateCoinData) Gas() int64 {
	switch len(data.Symbol.String()) {
	case 3:
		return 1000000000 // 1mln bips
	case 4:
		return 100000000 // 100k bips
	case 5:
		return 10000000 // 10k bips
	case 6:
		return 1000000 // 1k bips
	}

	return 100000 // 100 bips
}

func (data CreateCoinData) Run(tx *Transaction, context state.Interface, rewardPool *big.Int, currentBlock uint64) Response {
	sender, _ := tx.Sender()

	var checkState *state.CheckState
	var isCheck bool
	if checkState, isCheck = context.(*state.CheckState); !isCheck {
		checkState = state.NewCheckState(context.(*state.State))
	}
	response := data.BasicCheck(tx, checkState)
	if response != nil {
		return *response
	}

	commissionInBaseCoin := tx.CommissionInBaseCoin()
	commission := big.NewInt(0).Set(commissionInBaseCoin)

	if tx.GasCoin != types.GetBaseCoinID() {
		coin := checkState.Coins().GetCoin(tx.GasCoin)

		errResp := CheckReserveUnderflow(coin, commissionInBaseCoin)
		if errResp != nil {
			return *errResp
		}

		if coin.Reserve().Cmp(commissionInBaseCoin) < 0 {
			return Response{
				Code: code.CoinReserveNotSufficient,
				Log:  fmt.Sprintf("Gas coin reserve balance is not sufficient for transaction. Has: %s %s, required %s %s", coin.Reserve().String(), types.GetBaseCoin(), commissionInBaseCoin.String(), types.GetBaseCoin()),
				Info: EncodeError(map[string]string{
					"code":           strconv.Itoa(int(code.CoinReserveNotSufficient)),
					"has_value":      coin.Reserve().String(),
					"required_value": commissionInBaseCoin.String(),
					"coin":           fmt.Sprintf("%s", types.GetBaseCoin()),
				}),
			}
		}

		commission = formula.CalculateSaleAmount(coin.Volume(), coin.Reserve(), coin.Crr(), commissionInBaseCoin)
	}

	if checkState.Accounts().GetBalance(sender, tx.GasCoin).Cmp(commission) < 0 {
		gasCoin := checkState.Coins().GetCoin(tx.GasCoin)

		return Response{
			Code: code.InsufficientFunds,
			Log:  fmt.Sprintf("Insufficient funds for sender account: %s. Wanted %s %s", sender.String(), commission.String(), gasCoin.GetFullSymbol()),
			Info: EncodeError(map[string]string{
				"code":         strconv.Itoa(int(code.InsufficientFunds)),
				"sender":       sender.String(),
				"needed_value": commission.String(),
				"coin":         gasCoin.GetFullSymbol(),
			}),
		}
	}

	if checkState.Accounts().GetBalance(sender, types.GetBaseCoinID()).Cmp(data.InitialReserve) < 0 {
		return Response{
			Code: code.InsufficientFunds,
			Log:  fmt.Sprintf("Insufficient funds for sender account: %s. Wanted %s %s", sender.String(), data.InitialReserve.String(), types.GetBaseCoin()),
			Info: EncodeError(map[string]string{
				"code":           strconv.Itoa(int(code.InsufficientFunds)),
				"sender":         sender.String(),
				"needed_reserve": data.InitialReserve.String(),
				"coin":           fmt.Sprintf("%s", types.GetBaseCoin()),
			}),
		}
	}

	if tx.GasCoin.IsBaseCoin() {
		totalTxCost := big.NewInt(0)
		totalTxCost.Add(totalTxCost, data.InitialReserve)
		totalTxCost.Add(totalTxCost, commission)

		if checkState.Accounts().GetBalance(sender, types.GetBaseCoinID()).Cmp(totalTxCost) < 0 {
			gasCoin := checkState.Coins().GetCoin(tx.GasCoin)

			return Response{
				Code: code.InsufficientFunds,
				Log:  fmt.Sprintf("Insufficient funds for sender account: %s. Wanted %s %s", sender.String(), totalTxCost.String(), gasCoin.GetFullSymbol()),
				Info: EncodeError(map[string]string{
					"code":         strconv.Itoa(int(code.InsufficientFunds)),
					"sender":       sender.String(),
					"needed_value": totalTxCost.String(),
					"coin":         gasCoin.GetFullSymbol(),
				}),
			}
		}
	}

	if deliverState, ok := context.(*state.State); ok {
		rewardPool.Add(rewardPool, commissionInBaseCoin)

		deliverState.Coins.SubReserve(tx.GasCoin, commissionInBaseCoin)
		deliverState.Coins.SubVolume(tx.GasCoin, commission)

		deliverState.Accounts.SubBalance(sender, types.GetBaseCoinID(), data.InitialReserve)
		deliverState.Accounts.SubBalance(sender, tx.GasCoin, commission)

		coinID := deliverState.App.GetCoinsCount() + 1
		deliverState.Coins.Create(
			types.CoinID(coinID),
			data.Symbol,
			data.Name,
			data.InitialAmount,
			data.ConstantReserveRatio,
			data.InitialReserve,
			data.MaxSupply,
			&sender,
		)

		deliverState.App.SetCoinsCount(coinID)
		deliverState.Accounts.AddBalance(sender, types.CoinID(coinID), data.InitialAmount)
		deliverState.Accounts.SetNonce(sender, tx.Nonce)
	}

	tags := kv.Pairs{
		kv.Pair{Key: []byte("tx.type"), Value: []byte(hex.EncodeToString([]byte{byte(TypeCreateCoin)}))},
		kv.Pair{Key: []byte("tx.from"), Value: []byte(hex.EncodeToString(sender[:]))},
		kv.Pair{Key: []byte("tx.coin"), Value: []byte(data.Symbol.String())},
	}

	return Response{
		Code:      code.OK,
		Tags:      tags,
		GasUsed:   tx.Gas(),
		GasWanted: tx.Gas(),
	}
}
