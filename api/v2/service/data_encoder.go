package service

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"github.com/MinterTeam/minter-go-node/coreV2/state/coins"
	"github.com/MinterTeam/minter-go-node/coreV2/transaction"
	pb "github.com/MinterTeam/node-grpc-gateway/api_pb"
	"github.com/golang/protobuf/ptypes/any"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	_struct "google.golang.org/protobuf/types/known/structpb"
)

func encode(data transaction.Data, rCoins coins.RCoins) (*any.Any, error) {
	var m proto.Message
	switch d := data.(type) {
	case *transaction.BuyCoinData:
		m = &pb.BuyCoinData{
			CoinToBuy: &pb.Coin{
				Id:     uint64(d.CoinToBuy),
				Symbol: rCoins.GetCoin(d.CoinToBuy).GetFullSymbol(),
			},
			ValueToBuy: d.ValueToBuy.String(),
			CoinToSell: &pb.Coin{
				Id:     uint64(d.CoinToSell),
				Symbol: rCoins.GetCoin(d.CoinToSell).GetFullSymbol(),
			},
			MaximumValueToSell: d.MaximumValueToSell.String(),
		}
	case *transaction.EditCoinOwnerData:
		m = &pb.EditCoinOwnerData{
			Symbol:   d.Symbol.String(),
			NewOwner: d.NewOwner.String(),
		}
	case *transaction.CreateCoinData:
		m = &pb.CreateCoinData{
			Name:                 d.Name,
			Symbol:               d.Symbol.String(),
			InitialAmount:        d.InitialAmount.String(),
			InitialReserve:       d.InitialReserve.String(),
			ConstantReserveRatio: uint64(d.ConstantReserveRatio),
			MaxSupply:            d.MaxSupply.String(),
		}
	case *transaction.CreateMultisigData:
		weights := make([]uint64, 0, len(d.Weights))
		for _, weight := range d.Weights {
			weights = append(weights, uint64(weight))
		}
		addresses := make([]string, 0, len(d.Addresses))
		for _, address := range d.Addresses {
			addresses = append(addresses, address.String())
		}
		m = &pb.CreateMultisigData{
			Threshold: uint64(d.Threshold),
			Weights:   weights,
			Addresses: addresses,
		}
	case *transaction.DeclareCandidacyData:
		m = &pb.DeclareCandidacyData{
			Address:    d.Address.String(),
			PubKey:     d.PubKey.String(),
			Commission: uint64(d.Commission),
			Coin: &pb.Coin{
				Id:     uint64(d.Coin),
				Symbol: rCoins.GetCoin(d.Coin).GetFullSymbol(),
			},
			Stake: d.Stake.String(),
		}
	case *transaction.DelegateData:
		m = &pb.DelegateData{
			PubKey: d.PubKey.String(),
			Coin: &pb.Coin{
				Id:     uint64(d.Coin),
				Symbol: rCoins.GetCoin(d.Coin).GetFullSymbol(),
			},
			Value: d.Value.String(),
		}
	case *transaction.EditCandidateData:
		m = &pb.EditCandidateData{
			PubKey:         d.PubKey.String(),
			RewardAddress:  d.RewardAddress.String(),
			OwnerAddress:   d.OwnerAddress.String(),
			ControlAddress: d.ControlAddress.String(),
		}
	case *transaction.EditCandidatePublicKeyData:
		m = &pb.EditCandidatePublicKeyData{
			PubKey:    d.PubKey.String(),
			NewPubKey: d.NewPubKey.String(),
		}
	case *transaction.EditMultisigData:
		weights := make([]uint64, 0, len(d.Weights))
		for _, weight := range d.Weights {
			weights = append(weights, uint64(weight))
		}
		addresses := make([]string, 0, len(d.Addresses))
		for _, address := range d.Addresses {
			addresses = append(addresses, address.String())
		}
		m = &pb.EditMultisigData{
			Threshold: uint64(d.Threshold),
			Weights:   weights,
			Addresses: addresses,
		}
	case *transaction.MultisendData:
		list := make([]*pb.SendData, 0, len(d.List))
		for _, item := range d.List {
			list = append(list, &pb.SendData{
				Coin: &pb.Coin{
					Id:     uint64(item.Coin),
					Symbol: rCoins.GetCoin(item.Coin).GetFullSymbol(),
				},
				To:    item.To.String(),
				Value: item.Value.String(),
			})
		}
		m = &pb.MultiSendData{
			List: list,
		}
	// case *transaction.PriceVoteData:
	// 	m = &pb.PriceVoteData{
	// 		Price: strconv.Itoa(int(d.Price)),
	// 	}
	case *transaction.RecreateCoinData:
		m = &pb.RecreateCoinData{
			Name:                 d.Name,
			Symbol:               d.Symbol.String(),
			InitialAmount:        d.InitialAmount.String(),
			InitialReserve:       d.InitialReserve.String(),
			ConstantReserveRatio: uint64(d.ConstantReserveRatio),
			MaxSupply:            d.MaxSupply.String(),
		}
	case *transaction.RedeemCheckData:
		m = &pb.RedeemCheckData{
			RawCheck: base64.StdEncoding.EncodeToString(d.RawCheck),
			Proof:    base64.StdEncoding.EncodeToString(d.Proof[:]),
		}
	case *transaction.SellAllCoinData:
		m = &pb.SellAllCoinData{
			CoinToSell: &pb.Coin{
				Id:     uint64(d.CoinToSell),
				Symbol: rCoins.GetCoin(d.CoinToSell).GetFullSymbol(),
			},
			CoinToBuy: &pb.Coin{
				Id:     uint64(d.CoinToBuy),
				Symbol: rCoins.GetCoin(d.CoinToBuy).GetFullSymbol(),
			},
			MinimumValueToBuy: d.MinimumValueToBuy.String(),
		}
	case *transaction.SellCoinData:
		m = &pb.SellCoinData{
			CoinToSell: &pb.Coin{
				Id:     uint64(d.CoinToSell),
				Symbol: rCoins.GetCoin(d.CoinToSell).GetFullSymbol(),
			},
			ValueToSell: d.ValueToSell.String(),
			CoinToBuy: &pb.Coin{
				Id:     uint64(d.CoinToBuy),
				Symbol: rCoins.GetCoin(d.CoinToBuy).GetFullSymbol(),
			},
			MinimumValueToBuy: d.MinimumValueToBuy.String(),
		}
	case *transaction.SendData:
		m = &pb.SendData{
			Coin: &pb.Coin{
				Id:     uint64(d.Coin),
				Symbol: rCoins.GetCoin(d.Coin).GetFullSymbol(),
			},
			To:    d.To.String(),
			Value: d.Value.String(),
		}
	case *transaction.SetHaltBlockData:
		m = &pb.SetHaltBlockData{
			PubKey: d.PubKey.String(),
			Height: d.Height,
		}
	case *transaction.SetCandidateOnData:
		m = &pb.SetCandidateOnData{
			PubKey: d.PubKey.String(),
		}
	case *transaction.SetCandidateOffData:
		m = &pb.SetCandidateOffData{
			PubKey: d.PubKey.String(),
		}
	case *transaction.UnbondData:
		m = &pb.UnbondData{
			PubKey: d.PubKey.String(),
			Coin: &pb.Coin{
				Id:     uint64(d.Coin),
				Symbol: rCoins.GetCoin(d.Coin).GetFullSymbol(),
			},
			Value: d.Value.String(),
		}
	case *transaction.AddLiquidityData:
		m = &pb.AddLiquidityData{
			Coin0: &pb.Coin{
				Id:     uint64(d.Coin0),
				Symbol: rCoins.GetCoin(d.Coin0).GetFullSymbol(),
			},
			Coin1: &pb.Coin{
				Id:     uint64(d.Coin1),
				Symbol: rCoins.GetCoin(d.Coin1).GetFullSymbol(),
			},
			Volume0:        d.Volume0.String(),
			MaximumVolume1: d.MaximumVolume1.String(),
		}
	case *transaction.RemoveLiquidity:
		m = &pb.RemoveLiquidityData{
			Coin0: &pb.Coin{
				Id:     uint64(d.Coin0),
				Symbol: rCoins.GetCoin(d.Coin0).GetFullSymbol(),
			},
			Coin1: &pb.Coin{
				Id:     uint64(d.Coin1),
				Symbol: rCoins.GetCoin(d.Coin1).GetFullSymbol(),
			},
			Liquidity:      d.Liquidity.String(),
			MinimumVolume0: d.MinimumVolume0.String(),
			MinimumVolume1: d.MinimumVolume1.String(),
		}
	case *transaction.BuySwapPoolData:
		var coinsInfo []*pb.Coin
		for _, coin := range d.Coins {
			coinsInfo = append(coinsInfo, &pb.Coin{
				Id:     uint64(coin),
				Symbol: rCoins.GetCoin(coin).GetFullSymbol(),
			})
		}
		m = &pb.BuySwapPoolData{
			Coins:              coinsInfo,
			ValueToBuy:         d.ValueToBuy.String(),
			MaximumValueToSell: d.MaximumValueToSell.String(),
		}
	case *transaction.SellSwapPoolData:
		var coinsInfo []*pb.Coin
		for _, coin := range d.Coins {
			coinsInfo = append(coinsInfo, &pb.Coin{
				Id:     uint64(coin),
				Symbol: rCoins.GetCoin(coin).GetFullSymbol(),
			})
		}
		m = &pb.SellSwapPoolData{
			Coins:             coinsInfo,
			ValueToSell:       d.ValueToSell.String(),
			MinimumValueToBuy: d.MinimumValueToBuy.String(),
		}
	case *transaction.SellAllSwapPoolData:
		var coinsInfo []*pb.Coin
		for _, coin := range d.Coins {
			coinsInfo = append(coinsInfo, &pb.Coin{
				Id:     uint64(coin),
				Symbol: rCoins.GetCoin(coin).GetFullSymbol(),
			})
		}
		m = &pb.SellAllSwapPoolData{
			Coins:             coinsInfo,
			MinimumValueToBuy: d.MinimumValueToBuy.String(),
		}
	case *transaction.CreateTokenData:
		m = &pb.CreateTokenData{
			Name:          d.Name,
			Symbol:        d.Symbol.String(),
			InitialAmount: d.InitialAmount.String(),
			MaxSupply:     d.MaxSupply.String(),
			Mintable:      d.Mintable,
			Burnable:      d.Burnable,
		}
	case *transaction.RecreateTokenData:
		m = &pb.RecreateTokenData{
			Name:          d.Name,
			Symbol:        d.Symbol.String(),
			InitialAmount: d.InitialAmount.String(),
			MaxSupply:     d.MaxSupply.String(),
			Mintable:      d.Mintable,
			Burnable:      d.Burnable,
		}
	case *transaction.BurnTokenData:
		m = &pb.BurnTokenData{
			Coin: &pb.Coin{
				Id:     uint64(d.Coin),
				Symbol: rCoins.GetCoin(d.Coin).GetFullSymbol(),
			},
			Value: d.Value.String(),
		}
	case *transaction.MintTokenData:
		m = &pb.MintTokenData{
			Coin: &pb.Coin{
				Id:     uint64(d.Coin),
				Symbol: rCoins.GetCoin(d.Coin).GetFullSymbol(),
			},
			Value: d.Value.String(),
		}
	case *transaction.EditCandidateCommission:
		m = &pb.EditCandidateCommission{
			PubKey:     d.PubKey.String(),
			Commission: uint64(d.Commission),
		}
	// case *transaction.MoveStakeData:
	// 	m = &pb.MoveStakeData{
	// 		From: d.From.String(),
	// 		To:   d.To.String(),
	// 		Coin: &pb.Coin{
	// 			Id:     uint64(d.Coin),
	// 			Symbol: rCoins.GetCoin(d.Coin).GetFullSymbol(),
	// 		},
	// 		Stake: d.Stake.String(),
	// 	}
	case *transaction.VoteCommissionData:
		m = priceCommissionData(d, rCoins.GetCoin(d.Coin))
	case *transaction.VoteUpdateData:
		m = &pb.VoteUpdateData{
			PubKey:  d.PubKey.String(),
			Height:  d.Height,
			Version: d.Version,
		}
	case *transaction.CreateSwapPoolData:
		m = &pb.CreateSwapPoolData{
			Coin0: &pb.Coin{
				Id:     uint64(d.Coin0),
				Symbol: rCoins.GetCoin(d.Coin0).GetFullSymbol(),
			},
			Coin1: &pb.Coin{
				Id:     uint64(d.Coin1),
				Symbol: rCoins.GetCoin(d.Coin1).GetFullSymbol(),
			},
			Volume0: d.Volume0.String(),
			Volume1: d.Volume1.String(),
		}
	default:
		return nil, errors.New("unknown tx type")
	}

	a, err := anypb.New(m)
	if err != nil {
		return nil, err
	}

	return a, nil
}

func priceCommissionData(d *transaction.VoteCommissionData, coin *coins.Model) proto.Message {
	return &pb.VoteCommissionData{
		PubKey: d.PubKey.String(),
		Height: d.Height,
		Coin: &pb.Coin{
			Id:     uint64(d.Coin),
			Symbol: coin.GetFullSymbol(),
		},
		PayloadByte:      d.PayloadByte.String(),
		Send:             d.Send.String(),
		BuyBancor:        d.BuyBancor.String(),
		SellBancor:       d.SellBancor.String(),
		SellAllBancor:    d.SellAllBancor.String(),
		BuyPoolBase:      d.BuyPoolBase.String(),
		BuyPoolDelta:     d.BuyPoolDelta.String(),
		SellPoolBase:     d.SellPoolBase.String(),
		SellPoolDelta:    d.SellPoolDelta.String(),
		SellAllPoolBase:  d.SellAllPoolBase.String(),
		SellAllPoolDelta: d.SellAllPoolDelta.String(),
		CreateTicker3:    d.CreateTicker3.String(),
		CreateTicker4:    d.CreateTicker4.String(),
		CreateTicker5:    d.CreateTicker5.String(),
		CreateTicker6:    d.CreateTicker6.String(),
		CreateTicker7_10: d.CreateTicker7to10.String(),
		CreateCoin:       d.CreateCoin.String(),
		CreateToken:      d.CreateToken.String(),
		RecreateCoin:     d.RecreateCoin.String(),
		RecreateToken:    d.RecreateToken.String(),
		DeclareCandidacy: d.DeclareCandidacy.String(),
		Delegate:         d.Delegate.String(),
		Unbond:           d.Unbond.String(),
		RedeemCheck:      d.RedeemCheck.String(),
		SetCandidateOn:   d.SetCandidateOn.String(),
		SetCandidateOff:  d.SetCandidateOff.String(),
		CreateMultisig:   d.CreateMultisig.String(),
		MultisendBase:    d.MultisendBase.String(),
		MultisendDelta:   d.MultisendDelta.String(),
		EditCandidate:    d.EditCandidate.String(),
		SetHaltBlock:     d.SetHaltBlock.String(),
		EditTickerOwner:  d.EditTickerOwner.String(),
		EditMultisig:     d.EditMultisig.String(),
		// PriceVote:               d.PriceVote.String(),
		EditCandidatePublicKey:  d.EditCandidatePublicKey.String(),
		CreateSwapPool:          d.CreateSwapPool.String(),
		AddLiquidity:            d.AddLiquidity.String(),
		RemoveLiquidity:         d.RemoveLiquidity.String(),
		EditCandidateCommission: d.EditCandidateCommission.String(),
		// MoveStake:               d.MoveStake.String(),
		MintToken:      d.MintToken.String(),
		BurnToken:      d.BurnToken.String(),
		VoteCommission: d.VoteCommission.String(),
		VoteUpdate:     d.VoteUpdate.String(),
	}
}

func encodeToStruct(b []byte) (*_struct.Struct, error) {
	dataStruct := &_struct.Struct{}
	if err := dataStruct.UnmarshalJSON(b); err != nil {
		return nil, err
	}

	return dataStruct, nil
}

func toStruct(d interface{}) (*_struct.Struct, error) {
	byteData, err := json.Marshal(d)
	if err != nil {
		return nil, err
	}

	data, err := encodeToStruct(byteData)
	if err != nil {
		return nil, err
	}
	return data, nil
}
