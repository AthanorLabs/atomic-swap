export interface Pair {
    asset: String,
    verified: Boolean,
    offers: Number,
    liquidityXMR: Number,
    liquidityETH: Number,
}

export interface NetPairResults {
    Pairs: Pair[]
}