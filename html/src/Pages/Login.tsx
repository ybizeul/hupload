import { Alert, Button, Container, FocusTrap, Paper, PasswordInput, TextInput } from "@mantine/core";
import { useEffect, useState } from "react";
import { APIServerError, AuthInfo, H } from "../APIClient";

import { IconExclamationCircle } from "@tabler/icons-react";
import { useNavigate } from "react-router-dom";
import { useAuthContext } from "../AuthContext";


export function Login() {
    // Initialize States
    const [username, setUsername] = useState<undefined|string>("")
    const [password, setPassword] = useState<undefined|string>("")
    const [error, setError] = useState<APIServerError|undefined>()
    const [showLoginForm, setShowLoginForm] = useState<boolean|undefined>(undefined)

    // Initialize hooks
    const navigate = useNavigate();

    // Initialize contexts
    const { authInfo, setAuthInfo } = useAuthContext()
    
    // Functions
    function authenticate(event:React.FormEvent<HTMLFormElement>) {
        event.preventDefault()
        if (username && password) {
            H.login(username,password)
            .then(() => {
                setError(undefined)
                navigate("/shares")
                if (setAuthInfo !== null) {
                    if (authInfo) {
                        setAuthInfo({...authInfo, ...{user: username}})
                    }
                }
            })
            .catch(e => {
                setError(e)
            })
        }
    }

    useEffect(() => {
        H.auth()
            .then((r) => {
                const resp = r as AuthInfo
                setShowLoginForm(resp.showLoginForm)
                if (resp.loginUrl !== document.location.pathname) {
                window.location.href = resp.loginUrl
                }
        })
    },[navigate])

    if (showLoginForm !== true) {
        return
    }

    return (
        <Container size={420} my="10%">
            <Paper withBorder shadow="md" p={30} mt={30} radius="md">
                {error &&
                    <Alert mb="md" variant="light" color="red" title="Error" icon={<IconExclamationCircle/>}>
                        {error.message}
                    </Alert>}
                <form onSubmit={authenticate}>
                    <FocusTrap active={true}>
                        <TextInput id="username" label="Username" placeholder="Username" value={username} onChange={(e) => setUsername(e.target.value)} required data-autofocus/>
                        <PasswordInput id="password" label="Password" placeholder="Your password" value={password} onChange={(e) => setPassword(e.target.value)} required mt="md" />
                        <Button type="submit" fullWidth mt="xl" disabled={!(username && password)}>
                            Login
                        </Button>
                    </FocusTrap>
                </form>
            </Paper>
        </Container>
      );
}