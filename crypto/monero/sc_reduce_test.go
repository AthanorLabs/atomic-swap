package mcrypto

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_scReduce32(t *testing.T) {
	type testData struct {
		in  string
		out string
	}
	// Reference output values generated using:
	// https://github.com/FiloSottile/edwards25519/blob/v1.0.0/scalar.go#L618-L624
	// All but first test was generated using random (not carefully crafted) inputs.
	tests := []testData{
		{
			in:  "0200000000000000000000000000000000000000000000000000000000000000",
			out: "0200000000000000000000000000000000000000000000000000000000000000",
		},
		{
			in:  "d9726fbaa1b5bf7a6b0fa17f88ee006861f3c33a8583779f529d636209ae9b09",
			out: "d9726fbaa1b5bf7a6b0fa17f88ee006861f3c33a8583779f529d636209ae9b09",
		},
		{
			in:  "3c3930b33f8e1becc2a12f53014c12fe651f6abc086a143e7fdef57fecafb803",
			out: "3c3930b33f8e1becc2a12f53014c12fe651f6abc086a143e7fdef57fecafb803",
		},
		{
			in:  "6e8d8a38aca6afabdfc6bf257955a7a51312ee52d21944386e77dba019442417",
			out: "81b994db91439d53092ac8829a5bc8901312ee52d21944386e77dba019442407",
		},
		{
			in:  "8da6ad17326fa830abee9f9466774e94f1c8ef88941154d068912de100831a1c",
			out: "a0d2b7ba170c96d8d451a8f1877d6f7ff1c8ef88941154d068912de100831a0c",
		},
		{
			in:  "feb1df738ee6dbcc9cd9a3647462d133ac67fc5168c482d617645cf02989c3b2",
			out: "cf964f756ca41104671c0064e3a53c4eab67fc5168c482d617645cf02989c302",
		},
		{
			in:  "c1e9ffe2178a4c22204604210aba0ec1fd3a428266c034b2014d16aef9c14f8f",
			out: "594a51fb4471b9616d5f470915eb161afd3a428266c034b2014d16aef9c14f0f",
		},
		{
			in:  "8b55cb70013cbc01da9a0def3f0d75d20afa4ca158b8404a86bd712a024ffedf",
			out: "82924fb8aa33cd88f7a37aa8f15c22c309fa4ca158b8404a86bd712a024ffe0f",
		},
		{
			in:  "9d3bff8d6c1191853db579c103fbd94cc5290b13e17dbcd77ba8d248f26b285d",
			out: "fc1732bde82135cd0da5a392aa197fe4c4290b13e17dbcd77ba8d248f26b280d",
		},
		{
			in:  "42baac6367a725b2f7fb76e4fbcae5f76d392c23af4860613ac3da94e4e91395",
			out: "ed46081f7a2b80996e78c22928020f3c6d392c23af4860613ac3da94e4e91305",
		},
		{
			in:  "2131571b7bad5df1ae8e0e0ad2eb49f9feb6af9269256b4e1df858e23e963937",
			out: "5ab575042c8426e92bb8272136feacbafeb6af9269256b4e1df858e23e963907",
		},
		{
			in:  "03301a01ce5c53e56f18f78d370a8ee82b401061741eff62b96b7325eee27aa2",
			out: "c1e87f5fc67d9b7410f84a308547d8172b401061741eff62b96b7325eee27a02",
		},
		{
			in:  "a6287895d66e513c222cb1d90b5e9b302fe0edfaf3136481bd3ffdfe2d75559a",
			out: "51b5d350e9f2ab2399a8fc1e3895c4742ee0edfaf3136481bd3ffdfe2d75550a",
		},
		{
			in:  "5d2649904ee52c95026435860a344cf2a1139984669c33a058072565068df328",
			out: "837e5dd6191f08e5552a46404d408ec8a1139984669c33a058072565068df308",
		},
		{
			in:  "12eb235246da7ece9e930dde9a08dcd44ffd30192a3c968e4f4e00b3fa52b23a",
			out: "4b6f423bf7b047c61bbd26f5fe1a3f964ffd30192a3c968e4f4e00b3fa52b20a",
		},
		{
			in:  "fbe9ad967243c656435fc1aa918505be0a0edcc436d5e5910f2074af08862b96",
			out: "a676095285c7203ebadb0cf0bdbc2e020a0edcc436d5e5910f2074af08862b06",
		},
		{
			in:  "572725c72d8b4c1577252a1632db89bb0def79e48a676a3e3f4f9354df33bbfa",
			out: "74bcbd54a2bc38ece7f4a789263779820cef79e48a676a3e3f4f9354df33bb0a",
		},
		{
			in:  "02213c2c966c75c429618e68f63cfce2ec677e060a3c0db39029d8caa9a83f6c",
			out: "742979fef71907b423b4c096be61c265ec677e060a3c0db39029d8caa9a83f0c",
		},
		{
			in:  "656e3024dfad72559e1174e4e32b0695bc5488166692f04149f279889f98b002",
			out: "656e3024dfad72559e1174e4e32b0695bc5488166692f04149f279889f98b002",
		},
		{
			in:  "60ad31223758c82a69200665bb435576dec47927fa90c53a3ead8158227b053d",
			out: "9931500be82e9122e6491f7c1f56b837dec47927fa90c53a3ead8158227b050d",
		},
	}
	for _, tt := range tests {
		inBytes, err := hex.DecodeString(tt.in)
		require.NoError(t, err)
		var s [32]byte
		copy(s[:], inBytes)
		reduced := scReduce32(s)
		assert.Equal(t, tt.out, hex.EncodeToString(reduced[:]))
	}
}
