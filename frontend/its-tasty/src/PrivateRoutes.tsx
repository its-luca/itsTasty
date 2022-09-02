import React from 'react'
import {Outlet,Navigate} from 'react-router-dom'
import {useAuthContext} from "./AuthContext";


export  function PrivateRoutes() {
    const authContext  = useAuthContext();
    let isAuthenticated = false;
    if( authContext !== undefined) {
        isAuthenticated = authContext.isAuthenticated()
    }
    console.log(`private route, isAuthenticated: ${isAuthenticated}`)
    return (
        isAuthenticated ? <Outlet/> : <Navigate to={"/login"}/>
    );

}