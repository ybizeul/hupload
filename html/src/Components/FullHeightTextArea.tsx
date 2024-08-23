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
        <Textarea resize="vertical" classNames={{root: classes.root, wrapper: classes.wrapper, input: classes.input}} {...props} 
            value={currentValue} 
            onChange={(v) => {setCurrentValue(v.target.value); onChange&&onChange(v)}}/>
    )
}