import {useEffect, useState} from "react"
import Row from "react-bootstrap/Row";
import Col from "react-bootstrap/Col";
import {Spinner} from "react-bootstrap"
import {Redirect} from "react-router-dom"


export function ResourceNotFound() {
    const [redirectTime, setRedirectTime] = useState(5000) // milliseconds you want to wait before redirect

    useEffect(() => {
        let ttt = null
        if (redirectTime === 0) {
            return
        }

        ttt = setTimeout(() => {
            setRedirectTime((rt) => {
                rt -= 1000
                if (rt < 0) {
                    return 0
                }
                return rt
            })
        }, 1000)

        return function cleanup() {
            if (ttt) {
                clearTimeout(ttt)
            }
        }
    }, [redirectTime])

    return (
        <>
            <Row style={{margin: "0px"}}>
                <Col xs={12}>
                    <h1 style={{textAlign: "center", paddingTop: "16px", paddingBottom: "20px", color: "#717070"}}>Could
                        Not Find Resource</h1>
                    <div style={{textAlign: "center"}}>
                        <Spinner animation="border" variant="primary"/>
                    </div>
                    <h5 style={{textAlign: "center", paddingTop: "0px", color: "#717070"}}>Redirecting to Home
                        in: {redirectTime / 1000} seconds...</h5>
                    {redirectTime === 0 ? (<Redirect to="/"/>) : (<></>)}
                </Col>
            </Row>
        </>
    )
}

export default ResourceNotFound