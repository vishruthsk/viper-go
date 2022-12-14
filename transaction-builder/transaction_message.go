package transactionbuilder

import (
	"encoding/hex"

	"github.com/vishruthsk/viper-network/crypto"
	coreTypes "github.com/vishruthsk/viper-network/types"
	appsType "github.com/vishruthsk/viper-network/x/apps/types"
	nodesTypes "github.com/vishruthsk/viper-network/x/nodes/types"
)

// TransactionMessage interface that represents message to be sent as transaction
type TransactionMessage interface {
	coreTypes.ProtoMsg
}

// NewSend returns message for send transaction
func NewSend(fromAddress, toAddress string, amount int64) (TransactionMessage, error) {
	decodedFromAddress, err := hex.DecodeString(fromAddress)
	if err != nil {
		return nil, err
	}

	decodedToAddress, err := hex.DecodeString(toAddress)
	if err != nil {
		return nil, err
	}

	return &nodesTypes.MsgSend{
		FromAddress: decodedFromAddress,
		ToAddress:   decodedToAddress,
		Amount:      coreTypes.NewInt(amount),
	}, nil
}

// NewStakeApp returns message for Stake App transaction
func NewStakeApp(publicKey string, chains []string, amount int64) (TransactionMessage, error) {
	cryptoPublicKey, err := crypto.NewPublicKey(publicKey)
	if err != nil {
		return nil, err
	}

	return &appsType.MsgStake{
		PubKey: cryptoPublicKey,
		Chains: chains,
		Value:  coreTypes.NewInt(amount),
	}, nil
}

// NewUnstakeApp returns message for Unstake App transaction
func NewUnstakeApp(address string) (TransactionMessage, error) {
	decodedAddress, err := hex.DecodeString(address)
	if err != nil {
		return nil, err
	}

	return &appsType.MsgBeginUnstake{
		Address: decodedAddress,
	}, nil
}

// NewUnjailApp returns message for Unjail App transaction
func NewUnjailApp(address string) (TransactionMessage, error) {
	decodedAddress, err := hex.DecodeString(address)
	if err != nil {
		return nil, err
	}

	return &appsType.MsgUnjail{
		AppAddr: decodedAddress,
	}, nil
}

// NewStakeNode returns message for Stake Node transaction
func NewStakeNode(publicKey, serviceURL, outputAddress string, chains []string, amount int64) (TransactionMessage, error) {
	cryptoPublicKey, err := crypto.NewPublicKey(publicKey)
	if err != nil {
		return nil, err
	}

	decodedAddress, err := hex.DecodeString(outputAddress)
	if err != nil {
		return nil, err
	}

	return &nodesTypes.MsgStake{
		PublicKey:  cryptoPublicKey,
		Chains:     chains,
		Value:      coreTypes.NewInt(amount),
		ServiceUrl: serviceURL,
		Output:     decodedAddress,
	}, nil
}

// NewUnstakeNode returns message for Unstake Node transaction
func NewUnstakeNode(fromAddress, operatorAddress string) (TransactionMessage, error) {
	decodedFromAddress, err := hex.DecodeString(fromAddress)
	if err != nil {
		return nil, err
	}

	decodedOperatorAddress, err := hex.DecodeString(operatorAddress)
	if err != nil {
		return nil, err
	}

	return &nodesTypes.MsgBeginUnstake{
		Address: decodedOperatorAddress,
		Signer:  decodedFromAddress,
	}, nil
}

// NewUnjailNode returns message for Unjail Node transaction
func NewUnjailNode(fromAddress, operatorAddress string) (TransactionMessage, error) {
	decodedFromAddress, err := hex.DecodeString(fromAddress)
	if err != nil {
		return nil, err
	}

	decodedOperatorAddress, err := hex.DecodeString(operatorAddress)
	if err != nil {
		return nil, err
	}

	return &nodesTypes.MsgUnjail{
		ValidatorAddr: decodedOperatorAddress,
		Signer:        decodedFromAddress,
	}, nil
}
