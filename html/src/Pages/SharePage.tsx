import { Anchor, Box, Button, Center, CopyButton, Group, Paper, rem, Stack, Text, Tooltip } from "@mantine/core";
import { Dropzone } from "@mantine/dropzone";
import { IconClock, IconDownload, IconFileZip, IconHelpHexagon, IconLink, IconMoodSad, IconUpload, IconX } from "@tabler/icons-react";
import { useCallback, useEffect, useState } from "react";
import { H } from "../APIClient";
import { UploadQueue, QueueItem } from "../UploadQueue";
import {ItemComponent} from "@/Components";
import { Item } from "../hupload";
import { useAuthContext } from "@/AuthContext";
import { Message } from "@/Components/Message";
import { useShare } from "@/hooks";
import { AxiosError } from "axios";
import { useTranslation } from "react-i18next";

export function SharePage() {
    const { t } = useTranslation();

    const [items, setItems] = useState<Item[]|undefined>(undefined)
    const [queueItems, setQueueItems] = useState<QueueItem[]>([])
    //const [expired,setExpired] = useState(false)
    const [error, setError] = useState<undefined|AxiosError>(undefined)

    // Initialize hooks
    const { authInfo } = useAuthContext()
    const [share, shareError] = useShare();

    // Check if share is expired
    const expired = (shareError?.response?.status === 410)

    const updateProgress = useCallback((progress: QueueItem[]) => {
        setQueueItems((currentQueue) => {
            const j = currentQueue.map((currentItem) => {
                const p = progress.find((p) => p.file.name === currentItem.file.name)
                if (p) {
                    return p
                }
                return currentItem
            })
            const k = progress.filter((p) => !j.some((i) => i.file.name === p.file.name))
            return [...k, ...j]
    })
    },[])

    const queue = new UploadQueue(H,"/shares/"+share?.name, updateProgress)

    // useEffects

    useEffect(() => {
        // Get items from share
        if (share) {
            H.get('/shares/' + share.name + '/items').then((res) => {
                setItems(res as Item[])
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
            <Center h="100vh">
            <Stack align="center" pb="10em">
                <IconHelpHexagon style={{ width: '10%', height: '10%' }} stroke={1.5}/>
                <Text size="xl" fw="700">{t("share_does_not_exists")}</Text>
                <Text>{t("please_check_link")}</Text>
            </Stack>
            </Center>
        )
    }

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
                        <Tooltip withArrow arrowOffset={10} arrowSize={4} label={copied?"Copied!":"Copy URL"}>
                            <Button justify="center" variant="outline" color={copied ? 'teal' : 'gray'} size="xs" onClick={copy}><IconLink style={{ width: '70%', height: '70%' }} stroke={1.5}/>{share.name}</Button>
                        </Tooltip>
                        )}
                    </CopyButton>
                    {canDownload() && items.length + queueItems.filter((i) => i.failed === false && i.finished === true ).length > 0 &&
                        <Button component="a" href={'/d/'+share.name} justify="center" variant="outline" size="xs"><IconDownload style={{ width: '70%', height: '70%' }} stroke={1.5}/>Download</Button>
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
                    // Filter out files that are already uploaded
                    const newItems = items.filter((i) => {
                        return !files.some((f) => f.name === i.Path.split("/")[1])
                    })

                    setItems(newItems)

                    queue.addFiles(files)
                        // .then((r) => {
                            
                        //     //const finishedItems = r as Item[]

                        //     //setQueueItems([])
                        //     setItems([...finishedItems, ...newItems])
                        // })
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
                        Drag files here or click to select files
                    </Text>
                    </div>
                </Group>
                </Dropzone>
            </>}

            {
                // Display upload queue items (queue items uploading or finished 
                // uploading)
                queueItems.map((q) => (
                <ItemComponent  download={false} canDelete={false} key={'up_' + q.file.name} queueItem={q} />
                ))
            }

            {
                // Display share items
                items.map((item) => (
                <ItemComponent download={canDownload()} canDelete={canDelete()} onDelete={deleteItem} key={item.Path} item={item} />
                ))
            }
        </>
    )
}