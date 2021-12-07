import React from 'react'
import './style.css'
import { Config } from '../../util';
import { useParams } from 'react-router';
import ContentPanel, { ContentPanelBody, ContentPanelTitle, ContentPanelTitleIcon } from '../../components/content-panel';
import FlexBox from '../../components/flexbox';
import {AiFillCode} from 'react-icons/ai';
import {useInstance} from 'direktiv-react-hooks';
import { FailState, RunningState, SuccessState } from '../instances';
import { Link } from 'react-router-dom';

function InstancePage(props) {

    const params = useParams()

    let {namespace} = props;
    let instanceID = params["id"];

    let {data, err, getInput, getOutput, getInstance} = useInstance(Config.url, true, namespace, instanceID);
    if (data === null) {
        return <></>
    }

    if (err !== null) {
        // TODO
        return <></>
    }

    console.log(data);
    let label = <></>;
    if (data.status === "complete") {
        label = <SuccessState />
    } else if (data.status === "failed") {
        label = <FailState />
    }  else  if (data.status === "running") {
        label = <RunningState />
    }

    let wfName = data.as.split(":")[0]
    let revName = data.as.split(":")[1]

    let linkURL = `/n/${namespace}/explorer/${wfName}`;
    if (revName) {
        linkURL = `/n/${namespace}/explorer/${wfName}?tab=1&revision=${revName}&revtab=0`;
    }

    return (<>

        <FlexBox className="col gap" style={{paddingRight: "8px"}}>
            <FlexBox>
                <ContentPanel style={{width: "100%"}}>
                    <ContentPanelTitle>
                        <ContentPanelTitleIcon>
                            <AiFillCode />
                        </ContentPanelTitleIcon>
                        <FlexBox className="gap" style={{alignItems:"center"}}>
                            <div>
                            Instance Details
                            </div>
                            {label} 
                        </FlexBox>
                    </ContentPanelTitle>
                    <ContentPanelBody>
                        <FlexBox className="col gap">
                            <div style={{padding: "4px", display: "flex", flexWrap:"wrap", gap: "8px"}}>
                                <FlexBox className="wrap gap">
                                    <InstanceTuple label={"Workflow"} value={data.as} linkTo={linkURL} />
                                    <InstanceTuple label={"ID"} value={data.id} />
                                    <InstanceTuple label={"Updated at"} value={data.updatedAt} />
                                    <InstanceTuple label={"Created at"} value={data.createdAt} />
                                </FlexBox>
                            </div>
                            { data.status === "failed" ? 
                            <div>
                                <FlexBox className="wrap gap">
                                    { data.errorCode ? <InstanceTuple label={"Error code"} value={data.errorCode} /> :<></>}
                                    { data.errorMessage ? <InstanceTuple label={"Error message"} value={data.errorMessage} /> :<></>}
                                    <InstanceTuple label={""} value={""} /><InstanceTuple label={""} value={""} />
                                </FlexBox>
                            </div> :<></>}
                            <FlexBox>
                                <div style={{width: "100%", minWidth: "100%", height: "100%", minHeight: "100%", backgroundColor: "#0e0e0e"}}>

                                </div>
                            </FlexBox>
                        </FlexBox>
                    </ContentPanelBody>
                </ContentPanel>
            </FlexBox>
            <FlexBox className="gap wrap">
                <FlexBox style={{minWidth: "300px"}}>
                    <ContentPanel style={{width: "100%"}}>
                        <ContentPanelTitle>
                            <ContentPanelTitleIcon>
                                <AiFillCode />
                            </ContentPanelTitleIcon>
                            <FlexBox className="gap">
                                <div>
                                Input Data
                                </div>
                            </FlexBox>
                        </ContentPanelTitle>
                        <ContentPanelBody>
                        </ContentPanelBody>
                    </ContentPanel>
                </FlexBox>
                <FlexBox style={{minWidth: "300px"}}>
                    <ContentPanel style={{width: "100%"}}>
                        <ContentPanelTitle>
                            <ContentPanelTitleIcon>
                                <AiFillCode />
                            </ContentPanelTitleIcon>
                            <FlexBox className="gap">
                                <div>
                                Output
                                </div>
                            </FlexBox>
                        </ContentPanelTitle>
                        <ContentPanelBody>
                        </ContentPanelBody>
                    </ContentPanel>
                </FlexBox>
            </FlexBox>
        </FlexBox>

    </>)
}

export default InstancePage;

function InstanceTuple(props) {
    
    let {label, value, linkTo} = props;

    let x = value;
    if (linkTo) {
        x = (
            <Link to={linkTo}>{value}</Link>
        )
    }

    return (<>
        <FlexBox className="instance-details-tuple col" style={{minWidth: "150px", flex: "1"}}>
            <div>
                <b>{label}</b>
            </div>
            <div title={value} style={{fontSize: "12px", overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap"}}>
                {x}
            </div>
        </FlexBox>
    </>)
}