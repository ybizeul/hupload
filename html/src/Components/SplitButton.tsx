import { Button, Group, ActionIcon, rem, useMantineTheme, Popover, Drawer, Flex} from '@mantine/core';
import { IconChevronDown } from '@tabler/icons-react';
import classes from './SplitButton.module.css';
import { useDisclosure } from '@mantine/hooks';

import React from 'react';

interface SplitButtonProps {
    onClick: () => void;
    value: string;
    children: React.ReactNode;
    withDrawer?: boolean;
}

export function SplitButton(props: SplitButtonProps) {
  const { onClick, value, children } = props;
  const child = children as React.ReactElement;

  const theme = useMantineTheme();

  const [opened, { close, open }] = useDisclosure(false);

  const actionIcon = (
    <ActionIcon
      variant="filled"
      color={theme.primaryColor}
      size={36}
      className={classes.menuControl}
      onClick={()=> {opened?close():open()}}
    >
      <IconChevronDown style={{ width: rem(16), height: rem(16) }} stroke={1.5} />
    </ActionIcon>
  )

  // Share properties is displayed as a Popover if screen size > xs
  const popover = (
    <Popover opened={opened} onClose={close} middlewares={{size: true}}>
        <Popover.Target>
          {actionIcon}
        </Popover.Target>
        <Popover.Dropdown>
            {child&&React.cloneElement(child, {close: close})}
          {/* <ShareEditor onChange={onChange} onClick={() => {onClick();close();}} options={options}/> */}
        </Popover.Dropdown>
      </Popover>
  )

  // Share properties is displayed as a full screen Drawer if screen size < xs
  const drawer = (
    <>
        {actionIcon}
        <Drawer.Root size="100%" opened={opened} onClose={close} position="top">
                <Drawer.Overlay />
                <Drawer.Content w="100%" style={{display: "flex", flexGrow: 1, flexDirection: "column", justifyContent: 'space-between'}}>
                <Drawer.Header>
                    <Drawer.Title>Share Properties</Drawer.Title>
                    <Drawer.CloseButton />
                </Drawer.Header>
                <Flex flex="1" align={"stretch"}>
                    <Drawer.Body flex="1" pt="0">
                    {child&&React.cloneElement(child, {close: close})}
                    {/* <ShareEditor onChange={onChange} onClick={() => {onClick();close();}} options={options}/> */}
                    </Drawer.Body>
                </Flex>
                </Drawer.Content>
        </Drawer.Root>
    </>
  )

  return (
    <Group wrap="nowrap" gap={0} justify='center'>
        <Button onClick={onClick} className={classes.button}>{value}</Button>
        {props.withDrawer?drawer:popover}
    </Group>
  );
}