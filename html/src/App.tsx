import "@mantine/core/styles.css";
import '@mantine/dropzone/styles.css';

import { useEffect, useState } from "react";
import { Container, MantineProvider } from "@mantine/core";
import { BrowserRouter, Route, Routes } from "react-router-dom";

import { H } from "./APIClient";
import { SharePage, Login, SharesPage } from "@/Pages";

import { LoggedInContext } from "@/LoggedInContext";
import { VersionComponent } from "@/Components";
import { Haffix } from "./Components/Haffix";
import { AxiosResponse } from "axios";

// Logged in user is passed to the context
interface LoggedIn {
  user: string
}

export default function App() {
  // Component state
  const [loggedIn, setLoggedIn ] = useState<string|null>(null)
  // Check with server current logged in state
  // This is typically executed once when Hupload is loaded
  // State is updated later on login page or logout button
  useEffect(() => {
    H.login('/login').then((r) => {
      console.log(r)
      const response = r as AxiosResponse
      if (response.status == 202) {
        window.location.href = "/login"
      }
      const l = response.data as LoggedIn
      setLoggedIn(l.user)
    })
    .catch((e) => {
      setLoggedIn(null)
      console.log(e)
    })
  },[])

  return (
  <MantineProvider defaultColorScheme='auto'>
    <Container flex={1} size="sm" w="100%" pt="md">
      <BrowserRouter>
        <LoggedInContext.Provider value={{loggedIn,setLoggedIn}}>
          <Routes>
            <Route path="/" element={<Login />}/>
            <Route path="/shares" element={<>{loggedIn&&<Haffix/>}<SharesPage owner={loggedIn}/></>} />
            <Route path=":share" element={<>{loggedIn&&<Haffix/>}<SharePage /></>} />
          </Routes>
        </LoggedInContext.Provider>
      </BrowserRouter>
    </Container>
    {loggedIn&&<VersionComponent/>}
  </MantineProvider>)
}
