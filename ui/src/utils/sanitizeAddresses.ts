export const sanitizeAddresses = (peers: string[][]): string[] => {

    const added: string[] = []
    const res: string[] = []

    peers.forEach(addresses => {
        // only add an address if it's unique
        addresses.forEach(add => {
            const splited = add.split("/")
            const multiAddress = splited[splited.length - 1]
            !added.some((addedAddress) => addedAddress === multiAddress) && added.push(multiAddress) && res.push(add)
        })
    })

    return res
} 