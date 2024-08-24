import { Share } from "@/hupload";
import { ActionIcon, Box, BoxComponentProps, Button, Flex, Input, InputWrapper, NumberInput, rem, SegmentedControl, Stack, TextInput, useMantineTheme } from "@mantine/core";
import { IconChevronLeft, IconChevronRight, IconEye } from "@tabler/icons-react";
import { Message } from "./Message";
import { useDisclosure, useMediaQuery } from "@mantine/hooks";
import { FullHeightTextArea } from "./FullHeightTextArea";
import { useState } from "react";
import classes from './ShareEditor.module.css';

interface ShareEditorProps {
  onChange: (options: Share["options"]) => void;
  onClick: () => void;
  options: Share["options"];
}

export function ShareEditor(props: ShareEditorProps&BoxComponentProps) {
    const { onChange, onClick, options } = props;

    const [_options, setOptions] = useState<Share["options"]>(options)

    const [mdPanel, mdPanelH ] = useDisclosure(false);

    const theme = useMantineTheme()
    const matches = useMediaQuery('(min-width: +' + theme.breakpoints.xs + ')');

    const notifyChange = (o: Share["options"]) => {
      setOptions(o)
      onChange(o)
    }

    return (
      <Box miw={rem(200)} h="100%" w="100%" display={"flex"}>
        <Flex direction="column" gap="sm" w="100%" justify={"space-between"}>
          <Flex gap="sm" w="100%" flex="1" direction={{base: 'column', xs: 'row'}}>
            <Stack display= "flex" style={{position:"relative"}} w={{base: '100%', xs: rem(250)}}>
              {matches&&
              <ActionIcon variant="light" radius="xl" onClick={mdPanelH.toggle} style={{position:"absolute", top: 0, right: 0}}>
                {mdPanel?
                  <IconChevronLeft style={{ width: rem(16), height: rem(16) }} stroke={1.5} />
                  :
                  <IconChevronRight style={{ width: rem(16), height: rem(16) }} stroke={1.5} />
                }
              </ActionIcon>
              }
              <Input.Wrapper label="Exposure" description="Guests users can :">
                  <SegmentedControl className={classes.segmented} value={_options.exposure} data={[{ label: 'Upload', value: 'upload' }, { label: 'Download', value: 'download' }, { label: 'Both', value: 'both' }]}
                    onChange={(v) => { notifyChange({..._options, exposure:v}); }} transitionDuration={0} />
              </Input.Wrapper>
              <NumberInput
                label="Validity"
                description="Number of days the share is valid"
                value={_options.validity}
                min={0}
                onChange={(v) => { notifyChange({..._options, validity:v as number}); }}
                />
              <TextInput label="Description" value={_options.description} onChange={(v) => { notifyChange({..._options, description:v.target.value}); }}/>
            </Stack>
            {(mdPanel||!matches)&&
            <MarkDownEditor 
              pl={matches?"sm":"0"} 
              style={{borderLeft: matches?"1px solid lightGray":""}} 
              onChange={(v) => { notifyChange({..._options, message:v}); }}
              message={_options.message}/>
            }
          </Flex>
          <Flex >
            <Button w="100%" onClick={onClick}>Create</Button>
          </Flex>
        </Flex>
      </Box>
    )
}

interface MarkDownEditorProps {
  onChange: (message: string) => void;
  message: string|undefined;
}

function MarkDownEditor(props: MarkDownEditorProps&BoxComponentProps) {
  const [message, setMessage] = useState<string|undefined>(props.message);
  const [preview, previewH] = useDisclosure(false);
  return(
    <Box display="flex" flex="1" w={{base: '100%', xs: rem(500)}} pl={props.pl} style={props.style} pos={"relative"}>
      <ActionIcon size="xs" variant={preview?"filled":"subtle"} m={rem(3)} radius="xl" onClick={previewH.toggle} style={{position:"absolute", top: 0, right: 0}}>
        <IconEye style={{ width: rem(16), height: rem(16) }} stroke={1.5} />
      </ActionIcon>
      {preview?
      <InputWrapper display="flex" style={{flexDirection:"column"}} label="Message" description="This markdown will be displayed to the user" w="100%">
        <Message mt="5" value={message?decodeURIComponent(message):""} />
      </InputWrapper>
      :
      <FullHeightTextArea w="100%" flex="1" label="Message" description="This markdown will be displayed to the user" value={message} onChange={(v) => {setMessage(v.target.value); props.onChange(v.target.value);}}/>
      }
    </Box>
  )
}