import { createContext, useContext } from "react";

// Logged in user is passed to the context
export interface LoggedIn {
  user: string
  loginPage: string
}

interface LoggedInContextValue {
    loggedIn: LoggedIn | null;
    setLoggedIn: React.Dispatch<React.SetStateAction<LoggedIn | null>>;
  }

export const LoggedInContext = createContext<LoggedInContextValue|undefined>(undefined);

export const useLoggedInContext = () => {
    const loggedInContext = useContext(LoggedInContext);
    if (loggedInContext === undefined) {
      throw new Error('useOnboardingContext must be inside a OnboardingProvider');
    }
    return loggedInContext;
  };