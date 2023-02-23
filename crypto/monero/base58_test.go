package mcrypto

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"
)

// Test addresses were taken from here with the integrated addresses, which we do not want
// to support, removed:
// https://github.com/monero-project/monero/blob/v0.18.1.0/tests/functional_tests/validate_address.py#L68-L71
// Hex values were computed here:
// https://xmr.llcoins.net/addresstests.html
var addressEncodingTests = []struct {
	name        string      // Address description
	network     Network     // Mainnet, Stagenet, Testnet
	addressType AddressType // Standard, Integrated, Subaddress
	addressHex  string      // Base16 encoded address
	address     string      // Base58 encoded address
}{
	{
		name:        "mainnet primary (1)",
		network:     Mainnet,
		addressType: Standard,
		addressHex:  "121b3bd040020d3712ab84992b773d0a965134eb2df0392fb84af95de8a17be2ab231c9bf8341c6a870d92e3fb98063a90a355fb8dbf74a8561b9d7f9273247e9956e6f1b0", //nolint:lll
		address:     "42ey1afDFnn4886T7196doS9GPMzexD9gXpsZJDwVjeRVdFCSoHnv7KPbBeGpzJBzHRCAs9UxqeoyFQMYbqSWYTfJJQAWDm",
	},
	{
		name:        "mainnet primary (2)",
		network:     Mainnet,
		addressType: Standard,
		addressHex:  "1247335f3ceae62690c602dc20cdb6bd461dfb409f7322844e0092dbb4000c796cb8954f72ccc4bf16d93600d0cfba6be32def0ca114bf7147c20c42769bef4cfc605c9bed", //nolint:lll
		address:     "44Kbx4sJ7JDRDV5aAhLJzQCjDz2ViLRduE3ijDZu3osWKBjMGkV1XPk4pfDUMqt1Aiezvephdqm6YD19GKFD9ZcXVUTp6BW",
	},
	{
		name:        "testnet primary",
		network:     Testnet,
		addressType: Standard,
		addressHex:  "3543ca04c0bac1fee7087d0779959c89c773e1d4d4a477f2a2316cb431018ee955dd951a02750dcaa7af680fd3fd148331cd980eda5e1d881d00bf1e35865f40052e237e80", //nolint:lll
		address:     "9ujeXrjzf7bfeK3KZdCqnYaMwZVFuXemPU8Ubw335rj2FN1CdMiWNyFV3ksEfMFvRp9L9qum5UxkP5rN9aLcPxbH1au4WAB",
	},
	{
		name:        "stagenet primary",
		network:     Stagenet,
		addressType: Standard,
		addressHex:  "18365f7c1aa6cc01def62e128fffd8e1035d64cea20211b5b85e313737f28e14941bbc4ca71a085b5bb8390ab800b53e81be2abab23f63740ef8a450804c6de96fe16e7338", //nolint:lll
		address:     "53teqCAESLxeJ1REzGMAat1ZeHvuajvDiXqboEocPaDRRmqWoVPzy46GLo866qRFjbNhfkNckyhST3WEvBviDwpUDd7DSzB",
	},
	{
		name:        "mainnet subaddress (1)",
		network:     Mainnet,
		addressType: Subaddress,
		addressHex:  "2ade13d5e57591933d61237e94bfca6b3fd239c9d53c5582a592eeeb8d8986c71b0d4d160c2c2ef02c3d0dd3b8646fbbef9dadc5d54002e69cb78b74ace989510210defc2a", //nolint:lll
		address:     "8AsN91rznfkBGTY8psSNkJBg9SZgxxGGRUhGwRptBhgr5XSQ1XzmA9m8QAnoxydecSh5aLJXdrgXwTDMMZ1AuXsN1EX5Mtm",
	},
	{
		name:        "mainnet subaddress (2)",
		network:     Mainnet,
		addressType: Subaddress,
		addressHex:  "2a71521fb4561485775aa8e3b2398e8b7e9a6dcfb70da20abb6f8dae06d43a3ab28c3572ef0c82ae42b38586c7404e1587373aef49d6d94ef190e8feec603fd1dc7643061d", //nolint:lll
		address:     "86kKnBKFqzCLxtK1Jmx2BkNBDBSMDEVaRYMMyVbeURYDWs8uNGDZURKCA5yRcyMxHzPcmCf1q2fSdhQVcaKsFrtGRsdGfNk",
	},
	{
		name:        "testnet subaddress (1)",
		network:     Testnet,
		addressType: Subaddress,
		addressHex:  "3f87b78ea79e0d8542392dd48f5f7bf81ef674947434bb8035ee96b69947af9c61fa2829886abb638b9fdbc89f17b4e28be1c47dc2d89c7a5ab09cf4d0fa465ea4c6ec3a0b", //nolint:lll
		address:     "BdKg9udkvckC5T58a8Nmtb6BNsgRAxs7uA2D49sWNNX5HPW5Us6Wxu8QMXrnSx3xPBQQ2iu9kwEcRGAoiz6EPmcZKbF62GS",
	},
	{
		name:        "testnet subaddress (2)",
		network:     Testnet,
		addressType: Subaddress,
		addressHex:  "3f6b9ed65b32362dacaa7c48c8a7f82526d96f599897fefd7f04cb8b7fd4ea5e2658a1145569ef5bf0c949048a6ede19484f1221a2ba01df6e83e95741d2dd0fbc69e7a9ec", //nolint:lll
		address:     "BcFvPa3fT4gVt5QyRDe5Vv7VtUFao9ci8NFEy3r254KF7R1N2cNB5FYhGvrHbMStv4D6VDzZ5xtxeKV8vgEPMnDcNFuwZb9",
	},
	{
		name:        "stagenet subaddress (1)",
		network:     Stagenet,
		addressType: Subaddress,
		addressHex:  "241ce1f6265ca2c855d32c7160ee0c648e108b11323d8edb1f808e4c7b4c6c7d8911e6a21cc33534e41f2592beafd78e1e7b30fe95c596dafdd869a7fd15f034c09cb4dcb9", //nolint:lll
		address:     "73LhUiix4DVFMcKhsPRG51QmCsv8dYYbL6GcQoLwEEFvPvkVvc7BhebfA4pnEFF9Lq66hwvLqBvpHjTcqvpJMHmmNjPPBqa",
	},
	{
		name:        "stagenet subaddress (2)",
		network:     Stagenet,
		addressType: Subaddress,
		addressHex:  "24ccc5703d9109e9c619bc427e9874f740ce43c25e5466e743e1cc4a6cf6d4908f3c79ff40b5b8fb281e7b379a652c36e0b74129684f43473be6cac960f124b9fe5d74bcfa", //nolint:lll
		address:     "7A1Hr63MfgUa8pkWxueD5xBqhQczkusYiCMYMnJGcGmuQxa7aDBxN1G7iCuLCNB3VPeb2TW7U9FdxB27xKkWKfJ8VhUZthF",
	},
}

func TestMoneroAddrBytesToBase58(t *testing.T) {
	for _, tt := range addressEncodingTests {
		addrBytes, err := hex.DecodeString(tt.addressHex)
		require.NoError(t, err, tt.name)
		addr := addrBytesToBase58(addrBytes)
		require.Equal(t, tt.address, addr, tt.name)
	}
}

func TestMoneroAddrBase58ToBytes(t *testing.T) {
	for _, tt := range addressEncodingTests {
		addrBytes, err := addrBase58ToBytes(tt.address)
		require.NoError(t, err, tt.name)
		addrHex := hex.EncodeToString(addrBytes)
		require.Equal(t, tt.addressHex, addrHex, tt.name)
	}
}

func TestMoneroAddrBase58ToBytes_BadLength(t *testing.T) {
	_, err := addrBase58ToBytes("")
	require.ErrorIs(t, err, errInvalidAddressLength)
}

func TestMoneroAddrBase58ToBytes_BadEncoding(t *testing.T) {
	_, err := addrBase58ToBytes("")
	require.ErrorIs(t, err, errInvalidAddressLength)

	// Different code handles invalid encoding in the last block, so we add a non-valid base58
	// character in both the first and last block
	validAddr := "42ey1afDFnn4886T7196doS9GPMzexD9gXpsZJDwVjeRVdFCSoHnv7KPbBeGpzJBzHRCAs9UxqeoyFQMYbqSWYTfJJQAWDm"
	firstBlockBad := "l" + validAddr[1:]
	lastBlockBad := validAddr[:encodedAddressLen-1] + "l"

	_, err = addrBase58ToBytes(firstBlockBad)
	require.ErrorIs(t, err, errInvalidAddressEncoding)

	_, err = addrBase58ToBytes(lastBlockBad)
	require.ErrorIs(t, err, errInvalidAddressEncoding)
}
