import {useContext, useState} from 'react'
import {Button, Modal, OverlayTrigger, Tooltip} from "react-bootstrap"
import {Trash} from 'react-bootstrap-icons'
import ServerContext from "components/app/context";

const renderTooltip = (props) => (
    <Tooltip id="button-tooltip" {...props}>
        Delete Namespace
    </Tooltip>
);

export function DeleteNamespace(props) {
    const {namespace, fetchNamespaces} = props
    const [show, setShow] = useState(false)
    const [err, setErr] = useState("")
    const context = useContext(ServerContext)

    const handleShow = () => {
        setErr("")
        setShow(true)
    }
    const handleClose = () => {
        setErr("")
        setShow(false)
    }
    const handleSubmit = async () => {
        try {
            let resp = await fetch(`${context.SERVER_BIND}/namespaces/${namespace}`, {method: "DELETE"})
            if (resp.ok) {
                fetchNamespaces()
                handleClose()
            } else {
                setErr(await resp.text())
            }
        } catch (e) {
            setErr(e.message)
        }
    }

    return (
        <div style={{display: 'flex'}} onClick={(ev) => {
            ev.stopPropagation()
        }}>
            <OverlayTrigger
                placement="right"
                delay={{show: 100, hide: 150}}
                overlay={renderTooltip}
            >
                <Button
                    style={{
                        marginLeft: "6px",
                        background: "none",
                        border: "none",
                        textAlign: "right",
                        padding: "0px"
                    }}
                    variant="light"
                    onClick={handleShow}
                >
                    <Trash width="1.00em" height="1.00em"/>
                </Button>
            </OverlayTrigger>
            <Modal show={show} onHide={handleClose}>
                <Modal.Header closeButton>
                    <Modal.Title>
                        Delete Namespace
                    </Modal.Title>
                </Modal.Header>
                <Modal.Body>
                    Are you sure you want to delete {namespace}? This will remove all workflows associated with it.
                    {err !== "" ?
                        <div className="text-danger" style={{fontSize: "0.8em"}}>
                            *{err}
                        </div>
                        : ""
                    }
                </Modal.Body>
                <Modal.Footer>
                    <Button variant="secondary" onClick={handleClose}>
                        Close
                    </Button>
                    <Button variant="danger" onClick={handleSubmit}>
                        Delete
                    </Button>
                </Modal.Footer>
            </Modal>
        </div>

    )

}

export default DeleteNamespace