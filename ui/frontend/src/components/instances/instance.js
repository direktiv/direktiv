import React, {useCallback, useContext, useEffect, useState} from 'react'
import {useParams} from 'react-router-dom'
import ServerContext from "components/app/context"
import logoColor from "img/logo-color.png";

import {Alert, Tab, Tabs} from 'react-bootstrap'
import Container from 'react-bootstrap/Container'
import Row from 'react-bootstrap/Row'
import Col from 'react-bootstrap/Col'
import {XCircle} from "react-bootstrap-icons"

import InstanceLogs from './instance-logs';
import BackToWorkflow from './back-to-workflow';
import InstanceDetails from './instance-details'


export default function Instance() {

    const context = useContext(ServerContext);
    const params = useParams()
    let namespace = params.namespace;
    let id = params.id;
    let workflow = params.workflow;

    const [tab, setTab] = useState('details')
    const [loader, setLoader] = useState(false)
    const [err, setError] = useState("")

    const [basicValue, setBasicValue] = useState({input: {}, output: {}, status: null})


    const clearError = () => {
        setError("")
    }
    const fetchWorkflow = useCallback(() => {
        async function fetchWF() {
            try {
                let resp = await context.Fetch(`/instances/${namespace}/${workflow}/${id}`, {})
                if (!resp.ok) {
                    let text = await resp.text()
                    throw (new Error(`Error fetching instance workflow: ${text}`))
                }
                let json = await resp.json()
                setBasicValue(json)
                return json
            } catch (e) {
                setError(e.message)
            }
        }

        fetchWF()
    }, [context.Fetch, id, namespace, workflow])

    const fetchWorkflowInstanceBasic = useCallback(() => {
        async function fetchWorkflowInstanceBasic() {
            setLoader(true)
            await fetchWorkflow()
            setLoader(false)
        }

        fetchWorkflowInstanceBasic()
    }, [fetchWorkflow])


    useEffect(() => {
        fetchWorkflowInstanceBasic()
    }, [fetchWorkflowInstanceBasic])


    return (
        <Container id="instance">
            <Row>
                <Col id="namespace-header" style={{marginBottom: "15px"}}>
                    <div className="namespace-actions">
                        <div id="namespace-actions-title" className="namespace-actions-box">
                            <h4>
                                {workflow} ({namespace})
                            </h4>
                        </div>
                        <div id="namespace-actions-options" className="namespace-actions-box">
                            <BackToWorkflow/>
                        </div>
                    </div>
                </Col>
                <Col xs={12} style={{marginBottom: "15px"}}>
                    {err !== "" ?
                        <Row style={{margin: "0px"}}>
                            <Col style={{marginBottom: "15px"}} xs={12} md={12}>
                                <Alert variant="danger">
                                    <Container>
                                        <Row>
                                            <Col sm={11}>
                                                Fetching Instances: {err}
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
                        :
                        <>
                            {loader ?
                                <div style={{
                                    minHeight: "500px",
                                    display: "flex",
                                    alignItems: "center",
                                    justifyContent: "center",
                                    width: "100%"
                                }}>
                                    <img
                                        alt="loading symbol"
                                        src={logoColor}
                                        height={200}
                                        className="animate__animated animate__bounce animate__infinite"/>
                                </div>
                                :
                                <>
                                    <Tabs
                                        activeKey={tab}
                                        mountOnEnter={true}
                                        unmountOnExit={true}
                                        onSelect={(tab) => setTab(tab)}
                                    >
                                        <Tab eventKey="details" title="Details">
                                            <InstanceDetails fetchBasicWorkflow={fetchWorkflow}
                                                             basicValue={basicValue}/>
                                        </Tab>
                                        <Tab eventKey="logs" title="Logs">
                                            <InstanceLogs params={params}/>
                                        </Tab>
                                    </Tabs>
                                </>

                            }
                        </>
                    }

                </Col>
            </Row>

        </Container>
    )
}