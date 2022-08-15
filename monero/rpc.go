package monero

import (
	"fmt"
	"strings"

	"github.com/MarinX/monerorpc/wallet"

	mcrypto "github.com/noot/atomic-swap/crypto/monero"
)

func (c *client) callGenerateFromKeys(
	sk *mcrypto.PrivateSpendKey,
	vk *mcrypto.PrivateViewKey,
	address mcrypto.Address,
	filename,
	password string,
) error {

	const (
		successMessage         = "Wallet has been generated successfully."
		viewOnlySuccessMessage = "Watch-only wallet has been generated successfully."
	)

	spendKeyHex := ""
	if sk != nil {
		spendKeyHex = sk.Hex()
	}

	res, err := c.rpc.Wallet.GenerateFromKeys(&wallet.GenerateFromKeysRequest{
		Filename: filename,
		Address:  string(address),
		Viewkey:  vk.Hex(),
		Spendkey: spendKeyHex,
		Password: password,
	})
	if err != nil {
		return err
	}

	// TODO: check for if we passed spend key or not
	if strings.Compare(successMessage, res.Info) == 0 || strings.Compare(viewOnlySuccessMessage, res.Info) == 0 {
		return nil
	}

	return fmt.Errorf("got unexpected Info string: %s", res.Info)
}

func (c *client) callSweepAll(to string, accountIdx uint64) (*wallet.SweepAllResponse, error) {
	return c.rpc.Wallet.SweepAll(&wallet.SweepAllRequest{
		AccountIndex: accountIdx,
		Address:      to,
	})
}

func (c *client) callTransfer(destinations []wallet.Destination, accountIdx uint64) (*wallet.TransferResponse, error) {
	return c.rpc.Wallet.Transfer(&wallet.TransferRequest{
		Destinations: destinations,
		AccountIndex: accountIdx,
		Priority:     0,
	})
}

func (c *client) callGetBalance(idx uint64) (*wallet.GetBalanceResponse, error) {
	return c.rpc.Wallet.GetBalance(&wallet.GetBalanceRequest{
		AccountIndex: idx,
	})
}

func (c *client) callGetAddress(idx uint64) (*wallet.GetAddressResponse, error) {
	return c.rpc.Wallet.GetAddress(&wallet.GetAddressRequest{
		AccountIndex: idx,
	})
}

func (c *client) callGetAccounts() (*wallet.GetAccountsResponse, error) {
	return c.rpc.Wallet.GetAccounts(&wallet.GetAccountsRequest{})
}

func (c *client) callOpenWallet(filename, password string) error {
	return c.rpc.Wallet.OpenWallet(&wallet.OpenWalletRequest{
		Filename: filename,
		Password: password,
	})
}

func (c *client) callCreateWallet(filename, password string) error {
	return c.rpc.Wallet.CreateWallet(&wallet.CreateWalletRequest{
		Filename: filename,
		Password: password,
		Language: "English",
	})
}

func (c *client) callGetHeight() (uint64, error) {
	res, err := c.rpc.Wallet.GetHeight()
	if err != nil {
		return 0, err
	}
	return res.Height, nil
}
