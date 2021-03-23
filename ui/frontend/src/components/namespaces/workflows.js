import React, {useContext, useState} from "react";
import {useHistory} from "react-router-dom";

import {Button, OverlayTrigger, Tooltip} from "react-bootstrap";
import {Trash} from "react-bootstrap-icons";


import {DeleteWorkflow} from "components/workflow/workflow";
import DeleteModal from "components/workflow/delete-modal";
import ServerContext from "components/app/context";

export default function WorkflowList(props) {
    const {namespace, workflows, fetchWorkflows} = props;

    let list = [];
    for (let i = 0; i < workflows.length; i++) {
        list.push(
            <WorkflowListItem
                key={i}
                namespace={namespace}
                workflow={workflows[i]}
                fetchWorkflows={fetchWorkflows}
            />
        );
    }

    return <div id="workflows-list">{list}</div>;
}

export function WorkflowListItem(props) {
    const {namespace, workflow, fetchWorkflows} = props;
    const history = useHistory();
    const context = useContext(ServerContext);
    const [showDelete, setShowDelete] = useState(false);
    const renderTooltip = (props) => (
        <Tooltip id="button-tooltip" {...props}>
            Delete Workflow
        </Tooltip>
    );

    return (
        <div
            className="workflows-list-item"
            onClick={(ev) => {
                history.push(`/p/${namespace}/w/${workflow.id}`);
            }}
        >
            <DeleteModal
                onDelete={async () => {
                    await DeleteWorkflow(context, namespace, workflow.uid);
                    fetchWorkflows();
                }}
                show={showDelete}
                handleHide={() => {
                    setShowDelete(false);
                }}
            />
            <div style={{display: "flex", flex: "auto", width: "100%"}}>
                <div style={{flexGrow: "1", overflow: "hidden"}}>
                    <div>
                        <span style={{color: "#2396d8"}}>{workflow.id}</span>
                        <br/>
                        {workflow.description ? (
                            <div
                                style={{
                                    whiteSpace: "nowrap",
                                    overflow: "hidden",
                                    textOverflow: "ellipsis",
                                }}
                            >
                                {workflow.description}
                            </div>
                        ) : (
                            <div style={{color: "#999999", fontStyle: "italic"}}>
                                No Description
                            </div>
                        )}
                    </div>
                </div>
                <div>
                    <div
                        style={{
                            float: "right",
                            paddingTop: "5px",
                            marginLeft: "8px",
                            borderColor: "#6c757d",
                        }}
                        id="action-btn"
                        onClick={(ev) => {
                            ev.preventDefault();
                            ev.stopPropagation();
                        }}
                    >
                        <Button
                            style={{
                                marginLeft: "6px",
                                borderColor: "#e0e0e0",
                                textAlign: "center",
                                padding: "0.375rem 0.375rem 0.375rem 0.375rem",
                            }}
                            variant="danger"
                            onClick={() => {
                                setShowDelete(true);
                            }}
                        >
                            <OverlayTrigger
                                placement="top"
                                delay={{show: 100, hide: 150}}
                                overlay={renderTooltip}
                            >
                                <Trash width="1.5em" height="1.5em"/>
                            </OverlayTrigger>
                        </Button>
                    </div>
                </div>
            </div>
        </div>
    );
}
