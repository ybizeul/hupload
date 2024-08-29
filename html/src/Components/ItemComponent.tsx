import { ActionIcon, Box, Button, Center, Flex, Paper, Popover, rem, RingProgress, Text, Tooltip } from "@mantine/core";
import { IconCheck, IconDownload, IconTrash, IconX } from "@tabler/icons-react";
import { QueueItem } from "../UploadQueue";
import { humanFileSize, Item } from "../hupload";
import classes from './ItemComponent.module.css';
import { ReactNode } from "react";

export function ItemComponent(props: {download: boolean, onDelete?: (item:string) => void, canDelete: boolean, item?: Item, queueItem?: QueueItem}) {

    // Initialize props
    const {download, canDelete, onDelete, item, queueItem} = props

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
                        <IconX style={{ width: rem(20), height: rem(20) }} />
                    </ActionIcon>
                </Center>) // queueItem.failed
                :
                ((queueItem.finished)?
                (<Center>
                    <ActionIcon color="teal" variant="light" radius="xl" size="xl">
                        <IconCheck style={{ width: rem(20), height: rem(20) }} />
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
                {canDelete &&
                    <Popover width={200} position="bottom" withArrow shadow="md">
                        <Popover.Target>
                            <Tooltip withArrow arrowOffset={10} arrowSize={4} label="Delete File">
                                <ActionIcon ml="sm" variant="light" color="red" >
                                    <IconTrash style={{ width: '70%', height: '70%' }} stroke={1.5}/>
                                </ActionIcon>
                            </Tooltip>
                        </Popover.Target>
                        <Popover.Dropdown className={classes.popover}>
                            <Text ta="center" size="xs" mb="xs">Delete this item ?</Text>
                            <Button aria-description="delete" w="100%" variant='default' c='red' size="xs" onClick={() => {onDelete&&onDelete(item?.Path.split("/")[1])}}>Delete</Button>
                        </Popover.Dropdown>
                    </Popover>
                }
                {download &&
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