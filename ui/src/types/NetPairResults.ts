export interface Pair {
    asset: String,
    verified: Boolean,
    offers: Number,
    liquidityXmr: Number,
    liquidityEth: Number,
}

export interface NetPairResults {
    Pairs: Pair[]
}