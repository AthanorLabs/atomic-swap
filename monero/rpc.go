package monero

import (
	"encoding/json"
	"fmt"
	"strings"
)

// defaultEndpoint is the default monero-wallet-rpc endpoint for stagenet
const defaultEndpoint = "http://127.0.0.1:18082/json_rpc"

// Address represents a base58-encoded string
type Address string

type generateFromKeysRequest struct {
	Filename string `json:"filename"`
	Address  string `json:"address"`
	SpendKey string `json:"spendkey"`
	ViewKey  string `json:"viewkey"`
	Password string `json:"password"`
}

type generateFromKeysResponse struct {
	Address string `json:"address"`
	Info    string `json:"info`
}

func (c *client) callGenerateFromKeys(kp *PrivateKeyPair, filename, password string) error {
	const (
		method         = "generate_from_keys"
		successMessage = "Wallet has been generated successfully."
	)

	req := &generateFromKeysRequest{
		Filename: filename,
		Address:  string(kp.Address()),
		SpendKey: kp.sk.Hex(),
		ViewKey:  kp.vk.Hex(),
		Password: password,
	}

	params, err := json.Marshal(req)
	if err != nil {
		return err
	}

	resp, err := postRPC(c.endpoint, method, string(params))
	if err != nil {
		return err
	}

	if resp.Error != nil {
		return resp.Error
	}

	var res *generateFromKeysResponse
	if err = json.Unmarshal(resp.Result, &res); err != nil {
		return err
	}

	if strings.Compare(successMessage, res.Info) == 0 {
		return nil
	}

	return fmt.Errorf("got unexpected Info string: %s", res.Info)
}

type Destination struct {
	Amount uint `json:"amount"`
	Address string `json:"address"`
}

type transferRequest struct {
	Destinations []Destination `json:"destinations"`
	// AccountIndex uint // optional
	Priority uint `json:"priority"`
	//Mixin uint  `json:"mixin"`
	//RingSize uint  `json:"ring_size"`
	//UnlockTime uint  `json:"unlock_time"`
	// GetTxKey bool
}

type transferResponse struct {
	Amount uint  `json:"amount"`
	Fee uint `json:"fee"`
	MultisigTxset interface{} `json:"multisig_txset"`
	TxBlob string `json:"tx_blob"`
	TxHash string `json:"tx_hash"`
	TxKey string `json:"tx_key"`
	TxMetadata string `json:"tx_metadata"`
	UnsignedTxset string `json:"unsigned_txset"`
}

func (c *client) callTransfer(destinations []Destination) (string, error) {
	const (
		method         = "transfer"
	)

	req := &transferRequest{
		Destinations: destinations,
		Priority: 0,
		//RingSize: 11,
	}

	params, err := json.Marshal(req)
	if err != nil {
		return "", err
	}

	resp, err := postRPC(c.endpoint, method, string(params))
	if err != nil {
		return "", err
	}

	if resp.Error != nil {
		return "", resp.Error
	}

	var res *transferResponse
	if err = json.Unmarshal(resp.Result, &res); err != nil {
		return "", err
	}	

	return res.TxHash, nil
}