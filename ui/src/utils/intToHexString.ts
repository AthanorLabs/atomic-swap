export const intToHexString = (input: number[]) => {
    const hexArray = input.map((n) => {
        const num = Number(n).toString(16).padStart(2, "0")
        // console.log(n, num)
        return num
    })

    return hexArray.join("")
}