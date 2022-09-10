import React, {useEffect, useState} from 'react'
import {Outlet, Navigate, useLocation} from 'react-router-dom'
import {ApiError, DefaultService} from "./services/userAPI";
import {lsKeyIsAuthenticated} from "./localStorageKeys";

export interface LocationState {
    from: {
        pathname: string;
    };
}

export  function PrivateRoutes() {

    const targetLocation = useLocation();
    const [authData,setAuthData] = useState(localStorage.getItem(lsKeyIsAuthenticated))
    const [verifiedAuthStatus,setVerifiedAuthStatus] = useState(authData === String(true))
    useEffect( () => {
        const fetchCurrentUser = async () => {
            try {
                await DefaultService.getUsersMe()
                setAuthData(String(true))
                localStorage.setItem(lsKeyIsAuthenticated,String(true))
            } catch (e) {
                if ( e instanceof ApiError) {
                    if( e.status === 401) {
                        setAuthData(String(false))
                        localStorage.setItem(lsKeyIsAuthenticated,String(false))
                    }
                    console.log(e.status)
                }
            }
            setVerifiedAuthStatus(true)
        }

        //If we have stored that user is authenticated, check if they are
        if( !verifiedAuthStatus ) {
            fetchCurrentUser()
        }
    },);

    if( !verifiedAuthStatus){
        return <p>Checking authentication...</p>
    }

    if( authData !== String(true) ) {
        //place original target in Navigate state allowing us to redirect there after user has logged in
        const locState : LocationState = {from: targetLocation}
        return <Navigate to={"/login"} replace={true} state={locState}/>
    }

    return (
      <Outlet/>
    )

}