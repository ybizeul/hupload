import { ActionIcon, TextInput } from '@mantine/core';
import { IconSearch, IconX } from '@tabler/icons-react';

export function SearchField(props: { value?: string, onChange?: (e: string) => void }) {
    const { value, onChange } = props;
    return (
      <TextInput
        leftSection={<IconSearch size={16} stroke={1.5} />}
        rightSection={<ActionIcon size={16} radius="xl" variant="subtle" color="gray" onClick={() => {onChange && onChange("")}}><IconX size={16} stroke={1.5} /></ActionIcon>}
        placeholder="Search"
        value={value}
        radius={30}
        w={{base: "100%",xs:"50%"}}
        onChange={(event) => {
          onChange && onChange(event.currentTarget.value);
        }}
      />
    );
  }