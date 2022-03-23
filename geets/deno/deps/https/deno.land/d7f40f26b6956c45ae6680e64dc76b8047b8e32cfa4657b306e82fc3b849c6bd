import { fetchLocal, fetchRemote } from "../mod.ts"

export const fetchAuto = (path: string, onlyData?: boolean) => {
    try {
        new URL(path)
        return fetchRemote(path, onlyData)
    } catch(e) {
        return fetchLocal(path, onlyData)
    }
}