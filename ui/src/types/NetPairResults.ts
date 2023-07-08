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
    liquidityXmr: Number,
    liquidityEth: Number,
}

export interface NetPairResults {
    Pairs: Pair[]
}
