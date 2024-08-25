import { ActionIcon, Box, BoxComponentProps, InputWrapper, rem } from "@mantine/core";
import { useDisclosure } from "@mantine/hooks";
import { IconEye } from "@tabler/icons-react";
import { useState } from "react";
import { Message } from "../Message";
import { FullHeightTextArea } from "./FullHeightTextArea";

interface MarkDownEditorProps {
    onChange: (message: string) => void;
    markdown: string|undefined;
  }
  
  export function MarkDownEditor(props: MarkDownEditorProps&BoxComponentProps) {
    // Initialize props
    const { onChange, markdown } = props;
  
    // Initialize state
    const [_markdown, setMarkdown] = useState<string|undefined>(markdown);
    const [preview, previewH] = useDisclosure(false);
  
    // Functions
    const notifyChange = (m: string) => {
        setMarkdown(m)
        onChange(m)
    }
  
    return(
        <Box display="flex" flex="1" w={{base: '100%', xs: rem(500)}} pl={props.pl} style={props.style} pos={"relative"}>
            <ActionIcon size="xs" variant={preview?"filled":"subtle"} m={rem(3)} radius="xl" onClick={previewH.toggle} style={{position:"absolute", top: 0, right: 0}}>
                <IconEye style={{ width: rem(16), height: rem(16) }} stroke={1.5} />
            </ActionIcon>
            {preview?
            <InputWrapper display="flex" style={{flexDirection:"column"}} label="Message" description="This markdown will be displayed to the user" w="100%">
                <Message mt="5" value={_markdown?decodeURIComponent(_markdown):""} />
            </InputWrapper>
            :
            <FullHeightTextArea w="100%" flex="1" label="Message" description="This markdown will be displayed to the user" value={markdown} onChange={(v) => { notifyChange(v.target.value); }}/>
            }
        </Box>
    )
  }