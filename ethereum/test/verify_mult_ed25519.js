const { expect } = require("chai");
const ed = require("noble-ed25519");

describe("Swap", function () {
  const s_a = '34c388ea5bdd494b10b1eaebfa68564463947ee48efe69c520d4a1fadc550c04';
  const s_b = 'a6e51afb9662bf2173d807ceaf88938d09a82d1ab2cea3eeb1706eeeb8b6ba03';
  const pubKey_a = 'e32ad36ce8e59156aa416da9c9f41a7abc59f6b5f1dd1c2079e8ff4e14503090';
  const pubKey_b = '57edf916a28c2a0a172565468564ab1c5c217d33ea63436f7604a96aa28ec471';

  function toHexString(byteArray) {
    return Array.from(byteArray, function (byte) {
      return ('0' + (byte & 0xFF).toString(16)).slice(-2);
    }).join('')
  }

  function hexToBytes(hex) {
    for (var bytes = [], c = 0; c < hex.length; c += 2)
      bytes.push(parseInt(hex.substr(c, 2), 16));
    return bytes;
  }

  it("Should generate public key correctly", async function () {
    // const s_a = ed.utils.randomPrivateKey();
    // console.log(toHexString(s_a));
    // const s_b = ed.utils.randomPrivateKey();
    // console.log(toHexString(s_b));

    const Swap = await ethers.getContractFactory("Swap");
    const swap = await Swap.deploy(hexToBytes(pubKey_b).reverse(), hexToBytes(pubKey_a).reverse());
    await swap.deployed();

    await swap.verifySecret(hexToBytes(s_a).reverse(), hexToBytes(pubKey_a).reverse());
    await swap.verifySecret(hexToBytes(s_b).reverse(), hexToBytes(pubKey_b).reverse());
  });

  it("Should selfdestruct", async function () {
    const Swap = await ethers.getContractFactory("Swap");
    const swap = await Swap.deploy(hexToBytes(pubKey_b).reverse(), hexToBytes(pubKey_a).reverse(), { value: 10 });
    await swap.deployed();

    await swap.set_ready();
    await swap.claim(hexToBytes(s_b).reverse());

    // TODO verify contract has actually self-destructed
  });
});
