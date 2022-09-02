import { Button, Col, Container, Row} from "react-bootstrap";

export  function LoginPage() {
    return (
        <Container  className="justify-content-center" >
            <Col sm={5}>
                <Row>
                    <h1 >
                        ITS (hopefully) Tasty
                    </h1>
                    <p>
                        Welcome! This site lets you rate the quality of the dishes in our beloved mensa and
                        UKSH bistro. Login to proceed.
                    </p>
                </Row>
                <Row >
                    <Button variant={"primary"} href={"/authAPI/login"} >Login</Button>
                </Row>
            </Col>

        </Container>
    );
}