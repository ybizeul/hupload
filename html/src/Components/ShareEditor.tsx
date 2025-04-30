import { Share } from "@/hupload";
import { ActionIcon, Box, BoxComponentProps, Button, Flex, Group, Input, NumberInput, rem, SegmentedControl, Stack, TextInput, Tooltip, useMantineTheme } from "@mantine/core";
import { IconChevronLeft, IconChevronRight, IconClockPlus } from "@tabler/icons-react";
import { useDisclosure, useMediaQuery, useUncontrolled } from "@mantine/hooks";
import classes from './ShareEditor.module.css';
import { MarkDownEditor } from "./MarkdownEditor";
import { TemplatesMenu } from "./TemplatesMenu";
import { useTranslation } from "react-i18next";

interface ShareEditorProps {
  onChange: (options: Share["options"]) => void;
  onClick: () => void;
  onRenew?: () => void;
  close?: () => void;
  options: Share["options"];
  buttonTitle: string;
}

export function ShareEditor(props: ShareEditorProps&BoxComponentProps) {
    const { t } = useTranslation()
    // Initialize props
    const { onChange, onClick, onRenew, close, buttonTitle } = props;

    // Initialize state
    const [options, setOptions] = useUncontrolled({
        value: props.options,
        defaultValue: {validity: 7, exposure: "upload"},
        finalValue: {},
        onChange,
      });
 //   const [_options, setOptions] = useState<Share["options"]>(options)

    // Initialize hooks
    const [showMessage, showMessageH ] = useDisclosure(false);
    const theme = useMantineTheme()
    const isInBrowser = useMediaQuery('(min-width: +' + theme.breakpoints.xs + ')');

    // Functions
    const notifyChange = (o: Share["options"]) => {
        setOptions(o)
        onChange(o)
    }

    const renew = () => {
        onRenew && onRenew();
    }
    return (
        <Box miw={rem(200)} h="100%" w="100%" display={"flex"}>
            <Flex direction="column" gap="sm" w="100%" justify={"space-between"}>
                
                {/* Left section */}
                <Flex gap="sm" w="100%" flex="1" direction={{base: 'column', xs: 'row'}}>
                    <Stack display= "flex" style={{position:"relative"}} w={{base: '100%', xs: rem(250)}}>

                        {/* Action icon to show message editor */}
                        {isInBrowser&&
                        <ActionIcon id="showEditor" variant="light" radius="xl" onClick={showMessageH.toggle} style={{position:"absolute", top: 0, right: 0}}>
                            {showMessage?
                            <IconChevronLeft style={{ width: rem(16), height: rem(16) }} stroke={1.5} />
                            :
                            <IconChevronRight style={{ width: rem(16), height: rem(16) }} stroke={1.5} />
                            }
                        </ActionIcon>
                        }

                        {/* Share exposure */}
                        <Input.Wrapper label={t("exposure")} description={t("you_want_to")}>
                            <SegmentedControl 
                                className={classes.segmented} 
                                value={options.exposure} 
                                data={[ { label: t("receive"), value: 'upload' }, 
                                        { label: t("send"), value: 'download' }, 
                                        { label: t("both"), value: 'both' },
                                    ]}
                                onChange={(v) => { notifyChange({...options, exposure:v}); }} transitionDuration={0} 
                            />
                        </Input.Wrapper>

                        {/* Share validity */}
                            <NumberInput
                                label={t("validity")}
                                description={t("number_of_days_the_share_is_valid")}
                                value={options.validity}
                                min={0}
                                classNames={{wrapper: classes.numberInput}}
                                onChange={(v) => { notifyChange({...options, validity:v as number}); }}
                                inputContainer={(children) => (
                                    <>
                                    <Group gap="xs" align="center">
                                        {children}
                                        <Tooltip label={t("renew_share")} withArrow>
                                        <ActionIcon mt="xs" variant="light" radius="xl" onClick={renew} >
                                            <IconClockPlus style={{ width: '70%', height: '70%' }} stroke={1.5}/>
                                        </ActionIcon>
                                        </Tooltip>
                                    </Group>
                                    </>)}
                            />
                        {/* Share description */}
                        <TextInput label={t("description")} value={options.description}
                            onChange={(v) => { notifyChange({...options, description:v.target.value}); }}
                        />
                    </Stack>

                {/* Right section */}
                {(showMessage||!isInBrowser)&&
                    <Box display="flex" flex="1" w={{base: '100%', xs: rem(500)}} pos={"relative"}>
                        <MarkDownEditor 
                            pl={isInBrowser?"sm":"0"} 
                            style={{borderLeft: isInBrowser?"1px solid lightGray":""}} 
                            onChange={(v) => { notifyChange({...options, message:v}); }}
                            message={options.message?options.message:""}
                        />
                        <TemplatesMenu onChange={(v) => { notifyChange({...options, message: v}); }}/>
                    </Box>
                    
                }
                </Flex>
                <Flex >
                    <Button w="100%" onClick={() => {onClick(); close && close();}}>{buttonTitle}</Button>
                </Flex>
            </Flex>
        </Box>
    )
}
