export const sanitizeAddresses = (addresses: string[]) => {
    return addresses.filter(add => {
        return !add.includes("127.0.0.1")
    })
} 