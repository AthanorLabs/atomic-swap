import { rpcRequest } from '../utils';
import { writable } from 'svelte/store';
import type { Pair, NetPairResults } from '../types/NetPairResults';

export const isLoadingPairs = writable(false)
export const pairs = writable<Pair[]>([], () => {
    getPairs()
});

export const liquidity = writable(0)
export const offers = writable(0)

export const getPairs = () => {
    isLoadingPairs.set(true)
    return rpcRequest<NetPairResults>('net_pairs', { searchTime: 3 })
        .then(({ result }) => {
            pairs.set(result.Pairs)
            liquidity.set(result.Pairs.reduce((acc, a) => acc += Number(a.reportedLiquidityXmr), 0))
            console.log(result, liquidity)
            offers.set(result.Pairs.reduce((acc, a) => acc += a.offers, 0))
        })
        .catch(console.error)
        .finally(() => {
            isLoadingPairs.set(false)
        })
}