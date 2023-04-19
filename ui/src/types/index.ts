import type { Currency } from "./NetQueryPeerResults"

export type { NetAddressesResult } from "./NetAddressResult"
export type { NetDiscoverResult } from "./NetDiscoverResults"
export type { OfferRaw, NetQueryPeerResult, Currency } from "./NetQueryPeerResults"

export interface Offer {
    peerID: String
    offerID: String
    provides: Currency
    minAmount: Number
    maxAmount: Number
    exchangeRate: Number
    version: String
    ethAsset: Currency
    nonce: Number
}