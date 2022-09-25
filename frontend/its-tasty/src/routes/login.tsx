import { useLocation} from "react-router-dom";
import {LocationState} from "../PrivateRoutes";
import urlJoin from 'url-join';
import Typography from "@mui/material/Typography";
import Button from "@mui/material/Button";
import {ReactComponent as Logo} from "../itsTastyGopher.svg";
import { Paper, useMediaQuery} from "@mui/material";
import Box from "@mui/material/Box";
import Grid2 from "@mui/material/Unstable_Grid2";
import * as React from "react";
import {useTheme} from "@mui/material/styles";

export  function LoginPage() {
    const location =  useLocation()
    const theme = useTheme();
    const isSmallScreen = useMediaQuery(theme.breakpoints.down("sm"));

    //
    //Check if we should redirect to a specific location
    //

    //base url without redirectTo param
    let loginURL =  new URL(urlJoin(process.env.REACT_APP_AUTH_API_BASE_URL!,'/login',));

    //check if we placed redirect info in location state. If so , set redirectTo param of login url accordingly
    if( location.state ) {
        const {from} = location.state as LocationState
        if( from.pathname && from.pathname != "" && from.pathname != "/") {
            const redirectURL =   urlJoin(process.env.REACT_APP_PUBLIC_URL!,from.pathname)
            loginURL = new URL(urlJoin(process.env.REACT_APP_AUTH_API_BASE_URL!,'login',`?redirectTo=${redirectURL}`))
        }
    }



    return (
        <Grid2
            container
            sx={{
                display:"flex",
                flexDirection:"column",
                justifyContent:"center",
                alignItems:"center",
                minHeight: '100vh',
                textAlign:"center",
            }}
        >
            <Grid2
                container
                sx={{
                    display:"flex",
                    flexDirection:"column",
                    justifyContent:"center",
                    alignItems:"center",
                    textAlign:"center",
                }}
            >
                <Box
                    component={Paper}
                    elevation={10}
                    sx={{
                        display:"flex",
                        flexDirection:"column",
                        justifyContent:"center",
                        alignItems:"center",
                        borderRadius:"50px",
                        padding:"10px",
                        textAlign:"center",
                        margin:"10px"
                    }}
                >
                    <Box  component={Logo} sx={{
                        maxWidth: isSmallScreen ? "256px" : "100%",
                        maxHeight: isSmallScreen ? "256px" : "100%",

                    }}/>
                    <Box
                        sx={{
                            display:"flex",
                            flexDirection:"column",
                            justifyContent:"center",
                            alignItems:"center",
                            borderRadius:"50px",
                            padding:"10px",
                            textAlign:"center",
                        }}
                    >
                        <Typography
                            variant={ isSmallScreen ? "h4" : "h2"}
                            sx={{
                                fontWeight: 700,
                                color: 'inherit',
                                textDecoration: 'none',
                            }}
                        >
                            ITS Tasty
                        </Typography>
                        <Button
                            variant={"contained"}
                            href={loginURL.href}
                            sx={{
                                minWidth:"250px",
                                minHeight:"60px",
                            }}
                        >
                            Login
                        </Button>

                    </Box>
                    <Box
                        sx={{pt:"10px",display:"flex",flexDirection:"column" ,justifyContent:"end",alignItems:"center",textAlign:"center",}}>
                        <Typography>
                            Gopher artwork by <a href={"https://twitter.com/ashleymcnamara"}>Ashley McNamara</a> inspired
                            by <a href={"http://reneefrench.blogspot.co.uk"}>Renee French</a><br/>
                            Further assets  <a href="http://www.freepik.com">by macrovector and upklyak / Freepik</a>
                        </Typography>
                    </Box>
                </Box>

            </Grid2>

        </Grid2>
    );
}