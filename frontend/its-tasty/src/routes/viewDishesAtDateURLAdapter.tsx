import {Navigate, useParams} from "react-router-dom";
import moment from "moment";
import {ViewDishesAtDate} from "../viewDishesAtDate";
import urlJoin from "url-join";

export function ViewDishesAtDateURLAdapter() {
    let {dateString} = useParams();
    if( dateString === undefined ) {
        return <div>
           <Navigate
               to={urlJoin('/dishesByDate', moment().format("DD-MM-YYYY") )}
           />
        </div>
    }
    const date = moment(dateString,"DD-MM-YYYY")

    return (
        <div>
            <ViewDishesAtDate date={date}/>
        </div>
    )
}