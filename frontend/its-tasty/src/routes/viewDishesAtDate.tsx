import {Badge, Button} from "react-bootstrap";
import {useEffect, useState} from "react";
import moment, {Moment} from 'moment'
import {useAuthContext} from "../AuthContext";
import {SimpleDishView, SimpleDishViewMock} from "../services/SimpleDishView";
import {ApiError, DefaultService} from "../services/userAPI";

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
    const [errorMessage,setErrorMessage] = useState("")
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
                    <div className={"d-flex row"}>
                       No Entries
                    </div>
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
            content = (<h1>Error: {errorMessage}</h1>);
            break;
    }

    return (
        <div>
            <div className={"d-flex row text-center"}><h2>{props.location}</h2></div>
            {content}
        </div>
    )
}


/**
 * Show Date selector and call DishByDateLocationView for the selected data for both Mensa and UKSH location
 * @constructor
 */
export function ViewDishesAtDate() {
    const [date,setDate] = useState<moment.Moment>(moment())
    return(
        <div className={"d-grid justify-content-center"} >
            <div id={"buttonRow"} className={"d-flex row justify-content-center"}>
                <div className={"d-flex col w-100 justify-content-start"}>
                    <Button
                        key={"prevDate"} variant={"primary"}
                        onClick={ () => {
                            setDate( moment(date).subtract(1,"day"))
                        }}
                    >
                        Previous Day
                    </Button>
                </div>
                <div className={"d-flex col justify-content-center"}>
                    <Badge pill bg={"success"} className={"d-flex align-items-center"}>{date.format('DD.MM.YY')}</Badge>
                </div>
                <div className={"d-flex col justify-content-end"}>
                    <Button
                        key={"nextDate"} variant={"primary"}
                        disabled={ date.isSame(moment(),'day')}
                        onClick={ () => {
                            setDate( moment(date).add(1,"day"))
                        }}
                    >
                        Next Day
                    </Button>
                </div>
            </div>
            <div className={"d-flex"}>
                <div className={"d-flex col"}>
                    <DishByDateLocationView date={date} location={"UKSH"}/>
                </div>
                <div className={"d-flex  col"}>
                    <DishByDateLocationView date={date} location={"Mensa"}/>
                </div>
            </div>

        </div>
    )
}