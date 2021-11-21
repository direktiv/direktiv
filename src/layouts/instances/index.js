import React from 'react';
import './style.css';
import ContentPanel, { ContentPanelBody, ContentPanelTitle, ContentPanelTitleIcon } from '../../components/content-panel';
import FlexBox from '../../components/flexbox';
import {VscFileCode} from 'react-icons/vsc';
import { BsDot } from 'react-icons/bs';
import HelpIcon from '../../components/help';

function InstancesPage(props) {
    return(
        <div style={{ paddingRight: "8px" }}>
            <ContentPanel>
                <ContentPanelTitle>
                    <ContentPanelTitleIcon>
                        <VscFileCode/>
                    </ContentPanelTitleIcon>
                    <FlexBox className="gap" style={{ alignItems: "center" }}>
                        <div>
                            Instances
                        </div>
                        <HelpIcon msg={"A list of instances run by workflows within the current active namespace."} />
                    </FlexBox>
                </ContentPanelTitle>
                <ContentPanelBody>
                    <InstancesTable />
                </ContentPanelBody>
            </ContentPanel>
        </div>
    );
}

export default InstancesPage;

function InstancesTable(props) {

    return(
        <table className="instances-table">
            <thead>
                <tr>
                    <th>
                        State
                    </th>
                    <th>
                        Name
                    </th>
                    <th>
                        Started at
                    </th>
                    <th>
                        Finished at
                    </th>
                </tr>
            </thead>
            <tbody>
                <InstanceRow state={success} name={"test-01"} started={""} finished={""}  />
                <InstanceRow state={fail} name={"test-02"} started={""} finished={""} />
                <InstanceRow state={cancelled} name={"test-03"} started={""} finished={""} />
                <InstanceRow state={running} name={"test-04"} started={""} finished={""} />
            </tbody>
        </table>
    );
}

const success = "success";
const fail = "fail";
const cancelled = "cancelled";
const running = "running";

function InstanceRow(props) {

    let {state, name, started, finished} = props;

    let label;
    if (state === success) {
        label = <SuccessState />
    } else if (state === fail) {
        label = <FailState />
    } else if (state === cancelled) {
        label = <CancelledState />
    } else  if (state === running) {
        label = <RunningState />
    }

    return(<tr className="instance-row">
        <td>
            {label}
        </td>
        <td>
            {name}
        </td>
        <td>
            {started}
        </td>
        <td>
            {finished}
        </td>
    </tr>)
}

function StateLabel(props) {

    let {className, label} = props;

    return (
        <div>
            <FlexBox className={className} style={{ alignItems: "center" }} >
                <BsDot style={{ height: "32px", width: "32px", marginRight: "-8px" }} />
                <div>{label}</div>
            </FlexBox>
        </div>
    )
}

function SuccessState() {
    return (
        <StateLabel className={"success-label"} label={"Successful"} />
    )
}

function FailState() {
    return (
        <StateLabel className={"fail-label"} label={"Failed"} />
    )
}

function CancelledState() {
    return (
        <StateLabel className={"cancel-label"} label={"Cancelled"} />
    )
}

function RunningState() {
    return (
        <StateLabel className={"running-label"} label={"Running"} />
    )
}

