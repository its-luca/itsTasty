import React, {useEffect, useState} from 'react'
import {Outlet,Navigate} from 'react-router-dom'
import {ApiError, DefaultService} from "./services/userAPI";
import {lsKeyIsAuthenticated} from "./localStorageKeys";


export  function PrivateRoutes() {

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

    return (
        (authData === String(true)) ? <Outlet/> : <Navigate to={"/login"}/>
    );

}