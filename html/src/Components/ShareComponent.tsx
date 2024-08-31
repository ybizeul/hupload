import { ActionIcon, ActionIconGroup, Anchor, Box, Button, CopyButton, Flex, Group, Paper, Popover, Progress, Stack, Text, Tooltip, useMantineTheme } from "@mantine/core";
import { humanFileSize, prettyfiedCount, Share } from "../hupload";
import { Link } from "react-router-dom";
import classes from './ShareComponent.module.css';
import { IconClock, IconDots, IconLink, IconTrash } from "@tabler/icons-react";
import { useState } from "react";
import { H } from "@/APIClient";
import { ResponsivePopover } from "./ResponsivePopover";
import { useMediaQuery } from "@mantine/hooks";
import { ShareEditor } from "./ShareEditor";
import { Dropzone } from "@mantine/dropzone";
import { QueueItem, UploadQueue } from "@/UploadQueue";

export function ShareComponent(props: {share: Share}) {
    // Initialize States
    const [share,setShare] = useState(props.share)
    const [deleted,setDeleted] = useState(false)
    const [newOptions, setNewOptions] = useState<Share["options"]>(share.options)
    const [uploadPercent, setUploadPercent] = useState(0)
    const [uploading, setUploading] = useState(false)

    const theme = useMantineTheme();
    const isBrowser = useMediaQuery('(min-width: +' + theme.breakpoints.xs + ')');

    // Other initializations
    const key = share.name
    const name = share.name
    const count = share.count
    const size = share.size
    const countString = prettyfiedCount(count,"item", "items","empty")
    const remaining = (share.options.validity===0||share.options.validity===undefined)?null:(new Date(share.created).getTime() + share.options.validity*1000*60*60*24 - Date.now()) / 1000 / 60 / 60 / 24

    // Function
    const deleteShare = () => {
        H.delete('/shares/'+name).then(() => {
            setDeleted(true)
        })
    }

    const updateShare = () => {
        H.patch('/shares/'+name, newOptions).then(() => {
            setShare({...share, options: newOptions})
        })
    }

    if (deleted) {
        return
    }
    
    return (
        <>
        <Paper key={key} withBorder shadow="xs" radius="md" mt={10} pos="relative" className={classes.paper}>
            <Dropzone p={0} m={0} activateOnClick={false} enablePointerEvents={true} w={"100%"} h={"100%"} style={{border:"none", backgroundColor:"transparent"}}
                onDrop={(files) => {
                    const updateProgress = (progress: QueueItem[]) => {
                        const loaded = progress.reduce((acc, item) => acc + item.loaded, 0)
                        const total = progress.reduce((acc, item) => acc + item.total, 0)
                        setUploadPercent(loaded/total*100)
                        console.log(progress)
                    }
                    const queue = new UploadQueue(H,"/shares/"+name, updateProgress)
                    setUploading(true)
                    queue.addFiles(files)
                        .then(() => {
                            setUploading(false)
                            H.get('/shares/'+name).then((r) => {
                                const s = r as Share
                                setShare(s)
                            })
                        })
                        .catch((e) => {
                            setUploading(false)
                            console.log(e)
                        })
                }}

                onReject={(files) => console.log('rejected files', files)}
                >
                    {/* Share component footer */}
                    <Group wrap="nowrap" flex={1} w="100%" pos="absolute" bottom="0.1em" style={{justifyContent:"center"}} align="center" gap="0.2em">
                        <IconClock color={(remaining===null || remaining > 0 )?"gray":"red"} size="0.8em"  width={"1em"}/>
                        <Text style={{ whiteSpace: "nowrap"}} size="xs" c="gray">{
                            (remaining === null)?
                            "Unlimited"
                            :
                            (remaining<0)?
                            "Expired"
                            :
                            prettyfiedCount(remaining,"day","days",null) + " left"} | {share.options.exposure==="download"?"Guests can download":(share.options.exposure==="both"?"Guests can upload & download":"Guests can upload")}
                        </Text>
                    </Group>
                    {(uploading || uploadPercent > 0) &&
                            <Progress pos="absolute" bottom="0" w="100%" size="xs" color={uploading?"blue":"green"} value={uploadPercent} />
                        }
                    {/* Share informations */}
                    <Box p="lg">
                        <Stack gap="0">
                            <Flex align={"center"}>

                                {/* Share name */}
                                <Group flex="1" gap="0" align="baseline">
                                    <Anchor style={{ whiteSpace: "nowrap"}} flex={"1"} component={Link} to={'/'+name}><Text>{name}</Text></Anchor>
                                    <Stack gap="0" align="flex-end">
                                        <Text mr="xs" size="xs" c="gray">{countString + (size?(' | ' + humanFileSize(size)):'')}</Text>
                                    </Stack>
                                </Group>

                                {/* Share component tail with actions */}
                                <ActionIconGroup >
                                    {/* Copy button */}
                                    <CopyButton value={window.location.protocol + '//' + window.location.host + '/' + name}>
                                    {({ copied, copy }) => (
                                        <Tooltip withArrow arrowOffset={10} arrowSize={4} label={copied?"Copied!":"Copy URL"}>
                                            <ActionIcon variant="light" color={copied ? 'teal' : 'blue'} onClick={copy} >
                                                <IconLink style={{ width: '70%', height: '70%' }} stroke={1.5}/>
                                            </ActionIcon>
                                        </Tooltip>
                                    )}
                                    </CopyButton>

                                    {/* Delete button with confirmation Popover */}
                                    <Popover width={200} position="bottom" withArrow shadow="md">
                                        <Popover.Target>
                                            <Tooltip withArrow arrowOffset={10} arrowSize={4} label="Delete Share">
                                                <ActionIcon variant="light" color="red" >
                                                    <IconTrash style={{ width: '70%', height: '70%' }} stroke={1.5}/>
                                                </ActionIcon>
                                            </Tooltip>
                                        </Popover.Target>
                                        <Popover.Dropdown className={classes.popover}>
                                            <Text ta="center" size="xs" mb="xs">Delete this share ?</Text>
                                            <Button aria-description="delete" w="100%" variant='default' c='red' size="xs" onClick={deleteShare}>Delete</Button>
                                        </Popover.Dropdown>
                                    </Popover>
                                    
                                    {/* Edit share properties button */}
                                    <ResponsivePopover withDrawer={!isBrowser}>
                                        <Tooltip withArrow arrowOffset={10} arrowSize={4} label="Edit Share">
                                            <ActionIcon variant="light" color="blue" >
                                                <IconDots style={{ width: '70%', height: '70%' }} stroke={1.5}/>
                                            </ActionIcon>
                                        </Tooltip>
                                        <ShareEditor buttonTitle="Update" onChange={setNewOptions} onClick={updateShare} options={newOptions}/>
                                    </ResponsivePopover>
                                </ActionIconGroup>
                            </Flex>
                            {share.options.description && <Text w="100%" size="xs" c="gray">{share.options.description}</Text>}</Stack>
                    </Box>
                </Dropzone>
            {/* Share component footer */}
        </Paper>
        </>
)}