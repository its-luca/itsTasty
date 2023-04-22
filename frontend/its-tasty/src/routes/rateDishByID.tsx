import {useParams} from "react-router-dom";
import {DishVIew} from "../DishView";
import {ViewDishMergeCandidates} from "../DishMergeCandidatesView";
import {MergedDishView} from "../mergedDishes/mergedDishView";
import { Container } from "@mui/material";

export function RateDishByID() {
    let {id} = useParams();

    if( id === undefined ) {
        return <Container>
            Invalid dish id
        </Container>
    }
    const idAsNumber = parseInt(id)
    if( isNaN(idAsNumber) ) {
        return <Container>
            Invalid dish id
        </Container>
    }

    return (
            <DishVIew dishID={idAsNumber} showRatingData={true} />
    )
}

export function MergeDishByIDAdapter() {
    let {id} = useParams();

    if( id === undefined ) {
        return <Container>
            Invalid dish id
        </Container>
    }
    const idAsNumber = parseInt(id)
    if( isNaN(idAsNumber) ) {
        return <Container>
            Invalid dish id
        </Container>
    }

    return (
            <ViewDishMergeCandidates dishID={idAsNumber} />
    )
}

export function MergedDishViewByIDAdapter() {
    let {id} = useParams();

    if( id === undefined ) {
        return <Container>
            Invalid dish id
        </Container>
    }
    const idAsNumber = parseInt(id)
    if( isNaN(idAsNumber) ) {
        return <Container>
            Invalid dish id
        </Container>
    }

    return (
            <MergedDishView mergedDishID={idAsNumber} />
    )
}