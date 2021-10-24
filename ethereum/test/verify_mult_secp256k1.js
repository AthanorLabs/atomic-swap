const { expect } = require("chai");

describe("Swap", function () {
  const s_a = 'D30519BCAE8D180DBFCC94FE0B8383DC310185B0BE97B4365083EBCECCD75759';
  const pubKey_a_x = '3AF1E1EFA4D1E1AD5CB9E3967E98E901DAFCD37C44CF0BFB6C216997F5EE51DF';
  const pubKey_a_y = 'E4ACAC3E6F139E0C7DB2BD736824F51392BDA176965A1C59EB9C3C5FF9E85D7A';
  const s_b = 'ebb2c082fd7727890a28ac82f6bdf97bad8de9f5d7c9028692de1a255cad3e0f';
  const pubKey_b_x = '779dd197a5df977ed2cf6cb31d82d43328b790dc6b3b7d4437a427bd5847dfcd';
  const pubKey_b_y = 'e94b724a555b6d017bb7607c3e3281daf5b1699d6ef4124975c9237b917d426f';
  const s_r = 'af416cb5879aa89e8cd19567142186b4d3003c4c37611fd1b4f4f9de7e77d60a';
  const pubKey_r_x = '0197329ba8956982ec141a45caedcba69d3289502be521f7f69a0bb3fbdea061';
  const pubKey_r_y = '2b28c2a0615acb9c41101e610d0adf6be0ed86db8da27470b16582d898646fb3';

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

    const EC = await ethers.getContractFactory("EC");
    const ec = await EC.deploy();
    await ec.deployed();

    await ec.publicKeyVerify(hexToBytes(s_a), hexToBytes(pubKey_a_x), hexToBytes(pubKey_a_y));
    await ec.publicKeyVerify(hexToBytes(s_b), hexToBytes(pubKey_b_x), hexToBytes(pubKey_b_y));
    await ec.publicKeyVerify(hexToBytes(s_r), hexToBytes(pubKey_r_x), hexToBytes(pubKey_r_y));
  });
});
