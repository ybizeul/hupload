import { Button, Group, ActionIcon, rem, useMantineTheme, Popover} from '@mantine/core';
import { IconChevronDown } from '@tabler/icons-react';
import classes from './SplitButton.module.css';
import { useDisclosure } from '@mantine/hooks';
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
          <ShareEditor onChange={onChange} onClick={onClick} options={options}/>
        </Popover.Dropdown>
        </Popover>
    </Group>
  );
}