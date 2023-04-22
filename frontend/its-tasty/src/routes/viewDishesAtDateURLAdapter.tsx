import {Navigate, useParams} from "react-router-dom";
import moment from "moment";
import {ViewDishesAtDate} from "../viewDishesAtDate";
import urlJoin from "url-join";
import { Box } from "@mui/material";

export function ViewDishesAtDateURLAdapter() {
    let {dateString} = useParams();
    if( dateString === undefined ) {
        return <Box>
            <Navigate
                to={urlJoin('/dishesByDate', moment().format("DD-MM-YYYY"))}
            />
        </Box>
           
    }
    const date = moment(dateString,"DD-MM-YYYY")

    return (
            <ViewDishesAtDate date={date}/>
    )
}