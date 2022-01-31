import type { Currency } from "./NetQueryPeerResults"

export type { NetAddressesResult } from "./NetAddressResult"
export type { NetDiscoverResult } from "./NetDiscoverResults"
export type { OfferRaw, NetQueryPeerResult, Currency } from "./NetQueryPeerResults"

export interface Offer {
    peer: string
    id: string
    provides: Currency
    minAmount: number
    maxAmount: number
    exchangeRate: number
}