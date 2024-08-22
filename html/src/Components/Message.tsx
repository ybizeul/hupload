import './Message.css';

import { Paper, TypographyStylesProvider } from "@mantine/core";
import { marked } from 'marked';

export function Message(props: {value: string, mb?: string, mt?: string}) {
    const { value, mb, mt } = props;
    return (
        <Paper withBorder mb={mb} mt={mt} p="sm" h="90%">
            <TypographyStylesProvider h={"100%"}>
                <div className="message" dangerouslySetInnerHTML={{ __html: marked.parse(value)}}></div>
            </TypographyStylesProvider>
        </Paper>
    )
}