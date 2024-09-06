import { H } from "@/APIClient";
import { Message } from "@/hupload";
import { ActionIcon, ActionIconProps, Menu, rem } from "@mantine/core";
import { IconListCheck } from "@tabler/icons-react";
import { useEffect, useState } from "react";

interface TemplatesMenuProps {
    onChange: (message: string) => void;
}

export function TemplatesMenu(props:TemplatesMenuProps&ActionIconProps) {
    // Initialize state
    const [messages, setMessages] = useState<string[]>([])

    // effects
    useEffect(() => {
        H.get('/messages').then((res) => {
            setMessages(res as string[])
        })
    },[])

    const selectMessage = (index: number) => {
        H.get('/messages/'+index).then((res) => {
            const m = res as Message
            props.onChange(m.message)
        })
    }

    if (messages.length === 0) {
        return
    }

    return (
        <Menu withArrow withinPortal={false}>
        <Menu.Target>
            <ActionIcon size="xs" id="template" variant={"subtle"} m={rem(3)} radius="xl" style={{position:"absolute", top:0, right: 25}}>
                <IconListCheck style={{ width: rem(16), height: rem(16) }} stroke={1.5} />
            </ActionIcon>
        </Menu.Target>
        <Menu.Dropdown>
            {messages.map((m, i) => (
                <Menu.Item key={i+1} onClick={() => {selectMessage(i+1)}}>
                    {m}
                </Menu.Item>
            ))}
        </Menu.Dropdown>
    </Menu>
    );
}