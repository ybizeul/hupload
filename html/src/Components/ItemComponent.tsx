import { ActionIcon, Box, Center, Flex, Paper, rem, RingProgress, Text, Tooltip } from "@mantine/core";
import { IconCheck, IconDownload, IconX } from "@tabler/icons-react";
import { QueueItem } from "../UploadQueue";
import { humanFileSize, Item } from "../hupload";
import { useLoggedInContext } from "../LoggedInContext";
import classes from './ItemComponent.module.css';
import { ReactNode } from "react";

export function ItemComponent(props: {item?: Item, queueItem?: QueueItem}) {

    // Initialize props
    const {item, queueItem} = props

    // Initialize hooks
    const {loggedIn} = useLoggedInContext()

    // Other initializations

    // key is item Path, or file name for files being added
    const key = item?item.Path:queueItem?.file.name

    // fileName is item Path last path component, or file name for files being 
    // added
    const fileName = item?item.Path.split('/')[1]:queueItem?.file.name
    
    // Function to add tooltip to an element, used to display error message
    const addTooltip = (tooltip: string,element: ReactNode) => {
        if (tooltip === '') {
            return (element)
        } else {
            return(
            <Tooltip label={tooltip}>
                {element}
            </Tooltip>
            )
        }
    }
    return (
    <Paper key={key} withBorder p="md" shadow="xs" radius="md" mt={10} className={classes.paper}>
        <Flex direction="row" align="center" h={45}>
            <Text truncate="end">{fileName}</Text>
            <Box flex={1} ta={"right"}>
            <Flex align={"center"} justify={"right"}>
            {queueItem?
            addTooltip(queueItem.failed?queueItem.error:"",<RingProgress
                size={45}
                thickness={3}
                sections={[
                { value: (queueItem.finished)?(100):(100*queueItem.loaded/queueItem.total), color: (queueItem.failed)?('red'):((queueItem.finished)?'teal':'blue')},
                ]}
                label={
                (queueItem.failed)?
                (<Center>
                    <ActionIcon color="red" variant="light" radius="xl" size="xl">
                        <IconX style={{ width: rem(22), height: rem(22) }} />
                    </ActionIcon>
                </Center>) // queueItem.failed
                :
                ((queueItem.finished)?
                (<Center>
                    <ActionIcon color="teal" variant="light" radius="xl" size="xl">
                        <IconCheck style={{ width: rem(22), height: rem(22) }} />
                    </ActionIcon>
                </Center>) // queueItem.finished
                :
                <Text c="blue" fw={700} size="xs" ta="center" >
                    {(100*queueItem.loaded/queueItem.total).toFixed(0) + '%'}
                </Text>) // queueItem not finished
                }
            />)
            :item&&
            <>
                <Text size="xs" c="gray" style={{whiteSpace: "nowrap"}}>{humanFileSize(item.ItemInfo.Size)}</Text>
                {loggedIn&&
                <ActionIcon ml="sm" component="a" href={'/api/v1/shares/'+item?.Path.split("/")[0]+'/items/'+item?.Path.split("/")[1]} aria-label="Download" variant="light" color="blue">
                    <IconDownload style={{ width: '70%', height: '70%' }} stroke={1.5} />
                </ActionIcon>
                }
                </>
           }
            </Flex>
            </Box>
            </Flex>
        </Paper>
)}