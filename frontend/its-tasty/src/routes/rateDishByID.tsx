import {useParams} from "react-router-dom";
import {DishVIew} from "../DishView";

export function RateDishByID() {
    let {id} = useParams();

    if( id == undefined ) {
        return <div>
            Invalid dish id
        </div>
    }
    const idAsNumber = parseInt(id)
    if( isNaN(idAsNumber) ) {
        return <div>
            Invalid dish id
        </div>
    }

    return (
        <div>
            <DishVIew dishID={idAsNumber} />
        </div>
    )
}