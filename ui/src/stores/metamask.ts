import {providers, Contract, utils} from "ethers"
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

let ethersProvider
let chainId
let contract
let signer
let from

const initialize = async () => {
	ethersProvider = new providers.Web3Provider(window.ethereum, 'any');
	window.ethersProvider = ethersProvider
	signer = ethersProvider.getSigner()
	console.log("signer:", await signer.getAddress())
}

export const sign = async(msg) => {
	let tx = JSON.parse(msg)
	let value
	if (tx.value != "") {
		value = utils.parseEther(tx.value)
	}

	let params = 
	  {
	    from: signer.getAddress(),
	    to: tx.to,
	    gasPrice: window.ethersProvider.getGasPrice(), 
	    value: value,
	    data: tx.data,
	  }

	console.log("sending transaction:", params)
	let res
	try {
	 	res = await signer.sendTransaction(params)		
		console.log(res)
	} catch (e) {
		console.error("tx failed", e)
		return ""
	}

	return res.hash
}