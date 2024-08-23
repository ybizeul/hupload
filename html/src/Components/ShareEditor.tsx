import { Share } from "@/hupload";
import { ActionIcon, Box, Button, Group, Input, InputWrapper, NumberInput, rem, SegmentedControl, Stack, TextInput } from "@mantine/core";
import { IconChevronLeft, IconChevronRight, IconEye } from "@tabler/icons-react";
import { Message } from "./Message";
import { useDisclosure } from "@mantine/hooks";
import { FullHeightTextArea } from "./FullHeightTextArea";
import { useState } from "react";
import classes from './ShareEditor.module.css';

export function ShareEditor(props: { onChange: (options: Share["options"]) => void, onClick: () => void, options: Share["options"] }) {
    const { onChange, onClick, options } = props;
    const { exposure, validity, description, message } = options;

    const [_exposure, setExposure] = useState<string|undefined>(exposure)
    const [_validity, setValidity] = useState<number|undefined>(validity)
    const [_description, setDescription] = useState<string|undefined>(description)
    const [_message, setMessage] = useState<string|undefined>(message)

    const [mdPanel, mdPanelH ] = useDisclosure(false);
    const [preview, previewH] = useDisclosure(false);

    const notifyChange = () => {
        onChange({
            exposure: _exposure as string,
            validity: _validity as number,
            description: _description as string,
            message: _message as string
        })
    }
    return (
        <>
        <Group align='stretch'>
            <>
            <Stack style={{position:"relative"}}>
            <ActionIcon variant="light" radius="xl" onClick={mdPanelH.toggle} style={{position:"absolute", top: 0, right: 0}}>
              {mdPanel?
                <IconChevronLeft style={{ width: rem(16), height: rem(16) }} stroke={1.5} />
                :
                <IconChevronRight style={{ width: rem(16), height: rem(16) }} stroke={1.5} />
              }
            </ActionIcon>
            <Input.Wrapper label="Exposure" description="Guests users can :">
                <SegmentedControl className={classes.segmented} value={_exposure} data={[{ label: 'Upload', value: 'upload' }, { label: 'Download', value: 'download' }, { label: 'Both', value: 'both' }]}
                  onChange={(v) => { setExposure(v); notifyChange();}} transitionDuration={0} />
            </Input.Wrapper>
            <NumberInput
              label="Validity"
              description="Number of days the share is valid"
              value={_validity}
              min={0}
              onChange={(v) => { setValidity(v as number); notifyChange(); }}
              />
            <TextInput label="Description" value={_description} onChange={(v) => {setDescription(v.target.value); notifyChange();}}/>
            </Stack>
            </>
            {mdPanel&&
            <Box pl="sm" w="30em" style={{borderLeft: "1px solid lightGray"}}>
              <ActionIcon variant={preview?"filled":"subtle"} radius="xl" onClick={previewH.toggle} style={{position:"absolute", top: 5, right: 5}}>
                <IconEye style={{ width: rem(16), height: rem(16) }} stroke={1.5} />
              </ActionIcon>
              {preview?
              <InputWrapper label="Message" h={"100%"}>
                <Message value={_message?decodeURIComponent(_message):""} />
              </InputWrapper>
              :
              <FullHeightTextArea label="Message" w="30em" h="90%" value={message} onChange={(v) => {setMessage(v.target.value); notifyChange();}}/>
    }
            </Box>
            }
            </Group>
            <Button mt="sm" w="100%" onClick={() => {onClick(); close();}}>Create</Button>
        </>
        )
}