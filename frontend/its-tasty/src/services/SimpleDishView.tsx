import {Badge, Card, ListGroup} from "react-bootstrap";
import {Rating} from "react-simple-star-rating";
import {useEffect, useState} from "react";
import {ApiError, DefaultService, GetDishResp, RateDishReq} from "./userAPI";
import {useAuthContext} from "../AuthContext";
import {Link} from "react-router-dom";
import urlJoin from "url-join";



export function SimpleDishViewMock() {
    return(
        <Card className={"loadingAnimation"}>
            <Card.Body>
                <Card.Title>
                    <Badge pill bg={"secondary"} className={"d-flex"}>{}</Badge>
                </Card.Title>
                <ListGroup>
                    <ListGroup.Item >
                        Average Rating <Rating readonly={true} ratingValue={0}></Rating>
                    </ListGroup.Item>
                    <ListGroup.Item>
                        Your Rating <Rating readonly={true} ratingValue={0}></Rating>
                    </ListGroup.Item>
                </ListGroup>
            </Card.Body>
        </Card>
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
    

    const apiRatingToUiRating = ( rating : GetDishResp.ratingOfUser) : number => {
        switch (rating) {
            case GetDishResp.ratingOfUser._1:
                return 20
            case GetDishResp.ratingOfUser._2:
                return 40
            case GetDishResp.ratingOfUser._3:
                return 60
            case GetDishResp.ratingOfUser._4:
                return 80
            case GetDishResp.ratingOfUser._5:
                return 100
        }
    }

    useEffect( () => {
        fetchDish()
    },[props.dishID])

    let content;
    if (state === State.loading) {
        content = (<SimpleDishViewMock/>)
    } else if ( state === State.error || dishData == undefined ) {
        content = (
            <Card>
                <Card.Body>
                    <Card.Title>Dish {props.dishID}</Card.Title>
                    <Card.Text>
                        Failed to fetch: {errorMessage}
                    </Card.Text>
                </Card.Body>
            </Card>
        )
    } else {
        let averageRating = NaN;
        if( dishData.avgRating !== undefined && dishData.avgRating !== 0) {
            averageRating = dishData.avgRating;
        }
        let userRating = NaN;
        if( dishData.ratingOfUser !== undefined ) {
            userRating = dishData.ratingOfUser
        }
        content = (
            <Card>
                <Card.Body>
                    <Card.Title >
                        <Link to={new URL(urlJoin(process.env.REACT_APP_PUBLIC_URL!,'dish',props.dishID.toString()))}>
                            {dishData.name}
                        </Link>
                    </Card.Title>
                    <ListGroup>
                        <ListGroup.Item>
                            <div className={"row"}>
                                <div className={"col"}>Average Rating</div>
                                <div className={"col"}> <Rating readonly={true} ratingValue={apiRatingToUiRating(averageRating)}></Rating></div>
                            </div>
                        </ListGroup.Item>
                        <ListGroup.Item>
                            <div className={"row"}>
                                <div className={"col"}>Your Rating</div>
                                <div className={"col"}> <Rating readonly={true} ratingValue={apiRatingToUiRating(userRating)}/></div>
                            </div>
                        </ListGroup.Item>
                    </ListGroup>
                </Card.Body>
            </Card>
        )
    }

    return(
       content
    )
}