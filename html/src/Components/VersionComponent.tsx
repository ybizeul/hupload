import { useEffect, useState } from "react"
import { H } from "../APIClient"
import { Text } from "@mantine/core"


interface VersionInterface {
    version: string
}

export function VersionComponent() {
    const [version, setVersion] = useState<string|null>(null)
    useEffect(() => {
        H.get('/version').then((r) => {
            const v = r as VersionInterface
            setVersion(v.version)
        })
    })

    return (
        <Text w="100%" mt="lg" mb="lg" size="xs" c="darkgray" ta="center" >{version}</Text>
    )
}