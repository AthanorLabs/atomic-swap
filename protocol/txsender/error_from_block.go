package txsender

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// errorFromBlock returns the error for why a transaction was reverted when mined.
// Normally these errors are detected when creating the transaction, as the contract
// call is simulated to estimate gas, but the state is different in the mined block
// and the transaction can fail (losing gas) after it already went out to the network.
// In this case, we simulate the call using the mined block to extract the error.
func errorFromBlock(
	contractCaller ethereum.ContractCaller,
	from common.Address,
	tx *types.Transaction,
	blockNum *big.Int,
) error {
	msg := ethereum.CallMsg{
		From:     from,
		To:       tx.To(),
		Gas:      tx.Gas(),
		GasPrice: tx.GasPrice(),
		Value:    tx.Value(),
		Data:     tx.Data(),
	}
	res, err := contractCaller.CallContract(context.Background(), msg, blockNum)
	if err != nil {
		return fmt.Errorf("failed to determine error in mined block, %w", err)
	}
	return unpackError(res)
}

func unpackError(result []byte) error {
	errorSig := []byte{0x08, 0xc3, 0x79, 0xa0} // Keccak256("Error(string)")[:4]
	abiString, _ := abi.NewType("string", "", nil)

	if !bytes.Equal(result[:4], errorSig) {
		// Should not happen, but being safe and not panic'ing
		return errors.New("failed to determine error in mined block, TX result not of type Error(string)")
	}
	vs, err := abi.Arguments{{Type: abiString}}.UnpackValues(result[4:])
	if err != nil {
		// In theory, we also should never see this error
		return fmt.Errorf("failed to UnpackValues in mined block, %w", err)
	}
	return errors.New(vs[0].(string))
}
