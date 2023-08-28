import { rpcRequest } from './rpcApi';
import type { TokenInfo } from '../types/PersonalTokenInfoResult';

export const getTokenInfo = async (address: String): Promise<TokenInfo | void | undefined> => {
  return rpcRequest<TokenInfo | undefined>('personal_tokenInfo', { "tokenAddr": address })
    .then(({ result }): TokenInfo | undefined => {
      return result
    })
    .catch(console.error)
}