import "@mantine/core/styles.css";
import '@mantine/dropzone/styles.css';

import { useEffect, useState } from "react";
import { Container, MantineProvider } from "@mantine/core";
import { BrowserRouter, Route, Routes } from "react-router-dom";

//import { H } from "./APIClient";
import { SharePage, Login, SharesPage } from "@/Pages";

import { LoggedInContext, LoggedIn } from "@/LoggedInContext";
import { VersionComponent } from "@/Components";
import { Haffix } from "./Components/Haffix";
import { H } from "./APIClient";
//import { AxiosResponse } from "axios";



export default function App() {
  // Component state
  const [loggedIn, setLoggedIn ] = useState<LoggedIn|null>(null)

  // Check with server current logged in state
  // This is typically executed once when Hupload is loaded
  // State is updated later on login page or logout button
  useEffect(() => {
    H.login('/login').then((r) => {
      const l = r as LoggedIn
      setLoggedIn(l)
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
            <Route path="/shares" element={<>{loggedIn&&<Haffix/>}<SharesPage owner={loggedIn?(loggedIn.user):null}/></>} />
            <Route path=":share" element={<>{loggedIn&&<Haffix/>}<SharePage /></>} />
          </Routes>
        </LoggedInContext.Provider>
      </BrowserRouter>
    </Container>
    {loggedIn&&<VersionComponent/>}
  </MantineProvider>)
}
