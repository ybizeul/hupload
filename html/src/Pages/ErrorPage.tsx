import { Center, Stack, Text } from "@mantine/core";
import { Icon, IconExclamationCircle, IconHandStop, IconHelpHexagon, IconProps } from "@tabler/icons-react";
import React, { ForwardRefExoticComponent, RefAttributes } from "react";
import { useSearchParams } from "react-router-dom";

interface ErrorPretty {
    title: string
    icon: ForwardRefExoticComponent<IconProps & RefAttributes<Icon>>
}

interface ErrorProps {
    text?: string
    subText?: string|null
    icon?: ForwardRefExoticComponent<IconProps & RefAttributes<Icon>>
}

export function ErrorPage(props: ErrorProps) {
    let { text, subText, icon } = props

    const [searchParams] = useSearchParams()

    const pretty_errors: { [id: string] : ErrorPretty; } = {
        "access_denied": { title: "Access denied", icon: IconHandStop}
    }

    function getError(error: string|null) {
        if (!error) {
            return { title: "Unknown error", icon: IconHelpHexagon }
        }
        return pretty_errors[error] || { title: error, icon: IconExclamationCircle }
    }

    if (!(text || subText || icon)) {
        const e = getError(searchParams.get("error"))
        text = e.title
        subText = searchParams.get("error_description")
        icon = e.icon
    }

    return (
        <Center h="100vh">
            <Stack align="center" pb="10em" flex={1}>
                {icon&&React.createElement(icon, { style: { width: '10%', height: '10%' }, stroke: 1.5 })}
                <Text size="xl" fw="700">{text}</Text>
                <Text ta="center">{subText}</Text>
            </Stack>
        </Center>
    )
}