import { Group, rem, Text } from "@mantine/core";
import { Dropzone } from "@mantine/dropzone";
import { IconFileZip, IconUpload, IconX } from "@tabler/icons-react";
import { useEffect, useState } from "react";
import { H } from "../APIClient";
import { UploadQueue, QueueItem } from "../UploadQueue";
import { useLocation, useNavigate } from "react-router-dom";
import {ItemComponent} from "@/Components";
import { Item } from "../hupload";

export function Share() {

  const [items, setItems] = useState<Item[]|undefined>(undefined)
  const [queue, setQueue] = useState<QueueItem[]>([])
  const location = useLocation()

  const s=location.pathname.split("/")
  const share = s[1]

  const navigate = useNavigate();

  useEffect(() => {
    H.get('/share/' +share).then(
    (res) => {
      console.log(res)
      setItems(res as Item[])
    })
    .catch((e) => {
      console.log(e)
      navigate('/')
    })
  },[share, navigate])

  return (
    items &&
      <>
      <Dropzone
        onDrop={(files) => {
          const U = new UploadQueue(H,"/share/"+share, setQueue)
          U.addFiles(files)
          const newItems = items.filter((i) => {
            return !files.some((f) => f.name === i.Path.split("/")[1])
          })
          setItems(newItems)
        }}

        onReject={(files) => console.log('rejected files', files)}
      >
        <Group justify="center" gap="xl" mih={100} style={{ pointerEvents: 'none' }}>
          <Dropzone.Accept>
            <IconUpload
              style={{ width: rem(52), height: rem(52), color: 'var(--mantine-color-blue-6)' }}
              stroke={1.5}
            />
          </Dropzone.Accept>
          <Dropzone.Reject>
            <IconX
              style={{ width: rem(52), height: rem(52), color: 'var(--mantine-color-red-6)' }}
              stroke={1.5}
            />
          </Dropzone.Reject>
          <Dropzone.Idle>  
            <IconFileZip
              style={{ width: rem(52), height: rem(52), color: 'var(--mantine-color-dimmed)' }}
              stroke={1.5}
            />
          </Dropzone.Idle>

          <div>
            <Text size="xl" inline>
              Drag files here or click to select files
            </Text>
          </div>
        </Group>
      </Dropzone>
      {
        queue.map((q) => (
          <ItemComponent key={q.file.name} queueItem={q} />
        ))
      }
      {
        items.map((item) => (
          <ItemComponent key={item.Path} item={item} />
        ))
      }
    </>
  )
}