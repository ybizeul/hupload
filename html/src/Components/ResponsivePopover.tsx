import { Popover, Drawer, Flex} from '@mantine/core';
import { useDisclosure } from '@mantine/hooks';

import React from 'react';

interface ResponsivePopoverProps {
    children: React.ReactElement[];
    withDrawer?: boolean;
}

export function ResponsivePopover(props: ResponsivePopoverProps) {
    const { children } = props;
    const [_actionIcon, child] = children;

    const [opened, { close, open }] = useDisclosure(false);

    const actionIcon = React.cloneElement(_actionIcon, {onClick: () => {opened?close():open()}})

    // Share properties is displayed as a Popover if withDrawer is false
    const popover = (
        <Popover opened={opened} onClose={close} middlewares={{size: true}} withArrow>
            <Popover.Target>
                {actionIcon}
            </Popover.Target>
            <Popover.Dropdown>
                {child && React.cloneElement(child, {close: close})}
            </Popover.Dropdown>
        </Popover>
    )

    // Share properties is displayed as a full screen Drawer if withDrawer is true
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
                        </Drawer.Body>
                    </Flex>
                    </Drawer.Content>
            </Drawer.Root>
        </>
    )

    return (
        props.withDrawer?drawer:popover
    );
}