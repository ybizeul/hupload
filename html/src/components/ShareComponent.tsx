import { Anchor, Group, Paper, Text } from "@mantine/core";
import { Share } from "../hupload";
import { Link } from "react-router-dom";
import classes from './ShareComponent.module.css';

export function ShareComponent(props: {share: Share}) {
    const {share} = props
    const key = share.name
    const name = share.name
    return (
    <Paper key={key} p="md" shadow="xs" radius="md" mt={10} className={classes.paper}>
        <Group justify="space-between" h={45}>
            <Anchor component={Link} to={'/'+name}><Text>{name}</Text></Anchor>
        </Group>
    </Paper>
)}