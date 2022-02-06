import { rpcRequest } from '../utils';
import { derived, Readable, writable } from 'svelte/store';
import { peers } from './peerStore'
import type { NetQueryPeerResult, Offer } from 'src/types';
import { intToHexString } from 'src/utils';

export const isLoadingPeers = writable(false)

export const offers = derived<Readable<string[]>, Offer[]>(
    peers,
    ($peers, set) => {
        // loop over all the peers the get their offers
        $peers.reduce(async (acc: Promise<Offer[]>, curr: string) => {
            const previousPeersOffers = await acc
            const currentOffers = await getOffers(curr) || []
            return [...previousPeersOffers, ...currentOffers]
        }
            , Promise.resolve([])
        )
            .then(off => set(off))
    },
    []
)

export const getOffers = async (peerAddress: string) => {
    isLoadingPeers.set(true)
    return rpcRequest<NetQueryPeerResult | undefined>('net_queryPeer', { "multiaddr": peerAddress })
        .then(({ result }): Offer[] => {

            return result?.offers.map(off => ({
                peer: peerAddress,
                id: intToHexString(off.ID),
                exchangeRate: off.ExchangeRate,
                maxAmount: off.MaximumAmount,
                minAmount: off.MinimumAmount,
                provides: off.Provides
            })) || []
        })
        .catch(console.error)
        .finally(() => isLoadingPeers.set(false))
}