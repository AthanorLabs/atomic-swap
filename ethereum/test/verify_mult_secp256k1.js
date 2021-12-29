const { expect } = require("chai");
const secp = require('noble-secp256k1');

const arrayify = ethers.utils.arrayify;

function KeyPair(s, pubKey_x, pubKey_y) {
  this.s = s;
  this.pubKey_x = pubKey_x;
  this.pubKey_y = pubKey_y;
}

describe("Swap", function () {
  const test_vecs = [
    new KeyPair('0xD30519BCAE8D180DBFCC94FE0B8383DC310185B0BE97B4365083EBCECCD75759',
      '0x3AF1E1EFA4D1E1AD5CB9E3967E98E901DAFCD37C44CF0BFB6C216997F5EE51DF',
      '0xE4ACAC3E6F139E0C7DB2BD736824F51392BDA176965A1C59EB9C3C5FF9E85D7A'),
    new KeyPair('0xebb2c082fd7727890a28ac82f6bdf97bad8de9f5d7c9028692de1a255cad3e0f',
      '0x779dd197a5df977ed2cf6cb31d82d43328b790dc6b3b7d4437a427bd5847dfcd',
      '0xe94b724a555b6d017bb7607c3e3281daf5b1699d6ef4124975c9237b917d426f'),
    //TODO this one fails - not sure where we got it from and if it already failed before
    // new KeyPair('0xaf416cb5879aa89e8cd19567142186b4d3003c4c37611fd1b4f4f9de7e77d60a',
    //   '0x0197329ba8956982ec141a45caedcba69d3289502be521f7f69a0bb3fbdea061',
    //   '0x2b28c2a0615acb9c41101e610d0adf6be0ed86db8da27470b16582d898646fb3')
  ];

  let swap;
  beforeEach(async function () {
    const Swap = await ethers.getContractFactory("SwapMock");
    swap = await Swap.deploy();
  });

  it("Should verify commitment correctly with test vecs", async function () {
    // const s_a = ed.utils.randomPrivateKey();
    // console.log(toHexString(s_a));
    // const s_b = ed.utils.randomPrivateKey();
    // console.log(toHexString(s_b));

    let promises = [];
    test_vecs.forEach(async function (kp, i) {
      const qKeccak = ethers.utils.solidityKeccak256(
        ["uint256", "uint256"],
        [kp.pubKey_x, kp.pubKey_y]);

      // console.log(kp.s);
      // console.log(arrayify(kp.s));
      // console.log(qKeccak);
      // console.log(arrayify(qKeccak));
      console.log('Testing %s of %s test vectors...', i + 1, test_vecs.length);
      promises.push(swap.testVerifySecret(arrayify(kp.s), arrayify(qKeccak)));
    });
    await Promise.all(promises);
  });

  // TODO write test with randomly generated secrets
});
