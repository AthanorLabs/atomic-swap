import { rpcRequest } from '../utils';
import { derived, writable } from 'svelte/store';
import type { Readable } from 'svelte/store'
import { peers } from './peerStore'
import type { NetQueryPeerResult, Offer } from '../types';
import { intToHexString } from '../utils';

export const isLoadingOffers = writable(false)
export const selectedOffer = writable<Offer | undefined>()

export const offers = derived<Readable<string[]>, Offer[]>(
    peers,
    ($peers, set) => {
        refreshOffers($peers).then(off => set(off))
    },
    []
)

// Loop over all the peers to get their offers
export const refreshOffers = ($peers: string[]) =>
    $peers.reduce(async (acc: Promise<Offer[]>, curr: string) => {
        const previousPeersOffers = await acc
        const currentOffers = await getOffers(curr) || []
        return [...previousPeersOffers, ...currentOffers]
    }, Promise.resolve([]))

export const getOffers = async (peerAddress: string) => {
    isLoadingOffers.set(true)
    return rpcRequest<NetQueryPeerResult | undefined>('net_queryPeer', { "peerID": peerAddress })
        .then(({ result }): Offer[] => {
            return result?.offers.map(offer => ({
                peerID: peerAddress,
                ...offer
            })) || []
        })
        .catch(console.error)
        .finally(() => isLoadingOffers.set(false))
}




