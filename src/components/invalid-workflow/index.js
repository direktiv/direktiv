import React from 'react';
import FlexBox from '../flexbox';

function InvalidWorkflow(props) {

    let { invalidWorkflow } = props;

    return (
        <>
            {invalidWorkflow ?
                <FlexBox className="col center-y" style={{ padding: "0px 50px" }}>
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