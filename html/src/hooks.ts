import { useEffect, useState } from "react"
import { Share } from "./hupload"
import { H } from "./APIClient"
import { AxiosError } from "axios"

export function useShare() : [Share|undefined, AxiosError|null] {
    const [share, setShare] = useState<Share|undefined>(undefined)
    const [error, setError] = useState<null|AxiosError>(null)

    useEffect(() => {
        const s=location.pathname.split("/")

        if (s.length > 0) {
            H.get('/shares/' + s[1]).then((res) => {
                setShare(res as Share)
            })
            .catch((e: AxiosError) => {
                console.log(e)
                setError(e)
            })
        }
    },[])

    return [share, error];
}