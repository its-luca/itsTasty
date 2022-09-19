import React, {useEffect, useState} from 'react';
import './App.css';
import {Outlet, Navigate, useLocation,} from "react-router-dom";
import {ApiError, DefaultService} from "./services/userAPI";
import {AuthContext} from "./AuthContext";
import {lsKeyIsAuthenticated} from "./localStorageKeys";
import {LocationState} from "./PrivateRoutes";
import {ResponsiveAppBar} from "./MyAppBar";



function App() {

    const location = useLocation();
    const [userEmail,setUserEmail] = useState<string|undefined>(undefined);

    const [authData,setAuthData] = useState(localStorage.getItem(lsKeyIsAuthenticated))
    const setAuthStatus = (isAuthenticated : boolean) => {
        localStorage.setItem(lsKeyIsAuthenticated,String(isAuthenticated))
        setAuthData(String(isAuthenticated))
        console.log(`Set isAuthenticated to ${String(isAuthenticated)}`)
    }
    const isAuthenticated = () => {
        return authData === String(true)
    }
    const defaultAuthContextValue = {
        setAuthStatus: setAuthStatus,
        isAuthenticated: isAuthenticated
    };


    useEffect( () => {
        const fetchCurrentUser = async () => {
            try {
                console.log("Querying getUserMe to refresh auth status")
                const reply = await DefaultService.getUsersMe()
                setAuthStatus(true)
                console.log(`Logged in as user ${reply.email}`)
                setUserEmail(reply.email)
            } catch (e) {
                if ( e instanceof ApiError) {
                    if( e.status === 401) {
                        setAuthStatus(false)
                        setUserEmail(undefined)
                    }
                   console.log(e.status)
                }
            }
        }
        fetchCurrentUser()
    },[authData]);

    if( !isAuthenticated() ) {
        const locState : LocationState = {from: location}
        return (
            //place original target in Navigate state allowing us to redirect there after user has logged in
            <Navigate to={"/login"} replace={true} state={locState}/>
        )
    }

    return (
    <AuthContext.Provider value={defaultAuthContextValue}>
        <ResponsiveAppBar userName={userEmail ? userEmail : "Error"}/>
        <Outlet />
    </AuthContext.Provider>
);
}

export default App;
