import { encode } from '../deps.ts'

export const fetchRemote = async (url: string, onlyData?: boolean) => {
    const resp = await fetch(url)
    if(resp.status !== 200) throw new Error(`Request Failed. Server responsed with code ${resp.status}`)
    const contentType = resp.headers.get('content-type') ?? 'application/octet-stream';
    const buff = await resp.arrayBuffer()
    const data = encode(buff)
    return onlyData ? data : `data:${contentType};base64,${data}`
}