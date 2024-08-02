import { ActionIcon, Anchor, Button, CopyButton, Group, Paper, Popover, Text, Tooltip } from "@mantine/core";
import { Share } from "../hupload";
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
            <Group justify="space-between" h={45}>
                <Anchor component={Link} to={'/'+name}><Text>{name}</Text></Anchor>
                <Group>
                    <CopyButton value={window.location.protocol + '//' + window.location.host + '/' + name}>
                    {({ copied, copy }) => (
                        <Tooltip withArrow arrowOffset={10} arrowSize={4} label="Copy URL">
                            <ActionIcon variant="light" color={copied ? 'teal' : 'blue'} onClick={copy} >
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
            </Group>
        </Paper>
        }
        </>
)}