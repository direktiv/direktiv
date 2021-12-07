import React, { useState } from 'react'
import './style.css'
import { Config, copyTextToClipboard } from '../../util';
import Button from '../../components/button';
import { useParams } from 'react-router';
import ContentPanel, { ContentPanelBody, ContentPanelTitle, ContentPanelTitleIcon } from '../../components/content-panel';
import FlexBox from '../../components/flexbox';
import {AiFillCode} from 'react-icons/ai';
import {useInstance, useInstanceLogs} from 'direktiv-react-hooks';
import { FailState, RunningState, SuccessState } from '../instances';
import { Link } from 'react-router-dom';
import { AutoSizer, List } from 'react-virtualized';
import { IoCopy, IoEye, IoEyeOff } from 'react-icons/io5';

function InstancePageWrapper(props) {

    let {namespace} = props;
    if (!namespace) {
        return <></>
    }

    return <InstancePage namespace={namespace} />

}

export default InstancePageWrapper;

function InstancePage(props) {

    let {namespace} = props;
    const [follow, setFollow] = useState(true)
    const [width, setWidth] = useState(window.innerWidth);
    const [clipData, setClipData] = useState(null)

    const params = useParams()
    let instanceID = params["id"];

    let {data, err, cancelInstance, getInput, getOutput} = useInstance(Config.url, true, namespace, instanceID);
    
    if (data === null) {
        return <></>
    }

    if (err !== null) {
        // TODO
        return <></>
    }

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

    let linkURL = `/n/${namespace}/explorer/${wfName}?tab=2`;
    if (revName) {
        if(revName !== "latest"){
            linkURL = `/n/${namespace}/explorer/${wfName}?tab=1&revision=${revName}&revtab=0`;
        }
    }

    

    return (<>
        <FlexBox className="col gap" style={{paddingRight: "8px"}}>
            <FlexBox className="gap wrap">
                <FlexBox style={{minWidth: "340px", flex: "5"}}>
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
                                <FlexBox style={{flex: "auto", justifyContent: "right", paddingRight: "6px"}}>
                                    <Link to={linkURL}>
                                        <Button className="small light">
                                            <span className="hide-on-small">View</span> Workflow
                                        </Button>
                                    </Link>
                                </FlexBox>
                            </FlexBox>
                        </ContentPanelTitle>
                        <ContentPanelBody>
                            <FlexBox className="col">
                                <FlexBox style={{flexGrow:1, backgroundColor:"#223848", color:"white"}}>
                                    <Logs namespace={namespace} instanceID={instanceID} follow={follow} setFollow={setFollow} />
                                </FlexBox>
                                {/* <div style={{padding: "4px", flexWrap:"wrap", gap: "8px"}}>
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
                                </div> :<></>} */}
                            <FlexBox style={{height:"40px",backgroundColor:"#223848", color:"white", maxHeight:"40px", paddingRight:"10px", paddingLeft:"10px", boxShadow:"0px 0px 3px 0px #fcfdfe", alignItems:'center'}}>
                                <FlexBox className="gap" style={{justifyContent:"flex-end"}}>
                                    {follow ? 
                                        <div onClick={(e)=>setFollow(!follow)} style={{display:"flex", alignItems:"center", gap:"3px",backgroundColor:"#355166", paddingTop:"3px", paddingBottom:"3px", paddingLeft:"6px", paddingRight:"6px", cursor:"pointer", borderRadius:"3px"}}>
                                            <IoEyeOff/> Stop {width > 999 ? <span>watching</span>: ""}
                                        </div>
                                        :
                                        <div onClick={(e)=>setFollow(!follow)} style={{display:"flex", alignItems:"center", gap:"3px",backgroundColor:"#355166", paddingTop:"3px", paddingBottom:"3px", paddingLeft:"6px", paddingRight:"6px", cursor:"pointer", borderRadius:"3px"}}>
                                            <IoEye/> Follow {width > 999 ? <span>logs</span>: ""}
                                        </div>
                                    }
                                    <div onClick={()=>{
                                        copyTextToClipboard(clipData)
                                    }} style={{display:"flex", alignItems:"center", gap:"3px", backgroundColor:"#355166",paddingTop:"3px", paddingBottom:"3px",  paddingLeft:"6px", paddingRight:"6px", cursor:"pointer", borderRadius:"3px"}}>
                                        <IoCopy/> Copy {width > 999 ? <span>to Clipboard</span>:""}
                                    </div>
                                </FlexBox>
                                </FlexBox>
                            </FlexBox>
                        </ContentPanelBody>
                    </ContentPanel>
                </FlexBox>
                <FlexBox className="gap wrap" style={{minWidth: "300px", flex: "2", flexWrap: "wrap-reverse"}}>
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
                </FlexBox>
            </FlexBox>
            <FlexBox className="gap wrap">
                <FlexBox style={{minWidth: "300px", flex: "5"}}>
                    <ContentPanel style={{width: "100%"}}>
                        <ContentPanelTitle>
                            <ContentPanelTitleIcon>
                                <AiFillCode />
                            </ContentPanelTitleIcon>
                            <FlexBox className="gap" style={{alignItems:"center"}}>
                                <div>
                                Logical Flow Graph
                                </div>
                            </FlexBox>
                        </ContentPanelTitle>
                        <ContentPanelBody>
                        </ContentPanelBody>
                    </ContentPanel>
                </FlexBox>
                <FlexBox style={{minWidth: "300px", flex: "2"}}>
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

function Logs(props){ 

    let {namespace, instanceID, follow} = props;
    let {data, err} = useInstanceLogs(Config.url, true, namespace, instanceID)
    console.log(data, err)
    if (!data) {
        return <></>
    }

    if (err) {
        return <></> // TODO 
    }

    function rowRenderer({index, key, style}) {
        console.log(index, key, style)
        return (
          <div key={key} style={style}>
            {data[index].node.msg}
          </div>
        );
    }
      

    return(
        <div style={{flex:"1 1 auto", paddingLeft:'10px'}}>
            <AutoSizer>
                {({height, width})=>(
                    <List
                        width={width}
                        height={height}
                        rowRenderer={rowRenderer}
                        scrollToIndex={follow ? data.length - 1: 0}
                        rowCount={data.length}
                        rowHeight={20}
                    />
                )}
            </AutoSizer>
        </div>
    )
}