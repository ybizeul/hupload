import './Message.css';

import { Paper, TypographyStylesProvider } from "@mantine/core";
import { marked } from 'marked';

export function Message(props: {value: string, mb?: string, mt?: string}) {
    const { value, mb, mt } = props;
    return (
        <Paper flex="1" withBorder mb={mb} mt={mt} pt="5.5" px="12" display="flex">
            <TypographyStylesProvider flex="1" fs="sm">
                <div className="message" dangerouslySetInnerHTML={{ __html: marked.parse(value)}}></div>
            </TypographyStylesProvider>
        </Paper>
    )
}