import { Alert, Button, Container, FocusTrap, Paper, PasswordInput, TextInput } from "@mantine/core";
import { useState } from "react";
import { APIServerError, H } from "../APIClient";

import { IconExclamationCircle } from "@tabler/icons-react";
import { useNavigate } from "react-router-dom";
import { useLoggedInContext } from "../LoggedInContext";

export function Login() {
    // Initialize States
    const [username, setUsername] = useState<undefined|string>("")
    const [password, setPassword] = useState<undefined|string>("")
    const [error, setError] = useState<APIServerError|undefined>()

    // Initialize hooks
    const navigate = useNavigate();
    const { setLoggedIn } = useLoggedInContext()
    
    // Functions
    function authenticate(event:React.FormEvent<HTMLFormElement>) {
        event.preventDefault()
        if (username && password) {
            H.login('/login',username,password)
            .then(() => {
                setError(undefined)
                navigate("/shares")
                if (setLoggedIn !== null) {
                  setLoggedIn(username)
                }
            })
            .catch(e => {
                setError(e)
            })
        }
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