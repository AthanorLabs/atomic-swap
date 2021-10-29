const { expect } = require("chai");
const rewire = require('rewire');
const ed = rewire("noble-ed25519"); // export all functions

var bytesToNumberLE = ed.__get__('bytesToNumberLE');
var hexToBytes = ed.__get__('hexToBytes');


describe("SwapOnChain", function () {
  const s_1 = '34c388ea5bdd494b10b1eaebfa68564463947ee48efe69c520d4a1fadc550c04';
  const s_2 = 'a6e51afb9662bf2173d807ceaf88938d09a82d1ab2cea3eeb1706eeeb8b6ba03';
  const pubKey_1 = 'e32ad36ce8e59156aa416da9c9f41a7abc59f6b5f1dd1c2079e8ff4e14503090';
  const pubKey_2 = '57edf916a28c2a0a172565468564ab1c5c217d33ea63436f7604a96aa28ec471';

  it("Should generate public key correctly with test vectors", async function () {
    const Swap = await ethers.getContractFactory("SwapOnChainConsole");
    // Constructor arguments are ignored in the test
    const swap = await Swap.deploy(Array(32).fill(0), Array(32).fill(0));
    await swap.deployed();

    await swap.verifySecret(hexToBytes(s_1).reverse(), hexToBytes(pubKey_1).reverse());
    await swap.verifySecret(hexToBytes(s_2).reverse(), hexToBytes(pubKey_2).reverse());
  });

  it("Should generate public key correctly with random secret", async function () {
    const Swap = await ethers.getContractFactory("SwapOnChainConsole");
    // Constructor arguments are ignored in the test
    const swap = await Swap.deploy(Array(32).fill(0), Array(32).fill(0));
    await swap.deployed();

    for (let i = 0; i < 10; i++) {
      const s = ed.utils.randomPrivateKey();
      const pK = ed.Point.BASE.multiply(bytesToNumberLE(s)).toRawBytes();

      await swap.verifySecret(s.reverse(), pK.reverse());
    }
  });
});
