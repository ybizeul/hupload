import './Message.css';

import { TypographyStylesProvider, TypographyStylesProviderProps } from "@mantine/core";
import { marked } from 'marked';

interface MessageProps {
    value: string;
}
export function Message(props: MessageProps&TypographyStylesProviderProps){
    const { value } = props;
    return (
        <TypographyStylesProvider {...props} flex="1" fs="sm" maw="100%">
            <div className="message" dangerouslySetInnerHTML={{ __html: marked.parse(value)}}></div>
        </TypographyStylesProvider>
    )
}