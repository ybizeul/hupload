import { Textarea } from "@mantine/core";

import classes from './FullHeightTextArea.module.css';
import { useUncontrolled } from "@mantine/hooks";

interface FullHeightTextAreaProps {
    value?: string;
    onChange?: (value: string) => void;
}
export function FullHeightTextArea(props: FullHeightTextAreaProps) {
    const { value, onChange } = props;

    const [_value, handleChange] = useUncontrolled({
        value,
        onChange,
      });

    return (
        <Textarea w="100%" flex="1" label="Message" description="This markdown will be displayed to the user"  resize="vertical" value={_value} 
            classNames={{root: classes.root, wrapper: classes.wrapper, input: classes.input}}
            onChange={(e) => { handleChange(e.currentTarget.value) }}
        />
    )
}