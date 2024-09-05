import { ActionIcon, Box, BoxComponentProps, InputWrapper, Paper, rem } from "@mantine/core";
import { useDisclosure } from "@mantine/hooks";
import { IconEye } from "@tabler/icons-react";
import { useState } from "react";
import { Message } from "../Message";
import { FullHeightTextArea } from "./FullHeightTextArea";

interface MarkDownEditorProps {
    onChange: (message: string) => void;
  }
  
  export function MarkDownEditor(props: MarkDownEditorProps&BoxComponentProps&{children: string}) {
    // Initialize props
    const { onChange, children } = props;
  
    // Initialize state
    const [markdown, setMarkdown] = useState<string>(children);
    const [preview, previewH] = useDisclosure(false);
  
    // Functions
    const notifyChange = (m: string) => {
        setMarkdown(m)
        onChange(m)
    }
  
    return(
        <Box display="flex" flex="1" w={{base: '100%', xs: rem(500)}} pl={props.pl} style={props.style} pos={"relative"}>
            <ActionIcon size="xs" id="preview" variant={preview?"filled":"subtle"} m={rem(3)} radius="xl" onClick={previewH.toggle} style={{position:"absolute", top: 0, right: 0}}>
                <IconEye style={{ width: rem(16), height: rem(16) }} stroke={1.5} />
            </ActionIcon>
            {preview?
                <InputWrapper display="flex" style={{flexDirection:"column"}} label="Message" description="This markdown will be displayed to the user" w="100%">
                    <Paper flex="1" withBorder mt="5" pt="5.5" px="12" display="flex">
                        <Message value={markdown} />
                    </Paper>
                </InputWrapper>
            :
            <FullHeightTextArea w="100%" flex="1" label="Message" description="This markdown will be displayed to the user" value={markdown} onChange={(v) => { notifyChange(v.target.value); }}/>
            }
        </Box>
    )
  }