import './Message.css';

import { BoxComponentProps, Paper, TypographyStylesProvider } from "@mantine/core";
import { marked } from 'marked';

interface MessageProps {
    value: string;
}
export function Message(props: MessageProps&BoxComponentProps){
    const { value } = props;
    return (
        <Paper {...props} flex="1" withBorder pt="5.5" px="12" display="flex">
            <TypographyStylesProvider flex="1" fs="sm" maw="100%">
                <div className="message" dangerouslySetInnerHTML={{ __html: marked.parse(value)}}></div>
            </TypographyStylesProvider>
        </Paper>
    )
}