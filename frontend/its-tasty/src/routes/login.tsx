import { Button} from "react-bootstrap";
import {useLocation} from "react-router-dom";
import {LocationState} from "../PrivateRoutes";
import urlJoin from 'url-join';

export  function LoginPage() {
    const location =  useLocation()

    //Check if we should redirect to a specific location
    let loginURL =  new URL(urlJoin(process.env.REACT_APP_AUTH_API_BASE_URL!,'/login',));
    if( location.state ) {
        const {from} = location.state as LocationState
        if( from.pathname && from.pathname != "" && from.pathname != "/") {
            const redirectURL =   urlJoin(process.env.REACT_APP_PUBLIC_URL!,from.pathname)
            loginURL = new URL(urlJoin(process.env.REACT_APP_AUTH_API_BASE_URL!,'login',`?redirectTo=${redirectURL}`))
        }
    }
    console.log('requested redirect: ' + loginURL)

    return (
        <div className={"d-flex vw-100 vh-100 justify-content-center align-items-center"}>
                <div className={"col-5 text-center"}>
                    <div className={"row"}>
                        <h1 >
                            ITS (hopefully) Tasty
                        </h1>
                        <p>
                            Welcome! This site lets you rate the quality of the dishes in our beloved mensa and
                            UKSH bistro. Login to proceed.
                        </p>
                    </div>
                    <div className={"row"}>
                        <Button variant={"primary"} href={loginURL.href} >Login</Button>
                    </div>
                </div>

        </div>
    );
}