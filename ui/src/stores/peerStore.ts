import { rpcRequest } from '../utils';
import { writable } from 'svelte/store';
import type { NetDiscoverResult } from '../types/NetDiscoverResults';
import { sanitizeAddresses } from '../utils';

export const isLoadingPeers = writable(false)
export const peers = writable<string[]>([], () => {
    getPeers()
});

export const getPeers = () => {
    isLoadingPeers.set(true)
    return rpcRequest<NetDiscoverResult>('net_discover', { searchTime: 3 })
        .then(({ result }) => {
            const sanitizePeers = sanitizeAddresses(result.peers)
            peers.set(sanitizePeers)
        })
        .catch(console.error)
        .finally(() => {
            isLoadingPeers.set(false)
        })
}