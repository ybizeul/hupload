import { createContext, useContext } from "react";
import { AuthInfo } from "./APIClient";



interface AuthContextValue {
    authInfo: AuthInfo | null;
    setAuthInfo: React.Dispatch<React.SetStateAction<AuthInfo | null>>;
}

export const AuthContext = createContext<AuthContextValue|undefined>(undefined);

export const useAuthContext = () => {
    const authContext = useContext(AuthContext);
    if (authContext === undefined) {
        throw new Error('useAuthContext must be inside a AuthContext.Provider');
    }
    return authContext;
};