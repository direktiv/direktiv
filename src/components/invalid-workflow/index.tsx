import React from 'react';
import FlexBox from '../flexbox';

export interface InvalidWorkflowProps {
    /**
    * Invalid workflow error message
    */
    invalidWorkflow?: string | null;
}

/**
* UI Component card used for displaying error messages thrown while interfacing with invalid workflow.
* Is only rendered if invalidWorkflow is not null.
*/
function InvalidWorkflow({ invalidWorkflow }: InvalidWorkflowProps) {
    return (
        <>
            {invalidWorkflow ?
                <FlexBox col center="y" style={{ padding: "0px 50px" }}>
                    <h3 style={{ marginBottom: "0px" }}>Invalid Workflow</h3>
                    <pre style={{ whiteSpace: "break-spaces" }}>
                        {invalidWorkflow}
                    </pre>
                </FlexBox>
                : <></>}
        </>
    );
}

export default InvalidWorkflow;