package tests

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"runtime"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/stretchr/testify/require"

	"github.com/noot/atomic-swap/common"
)

/*
 * Golang packages can be tested in parallel, but in our case, they are sharing a
 * common ganache Ethereum simulator. To avoid transaction conflicts, we allocate
 * each package listed in `testPackageNames` below 2 keys. One taker and one maker
 * (or however the package chooses to use the keys). When different packages are tested
 * in parallel, there is no shared global state, so we can't use any kind of pool or
 * non-deterministic map traversals.
 */

// testPackageNames is the list of packages with _test.go code that requires access to
// one or more prefunded ganache wallets.
var testPackages = []struct {
	name    string
	numKeys int
}{
	{"cmd/daemon", 2},
	{"ethereum/block", 2},
	{"protocol/backend", 2},
	{"protocol/xmrmaker", 2},
	{"protocol/xmrtaker", 2},
	{"recover", 2},
	{"swapfactory", 16},
}

const (
	repoName = "github.com/noot/atomic-swap/"
)

// `ganache --deterministic --accounts=50` provides the following keys with
// 100 ETH on startup. The first 2 keys can be found in const.go and reserved
// for use in non-test files (files without the _test.go suffix).
var ganacheTestKeys = []string{
	"6370fd033278c143179d81c5526140625662b8daa446c22ee2d73db3707e620c", // ganache key #2
	"646f1ce2fdad0e6deeeb5c7e8e5543bdde65e86029e2fd9fc169899c440a7913", // ganache key #3
	"add53f9a7e588d003326d1cbf9e4a43c061aadd9bc938c843a79e7b4fd2ad743", // ganache key #4
	"395df67f0c2d2d9fe1ad08d1bc8b6627011959b79c53d7dd6a3536a33ab8a4fd", // ganache key #5
	"e485d098507f54e7733a205420dfddbe58db035fa577fc294ebd14db90767a52", // ganache key #6
	"a453611d9419d0e56f499079478fd72c37b251a94bfde4d19872c44cf65386e3", // ganache key #7
	"829e924fdf021ba3dbbc4225edfece9aca04b929d6e75613329ca6f1d31c0bb4", // ganache key #8
	"b0057716d5917badaf911b193b12b910811c1497b5bada8d7711f758981c3773", // ganache key #9
	"77c5495fbb039eed474fc940f29955ed0531693cc9212911efd35dff0373153f", // ganache key #10
	"d99b5b29e6da2528bf458b26237a6cf8655a3e3276c1cdc0de1f98cefee81c01", // ganache key #11
	"9b9c613a36396172eab2d34d72331c8ca83a358781883a535d2941f66db07b24", // ganache key #12
	"0874049f95d55fb76916262dc70571701b5c4cc5900c0691af75f1a8a52c8268", // ganache key #13
	"21d7212f3b4e5332fd465877b64926e3532653e2798a11255a46f533852dfe46", // ganache key #14
	"47b65307d0d654fd4f786b908c04af8fface7710fc998b37d219de19c39ee58c", // ganache key #15
	"66109972a14d82dbdb6894e61f74708f26128814b3359b64f8b66565679f7299", // ganache key #16
	"2eac15546def97adc6d69ca6e28eec831189baa2533e7910755d15403a0749e8", // ganache key #17
	"2e114163041d2fb8d45f9251db259a68ee6bdbfd6d10fe1ae87c5c4bcd6ba491", // ganache key #18
	"ae9a2e131e9b359b198fa280de53ddbe2247730b881faae7af08e567e58915bd", // ganache key #19
	"d09ba371c359f10f22ccda12fd26c598c7921bda3220c9942174562bc6a36fe8", // ganache key #20
	"2d2719c6a828911ed0c50d5a6c637b63353e77cf57ea80b8e90e630c4687e9c5", // ganache key #21
	"d353907ab062133759f149a3afcb951f0f746a65a60f351ba05a3ebf26b67f5c", // ganache key #22
	"971c58af72fd8a158d4e654cfbe98f5de024d28547005909684f58c9c46a25c4", // ganache key #23
	"85d168288e7fcf84b1841e447fc7945b1e27bfe9a3776367079a6427405eac66", // ganache key #24
	"f3da3ac70552606ed09d16dd2808c924826094f0c5cbfcb4f2e0e1cfc70ff8dd", // ganache key #25
	"bf20e9c05d70ce59a6b125eab3b4122eb75044a33749c4c5a77e3b0b86fa091e", // ganache key #26
	"647442126fdb80c6aec75a0d75a6fe1b31a4e204d29a2c446f550c4115cac139", // ganache key #27
	"ef78746d079c9d72d2e9a3c10447d1d4aaae6a51541d0296da4fc9ec7e060aff", // ganache key #28
	"c95286117cd74213417aeca52118ccd03ec240582f0a9a3e4ef7b434523179f3", // ganache key #29
	"21118f9a6de181061a2abd549511105adb4877cf9026f271092e6813b7cf58ab", // ganache key #30
	"1166189cdf129cdcb011f2ad0e5be24f967f7b7026d162d7c36073b12020b61c", // ganache key #31
	"1aa14c63d481dcc1185a654eb52c9c0749d07ac8f30ef17d45c3c391d9bf68eb", // ganache key #32
	"4a23fe455a34bb47f8f3282a4f6d36c22987275f0bb9aacb251568df7d038385", // ganache key #33
	"2450bb2893d0bddf92f4ac88cb65a8e94b56e89f7ec3e46c9c88b2b46ebe3ca5", // ganache key #34
	"f934aded8693d6b2b61ccbb3bc1f86a86afbbd8622a5eb3401b2f8de9863b07b", // ganache key #35
	"c8eea9d162fe9d6852afa0d55ebe6b14b8e6fc9b0e93ae13209e2b4db48a6482", // ganache key #36
	"be146cdb15d4069e0249da35c928819cbde563dd4fe3d1ccfeda7885a52e0754", // ganache key #37
	"74ae0c3d566d7e73613d4ebb814b0f37a2d040060814f75e115d28469d22f4c2", // ganache key #38
	"b2b19df163d1f952df31e32c694d592e530c0b3d54c6276015bc9b0acaf982de", // ganache key #39
	"86117111fcb34df8d0e58505969021b9308513c6e94d16172f0c8789a7130a43", // ganache key #40
	"dcb8686c211c231be763f0a95cc02227a707643fd2631bda99fcdbd03cd9ca3d", // ganache key #41
	"b74ffec4abd7e93889196054d5e6ed8ea9c1c3314e77a74c00f851c47f5268fd", // ganache key #42
	"ba30972105ec13423116d2e5c11a8d282805ac3654bb4c1c2f5fa63f4da42dad", // ganache key #43
	"87ad1798a2d32434f72598575237528a435416da1bdc900025c415903647957e", // ganache key #44
	"5d4af11a54d4a5196b0073ba26a1114cb113e1339d9354c8165b8e181c89cad9", // ganache key #45
	"a03bf2b145b0154c2e788a1d4642d235f6ff1c8aceeb41d0d7232525da8bdb77", // ganache key #46
	"b1f4063952ebc0785bbc201520ed7f0c5fc15298099e60e62f8cfa456bbc2705", // ganache key #47
	"41d647879d53baddb93cfadc3f5ef4d5bdc330bec4b4ef9caace19c70a385856", // ganache key #48
	"87c546d6cb8ec705bea47e2ab40f42a768b1e5900686b0cecc68c0e8b74cd789", // ganache key #49
}

func init() {
	totalKeys := 0
	for _, pkg := range testPackages {
		totalKeys += pkg.numKeys
	}
	if totalKeys > len(ganacheTestKeys) {
		panic("Insufficient ganache test keys")
	}
}

// minPackageName takes a long-form package+function name (example:
// "github.com/noot/atomic-swap/protocol/xmrtaker.newBackend") and returns
// just the package name without the repository prefix ("protocol/xmrtaker").
func minPackageName(t *testing.T, pkgAndFunc string) string {
	minPkgAndFunc := strings.TrimPrefix(pkgAndFunc, repoName)
	if minPkgAndFunc == pkgAndFunc {
		t.Fatalf("%q does not have the repo prefix %q", pkgAndFunc, repoName)
	}
	// with the domain name gone, the minimal package is everything before the first period.
	return strings.Split(minPkgAndFunc, ".")[0]
}

func getCallingPackageName(t *testing.T) string {
	// Determine the test package that requested the key from the call stack. We skip 2 callers
	// (1) this function and (2) the public function from this package that invoked it.
	pc, _, _, ok := runtime.Caller(2)
	if !ok {
		t.Fatalf("Failed to get caller info")
	}
	fullPackageName := runtime.FuncForPC(pc).Name()
	return minPackageName(t, fullPackageName)
}

func getPackageKeys(t *testing.T, packageName string) []string {
	startIndex := 0
	for _, pkg := range testPackages {
		if pkg.name == packageName {
			return ganacheTestKeys[startIndex : startIndex+pkg.numKeys]
		}
		startIndex += pkg.numKeys
	}
	t.Fatalf("Package %q does not have reserved test keys", packageName)
	panic("unreachable code")
}

func getPackageTestKey(t *testing.T, pkgName string, index int) *ecdsa.PrivateKey {
	keys := getPackageKeys(t, pkgName)
	require.Lessf(t, index, len(keys), "insufficient keys allocated to package %q", pkgName)
	pk, err := ethcrypto.HexToECDSA(keys[index])
	require.NoError(t, err)
	return pk
}

// GetTestKeyByIndex returns the ganache test key allocated to a package by index
func GetTestKeyByIndex(t *testing.T, index int) *ecdsa.PrivateKey {
	pkgName := getCallingPackageName(t)
	return getPackageTestKey(t, pkgName, index)
}

// GetMakerTestKey returns the first ganache test key allocated to a package
func GetMakerTestKey(t *testing.T) *ecdsa.PrivateKey {
	pkgName := getCallingPackageName(t)
	return getPackageTestKey(t, pkgName, 0)
}

// GetTakerTestKey returns the second ganache test key allocated to a package
func GetTakerTestKey(t *testing.T) *ecdsa.PrivateKey {
	pkgName := getCallingPackageName(t)
	return getPackageTestKey(t, pkgName, 1)
}

// NewEthClient returns a connection to the local ganache instance for unit tests along
// with its chain ID. The connection is automatically closed when the test completes.
func NewEthClient(t *testing.T) (*ethclient.Client, *big.Int) {
	ec, err := ethclient.Dial(common.DefaultEthEndpoint)
	require.NoError(t, err)
	t.Cleanup(func() {
		ec.Close()
	})
	chainID, err := ec.ChainID(context.Background())
	require.NoError(t, err)
	return ec, chainID
}

// MineTransaction is a test helper that blocks until the transaction is included in a block
// and returns the receipt. Errors are checked including the status.
func MineTransaction(t *testing.T, ec bind.DeployBackend, tx *ethtypes.Transaction) *ethtypes.Receipt {
	ctx := context.Background() // Create a MineTransactionWithCtx if a future test needs a custom context
	receipt, err := bind.WaitMined(ctx, ec, tx)
	require.NoError(t, err)
	require.Equal(t, ethtypes.ReceiptStatusSuccessful, receipt.Status) // Make sure the transaction was not reverted
	return receipt
}
