import React from "react";


import Modal from "react-bootstrap/Modal"
import Button from "react-bootstrap/Button";


export default function DeleteModal(props) {
    let {onDelete, handleHide, show} = props;

    const handleDelete = () => {
        onDelete()
        handleHide()
    };

    return (
        <div onClick={(ev) => ev.stopPropagation()}>
            <Modal show={show} onHide={handleHide}>
                <Modal.Header closeButton>
                    <Modal.Title>Delete Confirmation</Modal.Title>
                </Modal.Header>
                <Modal.Body>Are you sure you want to delete this Workflow?</Modal.Body>
                <Modal.Footer>
                    <Button variant="secondary" onClick={() => {
                        handleHide()
                    }}>
                        Cancel
                    </Button>
                    <Button variant="danger" onClick={() => {
                        handleDelete()
                    }}>
                        Delete Workflow
                    </Button>
                </Modal.Footer>
            </Modal>
        </div>

    )
}