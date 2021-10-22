package monero

import (
	"encoding/json"
	"fmt"
	"strings"
)

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

func callGenerateFromKeys(kp *PrivateKeyPair, filename, password string) error {
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

	resp, err := postRPC(defaultEndpoint, method, string(params))
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
