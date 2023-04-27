import { rpcRequest } from '../utils';
import { writable } from 'svelte/store';
import type { NetDiscoverResult } from '../types/NetDiscoverResults';

export const isLoadingPeers = writable(false)
export const peers = writable<string[]>([], () => {
    getPeers()
});

export const getPeers = () => {
    isLoadingPeers.set(true)
    return rpcRequest<NetDiscoverResult>('net_discover', { searchTime: 3 })
        .then(({ result }) => { peers.set([...new Set(result.peerIDs)]) })
        .catch(console.error)
        .finally(() => {
            isLoadingPeers.set(false)
        })
}