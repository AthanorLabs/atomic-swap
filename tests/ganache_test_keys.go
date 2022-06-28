package tests

import (
	"runtime"
	"strings"
	"testing"
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
var testPackageNames = []string{
	"cmd/daemon",
	"protocol/backend",
	"protocol/xmrmaker",
	"protocol/xmrtaker",
	"recover",
	"swapfactory",
}

const (
	ethKeysPerPackage = 2
	repoName          = "github.com/noot/atomic-swap/"
)

// `ganache-cli --deterministic --accounts=20` provides the following keys with
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
}

func init() {
	if len(testPackageNames)*ethKeysPerPackage > len(ganacheTestKeys) {
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

func getPackageIndex(t *testing.T) uint {
	// Determine the test package that requested the key from the call stack
	pc, _, _, ok := runtime.Caller(2) // skipping this function and GetMakerTestKey/GetTakerTestKey
	if !ok {
		t.Fatalf("Failed to get caller info")
	}
	// returns the package and function name from the program counter
	// example: "github.com/noot/atomic-swap/protocol/xmrtaker.newBackend"
	packageName := minPackageName(t, runtime.FuncForPC(pc).Name())

	for i, name := range testPackageNames {
		if name == packageName {
			return uint(i)
		}
	}
	t.Fatalf("Package %q does not have reserved test keys", packageName)
	panic("unreachable code")
}

// GetMakerTestKey returns a unique Ethereum/ganache maker key per test package
func GetMakerTestKey(t *testing.T) string {
	return ganacheTestKeys[getPackageIndex(t)*ethKeysPerPackage]
}

// GetTakerTestKey returns a unique Ethereum/ganache taker key per test package
func GetTakerTestKey(t *testing.T) string {
	return ganacheTestKeys[getPackageIndex(t)*ethKeysPerPackage+1]
}
