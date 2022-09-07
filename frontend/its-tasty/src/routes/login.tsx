import { Button} from "react-bootstrap";

export  function LoginPage() {
    return (
        <div className={"d-flex vw-100 vh-100 justify-content-center align-items-center"}>
                <div className={"col-5 text-center"}>
                    <div className={"row"}>
                        <h1 >
                            ITS (hopefully) Tasty
                        </h1>
                        <p>
                            Welcome! This site lets you rate the quality of the dishes in our beloved mensa and
                            UKSH bistro. Login to proceed.
                        </p>
                    </div>
                    <div className={"row"}>
                        <Button variant={"primary"} href={`${process.env.REACT_APP_AUTH_API_BASE_URL}/authAPI/login`} >Login</Button>
                    </div>
                </div>

        </div>
    );
}