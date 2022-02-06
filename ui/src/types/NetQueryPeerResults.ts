export type Currency = 'ETH' | 'XMR'

export interface OfferRaw {
    ID: number[]
    Provides: Currency
    MinimumAmount: number
    MaximumAmount: number
    ExchangeRate: number
}

export interface NetQueryPeerResult {
    offers: OfferRaw[]
}