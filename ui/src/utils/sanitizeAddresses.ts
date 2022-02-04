export const sanitizeAddresses = (addresses: string[]): string[] => {
    const added: string[] = []
    const res: string[] = []

    // only add an address if it's unique
    addresses.forEach(add => {
        const splited = add.split("/")
        const multiAddress = splited[splited.length - 1]
        !added.some((addedAddress) => addedAddress === multiAddress) && added.push(multiAddress) && res.push(add)
    })

    return res
} 