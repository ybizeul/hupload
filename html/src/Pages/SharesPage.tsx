
import { useCallback, useEffect, useState } from "react";
import { H } from "../APIClient";
import { useNavigate } from "react-router-dom";
import { Share } from "../hupload";
import {ShareComponent, SplitButton} from "@/Components";
import { Box, Text } from "@mantine/core";

export function SharesPage(props: {owner: string|null}) {
  // Initialize props
  const { owner } = props

  // Initialize state
  const [shares, setShares] = useState<Share[]|undefined>(undefined)
  const [exposure,setExposure] = useState<string>("upload")
  const [validity,setValidity] = useState<number>(7)
  // Initialize hooks
  const navigate = useNavigate();

  const updateShareProperties = (exposure: string, validity: number|string) => {
    setExposure(exposure)
    if (typeof validity === 'number') {
      setValidity(validity)
    }
  }

  const createShare = () => {
    const data = {
      exposure: exposure,
      validity: validity
    }
    H.post('/shares', data).then(
      () => {
        updateShares()
      })
      .catch((e) => {
        console.log(e)
        navigate('/')
      })
  }
  const updateShares = useCallback(() => {
    H.get('/shares').then(
      (res) => {
        setShares(res as Share[])
      })
      .catch((e) => {
        console.log(e)
        navigate('/')
      })
  },[navigate])

  useEffect(() => {
    updateShares()
  },[updateShares])


  return (
    shares &&
      <>
      <Box ta="center" mt="xl" mb="xl">
      <SplitButton exposure={exposure} validity={validity} onChange={updateShareProperties} onClick={() => createShare()}>Create Share</SplitButton>
      </Box>
      {
        shares.length == 0 ?
        <Text ta="center" mt="xl">No shares</Text>
        :
        <>
        {shares.some((s) => s.owner === owner) &&
          <>
          <Text size="xl" fw="700">Your Shares</Text>
          {shares.map((s) => (
            s.owner === owner &&
            <ShareComponent key={s.name} share={s} />
          ))}
          </>
        }
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