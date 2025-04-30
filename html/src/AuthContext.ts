import { createContext, useContext, useEffect, useState } from "react";
import { AuthInfo, H } from "./APIClient";


export const useAuthInfo = ():[AuthInfo|undefined,() => void] => {
    const [authInfo, setAuthInfo] = useState<AuthInfo|undefined>(undefined)

    const check = () => {
        H.auth().then((r) => {
            const l = r as AuthInfo
            setAuthInfo(l)
        })
        .catch((e) => {
            setAuthInfo(undefined)
            console.log(e)
        })
    }

    useEffect(() => {
        check()
    },[])

    return [authInfo, check]
}

interface AuthContextValue {
    authInfo?: AuthInfo | undefined;
    check?: () => void;
}

export const AuthContext = createContext<AuthContextValue|undefined>(undefined);

export const useAuthContext = () => {
    const authContext = useContext(AuthContext);
    if (authContext === undefined) {
        throw new Error('useAuthContext must be inside a AuthContext.Provider');
    }
    return authContext;
};