import React, {useCallback, useContext, useState} from "react";
import {Link, Redirect, useParams} from "react-router-dom";
import "css/workflow.css";
// import prettyYAML from "json-to-pretty-yaml"
import YAML from "js-yaml";
import logoColor from "img/logo-color.png";
import Row from "react-bootstrap/Row";
import Col from "react-bootstrap/Col";
import Alert from "react-bootstrap/Alert";
import Container from "react-bootstrap/Container";
import {XCircle} from "react-bootstrap-icons";


import Editor from "components/workflow/editor";
import WorkflowActions from "components/workflow/actions";
import ServerContext from "components/app/context";
import {RemoteResourceState} from "util/utils";
import JSONInputModal from "./json-input-modal";


// Delete a workflow
export async function DeleteWorkflow(context, namespace, workflow) {
    let resp = await context.Fetch(
        `/namespaces/${namespace}/workflows/${workflow}`,
        {
            method: "delete",
        }
    );
    if (!resp.ok) {
        throw resp;
    }
}

export default function Workflow(props) {
    let params = useParams();
    let {namespace} = params;

    const context = useContext(ServerContext);
    const [val, setVal] = useState(""); // Workflow Value
    const [error, setError] = useState({message: "", lang: ""});
    const [rrs, setRRS] = useState(RemoteResourceState.fetching);
    const [active, setActive] = useState();

    const [showInput, setShowInput] = useState(false);
    const [workflow, setWorkflow] = useState({id: params.workflow, uid: ""});

    const clearError = () => {
        setError({message: "", lang: ""});
    };

    // Start a workflow pipepline
    async function runWorkflow(inputJSON, args) {
        let resp = await args.context.Fetch(
            `/namespaces/${args.namespace}/workflows/${args.workflow}/execute`,
            {
                method: "post",
                body: inputJSON,
            }
        );
        if (!resp.ok) {
            let text = await resp.text();
            if (text === "workflow is inactive") {
                setActive(false)
            }
            throw {message: text, type: "invalidArg"};
        } else {
            let json = await resp.json();
            args.context.AddToast(
                `Started Workflow Instance`,
                <>
          <span>
            Workflow {args.workflow} successfully started instance
            <Link to={`/i/${json["instanceId"]}`}>
              {` ${json["instanceId"]} `}
            </Link>
          </span>
                </>
            );
            args.history.push(`/i/${json["instanceId"]}`);
        }
    }

    const fetchWorkflow = useCallback(() => {
        async function fetchWorkflow() {
            let resp = await context.Fetch(
                `/namespaces/${namespace}/workflows/${workflow.id}?name`,
                {}
            );
            if (!resp.ok) {
                setRRS(RemoteResourceState.failed);
                try {
                    console.log("failed to fetch workflow: " + (await resp.text()));
                } catch (e) {
                    console.log("failed to fetch workflow: unknown error", e);
                }
            } else {
                let json = await resp.json();
                let wf = atob(json.workflow);
                let yamlWF = YAML.load(wf);
                let startType = "default";

                if (
                    yamlWF.start &&
                    yamlWF.start.type
                ) {
                    startType = yamlWF.start.type;
                }

                setWorkflow((workflow) => {
                    workflow.id = json.id;
                    workflow.uid = json.uid;
                    workflow.active = json.active;
                    workflow.description = json.description;
                    workflow.createdAt = json.createdAt;
                    workflow.startType = startType;
                    return workflow;
                });

                setActive(json.active)

                setVal(wf);
                setRRS(RemoteResourceState.successful);
            }
        }

        fetchWorkflow();
    }, []);
    // Fetch data on mount
    React.useEffect(() => {
        fetchWorkflow();
    }, [context.Fetch, workflow, namespace]);

    function renderSwitch(state) {
        switch (state) {
            case RemoteResourceState.successful:
                return (
                    <>
                        <Row style={{margin: "0px"}}>
                            <Col xs={12} id="workflow-header">
                                <div className="workflow-actions">
                                    <div
                                        id="workflow-actions-title"
                                        className="workflow-actions-box"
                                    >
                                        <div className="workflow-actions-header">
                                            <Row>
                                                <h4>{workflow.id}</h4>
                                            </Row>
                                        </div>
                                    </div>
                                    <div
                                        id="workflow-actions-options"
                                        className="workflow-actions-box"
                                    >
                                        <WorkflowActions
                                            active={active}
                                            startType={workflow.startType}
                                            onRun={() => {
                                                setShowInput(true);
                                            }}
                                        />
                                    </div>
                                </div>
                            </Col>
                            <Col xs={12}>
                                <Col xs={12} style={{padding: "0px"}}>
                                    {error.type === "workflow" ? (
                                        <Alert variant="danger">
                                            <Container>
                                                <Row>
                                                    <Col sm={11}>{error.message}</Col>
                                                    <Col
                                                        sm={1}
                                                        style={{
                                                            textAlign: "right",
                                                            paddingRight: "0",
                                                        }}
                                                    >
                                                        <XCircle
                                                            style={{
                                                                cursor: "pointer",
                                                                fontSize: "large",
                                                            }}
                                                            onClick={() => {
                                                                clearError();
                                                            }}
                                                        />
                                                    </Col>
                                                </Row>
                                            </Container>
                                        </Alert>
                                    ) : (
                                        <></>
                                    )}
                                    <Editor
                                        value={val}
                                        onChange={(e) => {
                                            setVal(e);
                                        }}
                                        readOnly={true}
                                        style={{
                                            width: "100%",
                                            borderRadius: "4px",
                                            marginTop: "10px",
                                        }}
                                    />
                                </Col>
                            </Col>
                        </Row>
                    </>
                );
            case RemoteResourceState.failed:
                return (
                    <>
                        <Redirect to="/404"/>
                    </>
                );
            default:
                return (
                    <>
                        <Row style={{margin: "0px"}}>
                            <Col xs={12}>
                                <div
                                    style={{
                                        minHeight: "500px",
                                        display: "flex",
                                        alignItems: "center",
                                        justifyContent: "center",
                                    }}
                                >
                                    <img
                                        alt="loading symbol"
                                        src={logoColor}
                                        height={200}
                                        className="animate__animated animate__bounce animate__infinite"
                                    />
                                </div>
                            </Col>
                        </Row>
                    </>
                );
        }
    }

    return (
        <>
            <JSONInputModal
                modalTitle="Workflow Input"
                modalDescription="Input JSON Value for Workflow Execution"
                modalExec={runWorkflow}
                show={showInput}
                setShow={setShowInput}
                cancelText="Cancel"
                confirmText="Execute Workflow"
            />
            {renderSwitch(rrs)}
        </>
    );
}
