import { Textarea } from "@mantine/core";

import classes from './FullHeightTextArea.module.css';
import { useUncontrolled } from "@mantine/hooks";
import { useTranslation } from "react-i18next";

interface FullHeightTextAreaProps {
    value?: string;
    onChange?: (value: string) => void;
}
export function FullHeightTextArea(props: FullHeightTextAreaProps) {
    const {t} = useTranslation()
    const { value, onChange } = props;

    const [_value, handleChange] = useUncontrolled({
        value,
        onChange,
      });

    return (
        <Textarea w="100%" flex="1" label="Message" description={t("markdown_description")}  resize="vertical" value={_value} 
            classNames={{root: classes.root, wrapper: classes.wrapper, input: classes.input}}
            onChange={(e) => { handleChange(e.currentTarget.value) }}
        />
    )
}