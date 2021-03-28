import React, {useContext, useState} from "react";
import {useHistory, useParams} from "react-router-dom";
import "css/workflow.css";

import Row from "react-bootstrap/Row";
import Col from "react-bootstrap/Col";
import Alert from "react-bootstrap/Alert";
import Modal from "react-bootstrap/Modal";
import Button from "react-bootstrap/Button";
import Container from "react-bootstrap/Container";
import {X} from "react-bootstrap-icons"


import Editor from "components/workflow/editor";
import ServerContext from 'components/app/context'

export default function JSONInputModal(props) {
    const {modalTitle, modalDescription, modalExec, cancelText, confirmText, show, setShow} = props
    const history = useHistory()
    const context = useContext(ServerContext);
    const [error, setError] = useState({message: "", type: ""});
    const [inputVal, setInputVal] = useState("{}\n\n\n\n\n\n\n\n\n\n\n\n");
    let params = useParams()
    let {workflow, namespace} = params

    const closeInputModal = (exec) => {
        setError({message: "", type: ""})
        if (exec) {
            try {
                JSON.parse(inputVal)
            } catch (e) {
                setError({message: `${e.message}`, type: "json"})
                return
            }

            modalExec(inputVal, {
                context: context,
                namespace: namespace,
                workflow: workflow,
                history: history
            }).then(() => {
                setShow(false)
                setInputVal("{}\n\n\n\n\n\n\n\n\n\n\n\n")
            }).catch((e) => {
                let type = e.type ? e.type : "server"
                if (e && e.message && e.message !== "") {
                    setError({message: e.message, type: type})
                } else {
                    console.log("Impossible error happened: ", e)
                }
            })
        } else {
            setShow(false)
            setInputVal("{}\n\n\n\n\n\n\n\n\n\n\n\n")
        }
    };


    return (
        < Modal dialogClassName="input-modal" show={show} onHide={() => {
            closeInputModal(false)
        }
        }>
            <Modal.Header closeButton>
                <Modal.Title>{modalTitle}</Modal.Title>
            </Modal.Header>
            <Modal.Body>
                {error.type === "json" ?
                    (<Container style={{margin: "0px", paddingLeft: "0px", minWidth: "100%"}}>
                        <Alert variant="danger" style={{paddingTop: "0px", paddingBottom: "0px"}}>
                            <Row>
                                <Col>
                                    <div style={{
                                        padding: "0.375rem 0.375rem 0.375rem 0.375rem",
                                    }}>
                                        Invalid JSON: {error.message}
                                    </div>
                                </Col>
                                <Col sm={1} style={{display: "flex", justifyContent: "end", paddingRight: "0px"}}>
                                    <Button
                                        style={{
                                            marginLeft: "6px",
                                            textAlign: "center",
                                            padding: "0.375rem 0.375rem 0.375rem 0.375rem",
                                            color: "#721c24"
                                        }}
                                        variant="link"
                                        onClick={() => {
                                            setError({message: "", type: ""})
                                        }}
                                    >
                                        <X width="1.5em" height="1.5em"/>
                                    </Button>
                                </Col>
                            </Row>
                        </Alert>
                    </Container>)
                    :
                    (<></>)
                }
                {error.type === "server" ?
                    (<Container style={{margin: "0px", paddingLeft: "0px", minWidth: "100%"}}>
                        <Alert variant="danger" style={{paddingTop: "0px", paddingBottom: "0px"}}>
                            <Row>
                                <Col>
                                    <div style={{
                                        padding: "0.375rem 0.375rem 0.375rem 0.375rem",
                                    }}>
                                        Invalid Body: {error.message}
                                    </div>
                                </Col>
                                <Col sm={1} style={{display: "flex", justifyContent: "end", paddingRight: "0px"}}>
                                    <Button
                                        style={{
                                            marginLeft: "6px",
                                            textAlign: "center",
                                            padding: "0.375rem 0.375rem 0.375rem 0.375rem",
                                            color: "#721c24"
                                        }}
                                        variant="link"
                                        onClick={() => {
                                            setError({message: "", type: ""})
                                        }}
                                    >
                                        <X width="1.5em" height="1.5em"/>
                                    </Button>
                                </Col>
                            </Row>
                        </Alert>
                    </Container>)
                    :
                    (<></>)
                }
                {error.type === "invalidArg" ?
                    (<Container style={{margin: "0px", paddingLeft: "0px", minWidth: "100%"}}>
                        <Alert variant="danger" style={{paddingTop: "0px", paddingBottom: "0px"}}>
                            <Row>
                                <Col>
                                    <div style={{
                                        padding: "0.375rem 0.375rem 0.375rem 0.375rem",
                                    }}>
                                        Could Not Execute Workflow: {error.message}
                                    </div>
                                </Col>
                                <Col sm={1} style={{display: "flex", justifyContent: "end", paddingRight: "0px"}}>
                                    <Button
                                        style={{
                                            marginLeft: "6px",
                                            textAlign: "center",
                                            padding: "0.375rem 0.375rem 0.375rem 0.375rem",
                                            color: "#721c24"
                                        }}
                                        variant="link"
                                        onClick={() => {
                                            setError({message: "", type: ""})
                                        }}
                                    >
                                        <X width="1.5em" height="1.5em"/>
                                    </Button>
                                </Col>
                            </Row>
                        </Alert>
                    </Container>)
                    :
                    (<></>)
                }
                <h3 style={{fontSize: "1em"}}>{modalDescription}</h3>
                <Editor
                    value={inputVal}
                    edit={true}
                    onChange={(e => {
                        setInputVal(e)
                    })}
                    readOnly={false}
                    mode="json"
                    style={{width: "100%", borderRadius: "4px", marginTop: "16px"}}
                />
            </Modal.Body>
            <Modal.Footer>
                <Button variant="secondary" onClick={() => {
                    closeInputModal(false)
                }}>
                    {cancelText ? cancelText : "Cancel"}
                </Button>
                <Button variant="primary" onClick={() => {
                    closeInputModal(true)
                }}>
                    {confirmText ? confirmText : "Confirm"}
                </Button>
            </Modal.Footer>
        </Modal>
    );
}