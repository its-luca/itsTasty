import {useEffect, useState} from "react";
import {ApiError, DefaultService, GetDishResp} from "./services/userAPI";
import {useAuthContext} from "./AuthContext";
import Typography from "@mui/material/Typography";
import {
    Paper,
    Rating, Skeleton, Stack,
    Table,
    TableBody,
    TableCell,
    TableContainer,
    TableRow,
} from "@mui/material";
import ErrorOutlineIcon from '@mui/icons-material/ErrorOutline';
import { Link as RRLink} from "react-router-dom";
import urlJoin from "url-join";


export function SimpleDishViewMock() {
    return(
        <Stack
            maxWidth="sm"
            sx={{
                marginTop:"10px"
            }}>
            <TableContainer component={Paper} elevation={10} >
                <Table>
                    <TableBody>
                        <TableRow >
                            <TableCell  align="center" colSpan={2} >
                               <Skeleton variant={"text"} height={"60px"}/>
                            </TableCell>
                        </TableRow>

                        <TableRow>
                            <Skeleton variant={"rectangular"} width={"330px"} height={"60px"}/>
                        </TableRow>

                        <TableRow>
                            <Skeleton variant={"rectangular"} width={"330px"} height={"60px"}/>
                        </TableRow>

                    </TableBody>
                </Table>
            </TableContainer>
        </Stack>
    )
}

export interface SimpleDishViewProps {
    dishID : number
}
export function SimpleDishView(props : SimpleDishViewProps) {
    const [dishData,setDishData] = useState<GetDishResp|undefined>(undefined)

    enum State {
        loading,
        error,
        success
    }
    const [state,setState] = useState<State>(State.loading)
    const [_,setErrorMessage] = useState("")
    const authContext = useAuthContext()
    if( authContext === undefined ) {
        console.log("authContext undefined")
        setState(State.error)
        setErrorMessage("Internal Error")
    }

    const fetchDish = async () => {
        try {
            const reply = await DefaultService.getDishes(props.dishID)
            setDishData(reply)
            setState(State.success)
        } catch (e) {
            if (e instanceof ApiError) {
                switch (e.status) {
                    case 400:
                        setErrorMessage("Failed to fetch Dish: Bad Input Data")
                        break;
                    case 500:
                        setErrorMessage( "Failed to fetch Dish: Internal Server Error")
                        break;
                    case 404:
                        setErrorMessage("Failed to fetch Dish: Did not find dish")
                        break;
                    case 401:
                        authContext?.setAuthStatus(false)
                        break;
                    default:
                        setErrorMessage("Failed to fetch Dish: Unknown Error")
                        break;
                }
            } else {
                setErrorMessage( "Failed to fetch Dish: Unknown Error")
            }
            setState(State.error)
        }
    }


    const updateUserVoting = async (rating :number) => {


        try {
            await DefaultService.postDishes(props.dishID,{rating:rating})
            setState(State.success)
        } catch (e) {
            console.log(`updateUserVoting error : ${e}`)
            if (e instanceof ApiError) {
                switch (e.status) {
                    case 400:
                        setErrorMessage("Failed to update Rating: Bad Input Data")
                        break;
                    case 500:
                        setErrorMessage( "Failed to update Rating: Internal Server Error")
                        break;
                    case 404:
                        setErrorMessage("Failed to update Rating: Did not find dish")
                        break;
                    case 401:
                        authContext?.setAuthStatus(false)
                        break;
                    default:
                        setErrorMessage("Failed to update Rating: Unknown Error")
                        break;
                }
            } else {
                setErrorMessage( "Failed to update Rating: Unknown Error")
            }
            setState(State.error)
        }
        //trigger reload of dish
        fetchDish()
    }

    useEffect( () => {
        console.log("useEffect is running")
        fetchDish()
    },[])

    if( state === State.loading ) {
        return(
           <SimpleDishViewMock/>
        )
    }
    if( state === State.error || dishData === undefined) {
        return (
            <Stack
                maxWidth="sm"
                sx={{
                    marginTop:"10px"
                }}>
                <Paper elevation={10} sx={{height:"180px",width:"330px",display:"flex",alignItems:"center",justifyContent:"center"}}>
                    <ErrorOutlineIcon fontSize={"large"}/>
                </Paper>
            </Stack>
        )
    }


    let ratings = new Array<number>()
    for(let rating = 1; rating <=5; rating++) {
        let count = dishData.ratings[rating.toString()]
        if (count === undefined) {
            count = 0;
        }
        ratings.push(count)
    }

    return (
        <Stack
            maxWidth="sm"
            sx={{
                marginTop:"10px"
            }}>
            <TableContainer component={Paper} elevation={10} >
                <Table>
                    <TableBody>
                        <TableRow >
                            <TableCell  align="center" colSpan={2} >
                                <Typography
                                    component={RRLink}
                                    to={new URL(urlJoin(process.env.REACT_APP_PUBLIC_URL!,'dish',props.dishID.toString()))}
                                    sx={{fontSize:"large",fontWeight:"bold",textAlign:"center"}} >
                                    {dishData.name}
                                </Typography>
                            </TableCell>
                        </TableRow>

                        <TableRow>
                            <TableCell >
                                <Typography sx={{fontSize:"large"}}>Average Rating</Typography>
                            </TableCell>
                            <TableCell >
                                <Rating
                                    max={5}
                                    precision={0.25}
                                    value={ (dishData.avgRating !== undefined  && dishData.avgRating !== 0) ? dishData.avgRating : null }
                                    readOnly={true}
                                />
                            </TableCell>
                        </TableRow>

                        <TableRow>
                            <TableCell>
                                <Typography sx={{fontSize:"large"}}>Your Rating</Typography>
                            </TableCell>
                            <TableCell>
                                <Rating
                                    max={5}
                                    value={dishData.ratingOfUser === undefined ? null : dishData.ratingOfUser}
                                    onChange={(_event, value) => value !== null && updateUserVoting(value)}
                                />
                            </TableCell>
                        </TableRow>
                    </TableBody>
                </Table>
            </TableContainer>
        </Stack>
    )
}