import {providers, utils} from "ethers"
import detectEthereumProvider from "@metamask/detect-provider"
import { writable } from 'svelte/store';

export const currentAccount = writable("");

// detect provider using @metamask/detect-provider
detectEthereumProvider()
.then((provider) => {
    if (!provider) {
		console.log('Please install MetaMask!');
return
	}
        provider.on('accountsChanged', handleAccountsChanged);
		provider.on('chainChanged', () => window.location.reload());

        //connect btn is initially disabled
        // $('#connect-btn').addEventListener('click', connectAccount);
        checkConnection();
})

export function connectAccount() {
	if (!window.ethereum) return

    window.ethereum
        .request({ method: 'eth_requestAccounts'})
        .then(handleAccountsChanged)
		.then(initialize)
        .catch((err: any) => {
            if (err.code === 4001) {
                console.log('Please connect to MetaMask.');
            } else {
                console.error(err);
            }
        });
}

function checkConnection() {
	if (!window.ethereum) return

    window.ethereum
        .request({ method: 'eth_accounts' })
        .then(handleAccountsChanged)
        .catch(console.error);
}


const handleAccountsChanged = (accounts: string[]) => {
  if (accounts.length === 0) {
    // MetaMask is locked or the user has not connected any accounts
    console.log('Please connect to MetaMask.');
	return 
  }

  currentAccount.set(accounts[0]);
}

const initialize = async () => {
	if (!window.ethersProvider) return

	const ethersProvider = new providers.Web3Provider(window.ethereum, 'any');
	window.ethersProvider = ethersProvider
}

export const sign = async (msg: string) => {
	const tx = JSON.parse(msg)
	const signer = window.ethersProvider.getSigner()
	let value

	if (tx.value != "") {
		value = utils.parseEther(tx.value)
	}

	let params = 
	  {
	    from: signer.getAddress(),
	    to: tx.to,
	    gasPrice: window.ethersProvider.getGasPrice(), 
	    gasLimit: "200000",
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
