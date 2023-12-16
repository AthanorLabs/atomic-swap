export { rpcRequest, getPort } from './rpcApi'
export { intToHexString } from './intToHexString'
export { getCorrespondingToken } from './getCorrespondingToken'
export { getTokenInfo } from './getTokenInfo'

import type { TokenInfo } from '../types/PersonalTokenInfoResult'

export const EthTokenInfo: TokenInfo = {
  address: "0x0000000000000000000000000000000000000000",
  decimals: 18,
  name: "Ether",
  symbol: "ETH",
}
