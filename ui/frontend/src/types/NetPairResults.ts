export interface Pair {
    ethAsset: String,
    token: {
        address: String,
        decimals: Number,
        name: String,
        symbol: String,
    },
    verified: Boolean,
    offers: Number,
    reportedLiquidityXmr: Number,
}

export interface NetPairResults {
    Pairs: Pair[]
}