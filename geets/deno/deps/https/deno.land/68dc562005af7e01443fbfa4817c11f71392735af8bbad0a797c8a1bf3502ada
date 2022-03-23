import { encode } from "../deps.ts"

export const fetchLocal = async (url: string, onlyData?: boolean) => {
    const file = await Deno.readFile(url)
    const data = encode(file)
    const contentType = `image/${url.split(".").reverse()[0]}`
    return onlyData ? data : `data:${contentType};base64,${data}`
}