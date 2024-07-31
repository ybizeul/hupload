import "@mantine/core/styles.css";
import '@mantine/dropzone/styles.css';
import { ActionIcon, Affix, Container, MantineProvider, Menu } from "@mantine/core";
import { theme } from "./theme";

import Home from "./Home";
import Share from "./Share";
import Login from "./Login";
import Shares from "./Shares";

import { BrowserRouter, Link, Route, Routes } from "react-router-dom";
import { IconMenu2 } from "@tabler/icons-react";
import { H } from "./APIClient";
import { useEffect, useState } from "react";
import { LoggedInContext } from "./LoggedInContext";
import VersionComponent from "./components/VersionComponent";

export default function App() {

  const [loggedIn, setLoggedIn ] = useState<boolean|null>(false)
  
  useEffect(() => {
    H.post('/login').then(() => {
      console.log("OK")
      setLoggedIn(true)
    })
    .catch((e) => {
      console.log(e)
    })
  })
  return (
  <MantineProvider theme={theme} defaultColorScheme='auto'>
    <Container mt="md" h="100%">
      <BrowserRouter>
        <LoggedInContext.Provider value={{loggedIn,setLoggedIn}}>
        {loggedIn &&
          <Affix position={{ bottom: 20, right: 20 }}>
            <Menu trigger="hover" openDelay={100} closeDelay={400} shadow="md" width={200}>
              <Menu.Target>
                <ActionIcon variant="filled" size="xl" radius="xl" aria-label="Settings">
                  <IconMenu2 style={{ width: '70%', height: '70%' }} stroke={1.5} />
                </ActionIcon>
              </Menu.Target>
              <Menu.Dropdown>
                <Menu.Item component={Link} to="/shares">
                  Shares
                </Menu.Item>
                <Menu.Item onClick={() => { H.logoutNow(); window.location.href='/'}}>
                Logout
                </Menu.Item>
              </Menu.Dropdown>
            </Menu>
          </Affix>
          }
          <Routes>
            <Route path="/" element={<Home />} />
            <Route path=":share" element={<Share />} />
            <Route path="/login" element={<Login />} />
            <Route path="/shares" element={<Shares />} />
          </Routes>
        </LoggedInContext.Provider>
      </BrowserRouter>
    </Container>
    {loggedIn&&<VersionComponent/>}
  </MantineProvider>)
}
