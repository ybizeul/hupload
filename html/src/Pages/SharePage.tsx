import { Box, Button, Center, CopyButton, Group, rem, Text, Tooltip } from "@mantine/core";
import { Dropzone } from "@mantine/dropzone";
import { IconFileZip, IconLink, IconUpload, IconX } from "@tabler/icons-react";
import { useEffect, useState } from "react";
import { H } from "../APIClient";
import { UploadQueue, QueueItem } from "../UploadQueue";
import { useLocation, useNavigate } from "react-router-dom";
import {ItemComponent} from "@/Components";
import { Item, Share } from "../hupload";
import { useLoggedInContext } from "@/LoggedInContext";

export function SharePage() {

  const [items, setItems] = useState<Item[]|undefined>(undefined)
  const [queue, setQueue] = useState<QueueItem[]>([])
  const [expired,setExpired] = useState(false)
  const [share,setShare] = useState<Share|undefined>(undefined)
  
  // Initialize hooks
  const location = useLocation()
  const {loggedIn} = useLoggedInContext()
  const navigate = useNavigate();

  // Functions

  const showDropZone = () => {
    if (loggedIn) {
      return true
    }
    // We are guest

    if (!share) {
      return false
    }

    return share.exposure === "" || share.exposure === "upload" || share.exposure === "both"
  }

  const canDownload = () => {
    if (loggedIn) {
      return true
    }
    // We are guest

    if (!share) {
      return false
    }

    return (share.exposure === "download" || share.exposure === "both")
  }

  useEffect(() => {
    const s=location.pathname.split("/")
    const shareSegment = s[1]
    H.get('/shares/' + shareSegment).then((res) => {
      setShare(res as Share)
      console.log(res)
      H.get('/shares/' + shareSegment + "/items").then(
      (res) => {
        setItems(res as Item[])
      })
      .catch((e) => {
        console.log(e)
        if (e.response.status === 410) {
          setExpired(true)
          return
        }
        navigate('/')
      })
    })
    .catch((e) => {
      console.log(e)
      if (e.response.status === 410) {
        setExpired(true)
        return
      }
      navigate('/')
    })
  },[navigate, location.pathname])

  return (
    expired?
      <Center><Text size="xl" ta="center">Sorry, this share has expired</Text></Center>
    :
    items &&
      <>
      {/* Top of page copy button */}
      <Box w="100%" ta="center">
          <CopyButton value={window.location.protocol + '//' + window.location.host + '/' + share?.name}>
            {({ copied, copy }) => (
              <Tooltip withArrow arrowOffset={10} arrowSize={4} label={copied?"Copied!":"Copy URL"}>
                <Button mb="sm" justify="center" variant="outline" color={copied ? 'teal' : 'gray'} size="xs" onClick={copy}><IconLink style={{ width: '70%', height: '70%' }} stroke={1.5}/>{share?.name}</Button>
              </Tooltip>
            )}
          </CopyButton>
      </Box>
      {/* Files drop zone */}
      {showDropZone() &&
      <>
        <Dropzone
          onDrop={(files) => {
            const U = new UploadQueue(H,"/shares/"+share?.name, setQueue)
            const newItems = items.filter((i) => {
              return !files.some((f) => f.name === i.Path.split("/")[1])
            })
            setItems(newItems)
            U.addFiles(files)
            .then((r) => {
              const finishedItems = r as Item[]
              setQueue([])
              setItems([...finishedItems,...newItems])
            })
            .catch((e) => {
              console.log(e)
            })
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
      </>}
      {
        // Display upload queue items
        queue.map((q) => (
          <ItemComponent download={false} key={'up_' + q.file.name} queueItem={q} />
        ))
      }
      {
        // Display share items
        items.map((item) => (
          <ItemComponent download={canDownload()} key={item.Path} item={item} />
        ))
      }
    </>
  )
}