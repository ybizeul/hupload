
import { useCallback, useEffect, useState } from "react";
import { H } from "../APIClient";
import { useNavigate } from "react-router-dom";
import { Share } from "../hupload";
import {ShareComponent, SplitButton} from "@/Components";
import { Anchor, Box, Center, Stack, Text } from "@mantine/core";
import { IconMoodSad } from "@tabler/icons-react";
import { AxiosError } from "axios";

export function SharesPage(props: {owner: string|null}) {
  // Initialize props
  const { owner } = props

  // Initialize state
  const [shares, setShares] = useState<Share[]|undefined>(undefined)
  const [newShareOptions, setNewShareOptions] = useState<Share["options"]>({exposure: "upload", validity: 7, description: "", message: ""})
  const [error, setError] = useState<AxiosError|undefined>(undefined)
  
  // Initialize hooks
  const navigate = useNavigate();

  const updateShareProperties = (props: Share["options"]) => {
    setNewShareOptions(props)
  }

  // Functions
  const createShare = () => {
    H.post('/shares', newShareOptions).then(
      () => {
        updateShares()
      })
      .catch((e) => {
        console.log(e)
        setError(e)
      })
  }

  const updateShares = useCallback(() => {
    H.get('/shares').then(
      (res) => {
        setShares(res as Share[])
      })
      .catch((e) => {
        console.log(e)
        setError(e)
      })
  },[])

  useEffect(() => {
    updateShares()
  },[updateShares])

  if (error) {
    if (error.response?.status === 401) {
      navigate('/')
      return
    }

    return (
      <Center h="100vh">
        <Stack align="center" pb="10em">
          <IconMoodSad style={{ width: '10%', height: '10%' }} stroke={1.5}/>
          <Text size="xl" fw="700">{error.message}</Text>
          <Anchor onClick={() => { window.location.reload()}}>Reload</Anchor>
        </Stack>
      </Center>
    )
  }

  if (!shares) {
    return
  }

  return (
      <>
      {/* Create share button */}
      <Box ta="center" mt="xl" mb="xl">
        <SplitButton options={newShareOptions} onChange={updateShareProperties} onClick={() => createShare()}>Create Share</SplitButton>
      </Box>

      { shares.length == 0 ?
        <Text ta="center" mt="xl">There are currently no shares</Text>
        :
        <>
          {/* Currently logged in user shares */}
          {shares.some((s) => s.owner === owner) &&
            <>
            <Text size="xl" fw="700">Your Shares</Text>
            {shares.map((s) => (
              s.owner === owner &&
              <ShareComponent key={s.name} share={s} />
            ))}
            </>
          }

          {/* Other users shares */}
          {shares.some((s) => s.owner !== owner) &&
            <>
            <Text mt="md" size="xl" fw="700">Other Shares</Text>
            {shares.map((s) => (
              s.owner === owner ||
              <ShareComponent key={s.name} share={s} />
            ))}
            </>
          }
        </>
      }
    </>
  )
}