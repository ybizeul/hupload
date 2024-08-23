import { Button, Group, ActionIcon, rem, useMantineTheme, Popover, Drawer} from '@mantine/core';
import { IconChevronDown } from '@tabler/icons-react';
import classes from './SplitButton.module.css';
import { useDisclosure, useMediaQuery } from '@mantine/hooks';
import { Share } from '@/hupload';
import { ShareEditor } from './ShareEditor';

interface SplitButtonProps {
    onClick: () => void;
    onChange: (props: Share["options"]) => void;
    options: Share["options"];
    children: React.ReactNode;
}

export function SplitButton(props: SplitButtonProps) {
  const { onClick, onChange, options, children } = props;
  const theme = useMantineTheme();

  const [opened, { close, open }] = useDisclosure(false);

  const matches = useMediaQuery('(min-width: +' + theme.breakpoints.xs + ')');

  return (
    <Group wrap="nowrap" gap={0} justify='center'>
      <Button onClick={onClick} className={classes.button}>{children}</Button>
      {matches?
      <Popover opened={opened} onClose={close} middlewares={{size: true}}>
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
          <ShareEditor onChange={onChange} onClick={() => {onClick();close();}} options={options}/>
        </Popover.Dropdown>
      </Popover>
      :
      <>
        <ActionIcon
          variant="filled"
          color={theme.primaryColor}
          size={36}
          className={classes.menuControl}
          onClick={()=> {opened?close():open()}}
        >
          <IconChevronDown style={{ width: rem(16), height: rem(16) }} stroke={1.5} />
        </ActionIcon>
        <Drawer size="100%" opened={opened} onClose={close} title="Share Properties" position="top">
          <ShareEditor onChange={onChange} onClick={() => {onClick();close();}} options={options}/>
        </Drawer>
      </>
        }
    </Group>
  );
}