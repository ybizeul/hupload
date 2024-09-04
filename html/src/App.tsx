import "@mantine/core/styles.css";
import '@mantine/dropzone/styles.css';

import { useEffect, useState } from "react";
import { Container, MantineProvider } from "@mantine/core";
import { BrowserRouter, Route, Routes } from "react-router-dom";

import { SharePage, Login, SharesPage } from "@/Pages";

import { AuthContext } from "@/AuthContext";
import { VersionComponent, Haffix } from "@/Components";
import { AuthInfo, H } from "./APIClient";



export default function App() {
    // Component state
    const [authInfo, setAuthInfo ] = useState<AuthInfo|null>(null)

    // Check with server current logged in state
    // This is typically executed once when Hupload is loaded
    // State is updated later on login page or logout button
    useEffect(() => {
        H.auth().then((r) => {
            const l = r as AuthInfo
            setAuthInfo(l)
        })
        .catch((e) => {
            setAuthInfo(null)
            console.log(e)
        })
    },[])

    return (
    <MantineProvider defaultColorScheme='auto'>
        <Container flex={1} size="sm" w="100%" pt="md">
        <BrowserRouter>
            <AuthContext.Provider value={{authInfo,setAuthInfo}}>
            <Routes>
                <Route path="/" element={<Login />}/>
                <Route path="/shares" element={<>{authInfo?.user&&<Haffix/>}<SharesPage owner={authInfo?.user?(authInfo?.user):null}/></>} />
                <Route path=":share" element={<>{authInfo?.user&&<Haffix/>}<SharePage /></>} />
            </Routes>
            </AuthContext.Provider>
        </BrowserRouter>
        </Container>
        {authInfo?.user&&<VersionComponent/>}
    </MantineProvider>)
}
