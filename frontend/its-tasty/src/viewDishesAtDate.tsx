import {useEffect, useState} from "react";
import moment, {Moment} from 'moment'
import {useAuthContext} from "./AuthContext";
import {SimpleDishView, SimpleDishViewMock} from "./SimpleDishView";
import {ApiError, DefaultService} from "./services/userAPI";
import Button from "@mui/material/Button";
import { Paper, Stack, useMediaQuery} from "@mui/material";
import Grid2 from '@mui/material/Unstable_Grid2';
import Container from "@mui/material/Container";
import Typography from "@mui/material/Typography"; // Grid version 2
import { useTheme } from '@mui/material/styles';
import ErrorOutlineIcon from '@mui/icons-material/ErrorOutline';
import { Link as RRLink} from "react-router-dom";
import urlJoin from "url-join";


interface DishByDateLocationViewProps {
    date: Moment
    location:string
}

/**
 * Fetch dishes for date and location given in props
 * @param props
 * @constructor
 */
function DishByDateLocationView(props: DishByDateLocationViewProps) {
    enum State {
        loading,
        error,
        success
    }
    const [state,setState] = useState<State>(State.loading)
    const [_,setErrorMessage] = useState("")
    const [dishIDs,setDishIDs] = useState<Array<number>>(new Array<number>());
    const authContext = useAuthContext()
    if( authContext === undefined ) {
        console.log("authContext undefined")
        setState(State.error)
        setErrorMessage("Internal Error")
    }

    const fetchDishIDs = async (when: moment.Moment, location : string) => {
        try {
            const reply = await DefaultService.postSearchDishByDate(
                {
                    date:when.format("YYYY-MM-DD"),
                    location:location
                }
            )
            setDishIDs(reply)
            setState(State.success)
        } catch (e) {
            if (e instanceof ApiError) {
                switch (e.status) {
                    case 500:
                        setErrorMessage( "Failed to query dishes: Internal Server Error")
                        break;
                    case 401:
                        authContext?.setAuthStatus(false)
                        break;
                    default:
                        setErrorMessage("Failed to query dishes: Unknown Error")
                        break;
                }
            } else {
                setErrorMessage( "Failed to query dishes: Unknown Error")
                console.log(`fetchDishIDs unexpected error : ${e}`);
            }
            setState(State.error)
        }
    }

    useEffect(() => {
        fetchDishIDs(props.date,props.location)
    },[props.date,props.location])


    let content;
    switch (state) {
        case State.success:
            if( dishIDs.length == 0) {
                content =   (
                    <Typography sx={{pt:"20px",textAlign:"center"}} variant={"h5"}>No Dishes</Typography>
                )
            } else {
                content = dishIDs.map( (id) => {
                    return (
                        <div className={"d-flex row "}>
                            <SimpleDishView dishID={id}/>
                        </div>
                    )
                });
            }
            break;
        case State.loading:
            content =(
                <div>
                    <div className={"d-flex row"}>
                        <SimpleDishViewMock/>
                    </div>
                    <div className={"d-flex row"}>
                        <SimpleDishViewMock/>
                    </div>
                    <div className={"d-flex row"}>
                        <SimpleDishViewMock/>
                    </div>
                    <div className={"d-flex row"}>
                        <SimpleDishViewMock/>
                    </div>
                </div>

            )
            break;
        case State.error:
            content =  (
                <Container sx={{display:"flex",justifyContent:"center"}}>
                    <ErrorOutlineIcon fontSize={"large"}/>
                </Container>
            )

            break;
    }

    return (
        <Stack>
            <Typography sx={{textAlign:"center"}} variant={"h3"}>{props.location}</Typography>
            {content}
        </Stack>

    )
}


interface ViewDishesAtDateProps {
    date: Moment
}

/**
 * Show Date selector and call DishByDateLocationView for the selected data for both Mensa and UKSH location
 * @constructor
 */
export function ViewDishesAtDate(props :ViewDishesAtDateProps) {
    const theme = useTheme();
    const isSmallScreen = useMediaQuery(theme.breakpoints.down("sm"));

    const buttonSize = isSmallScreen ? "50px" : "150px";

    const prevButton = (
        <Button
            component={RRLink}
            to={urlJoin('/dishesByDate', moment(props.date).subtract(1,"day").format("DD-MM-YYYY") )}
            sx={{width: buttonSize}}
            variant={"contained"}
            key={"prevDate"}
            /*onClick={ () => {
                setDate( moment(date).subtract(1,"day"))
            }}*/
        >
            {isSmallScreen ? "Prev" : "Previous Day"}
        </Button>
    )

    const nextButton = (
        <Button
            component={RRLink}
            to={urlJoin('/dishesByDate', moment(props.date).add(1,"day").format("DD-MM-YYYY") )}
            sx={{width: buttonSize}}
            variant={"contained"}
            key={"nextDate"}
            disabled={ props.date.isSame(moment(),'day')}
            /*onClick={ () => {
                setDate( moment(date).add(1,"day"))
            }}*/
        >
            {isSmallScreen ? "Next" : "Next Day"}
        </Button>
    );

    const dateDisplay = (
        <Paper elevation={10} sx={{display:"flex",justifyContent:"center",alignItems:"center",pl:"10px",pr:"10px"}}>
            <Typography fontSize={"larger"} fontWeight={"bold"}>
                {props.date.format("DD.MM.YY")}
            </Typography>
        </Paper>
    );

    return(
        <Container maxWidth={"xl"} >
            <Container   sx={{mt:"10px",display:"flex",justifyContent:"space-between"}} >
                {prevButton}
                {dateDisplay}
                {nextButton}
            </Container>
            <Grid2 container sx={{mt:"10px",display:"flex",justifyContent:"space-around",alignItems:"start"}}>
                <Grid2>
                       <DishByDateLocationView date={props.date} location={"UKSH"}/>
                </Grid2>
                <Grid2>
                        <DishByDateLocationView date={props.date} location={"Mensa"}/>
                </Grid2>
            </Grid2>
        </Container>
    )
}