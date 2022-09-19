import {ApiError, DefaultService, GetDishResp} from "./services/userAPI";
import {useEffect, useState} from "react";
import {useAuthContext} from "./AuthContext";
import {
    Container,
    Paper,
    Table,
    TableBody, TableCell,
    TableContainer,
    TableRow, Rating, SxProps, Theme
} from "@mui/material";
import moment from "moment";
import Box from "@mui/material/Box";
import Typography from "@mui/material/Typography";


interface DishViewProps {
    dishID : number
}

export function DishVIew(props : DishViewProps) {
   const [dishData,setDishData] = useState<GetDishResp|undefined>(undefined)

    enum State {
       loading,
        error,
        success
    }
    const [state,setState] = useState<State>(State.loading)
    const [errorMessage,setErrorMessage] = useState("")
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
            <Container >
                <p>Loading...</p>
            </Container>
        )
    }
    if( state === State.error || dishData === undefined) {
        return (
            <Container>
                <p>
                    Error : {errorMessage}. Maybe try loading again?
                </p>
            </Container>
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

    //occurrenceString contains up to the two most recent occurrences or is set to "never"
    let occurrenceString : string;
    if( dishData.recentOccurrences.length === 0 ) {
        occurrenceString = "never"
    } else if (dishData.recentOccurrences.length === 1 ) {
        //occurrenceString = moment().subtract(moment(dishData.recentOccurrences[0])).humazinze();
        const now = moment();
        const occurrence = moment(dishData.recentOccurrences[0]);
        occurrenceString = moment.duration(occurrence.diff(now)).humanize() + " ago"
    } else { // >= 2
        const now = moment();
        const o1 = moment(dishData.recentOccurrences[0]);
        const o2 = moment(dishData.recentOccurrences[1])
        occurrenceString =  moment.duration(o1.diff(now)).humanize() + " ago and "
            + moment.duration(o2.diff(now)).humanize() + " ago";
    }


    const variantLeftColumn = "h6"

    const sxLeftColum : SxProps<Theme> = {
        fontWeight:"medium",
    };

    return (
        <Container
            maxWidth="sm"
            sx={{
            marginTop:"10px"
        }}>
            <TableContainer component={Paper} elevation={10} >
                <Table>
                    <TableBody>
                        <TableRow>
                            <TableCell  component={Paper}>
                                <Typography variant={variantLeftColumn}>Name</Typography>
                            </TableCell>
                            <TableCell component={Paper}>
                                <Typography sx={sxLeftColum} >{dishData.name}</Typography>
                            </TableCell>
                        </TableRow>

                        <TableRow>
                            <TableCell component={Paper}>
                                <Typography variant={variantLeftColumn}>Served at</Typography>
                            </TableCell>
                            <TableCell component={Paper}>
                                <Typography sx={sxLeftColum} >{dishData.servedAt}</Typography>
                            </TableCell>
                        </TableRow>

                        <TableRow>
                            <TableCell sx={{width:"50%"}} component={Paper}>
                                <Typography variant={variantLeftColumn}>Occurrence Count</Typography>
                            </TableCell>
                            <TableCell sx={{width:"50%"}} component={Paper}>
                                <Typography sx={sxLeftColum} >{dishData.occurrenceCount}</Typography>
                            </TableCell>
                        </TableRow>

                        <TableRow>
                            <TableCell component={Paper}>
                                <Typography variant={variantLeftColumn}>Recent servings</Typography>
                            </TableCell>
                            <TableCell component={Paper}>
                                <Typography sx={sxLeftColum} >{occurrenceString}</Typography>
                            </TableCell>
                        </TableRow>

                        <TableRow>
                            <TableCell component={Paper}>
                                <Typography variant={variantLeftColumn}>Average Rating</Typography>
                            </TableCell>
                            <TableCell component={Paper}>
                                <Rating
                                    max={5}
                                    precision={0.25}
                                    value={ (dishData.avgRating !== undefined  && dishData.avgRating !== 0) ? dishData.avgRating : null }
                                    readOnly={true}
                                />
                            </TableCell>
                        </TableRow>

                        <TableRow>
                            <TableCell component={Paper}>
                                <Typography variant={variantLeftColumn}>Your Rating</Typography>
                            </TableCell>
                            <TableCell component={Paper}>
                              <Rating
                                  max={5}
                                  value={dishData.ratingOfUser === undefined ? null : dishData.ratingOfUser}
                                 onChange={(_event, value) => value !== null && updateUserVoting(value)}
                              />
                            </TableCell>
                        </TableRow>

                        <TableRow>
                            <TableCell component={Paper}>
                                <Typography variant={variantLeftColumn}>Ratings</Typography>
                            </TableCell>
                            <TableCell component={Paper}>
                            {
                                ratings.map( (value,index) => (
                                        <Box sx={{
                                            display:"flex",
                                            alignItems:"center",
                                        }}>
                                            <Typography sx={sxLeftColum} >{value}</Typography>
                                            <Rating name="read-only" value={value} max={index+1} readOnly />
                                        </Box>
                                ))}
                            </TableCell>
                        </TableRow>

                    </TableBody>
                </Table>
            </TableContainer>
        </Container>

    )
}