import { Textarea, TextareaProps } from "@mantine/core";
import { ChangeEvent, useState } from "react";

import classes from './FullHeightTextArea.module.css';

// interface FullHeightTextAreaProps extends TextareaProps {
//     value: string | number | readonly string[];
//     label: string;
//     onChange: ChangeEventHandler<HTMLTextAreaElement>;
// }
export function FullHeightTextArea(props: TextareaProps) {
    const { value, onChange } = props;
    const [_value, setValue] = useState<string | number | readonly string[] | undefined>(value);

    function notifyChange(v: ChangeEvent<HTMLTextAreaElement>) {
        setValue(v.target.value)
        onChange && onChange(v)
    }

    return (
        <Textarea {...props} resize="vertical" value={_value} 
            classNames={{root: classes.root, wrapper: classes.wrapper, input: classes.input}}
            onChange={(v) => { notifyChange(v) }}
        />
    )
}