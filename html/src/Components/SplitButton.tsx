import { Button, Group, ActionIcon, rem, useMantineTheme, Popover, NumberInput, SegmentedControl, Input, Stack, TextInput, Box, InputWrapper } from '@mantine/core';
import { IconChevronDown, IconChevronLeft, IconChevronRight, IconEye } from '@tabler/icons-react';
import classes from './SplitButton.module.css';
import { useState } from 'react';
import { useDisclosure } from '@mantine/hooks';
import { Share } from '@/hupload';
import '@mantine/tiptap/styles.css';
import { FullHeightTextArea } from './FullHeightTextArea';
import { Message } from './Message';
interface SplitButtonProps {
    onClick: () => void;
    onChange: (props: Share["options"]) => void;
    exposure: string;
    validity: number;
    description: string;
    message: string;
    children: React.ReactNode;
}
export function SplitButton(props: SplitButtonProps) {
  const { onClick, onChange, children } = props;
  const theme = useMantineTheme();

  const [opened, { close, open }] = useDisclosure(false);
  const [expand, handlers] = useDisclosure(false);
  const [preview, previewH] = useDisclosure(false);

  const [exposure, setExposure] = useState<string>(props.exposure)
  const [validity, setValidity] = useState<number>(props.validity)
  const [description, setDescription] = useState<number|string>(props.description)
  const [message, setMessage] = useState<string>(props.message)

  return (
    <Group wrap="nowrap" gap={0} justify='center'>
      <Button onClick={onClick} className={classes.button}>{children}</Button>
      <Popover opened={opened} onClose={close}>
        <Popover.Target>
          <ActionIcon
            variant="filled"
            color={theme.primaryColor}
            size={36}
            className={classes.menuControl}
            onClick={()=> {opened?close():open()}}
          >
            <IconChevronDown style={{ width: rem(16), height: rem(16) }} stroke={1.5} />
          </ActionIcon>
        </Popover.Target>
        <Popover.Dropdown>
          <Group align='stretch'>
            <>
            <Stack style={{position:"relative"}}>
            <ActionIcon variant="light" radius="xl" onClick={handlers.toggle} style={{position:"absolute", top: 0, right: 0}}>
              {expand?
                <IconChevronLeft style={{ width: rem(16), height: rem(16) }} stroke={1.5} />
                :
                <IconChevronRight style={{ width: rem(16), height: rem(16) }} stroke={1.5} />
              }
            </ActionIcon>
            <Input.Wrapper label="Exposure" description="Guests users can :">
                <SegmentedControl className={classes.segmented} value={exposure} data={[{ label: 'Upload', value: 'upload' }, { label: 'Download', value: 'download' }, { label: 'Both', value: 'both' }]}
                  onChange={(v) => { setExposure(v); onChange({exposure: v})}} transitionDuration={0} />
            </Input.Wrapper>
            <NumberInput
              label="Validity"
              description="Number of days the share is valid"
              value={validity}
              min={0}
              onChange={(v) => { setValidity(v as number); onChange({exposure: v.toString()}) }}
              />
            <TextInput label="Description" value={description} onChange={(v) => {setDescription(v.target.value); onChange({description: v.target.value})}}/>
            </Stack>
            </>
            {expand&&
            <Box pl="sm" w="30em" style={{borderLeft: "1px solid lightGray"}}>
              <ActionIcon variant={preview?"filled":"subtle"} radius="xl" onClick={previewH.toggle} style={{position:"absolute", top: 5, right: 5}}>
                <IconEye style={{ width: rem(16), height: rem(16) }} stroke={1.5} />
              </ActionIcon>
              {preview?
              <InputWrapper label="Message" h={"100%"}>
                <Message value={decodeURIComponent(message)} />
              </InputWrapper>
              :
              <FullHeightTextArea label="Message" w="30em" h="90%" value={message} onChange={(v) => {setMessage(v.target.value); onChange({message: v.target.value})}}/>
    }
            </Box>
            }
            </Group>
            <Button mt="sm" w="100%" onClick={() => {onClick(); close();}}>Create</Button>
        </Popover.Dropdown>
        </Popover>
    </Group>
  );
}