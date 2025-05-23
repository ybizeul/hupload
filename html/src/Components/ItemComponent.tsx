import { ActionIcon, Box, Button, Center, Flex, Paper, Popover, rem, RingProgress, Text, ThemeIcon, Tooltip } from "@mantine/core";
import { IconCheck, IconDownload, IconTrash, IconX } from "@tabler/icons-react";
import { humanFileSize, UploadableItem } from "../hupload";
import classes from './ItemComponent.module.css';
import { ReactNode } from "react";
import { useTranslation } from "react-i18next";

export function ItemComponent(props: {download: boolean, onDelete?: (item:string) => void, canDelete: boolean, item: UploadableItem}) {
    const { t } = useTranslation();

    // Initialize props
    const {download, canDelete, onDelete, item} = props

    // key is item Path, or file name for files being added
    const key = item.Path

    // fileName is item Path last path component, or file name for files being 
    // added
    const fileName = item.Path.split('/')[1]
    
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
            {item.QueueItem?
            addTooltip(item.QueueItem.failed?item.QueueItem.error:"",<RingProgress
                size={45}
                thickness={3}
                sections={[
                { value: (item.QueueItem.finished)?(100):(100*item.QueueItem.loaded/item.QueueItem.total), color: (item.QueueItem.failed)?('red'):((item.QueueItem.finished)?'teal':'blue')},
                ]}
                label={
                (item.QueueItem.failed)?
                (<Center>
                    <ThemeIcon color="red" variant="light" radius="xl" size="xl">
                        <IconX style={{ width: rem(20), height: rem(20) }} />
                    </ThemeIcon>
                </Center>) // item.QueueItem.failed
                :
                ((item.QueueItem.finished)?
                (<Center>
                    <ThemeIcon color="teal" variant="light" radius="xl" size="xl">
                        <IconCheck style={{ width: rem(20), height: rem(20) }} />
                    </ThemeIcon>
                </Center>) // item.QueueItem.finished
                :
                <Text c="blue" fw={700} size="xs" ta="center" >
                    {(100*item.QueueItem.loaded/item.QueueItem.total).toFixed(0) + '%'}
                </Text>) // item.QueueItem not finished
                }
            />)
            :
            <>
                <Text size="xs" c="gray" style={{whiteSpace: "nowrap"}}>{humanFileSize(item.ItemInfo.Size)}</Text>
                {canDelete &&
                    <Popover width={200} position="bottom" withArrow shadow="md">
                        <Popover.Target>
                            <Tooltip withArrow arrowOffset={10} arrowSize={4} label={t("delete_file")}>
                                <ActionIcon ml="sm" variant="light" color="red" >
                                    <IconTrash style={{ width: '70%', height: '70%' }} stroke={1.5}/>
                                </ActionIcon>
                            </Tooltip>
                        </Popover.Target>
                        <Popover.Dropdown className={classes.popover}>
                            <Text ta="center" size="xs" mb="xs">{t("delete_this_item")}</Text>
                            <Button aria-description="delete" w="100%" variant='default' c='red' size="xs" onClick={() => {onDelete&&onDelete(item?.Path.split("/")[1])}}>{t("delete_file")}</Button>
                        </Popover.Dropdown>
                    </Popover>
                }
                {download &&
                <ActionIcon ml="sm" component="a" href={'/d/'+item?.Path.split("/")[0]+'/'+item?.Path.split("/")[1]} aria-label="Download" variant="light" color="blue">
                    <Tooltip withArrow arrowOffset={10} arrowSize={4} label={t("download_button")}>
                        <IconDownload style={{ width: '70%', height: '70%' }} stroke={1.5} />
                    </Tooltip>
                </ActionIcon>
                }
                </>
           }
            </Flex>
            </Box>
            </Flex>
        </Paper>
)}