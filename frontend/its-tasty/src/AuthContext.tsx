import React, {useContext} from "react";

interface AuthContextValue {
    isAuthenticated() : boolean,
    setAuthStatus(status : boolean) :void
}

export const AuthContext = React.createContext<AuthContextValue|undefined>(undefined);

export function useAuthContext() {
    return useContext(AuthContext)
}