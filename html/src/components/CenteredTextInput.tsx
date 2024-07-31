import { TextInput } from "@mantine/core";

import classes from "./CenteredTextInput.module.css";

export default function CenteredTextInput(props: {value: string, onChange: (e: React.ChangeEvent<HTMLInputElement>) => void}) {
    return (
        <TextInput size="xl" fw={700} w="14em" classNames={{input: classes.input}} value={props.value} onChange={(e) => {props.onChange(e)}}/>
    )
}