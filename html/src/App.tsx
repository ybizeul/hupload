import "@mantine/core/styles.css";
import '@mantine/dropzone/styles.css';
import "./i18n/config.ts";

import { Container, MantineProvider } from "@mantine/core";
import { BrowserRouter, Route, Routes } from "react-router-dom";

import { SharePage, Login, SharesPage } from "@/Pages";

import { AuthContext, useAuthInfo } from "@/AuthContext";
import { VersionComponent, Haffix } from "@/Components";
import { ErrorPage } from "./Pages/ErrorPage.tsx";



export default function App() {
    const [authInfo, check] = useAuthInfo()

    return (
    <MantineProvider defaultColorScheme='auto'>
        <Container flex={1} size="sm" w="100%" pt="md">
        <BrowserRouter>
            <AuthContext.Provider value={{authInfo,check}}>
                <Routes>
                    <Route path="/" element={<Login />}/>
                    <Route path="/error" element={<ErrorPage/>}/>
                    <Route path="/shares" element={<>{authInfo?.user&&<Haffix/>}<SharesPage owner={authInfo?.user?(authInfo?.user):null}/></>} />
                    <Route path=":share" element={<>{authInfo?.user&&<Haffix/>}<SharePage /></>} />
                </Routes>
            </AuthContext.Provider>
        </BrowserRouter>
        </Container>
        {authInfo?.user&&<VersionComponent/>}
    </MantineProvider>)
}
