import { ActionIcon, Anchor, Box, Button, CopyButton, Flex, Group, Paper, Popover, Stack, Text, Tooltip } from "@mantine/core";
import { humanFileSize, prettyfiedCount, Share } from "../hupload";
import { Link } from "react-router-dom";
import classes from './ShareComponent.module.css';
import { IconClock, IconLink, IconTrash } from "@tabler/icons-react";
import { useState } from "react";
import { H } from "@/APIClient";

export function ShareComponent(props: {share: Share}) {
    // Initialize props
    const { share } = props

    // Initialize States
    const [deleted,setDeleted] = useState(false)

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

    if (deleted) {
        return
    }

    return (
        <>
        <Paper key={key} withBorder shadow="xs" radius="md" mt={10} pos="relative" className={classes.paper}>

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
                        <Group justify="flex-end" gap="xs" wrap="nowrap" align="baseline">

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

                        </Group>
                    </Flex>
                    {share.options.description && <Text w="100%" size="xs" c="gray">{share.options.description}</Text>}</Stack>
            </Box>
        </Paper>
        </>
)}