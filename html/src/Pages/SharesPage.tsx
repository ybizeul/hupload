
import { useCallback, useEffect, useState } from "react";
import { H } from "../APIClient";
import { useNavigate } from "react-router-dom";
import { Share, ShareDefaults } from "../hupload";
import {ShareComponent, ResponsivePopover} from "@/Components";
import { ActionIcon, Anchor, Box, Button, Center, Group, rem, Stack, Text, useMantineTheme } from "@mantine/core";
import { IconChevronDown, IconMoodSad } from "@tabler/icons-react";
import { AxiosError } from "axios";
import { ShareEditor } from "@/Components/ShareEditor";
import { useMediaQuery } from "@mantine/hooks";

import classes from './SharesPage.module.css';

import { useTranslation } from "react-i18next";

export function SharesPage(props: {owner: string|null}) {
    const { t } = useTranslation();

    // Initialize props
    const { owner } = props

    // Initialize state
    const [shares, setShares] = useState<Share[]|undefined>(undefined)
    const [newShareOptions, setNewShareOptions] = useState<Share["options"]>({exposure: "upload", validity: 7, description: "", message: ""})
    const [error, setError] = useState<AxiosError|undefined>(undefined)
    
    // Initialize hooks
    const navigate = useNavigate();
    const theme = useMantineTheme();
    const isBrowser = useMediaQuery('(min-width: +' + theme.breakpoints.xs + ')');

    const updateShareProperties = (props: Share["options"]) => {
        setNewShareOptions(props)
    }

    // Functions
    const createShare = () => {
        H.post('/shares', newShareOptions).then(
        () => {
            updateShares()
        })
        .catch((e) => {
            console.log(e)
            setError(e)
        })
    }

    const updateShares = useCallback(() => {
        H.get('/shares').then(
        (res) => {
            setShares(res as Share[])
        })
        .catch((e) => {
            console.log(e)
            if (e.response?.status === 401) {
                navigate('/')
                return
            }
        })
    },[navigate])

    useEffect(() => {
        updateShares()
    },[updateShares])

    useEffect(() => {
        H.get("/defaults").then(
        (res) => {
            const newOptions = {...newShareOptions,...res as ShareDefaults}
            setNewShareOptions(newOptions)
        })
    },[])

    if (error) {
        return (
        <Center h="100vh">
            <Stack align="center" pb="10em">
            <IconMoodSad style={{ width: '10%', height: '10%' }} stroke={1.5}/>
            <Text size="xl" fw="700">{error.message}</Text>
            <Anchor onClick={() => { window.location.reload()}}>Reload</Anchor>
            </Stack>
        </Center>
        )
    }

    if (!shares) {
        return
    }

    return (
        <>
            {/* Create share button */}
            <Box ta="center" mt="xl" mb="xl">
                <Group wrap="nowrap" gap={0} justify='center'>
                    <Button onClick={() => {createShare()}} className={classes.button}>{t("create_share")}</Button>
                    <ResponsivePopover withDrawer={!isBrowser} >
                        <ActionIcon
                            variant="filled"
                            color={theme.primaryColor}
                            size={36}
                            className={classes.menuControl}
                        >
                            <IconChevronDown style={{ width: rem(16), height: rem(16) }} stroke={1.5} />
                        </ActionIcon>
                        <ShareEditor buttonTitle={t("create")} onChange={updateShareProperties}
                            onClick={() => {createShare()}} 
                            options={newShareOptions}
                        />
                    </ResponsivePopover>
                </Group>
            </Box>

            { shares.length == 0 ?
                <Text ta="center" mt="xl">There are currently no shares</Text>
                :
                <>
                {/* Currently logged in user shares */}
                {shares.some((s) => s.owner === owner) &&
                    <>
                    <Text size="xl" fw="700">{t("your_shares")}</Text>
                    {shares.map((s) => (
                    s.owner === owner &&
                    <ShareComponent key={s.name} share={s} />
                    ))}
                    </>
                }

                {/* Other users shares */}
                {shares.some((s) => s.owner !== owner) &&
                    <>
                    <Text mt="md" size="xl" fw="700">{t("other_shares")}</Text>
                    {shares.map((s) => (
                    s.owner === owner ||
                    <ShareComponent key={s.name} share={s} />
                    ))}
                    </>
                }
                </>
            }
        </>
    )
}