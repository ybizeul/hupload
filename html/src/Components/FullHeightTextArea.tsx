import { Textarea, TextareaProps } from "@mantine/core";
import { useState } from "react";

import classes from './FullHeightTextArea.module.css';

// interface FullHeightTextAreaProps extends TextareaProps {
//     value: string | number | readonly string[];
//     label: string;
//     onChange: ChangeEventHandler<HTMLTextAreaElement>;
// }
export function FullHeightTextArea(props: TextareaProps) {
    const { value, onChange } = props;
    const [currentValue, setCurrentValue] = useState<string | number | readonly string[] | undefined>(value);

    return (
        <Textarea classNames={{wrapper: classes.wrapper, input: classes.input}} w="30em" {...props} value={currentValue} onChange={(v) => {setCurrentValue(v.target.value); onChange&&onChange(v)}}/>
    )
}