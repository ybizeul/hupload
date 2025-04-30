import { Anchor, Box, Button, Center, CopyButton, Group, Paper, rem, Stack, Text, Tooltip } from "@mantine/core";
import { Dropzone } from "@mantine/dropzone";
import { IconClock, IconDownload, IconFileZip, IconHelpHexagon, IconLink, IconMoodSad, IconUpload, IconX } from "@tabler/icons-react";
import { useCallback, useEffect, useState } from "react";
import { H } from "../APIClient";
import { UploadQueue, QueueItem } from "../UploadQueue";
import {ItemComponent} from "@/Components";
import { UploadableItem } from "../hupload";
import { useAuthContext } from "@/AuthContext";
import { Message } from "@/Components/Message";
import { useShare } from "@/hooks";
import { AxiosError } from "axios";
import { useTranslation } from "react-i18next";
import { ErrorPage } from "./ErrorPage";

export function SharePage() {
    const { t } = useTranslation();

    const [items, setItems] = useState<UploadableItem[]>([])
    const [error, setError] = useState<undefined|AxiosError>(undefined)

    // Initialize hooks
    const { authInfo } = useAuthContext()
    const [share, shareError] = useShare();

    // Check if share is expired
    const expired = (shareError?.response?.status === 410)

    const updateProgress = useCallback((progress: QueueItem[]) => {
        setItems((currentItems) => {
            const i = currentItems.map((currentItem) => {
                const p = progress.find((p) => p.file.name === currentItem.Path.split("/")[1])
                if (p) {
                    if (p.finished) {
                        setTimeout(() => {
                            setItems((currentItems) => {
                                return currentItems.map((currentItem) => {
                                    if (currentItem.ItemInfo.Name === p.file.name) {
                                        return {...currentItem, QueueItem: undefined}
                                    }
                                    return currentItem
                                })
                         })
                        },2000)
                    }
                    return {...currentItem, QueueItem: p}
                }
                return currentItem
            })
            return i
        })


    //     setQueueItems((currentQueue) => {
    //         const j = currentQueue.map((currentItem) => {
    //             const p = progress.find((p) => p.file.name === currentItem.file.name)
    //             if (p) {
    //                 return p
    //             }
    //             return currentItem
    //         })
    //         const k = progress.filter((p) => !j.some((i) => i.file.name === p.file.name))
    //         return [...k, ...j]
    // })
    },[])

    const queue = new UploadQueue(H,"/shares/"+share?.name, updateProgress)

    // useEffects

    useEffect(() => {
        // Get items from share
        if (share) {
            H.get('/shares/' + share.name + '/items').then((res) => {
                setItems(res as UploadableItem[])
            })
            .catch((e) => {
                setError(e)
            })
        }
    },[share])

    // Catch any errors

    if (expired) {
        return (
            <Center h="100vh">
                <Stack align="center" pb="10em">
                <IconClock style={{ width: '10%', height: '10%' }} stroke={1.5}/>
                <Text size="xl" fw="700">{t("sorry_share_expired")}</Text>
                </Stack>
            </Center>
        )
    }

    if (shareError?.response?.status === 404) {
        return (
            <ErrorPage text={t("share_does_not_exists")} subText={t("please_check_link")} icon={IconHelpHexagon}/>
        )
    }

    if (error) {
        return (
            <Center h="100vh">
                <Stack align="center" pb="10em">
                <IconMoodSad style={{ width: '10%', height: '10%' }} stroke={1.5}/>
                <Text size="xl" fw="700">{error.message}</Text>
                <Anchor onClick={() => { window.location.reload()}}>{t("reload")}</Anchor>
                </Stack>
            </Center>
        )
    }

    if (!share) {
        return
    }

    if (!items) {
        return
    }

    // Functions

    // showDropZone returns if the file drop zone should be displayed
    // drop zone is not displayed if the user is not logged in and the share is 
    // of type "download".
    const showDropZone = () => {
        if (authInfo?.user) {
            return true
        }

        // We are guest

        if (!share) {
            return false
        }

        return share.options.exposure === "" || share.options.exposure === "upload" || share.options.exposure === "both"
    }

    // canDownload returns if the user can download files from the share.
    // the user can download if they are logged in or if the share is of type
    // "download" or "both".
    const canDownload = () => {
        if (authInfo?.user){
            return true
        }

        // We are guest

        if (!share) {
            return false
        }

        return (share.options.exposure === "download" || share.options.exposure === "both")
    }

    // canDelete returns if the user can delete files from the share.
    // the user can delete if they are logged in or if the share is of type
    // "upload" (i.e. they uploaded the files).
    const canDelete = () => {
        if (authInfo?.user) {
            return true
        }

        // We are guest

        if (!share) {
            return false
        }

        return (share.options.exposure === "upload")
    }

    // deleteItem deletes an item from the share.
    const deleteItem = (item: string) => {
        H.delete('/shares/' + share.name + '/items/' + item).then(() => {
            setItems(items?.filter((i) => i.Path !== share.name + "/" + item))
        })
        .catch((e) => {
            console.log(e)
        })
    }

    return (
        <>
            {/* Top of page copy button */}
            <Box w="100%" ta="center">
                <Group gap="md" justify={"center"} w="100%" mb="sm">
                    <CopyButton value={window.location.protocol + '//' + window.location.host + '/' + share.name}>
                        {({ copied, copy }) => (
                        <Tooltip withArrow arrowOffset={10} arrowSize={4} label={copied?t("copied"):t("copy_url")}>
                            <Button justify="center" variant="outline" color={copied ? 'teal' : 'gray'} size="xs" onClick={copy}><IconLink style={{ width: '70%', height: '70%' }} stroke={1.5}/>{share.name}</Button>
                        </Tooltip>
                        )}
                    </CopyButton>
                    {canDownload() && items.length &&
                        <Tooltip withArrow arrowOffset={10} arrowSize={4} label={t("download_all")}>
                            <Button component="a" href={'/d/'+share.name} justify="center" variant="outline" size="xs"><IconDownload style={{ width: '70%', height: '70%' }} stroke={1.5}/>{t("download_button")}</Button>
                        </Tooltip>
                    }
                </Group>
            </Box>

            {/* Message */}
            {share.options.message &&
                <Paper withBorder p="sm" mb="sm">
                <Message value={share.options.message} />
                </Paper>
            }


            {/* Files drop zone */}
            {showDropZone() &&
            <>
                <Dropzone
                onDrop={(files) => {
                    const uploadableItems = files.map((f) => {
                        return {
                            Path: share.name + "/" + f.name,
                            ItemInfo: {
                                Name: f.name,
                                Size: f.size,
                            },
                            QueueItem: {},
                        } as UploadableItem
                    })

                    setItems((currentItems) => {
                        const i = [...currentItems, ...uploadableItems]
                        return i
                    })
                    // // Filter out files that are already uploaded
                    // const newItems = items.filter((i) => {
                    //     return !files.some((f) => f.name === i.Path.split("/")[1])
                    // })

                    // setItems(newItems)

                    queue.addFiles(files)
                    .catch((e) => {
                        console.log(e)
                    })
                }}

                onReject={(files) => console.log('rejected files', files)}
                >
                <Group justify="center" gap="xl" mih={100} style={{ pointerEvents: 'none' }}>
                    <Dropzone.Accept>
                    <IconUpload
                        style={{ width: rem(52), height: rem(52), color: 'var(--mantine-color-blue-6)' }}
                        stroke={1.5}
                    />
                    </Dropzone.Accept>
                    <Dropzone.Reject>
                    <IconX
                        style={{ width: rem(52), height: rem(52), color: 'var(--mantine-color-red-6)' }}
                        stroke={1.5}
                    />
                    </Dropzone.Reject>
                    <Dropzone.Idle>  
                    <IconFileZip
                        style={{ width: rem(52), height: rem(52), color: 'var(--mantine-color-dimmed)' }}
                        stroke={1.5}
                    />
                    </Dropzone.Idle>
                    <div>
                    <Text size="xl" inline>
                        {t("drag_area")}
                    </Text>
                    </div>
                </Group>
                </Dropzone>
            </>}

            {
                // Display share items
                items.map((item) => (
                <ItemComponent download={canDownload()} canDelete={canDelete()} onDelete={deleteItem} key={item.Path} item={item} />
                ))
            }
        </>
    )
}