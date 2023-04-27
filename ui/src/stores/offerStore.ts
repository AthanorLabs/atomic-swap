import { rpcRequest } from '../utils';
import { derived, Readable, writable } from 'svelte/store';
import { peers } from './peerStore'
import type { NetQueryPeerResult, Offer } from 'src/types';
import { intToHexString } from 'src/utils';

export const isLoadingOffers = writable(false)
export const selectedOffer = writable<Offer | undefined>()

export const offers = derived<Readable<string[]>, Offer[]>(
    peers,
    ($peers, set) => {
        refreshOffers($peers)
            .then(off => set(off))
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
        .then(({ result }): Offer[] =>
            result?.offers.map(offer => ({
                peerID: peerAddress,
                ...offer
            })) || []
        )
        .catch(console.error)
        .finally(() => isLoadingOffers.set(false))
}