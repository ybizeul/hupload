import { useEffect, useState } from "react"
import { Share } from "./hupload"
import { H } from "./APIClient"
import { AxiosError } from "axios"

export function useShare(user?: string) : [Share|undefined|null, AxiosError|null] {
    const [share, setShare] = useState<Share|undefined|null>(undefined)
    const [error, setError] = useState<null|AxiosError>(null)

    useEffect(() => {
        let cancelled = false
        const s=location.pathname.split("/")

        if (s.length > 0) {
            setError(null)
            setShare(undefined)
            H.get('/shares/' + s[1]).then((res) => {
                if (cancelled) {
                    return
                }
                setError(null)
                setShare(res as Share)
            })
            .catch((e: AxiosError) => {
                if (cancelled) {
                    return
                }
                console.log(e)
                setShare(null)
                setError(e)
            })
        }

        return () => {
            cancelled = true
        }
    },[user])

    return [share, error];
}