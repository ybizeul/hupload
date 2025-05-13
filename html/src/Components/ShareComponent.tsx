import { ActionIcon, ActionIconGroup, Anchor, Box, Button, CopyButton, Flex, Group, Paper, Popover, Progress, Stack, Text, Tooltip, useMantineTheme } from "@mantine/core";
import { humanFileSize, prettyfiedCount, Share, ShareDefaults } from "../hupload";
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

import { useTranslation } from "react-i18next";

export function ShareComponent(props: {share: Share, onDelete?: (name: string) => void}) {
    const { t } = useTranslation();
    
    // Initialize States
    const [share,setShare] = useState(props.share)
    const [newOptions, setNewOptions] = useState<Share["options"]>(share.options)
    const [uploadPercent, setUploadPercent] = useState(0)
    const [uploading, setUploading] = useState(false)
    const [error, setError] = useState(false)

    const theme = useMantineTheme();
    const isBrowser = useMediaQuery('(min-width: +' + theme.breakpoints.xs + ')');

    // Other initializations
    const key = share.name
    const name = share.name
    const count = share.count
    const size = share.size
    const countString = prettyfiedCount(count,t("item"), t("items"),t("empty"))
    const remaining = (share.options.validity===0||share.options.validity===undefined)?null:(new Date(share.created).getTime() + share.options.validity*1000*60*60*24 - Date.now()) / 1000 / 60 / 60 / 24

    // Functions

    const updateShare = () => {
        H.patch('/shares/'+name, newOptions).then(() => {
            setShare({...share, options: newOptions})
        })
    }

    const renew = () => {
        H.get('/defaults').then((r) => {
            const d = r as ShareDefaults
            const shareStartDate = new Date(share.created).getTime()
            const shareEndDate = Date.now() + d.validity*1000*60*60*24
            const newValidity = Math.ceil((shareEndDate-shareStartDate) / 1000 / 60 / 60 / 24)

            setNewOptions({ ...newOptions, validity: newValidity})
        })
    }

    const linkify = (text: string) => {
        const urlRegex = /^https?:\/\/(?:www\.)?[-a-zA-Z0-9@:%._+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}\b(?:[-a-zA-Z0-9()@:%_+.~#?&/=]*)$/
        if (urlRegex.test(text)) {
            return <Anchor href={text} target="_blank" rel="noopener noreferrer">{text}</Anchor>
        }
        else {
            return text
        }
    }
    
    const localeStorage = window.localStorage.getItem("i18nextLng")
    const locales: string[] = localeStorage===null?[]:[localeStorage]
    
    return (
        <>
        <Paper id={share.name} key={key} withBorder shadow="xs" radius="md" 
            mt={10} pos="relative" className={classes.paper} 
            style={(uploading || uploadPercent > 0)?{ borderBottomLeftRadius: "0", borderBottomRightRadius: "0"}:{}}>
            <Dropzone p={0} m={0} activateOnClick={false} enablePointerEvents={true} w={"100%"} h={"100%"} style={{border:"none", backgroundColor:"transparent"}}
                onDrop={(files) => {
                    const updateProgress = (progress: QueueItem[]) => {
                        const loaded = progress.reduce((acc, item) => acc + item.loaded, 0)
                        const total = progress.reduce((acc, item) => acc + item.total, 0)
                        const finished = Object.values(progress).every((item) => item.finished)
                        setUploadPercent(finished?100:loaded/total*100)
                        console.log(progress)
                    }
                    const queue = new UploadQueue(H,"/shares/"+name, updateProgress)
                    setUploading(true)
                    queue.addFiles(files)
                        .then(() => {
                            setUploading(false)
                            setError(false)
                            H.get('/shares/'+name).then((r) => {
                                const s = r as Share
                                setShare(s)
                            })
                        })
                        .catch((e) => {
                            setUploading(false)
                            setError(true)
                            console.log(e)
                        })
                }}

                onReject={(files) => console.log('rejected files', files)}
                > 
                    <Flex justify={"center"}>
                        {/* Share component footer */}
                        <Group wrap="nowrap" flex={1} w="100%" pos="absolute" bottom="0.1em" style={{justifyContent:"center"}} align="center" gap="0.2em">
                            <IconClock color={(remaining===null || remaining > 0 )?"gray":"red"} size="0.8em"  width={"1em"}/>
                            <Text style={{ whiteSpace: "nowrap"}} size="xs" c="gray">{
                                (remaining === null)?
                                t("unlimited")
                                :
                                (remaining<0)?
                                t("expired")
                                :
                                prettyfiedCount(remaining,t("day_left"),t("days_left"),null)} | {share.options.exposure==="download"?t("guests_can_download"):(share.options.exposure==="both"?t("guests_can_upload_and_download"):t("guests_can_upload"))}
                            </Text>
                        </Group>
                        {(uploading || uploadPercent > 0) &&
                                <Progress radius="0" w="100%" pos="absolute" bottom="0" size="xs" color={error?"red":(uploading?"blue":"green")} value={uploadPercent} />
                            }
                        {/* Share informations */}
                        <Box w="100%" flex="1" p="md" >
                            <Stack w="100%" gap="0" >
                                <Flex w="100%" align={"center"} >

                                    {/* Share name */}
                                    <Group wrap="nowrap" w="100%" flex="1" gap="0" align="center" style={{overflow:"hidden"}}>
                                        <Stack w="100%" flex="1" gap={0} align="flex-start" style={{overflow:"hidden"}}>
                                            <Anchor style={{ whiteSpace: "nowrap"}} component={Link} to={'/'+name}><Text>{name}</Text></Anchor>
                                            <Text w="100%" flex="1" size="xs" c="gray" style={{overflow:"hidden",textOverflow: "ellipsis",whiteSpace: "nowrap"}}>
                                                {share.options.description?linkify(share.options.description)
                                            :
                                                t("created") + " " + new Date(share.created).toLocaleString(locales,{dateStyle:"long",timeStyle:"short"})
                                            }
                                            </Text>
                                        </Stack>
                                        <Stack gap="0" align="flex-end">
                                            <Text mr="xs" size="xs" c="gray">{countString + (size?(' | ' + humanFileSize(size)):'')}</Text>
                                        </Stack>
                                    </Group>

                                    {/* Share component tail with actions */}
                                    <ActionIconGroup >
                                        {/* Copy button */}
                                        <CopyButton value={window.location.protocol + '//' + window.location.host + '/' + name}>
                                        {({ copied, copy }) => (
                                            <Tooltip withArrow arrowOffset={10} arrowSize={4} label={copied?t("copied"):t("copy_url")}>
                                                <ActionIcon id="copy" variant="light" color={copied ? 'teal' : 'blue'} onClick={copy} >
                                                    <IconLink style={{ width: '70%', height: '70%' }} stroke={1.5}/>
                                                </ActionIcon>
                                            </Tooltip>
                                        )}
                                        </CopyButton>

                                        {/* Delete button with confirmation Popover */}
                                        <Popover width={200} position="bottom" withArrow shadow="md">
                                            <Popover.Target>
                                                <Tooltip withArrow arrowOffset={10} arrowSize={4} label={t("delete_share")}>
                                                    <ActionIcon id="delete" variant="light" color="red" >
                                                        <IconTrash style={{ width: '70%', height: '70%' }} stroke={1.5}/>
                                                    </ActionIcon>
                                                </Tooltip>
                                            </Popover.Target>
                                            <Popover.Dropdown className={classes.popover}>
                                                <Text ta="center" size="xs" mb="xs">{t("delete_this_share")}</Text>
                                                <Button aria-description="delete" w="100%" variant='default' c='red' size="xs" onClick={() => props.onDelete&&props.onDelete(share.name)}>Delete</Button>
                                            </Popover.Dropdown>
                                        </Popover>
                                        
                                        {/* Edit share properties button */}
                                        <ResponsivePopover withDrawer={!isBrowser}>
                                            <Tooltip withArrow arrowOffset={10} arrowSize={4} label={t("edit_share")}>
                                                <ActionIcon id="edit" variant="light" color="blue" >
                                                    <IconDots style={{ width: '70%', height: '70%' }} stroke={1.5}/>
                                                </ActionIcon>
                                            </Tooltip>
                                            <ShareEditor buttonTitle={t("update")} onChange={setNewOptions} onClick={updateShare} onRenew={renew} options={newOptions}/>
                                        </ResponsivePopover>
                                    </ActionIconGroup>
                                </Flex>
                                
                            </Stack>
                        </Box>
                    </Flex>
                </Dropzone>
            {/* Share component footer */}
        </Paper>
        </>
)}