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

        checkConnection();
})

export function connectAccount() {
	if (!window.ethereum) return

    window.ethereum
        .request({ method: 'eth_requestAccounts'})
        .then(handleAccountsChanged)
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

export const sign = async (msg: string) => {
	if(!window.ethereum){
		console.error('no window.ethereum')
		return
	}

	const ethersProvider = new providers.Web3Provider(window.ethereum, 'any');
	const tx = JSON.parse(msg)
	const signer = ethersProvider.getSigner()
	console.log('signer...', signer)
	let value

	if (tx.value != "") {
		value = utils.parseEther(tx.value)
	}

	let params = 
	  {
	    from: signer.getAddress(),
	    to: tx.to,
	    gasPrice: ethersProvider.getGasPrice(), 
	    gasLimit: "200000",
	    value,
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
