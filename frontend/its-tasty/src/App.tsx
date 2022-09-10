import React, {useEffect, useState} from 'react';
import './App.css';
import {Outlet, Link, Navigate,} from "react-router-dom";
import {ApiError, DefaultService} from "./services/userAPI";
import { Nav, Navbar} from "react-bootstrap";
import {AuthContext} from "./AuthContext";
import {lsKeyIsAuthenticated} from "./localStorageKeys";



function App() {

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
        return (
            <Navigate to={"/login"}/>
        )
    }

    return (
    <AuthContext.Provider value={defaultAuthContextValue}>
        <div id={"App.tsx main container"}>
            <Navbar bg={"light"} expand={"lg"}>
                <Navbar.Brand as={Link} to={"/welcome"} >ITS (hopefully) Tasty</Navbar.Brand>
                <div className=" d-flex w-100 h-100 justify-content-end" >
                    <div className={"row"}>
                        <div className={"col"}>
                            <Navbar.Text>User: {userEmail !== undefined ? userEmail: "error"} </Navbar.Text>
                        </div>
                        <div className={"d-flex col align-items-center"}>
                            {isAuthenticated() &&
                                <Nav.Link  href={`${process.env.REACT_APP_AUTH_API_BASE_URL}/authAPI/logout`}
                                           onClick={ () => { localStorage.setItem(lsKeyIsAuthenticated,String(false))}}>
                                    Logout
                                </Nav.Link>}
                        </div>
                    </div>

                </div>
            </Navbar>
            <Outlet />
        </div>
    </AuthContext.Provider>
);
}

export default App;
