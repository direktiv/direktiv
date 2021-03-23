import React from 'react'

import Button from "react-bootstrap/Button";

import {PlayFill} from "react-bootstrap-icons";
import {OverlayTrigger, Tooltip} from "react-bootstrap";

export default function WorkflowActions(props) {
    return (
        <div id="actions-container">
            <RunButton props={props}/>
        </div>
    );
}

function RunButton(props) {
    let {onRun, active, startType} = props.props;

    const renderTooltipStart = (props) => (
        <Tooltip id="button-tooltip" {...props}>
            Only workflows with start type "default" can be directly invoked.
        </Tooltip>
    );

    const renderTooltipDisabled = (props) => (
        <Tooltip id="button-tooltip" {...props}>
            Disabled workflows cannot be invoked.
        </Tooltip>
    );

    if (!active) {
        return (
            <OverlayTrigger
                placement="top"
                delay={{show: 100, hide: 150}}
                overlay={renderTooltipDisabled}
            >
                <Button
                    style={{marginLeft: "6px"}}
                    variant="success"
                    onClick={onRun}
                    disabled={true}
                >
                    <PlayFill/> Run
                </Button>
            </OverlayTrigger>
        );
    }

    if (startType != "default") {
        return (
            <OverlayTrigger
                placement="top"
                delay={{show: 100, hide: 150}}
                overlay={renderTooltipStart}
            >
                <Button
                    style={{marginLeft: "6px"}}
                    variant="success"
                    onClick={onRun}
                    disabled={true}
                >
                    <PlayFill/> Run
                </Button>
            </OverlayTrigger>
        );
    }

    return (
        <Button style={{marginLeft: "6px"}} variant="success" onClick={onRun}>
            <PlayFill/> Run
        </Button>
    );
}
