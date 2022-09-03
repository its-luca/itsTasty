import { Container} from "react-bootstrap";
import {ApiError, DefaultService, GetDishResp, RateDishReq} from "./services/userAPI";
import {useEffect, useState} from "react";
import {useAuthContext} from "./AuthContext";
import {Rating} from 'react-simple-star-rating'

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
    if( authContext == undefined ) {
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

    const uiRatingToReqParam = (rating :number) : RateDishReq.rating|undefined => {
        switch (rating) {
            case 20:
                return RateDishReq.rating._1
            case 40:
                return RateDishReq.rating._2
            case 60:
                return RateDishReq.rating._3
            case 80:
                return RateDishReq.rating._4
            case 100:
                return RateDishReq.rating._5
            default:
                return undefined
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

    const updateUserVoting = async (rating :number) => {

        const validatedRating = uiRatingToReqParam(rating)
        if( validatedRating === undefined ) {
            console.log(`updateUserVoting called with invalid rating ${rating}`)
            return
        }

        try {
            await DefaultService.postDishes(props.dishID,{rating:validatedRating})
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

    if( state == State.loading ) {
        return(
            <Container >
                <p>Loading...</p>
            </Container>
        )
    }
    if( state == State.error || dishData == undefined) {
        return (
            <Container>
                <p>
                    Error : {errorMessage}. Maybe try loading again?
                </p>
            </Container>
        )
    }


    let ratings = new Array<string>()
    for( let rating in dishData.ratings ) {
        let count = dishData.ratings[rating]
        ratings.push( rating + " Stars: " + count.toString())
    }

    //occurrenceString contains up to the two most recent occurrences or is set to "never"
    let occurrenceString : string;
    if( dishData.recentOccurrences.length == 0 ) {
        occurrenceString = "never"
    } else if (dishData.recentOccurrences.length === 1 ) {
        occurrenceString = dishData.recentOccurrences[0];
    } else { // >= 2
            occurrenceString = dishData.recentOccurrences[0] + ", " + dishData.recentOccurrences[1];
    }


    return (
        <div className="d-flex justify-content-center">
            <div className={"col-lg-4 col-md"}>
                <div className={"row"}>
                    <div className={"col"}>Name</div>
                    <div className={"col"}>{dishData.name}</div>
                </div>
                <div className={"row"}>
                    <div className={"col"}>Served at</div>
                    <div className={"col"}>{dishData.servedAt}</div>
                </div>
                <div className={"row"}>
                    <div className={"col"}>Occurrence Count</div>
                    <div className={"col"}>{dishData.occurrenceCount}</div>
                </div>
                <div className={"row"}>
                    <div className={"col"}>Most recent servings</div>
                    <div className={"col"}>{occurrenceString}</div>
                </div>
                <div className={"row"}>
                    <div className={"col"}>Ratings</div>
                    <div className={"col"}>
                        { ratings.length === 0 ? "No Ratings yet" : ratings.join(", ") }
                    </div>
                </div>
                <div className={"row"}>
                    <div className={"col"}>Average Rating</div>
                    <div className={"col"}>
                        {
                            (dishData.avgRating !== undefined  && dishData.avgRating !== 0)? dishData.avgRating : "No votes yet"
                        }
                    </div>
                </div>
                <div className={"row"}>
                    <div className={"col d-flex align-items-center"}>Your Rating</div>
                    <div className={"col"}>
                        <Rating
                            ratingValue={ ( dishData.ratingOfUser === undefined) ? NaN : apiRatingToUiRating(dishData.ratingOfUser)}
                            onClick={ (value: number) => updateUserVoting(value)}
                        />
                    </div>
                </div>
            </div>
        </div>
    )
}