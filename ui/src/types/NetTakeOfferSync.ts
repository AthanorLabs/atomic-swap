
export type NetTakeOfferSyncResult = {
    status: 'success' | 'aborted' | 'refunded'
    id: number
}