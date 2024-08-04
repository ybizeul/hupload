import { ActionIcon, Anchor, Button, CopyButton, Flex, Group, Paper, Popover, Text, Tooltip } from "@mantine/core";
import { humanFileSize, Share } from "../hupload";
import { Link } from "react-router-dom";
import classes from './ShareComponent.module.css';
import { IconLink, IconTrash } from "@tabler/icons-react";
import { useState } from "react";
import { H } from "@/APIClient";

export function ShareComponent(props: {share: Share}) {
    // Initialize props
    const {share} = props

    // State
    const [deleted,setDeleted] = useState(false)

    // Constants
    const key = share.name
    const name = share.name
    const count = share.count
    const size = share.size
    const countString = count?(count + ' item' + (count > 1 ? 's' : '')):"empty"

    // Function
    const deleteShare = () => {
        H.delete('/share/'+name).then(() => {
            setDeleted(true)
        })
    }
    return (
        <>
        {!deleted &&
        <Paper key={key} p="md" shadow="xs" radius="md" mt={10} className={classes.paper}>
            <Flex h={45} align={"center"}>
                <Group flex="1" gap="0" align="baseline">
                    <Anchor style={{ whiteSpace: "nowrap"}} flex={"1"} component={Link} to={'/'+name}><Text>{name}</Text></Anchor>
                    <Text size="xs" c="gray">{countString + (size?(' | ' + humanFileSize(size)):'')}</Text>
                </Group>
                <Group justify="flex-end">
                    <CopyButton value={window.location.protocol + '//' + window.location.host + '/' + name}>
                    {({ copied, copy }) => (
                        <Tooltip withArrow arrowOffset={10} arrowSize={4} label="Copy URL">
                            <ActionIcon ml="sm" variant="light" color={copied ? 'teal' : 'blue'} onClick={copy} >
                            <IconLink style={{ width: '70%', height: '70%' }} stroke={1.5}/>
                            </ActionIcon>
                        </Tooltip>
                    )}
                    </CopyButton>
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
        </Paper>
        }
        </>
)}