
import { useCallback, useEffect, useState } from "react";
import { H } from "../APIClient";
import { useNavigate } from "react-router-dom";
import { Share } from "../hupload";
import {ShareComponent} from "@/Components";
import { Box, Button, Text } from "@mantine/core";

export function Shares(props: {owner: string|null}) {
  const { owner } = props
  const [shares, setShares] = useState<Share[]|undefined>(undefined)

  const navigate = useNavigate();

  const createShare = () => {
    H.post('/share').then(
      () => {
        updateShares()
      })
      .catch((e) => {
        console.log(e)
        navigate('/')
      })
  }
  const updateShares = useCallback(() => {
    H.get('/share').then(
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
      <Button onClick={() => createShare()}>Create Share</Button>
      </Box>
      {
        shares.length == 0 ?
        <Text ta="center" mt="xl">No shares</Text>
        :
        <>
        <Text size="xl" fw="700">You Shares</Text>
        {shares.map((s) => (
          s.owner === owner &&
          <ShareComponent key={s.name} share={s} />
        ))}
        <Text mt="md" size="xl" fw="700">Other Shares</Text>
        {shares.map((s) => (
          s.owner === owner ||
          <ShareComponent key={s.name} share={s} />
        ))}
        </>
      }
    </>
  )
}