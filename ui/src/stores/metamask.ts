import {providers, Contract} from "ethers"
import detectEthereumProvider from "@metamask/detect-provider"
import { writable } from 'svelte/store';
import SwapFactory from "../../../ethereum/artifacts/contracts/SwapFactory.sol/SwapFactory.json"

ethereum.on('chainChanged', (_chainId) => window.location.reload());
ethereum.on('accountsChanged', handleAccountsChanged);

export const currentAccount = writable(null);

export const connectAccount = async () => {
	const provider = (await detectEthereumProvider()) as any
	if (provider) {
		startApp(provider);
		await provider.request({ method: "eth_accounts" }).then(handleAccountsChanged)
		  .catch((err) => {
		    // Some unexpected error.
		    // For backwards compatibility reasons, if no accounts are available,
		    // eth_accounts will return an empty array.
		    console.error(err);
		  });
		  await initialize()
	} else {
		console.error("Metamask is not installed")
	}
}

function startApp(provider) {
  // If the provider returned by detectEthereumProvider is not the same as
  // window.ethereum, something is overwriting it, perhaps another wallet.
  if (provider !== window.ethereum) {
    console.error('Do you have multiple wallets installed?');
  }
  // Access the decentralized web!
}

// Note that this event is emitted on page load.
// If the array of accounts is non-empty, you're already
// connected.
ethereum.on('accountsChanged', handleAccountsChanged);

// For now, 'eth_accounts' will continue to always return an array
function handleAccountsChanged(accounts) {
  if (accounts.length === 0) {
    // MetaMask is locked or the user has not connected any accounts
    console.log('Please connect to MetaMask.');
  } else if (accounts[0] !== currentAccount) {
    currentAccount.set(accounts[0]);
   	console.log(currentAccount)
   	console.log(accounts[0])
    // Do any other work!
  }
}

const swapContractAddrs = {
	goerli: "0xe532f0C720dCD102854281aeF1a8Be01f464C8fE",
}

let ethersProvider
let chainId
let contract
let signer

const initialize = async () => {
	ethersProvider = new providers.Web3Provider(window.ethereum, 'any');
	signer = ethersProvider.getSigner()
	console.log("signer:", await signer.getAddress())
	chainId = await ethereum.request({ method: 'eth_chainId' });
	if (chainId == 5) {
		contract = new Contract(swapContractAddrs.goerli, SwapFactory.abi).connect(signer)
		console.log("instantiated contract on Goerli at", swapContractAddrs.goerli)
	}
}

export const newSwap = async() => {
	let tx = await contract.new_swap();
	console.log(tx)
}