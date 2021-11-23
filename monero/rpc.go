package monero

import (
	"encoding/json"
	"fmt"
	"strings"
)

const (
	// defaultEndpointWalletFile is the default monero-wallet-rpc endpoint with a provided --wallet-file
	defaultEndpointWalletFile = "http://127.0.0.1:18083/json_rpc"

	// defaultEndpointWalletDir is the default monero-wallet-rpc endpoint with a provided --wallet-dir
	defaultEndpointWalletDir = "http://127.0.0.1:18084/json_rpc"
)

const defaultDaemonEndpoint = "http://127.0.0.1:18081/json_rpc"

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
	Info    string `json:"info"`
}

func (c *client) callGenerateFromKeys(sk *PrivateSpendKey, vk *PrivateViewKey, address Address, filename, password string) error {
	const (
		method                 = "generate_from_keys"
		successMessage         = "Wallet has been generated successfully."
		viewOnlySuccessMessage = "Watch-only wallet has been generated successfully."
	)

	req := &generateFromKeysRequest{
		Filename: filename,
		Address:  string(address),
		ViewKey:  vk.Hex(),
		Password: password,
	}

	if sk != nil {
		req.SpendKey = sk.Hex()
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

	// TODO: check for if we passed spend key or not
	if strings.Compare(successMessage, res.Info) == 0 || strings.Compare(viewOnlySuccessMessage, res.Info) == 0 {
		return nil
	}

	return fmt.Errorf("got unexpected Info string: %s", res.Info)
}

type Destination struct {
	Amount  uint   `json:"amount"`
	Address string `json:"address"`
}

type transferRequest struct {
	Destinations []Destination `json:"destinations"`
	AccountIndex uint          // optional
	Priority     uint          `json:"priority"`
	// Mixin uint  `json:"mixin"`
	// RingSize uint  `json:"ring_size"`
	// UnlockTime uint  `json:"unlock_time"`
	// GetTxKey bool
}

type transferResponse struct {
	Amount        uint        `json:"amount"`
	Fee           uint        `json:"fee"`
	MultisigTxset interface{} `json:"multisig_txset"`
	TxBlob        string      `json:"tx_blob"`
	TxHash        string      `json:"tx_hash"`
	TxKey         string      `json:"tx_key"`
	TxMetadata    string      `json:"tx_metadata"`
	UnsignedTxset string      `json:"unsigned_txset"`
}

func (c *client) callTransfer(destinations []Destination, accountIdx uint) (string, error) {
	const (
		method = "transfer"
	)

	req := &transferRequest{
		Destinations: destinations,
		AccountIndex: accountIdx,
		Priority:     0,
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

type getBalanceRequest struct {
	AccountIndex uint `json:"account_index"`
}

type getBalanceResponse struct {
	Balance         float64                  `json:"balance"`
	BlocksToUnlock  uint                     `json:"blocks_to_unlock"`
	UnlockedBalance float64                  `json:"unlocked_balance"`
	PerSubaddress   []map[string]interface{} `json:"per_subaddress"`
}

func (c *client) callGetBalance(idx uint) (*getBalanceResponse, error) {
	const method = "get_balance"

	req := &getBalanceRequest{
		AccountIndex: idx,
	}

	params, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := postRPC(c.endpoint, method, string(params))
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, resp.Error
	}

	var res *getBalanceResponse
	if err = json.Unmarshal(resp.Result, &res); err != nil {
		return nil, err
	}

	return res, nil
}

type getAddressRequest struct {
	AccountIndex uint `json:"account_index"`
}

type getAddressResponse struct {
	Address string `json:"address"`
}

func (c *client) callGetAddress(idx uint) (*getAddressResponse, error) {
	const method = "get_address"

	req := &getAddressRequest{
		AccountIndex: idx,
	}

	params, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := postRPC(c.endpoint, method, string(params))
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, resp.Error
	}

	var res *getAddressResponse
	if err = json.Unmarshal(resp.Result, &res); err != nil {
		return nil, err
	}

	return res, nil
}

type getAccountsResponse struct {
	SubaddressAccounts []map[string]interface{} `json:"subaddress_accounts"`
}

func (c *client) callGetAccounts() (*getAccountsResponse, error) {
	const method = "get_accounts"

	resp, err := postRPC(c.endpoint, method, "{}")
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, resp.Error
	}

	var res *getAccountsResponse
	if err = json.Unmarshal(resp.Result, &res); err != nil {
		return nil, err
	}

	return res, nil
}

type openWalletRequest struct {
	Filename string `json:"filename"`
	Password string `json:"password"`
}

func (c *client) callOpenWallet(filename, password string) error {
	const method = "open_wallet"

	req := &openWalletRequest{
		Filename: filename,
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

	return nil
}
