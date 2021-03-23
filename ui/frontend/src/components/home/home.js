import React, {useCallback, useContext, useEffect, useState} from "react";
import {Alert, Container} from "react-bootstrap"
import {XCircle} from "react-bootstrap-icons"

import Row from "react-bootstrap/Row";
import Col from "react-bootstrap/Col";


import HomepageSidebar from "components/home/home-sidebar";
import HomepageActivities from "components/home/home-activities";
import ServerContext from 'components/app/context'

export default function Home(props) {

    const context = useContext(ServerContext);
    const [namespaces, setNamespaces] = useState([]);
    const [loading, setLoading] = useState(false)
    const [err, setErr] = useState("")

    const clearError = () => {
        setErr("")
    }
    const fetchNamespaces = useCallback(
        () => {

            async function fetchNamespaces() {
                try {
                    setLoading(true)
                    let resp = await context.Fetch(`/namespaces`, {
                        method: `GET`
                    })
                    if (resp.ok) {
                        let json = await resp.json()
                        setNamespaces(json.namespaces)
                    } else {
                        setErr(await resp.text())
                    }
                } catch (e) {
                    setErr(e.message)
                }
                setLoading(false)
            }

            fetchNamespaces()
        },
        [context.Fetch],
    )

    useEffect(() => {
        fetchNamespaces()
    }, [fetchNamespaces])

    // if were loading from the API still
    if (loading) {
        return (
            <Row style={{margin: "0px"}}>
                <Col style={{marginBottom: "15px"}} xs={12} md={12}>
                    Loading...
                </Col>
            </Row>
        )
    }

    if (err !== "") {
        return (
            <Row style={{margin: "0px"}}>
                <Col style={{marginBottom: "15px"}} xs={12} md={12}>
                    <Alert variant="danger">
                        <Container>
                            <Row>
                                <Col sm={11}>
                                    Fetching Namespaces: {err}
                                </Col>
                                <Col sm={1} style={{textAlign: "right", paddingRight: "0"}}>
                                    <XCircle style={{cursor: "pointer", fontSize: "large"}} onClick={() => {
                                        clearError()
                                    }}/>
                                </Col>
                            </Row>
                        </Container>
                    </Alert>
                </Col>
            </Row>
        )
    }

    // Show the entire page
    return (
        <Row style={{margin: "0px"}}>
            <Col style={{marginBottom: "15px"}} xs={12} md={3}>
                <HomepageSidebar fetchNamespaces={fetchNamespaces} namespaces={namespaces}/>
            </Col>
            <Col className="no-left-pad-above-sm" style={{marginBottom: "15px"}} xs={12} md={9}>
                <HomepageActivities namespaces={namespaces}/>
            </Col>
        </Row>
    );

}
