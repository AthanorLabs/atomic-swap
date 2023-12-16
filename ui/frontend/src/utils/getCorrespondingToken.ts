import type { Currency } from "src/types";

export const getCorrespondingToken = (currency: Currency): Currency => currency === 'ETH' ? 'XMR' : 'ETH'