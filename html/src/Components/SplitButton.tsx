import { Button, Group, ActionIcon, rem, useMantineTheme, Popover, NumberInput, SegmentedControl, Input, Stack } from '@mantine/core';
import { IconChevronDown } from '@tabler/icons-react';
import classes from './SplitButton.module.css';
import { useState } from 'react';
import { useDisclosure } from '@mantine/hooks';

interface SplitButtonProps {
    onClick: () => void;
    onChange: (exposure: string, validity: number|string) => void;
    exposure: string;
    validity: number|string;
    children: React.ReactNode;
}
export function SplitButton(props: SplitButtonProps) {
  const { onClick, onChange, children } = props;
  const theme = useMantineTheme();

  const [opened, { close, open }] = useDisclosure(false);

  const [exposure, setExposure] = useState<string>(props.exposure)
  const [validity, setValidity] = useState<number|string>(props.validity)

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
            <Stack>
            <Input.Wrapper label="Exposure" description="Guests users can :">
                <SegmentedControl className={classes.segmented} value={exposure} data={[{ label: 'Upload', value: 'upload' }, { label: 'Download', value: 'download' }, { label: 'Both', value: 'both' }]} onChange={(v) => { setExposure(v); onChange(v,validity)}} transitionDuration={0} />
            </Input.Wrapper>
                <NumberInput
                label="Validity"
                description="Number of days the share is valid"
                value={validity}
                min={0}
                onChange={(v) => { setValidity(v); onChange(exposure,v)}}
                />
            </Stack>
            <Button mt="sm" w="100%" onClick={() => {onClick(); close();}}>Create</Button>
        </Popover.Dropdown>
        </Popover>
    </Group>
  );
}