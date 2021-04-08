import {Button, Form, Modal, OverlayTrigger, Tooltip} from "react-bootstrap";
import {Plus} from "react-bootstrap-icons";
import {useContext, useState} from "react";
import ServerContext from "components/app/context";
import {ResourceRegex} from "util/utils";

const renderTooltip = (props) => (
    <Tooltip id="button-tooltip" {...props}>
        Create Namespace
    </Tooltip>
);

function validationName(name) {
    if (!name || name === "") {
        return "namespace can not be empty";
    }

    if (name.length < 3) {
        return "namespace must be atleast three characters long";
    }

    if (name.match(/^\d/)) {
        return "namespace must start with lowercase letter";
    }

    if (!ResourceRegex.test(name)) {
        return "namespace must be less than 36 characters and may only use lowercase letters, numbers, and “-_”";
    }
    return null;
}

export function NewNamespace(props) {
    const {fetchNamespaces} = props;
    const context = useContext(ServerContext);
    const [show, setShow] = useState(false);
    const [err, setErr] = useState("");

    const handleClose = () => {
        setShow(false);
        setErr("");
    };

    const handleShow = () => {
        setShow(true);
        setTimeout(() => {
            document.getElementById("namespace-id").focus();
        }, 400);
    };

    const handleKeyPress = async (target) => {
        if (target.charCode === 13) {
            await handleSubmit();
        }
    };

    const handleSubmit = async () => {
        if (document.getElementById("namespace-id").value !== "") {
            try {
                let namespace = document.getElementById("namespace-id").value;
                // Go style errors lol
                let err = validationName(namespace);
                if (err != null) {
                    let error = new Error();
                    error = { ...error, message: err};
                    throw error;
                }

                let resp = await fetch(`${context.SERVER_BIND}/namespaces/${namespace}`, {
                    method: "POST",
                });
                let text = await resp.text();
                if (!resp.ok) {
                    throw new Error(`Error creating namespace: ${text}`);
                }
                fetchNamespaces();
            } catch (e) {
                setErr(e.message);
            }
        } else {
            setErr("You must enter a name to create a namespace");
        }
    };

    return (
        <>
            <div
                style={{
                    display: "flex",
                    flex: "auto",
                    width: "100%",
                    flexDirection: "column-reverse",
                    alignItems: "center",
                }}
            >
                <div style={{alignSelf: "flex-end", paddingBottom: "3px"}}>
                    <OverlayTrigger
                        placement="left"
                        delay={{show: 100, hide: 150}}
                        overlay={renderTooltip}
                    >
                        <Button
                            style={{
                                marginLeft: "6px",
                                background: "#e9ecef",
                                border: "none",
                                textAlign: "right",
                                padding: "0px",
                            }}
                            variant="light"
                            onClick={handleShow}
                        >
                            <Plus width="1.5em" height="1.5em"/>
                        </Button>
                    </OverlayTrigger>
                    <Modal show={show} onHide={handleClose}>
                        <Modal.Header closeButton>
                            <Modal.Title>Create a Namespace</Modal.Title>
                        </Modal.Header>
                        <Modal.Body>
                            <Form onSubmit={(e) => e.preventDefault()} autocomplete="off">
                                <Form.Label>Name</Form.Label>
                                <Form.Control
                                    id="namespace-id"
                                    type="text"
                                    placeholder="Enter name for namespace"
                                    onKeyPress={handleKeyPress}
                                />
                                {err !== "" ? (
                                    <div className="text-danger" style={{fontSize: "0.8em"}}>
                                        *{err}
                                    </div>
                                ) : (
                                    ""
                                )}
                            </Form>
                        </Modal.Body>
                        <Modal.Footer>
                            <Button variant="secondary" onClick={handleClose}>
                                Close
                            </Button>
                            <Button variant="primary" onClick={handleSubmit}>
                                Create
                            </Button>
                        </Modal.Footer>
                    </Modal>
                </div>
            </div>
        </>
    );
}

export default NewNamespace;
