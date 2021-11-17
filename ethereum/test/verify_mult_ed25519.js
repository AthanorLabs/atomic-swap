const { expect } = require("chai");
const ed = require("noble-ed25519");

const arrayify = ethers.utils.arrayify;


describe("SwapOnChain", function () {
    const s_1 = '0x34c388ea5bdd494b10b1eaebfa68564463947ee48efe69c520d4a1fadc550c04';
    const s_2 = '0xa6e51afb9662bf2173d807ceaf88938d09a82d1ab2cea3eeb1706eeeb8b6ba03';
    const pubKey_1 = '0xe32ad36ce8e59156aa416da9c9f41a7abc59f6b5f1dd1c2079e8ff4e14503090';
    const pubKey_2 = '0x57edf916a28c2a0a172565468564ab1c5c217d33ea63436f7604a96aa28ec471';

    let swap;
    beforeEach(async function () {
        const Swap = await ethers.getContractFactory("SwapOnChainMock");
        swap = await Swap.deploy();
    });

    it("Should generate public key correctly with test vectors", async function () {
        console.log('Testing 1 of 2 test vectors...');
        await swap.testVerifySecret(arrayify(s_1).reverse(), arrayify(pubKey_1).reverse());
        console.log('Testing 2 of 2 test vectors...');
        await swap.testVerifySecret(arrayify(s_2).reverse(), arrayify(pubKey_2).reverse());
    });

    it("Should generate public key correctly with random secret", async function () {
        const n = 3;
        // pK derivation time is negligible vs. contract call
        for (let i = 0; i < n; i++) {
            const s = ed.utils.randomPrivateKey();
            const pK = ed.Point.BASE.multiply(ethers.BigNumber.from(s).toBigInt()).toRawBytes();

            console.log('Testing %s of %s randomly generated key pairs...', i + 1, n);
            await swap.testVerifySecret(s, pK.reverse());
        }

        // These do take time to verify but if everything is timing out, it's not the issue.
        // See https://stackoverflow.com/q/44149096 - not sure about a fix, I think
        // errors in any of the hardhat tests can cause all to fail that way :(
    });
});
