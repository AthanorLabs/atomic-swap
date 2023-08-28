import { rpcRequest } from '../utils';
import { writable } from 'svelte/store';
import type { NetDiscoverResult } from '../types/NetDiscoverResults';

export const isLoadingPeers = writable(false)
export const peers = writable<string[]>([], () => {
    getPeers()
});

export const getPeers = async () => {
    try {
      isLoadingPeers.set(true)
      const resp = await rpcRequest<NetDiscoverResult>('net_discover', { searchTime: 3 })
      peers.set([...new Set(resp.result.peerIDs)])
    } catch (e) {
      console.error(e)
    } finally {
      isLoadingPeers.set(false)
    }
}