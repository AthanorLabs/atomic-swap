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

const swapContractAddrs = {
	goerli: "0xe532f0C720dCD102854281aeF1a8Be01f464C8fE",
	dev: "0xe78A0F7E598Cc8b0Bb87894B0F60dD2a88d6a8Ab",
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
	chainId = await ethereum.request({ method: 'eth_chainId' });
	if (chainId == 5) {
		contract = new Contract(swapContractAddrs.goerli, SwapFactory.abi).connect(signer)
		console.log("instantiated contract on Goerli at", swapContractAddrs.goerli)
	} else if (chainId == 1337) {
		contract = new Contract(swapContractAddrs.dev, SwapFactory.abi).connect(signer)
	}
}

export const sign = async(msg) => {
	let tx = JSON.parse(msg)
	let value
	if tx.value != "" {
		value = utils.parseEther(tx.value)
	}

	let params = 
	  {
	    from: signer.getAddress(),
	    to: tx.to,
	    gasPrice: window.ethersProvider.getGasPrice(), // 10000000000000
	    value: value,
	    data: tx.data,
	  }
	console.log("sending tx request...")
	// let res = await window.ethereum.request({
	// 	method: "eth_sendTransaction",
	// 	params,
	// })
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