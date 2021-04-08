import React, {useCallback, useContext, useState} from 'react'

import ServerContext from "../app/context";
import {Link, useHistory} from "react-router-dom";
import {useDropzone} from "react-dropzone";
import YAML from "yaml";
import {ResourceRegex} from "../../util/utils";
import {Button, Container, Form, Modal, OverlayTrigger, Tooltip} from "react-bootstrap";
import {Upload} from "react-bootstrap-icons";

export default function NamespaceActions(props) {
    return (
        <div id="actions-container">
            <CreateWorkflowButton namespace={props.namespace} onNew={props.onNew} onRun={props.onRun}/>
        </div>
    );
}

function CreateWorkflowButton(props) {
    const context = useContext(ServerContext);
    const history = useHistory();
    const [show, setShow] = useState(false);
    const [error, setError] = useState();
    const [myFiles, setMyFiles] = useState([]);

    // Upload Workflow
    const onDrop = useCallback(
        (acceptedFiles, fileRejections) => {
            if (fileRejections.length > 0) {
                setError(`File: '${fileRejections[0].file.name}' is not supported, ${fileRejections[0].errors[0].message}`);
            } else {
                setError(null);
                setMyFiles([...acceptedFiles]);
            }
        },
        []
    );

    const {getRootProps, getInputProps} = useDropzone({
        onDrop,
        accept: "application/x-yaml, .yaml, .yml",
        maxFiles: 1,
        disabled: false,
    });

    // const removeFile = (file) => () => {
    //   const newFiles = [...myFiles];
    //   newFiles.splice(newFiles.indexOf(file), 1);
    //   setMyFiles(newFiles);
    // };

    const removeAll = () => {
        setMyFiles([]);
    };


    function readFile(file) {
        return new Promise((resolve, reject) => {
            const reader = new FileReader();

            reader.onload = (res) => {
                resolve(res.target.result);
            };
            reader.onerror = (err) => reject(err);

            reader.readAsText(file);
        });
    }

    const handleClose = () => {
        setShow(false);
        setError(null);
        removeAll();
    };
    const handleShow = () => {
        setShow(true);
    };
    const handleSubmit = async () => {
        setError(null);
        let workflowBody = null;
        let name = "NamePlaceholder";
        // Read / Validate File
        let rawFile = await readFile(myFiles[0]);
        let validJSON = false;
        let validYAML = false;
        let err = null;

        try {
            let workflow = YAML.parse(rawFile);
            if (workflow.id) {
                name = workflow.id;
                validYAML = true;
                workflowBody = rawFile;
            } else {
                err = "Workflow 'id' is missing from file";
            }
        } catch (e) {
        }

        if (!(validJSON || validYAML)) {
            // generic error
            if (!err) {
                setError("Invalid Workflow File");
            } else {
                setError(err);
            }
            return;
        }

        err = validationName(name);
        if (err != null) {
            setError(err);
            return;
        }


        fetch(`${context.SERVER_BIND}/namespaces/${props.namespace}/workflows`, {
                method: "POST",
                body: workflowBody,
            })
            .then((resp) => {
                // Real
                if (!resp.ok) {
                    throw resp;
                }

                props.onNew();
                context.AddToast(
                    `Workflow ${name} created`,
                    <>
            <span>
              Workflow
              <Link to={`/p/${props.namespace}/w/${name}`}>{` ${name} `}</Link>
              successfully created.
            </span>
                    </>
                );
                history.push(`/p/${props.namespace}/w/${name}`);

                handleClose();
            })
            .catch(async (e) => {
                try {
                    let err = "Failed to create workflow: " + (await e.text());
                    setError(err);
                } catch (unknownE) {
                    setError(unknownE);
                }
            });
    };

    function validationName(name) {
        if (!name || name === "") {
            return "Workflow name can not be empty";
        }

        if (name.length < 3) {
            return "Workflow name must be atleast three characters long";
        }

        if (name.match(/^\d/)) {
            return "Workflow name must start with lowercase letter";
        }

        if (!ResourceRegex.test(name)) {
            return "Workflow name may only use lowercase letters, numbers, and “-_”";
        }
        return null;
    }

    const renderTooltip = (props) => (
        <Tooltip id="button-tooltip" {...props}>
            Upload Workflow
        </Tooltip>
    );

    return (
        <>
            <OverlayTrigger
                placement="left"
                delay={{show: 100, hide: 150}}
                overlay={renderTooltip}
            >
                <Button
                    style={{
                        marginLeft: "6px",
                        background: "#e9ecef",
                        borderColor: "#e0e0e0",
                        textAlign: "center",
                        padding: "0.375rem 0.375rem 0.375rem 0.375rem",
                    }}
                    variant="light"
                    onClick={() => {
                        handleShow();
                    }}
                >
                    <Upload width="1.5em" height="1.5em"/>
                </Button>
            </OverlayTrigger>
            <Modal show={show} onHide={handleClose}>
                <Modal.Header closeButton>
                    <Modal.Title>
                        Upload a new Workflow
                        <br/>
                        {error ? (
                            <Form.Text className="text-danger" style={{fontSize: "0.6em"}}>
                                *{error}
                            </Form.Text>
                        ) : (
                            <></>
                        )}
                    </Modal.Title>
                </Modal.Header>
                <Modal.Body style={{padding: "0"}}>
                    <Container style={{padding: "1rem 1rem 1rem 1rem"}}>
                        <section className="workflowUpload">
                            <div {...getRootProps({className: "dropzone"})}>
                                <input {...getInputProps()} />
                                {myFiles && myFiles.length > 0 ? (
                                    <>
                                        <p style={{margin: "1rem 0 1rem 0"}}>
                                            Workflow: {myFiles[0].path}
                                        </p>
                                    </>
                                ) : (
                                    <div style={{padding: "0.7rem 0 0.7rem 0"}}>
                                        <p style={{margin: "0"}}>
                                            {" "}
                                            Drag 'n' drop or click upload workflow
                                        </p>
                                        <p style={{margin: "0"}}>(yaml)</p>
                                    </div>
                                )}
                            </div>
                        </section>
                    </Container>
                </Modal.Body>
                <Modal.Footer>
                    <Button variant="secondary" onClick={handleClose}>
                        Close
                    </Button>
                    <Button variant="primary" onClick={handleSubmit} disabled={myFiles.length === 0}>
                        Upload
                    </Button>
                </Modal.Footer>
            </Modal>
        </>
    );
}
