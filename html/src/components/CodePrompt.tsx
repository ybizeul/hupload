import { Button, Stack, Text } from "@mantine/core";

import CenteredTextInput from "./CenteredTextInput";
import { useState } from "react";
import { useNavigate } from "react-router-dom";

export default function CodePrompt() {
    const [code, setCode] = useState("")
    const navigate = useNavigate()
    return (
        <Stack mb="20%" gap="lg" align="center">
            <Text size="xl" fw={700} c="gray">Please use your invitation code :</Text>
            <Stack gap="sm" align="center">
                <CenteredTextInput value={code} onChange={(e) => {setCode(e.currentTarget.value)}}/>
                <Button size="lg" color="blue" variant="light" onClick={() => {navigate('/'+code)}}>Go</Button>
            </Stack>
        </Stack>
    )
}