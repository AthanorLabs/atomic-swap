export const intToHexString = (input: number[]) => {
    const hexArray = input.map((n) => {
        return Number(n).toString(16)
    })

    return hexArray.join("")
}