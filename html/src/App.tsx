import "@mantine/core/styles.css";
import '@mantine/dropzone/styles.css';

import { useEffect, useState } from "react";
import { ActionIcon, Affix, Container, MantineProvider, Menu } from "@mantine/core";
import { BrowserRouter, Link, Route, Routes } from "react-router-dom";

import { H } from "./APIClient";
import { Share, Login, Shares } from "@/Pages";

import { IconMenu2 } from "@tabler/icons-react";
import { LoggedInContext } from "@/LoggedInContext";
import { VersionComponent } from "@/Components";

// Logged in user is passed to the context
interface LoggedIn {
  user: string
}

export default function App() {
  // Component state
  const [loggedIn, setLoggedIn ] = useState<string|null>(null)
  
  // Component Hooks

  // Check with server current logged in state
  // This is typically executed once when Hupload is loaded
  // State is updated later on login page or logout button
  useEffect(() => {
    H.post('/login').then((r) => {
      const l = r as LoggedIn
      setLoggedIn(l.user)
    })
    .catch((e) => {
      setLoggedIn(null)
      console.log(e)
    })
  })

  return (
  <MantineProvider defaultColorScheme='auto'>
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
            <Route path="/" element={<Login />} />
            <Route path=":share" element={<Share />} />
            <Route path="/shares" element={<Shares owner={loggedIn}/>} />
          </Routes>
        </LoggedInContext.Provider>
      </BrowserRouter>
    </Container>
    {loggedIn&&<VersionComponent/>}
  </MantineProvider>)
}
