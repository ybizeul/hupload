import { ActionIcon, Anchor, Center, Group, Paper, rem, RingProgress, Text } from "@mantine/core";
import { IconCheck } from "@tabler/icons-react";
import { QueueItem } from "../UploadQueue";
import { humanFileSize, Item } from "../hupload";
import { useLoggedInContext } from "../LoggedInContext";
import classes from './ItemComponent.module.css';

export function ItemComponent(props: {item?: Item, queueItem?: QueueItem}) {
    const {item, queueItem} = props
    const key = item?item.Path:queueItem?.file.name
    const fileName = item?item.Path.split('/')[1]:queueItem?.file.name
    const {loggedIn} = useLoggedInContext()
    
    return (
    <Paper key={key} p="md" shadow="xs" radius="md" mt={10} className={classes.paper}>
        <Group justify="space-between" h={45}>
            {loggedIn&&!queueItem?<Anchor href={'/api/v1/share/'+item?.Path}>{fileName}</Anchor>:<Text>{fileName}</Text>}
            {queueItem?
            <RingProgress
                size={45}
                thickness={4}
                sections={[
                { value: queueItem.finished?100:100*queueItem.loaded/queueItem.total, color: (queueItem.finished)?'teal':'blue'},
                ]}
                label={
                (queueItem.finished)?
                <Center>
                    <ActionIcon color="teal" variant="light" radius="xl" size="xl">
                    <IconCheck style={{ width: rem(22), height: rem(22) }} />
                    </ActionIcon>
                </Center>
                :
                <Text c="blue" fw={700} ta="center" size={rem(10)}>
                    {(100*queueItem.loaded/queueItem.total).toFixed(0) + '%'}
                </Text>
                }
            />
            :item&&
            <Text>{humanFileSize(item.ItemInfo.Size)}</Text>}
        </Group>
    </Paper>
)}