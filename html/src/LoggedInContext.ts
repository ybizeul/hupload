import { createContext, useContext } from "react";

interface LoggedInContextValue {
    loggedIn: string | null;
    setLoggedIn: React.Dispatch<React.SetStateAction<string | null>>;
  }

export const LoggedInContext = createContext<LoggedInContextValue|undefined>(undefined);

export const useLoggedInContext = () => {
    const loggedInContext = useContext(LoggedInContext);
    if (loggedInContext === undefined) {
      throw new Error('useOnboardingContext must be inside a OnboardingProvider');
    }
    return loggedInContext;
  };