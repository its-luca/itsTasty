import {useParams} from "react-router-dom";
import {Container} from "react-bootstrap";

export function RateDishByID() {
    let {id} = useParams();

    if( id == undefined ) {
        return <Container>
            Invalid dish id
        </Container>
    }

    return (
        <Container>
            Voting site for dish id {id}
        </Container>
    )
}