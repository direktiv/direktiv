import React, { useEffect, useState } from 'react'
import './style.css'
import { Config, copyTextToClipboard, GenerateRandomKey } from '../../util';
import Button from '../../components/button';
import { useParams } from 'react-router';
import ContentPanel, { ContentPanelBody, ContentPanelHeaderButton, ContentPanelHeaderButtonIcon, ContentPanelTitle, ContentPanelTitleIcon } from '../../components/content-panel';
import FlexBox from '../../components/flexbox';
import {useInstance, useInstanceLogs, useWorkflow} from 'direktiv-react-hooks';
import { CancelledState, FailState, InstancesTable, RunningState, SuccessState } from '../instances';

import { Link, useNavigate, useSearchParams } from 'react-router-dom';
import { AutoSizer, List, CellMeasurer, CellMeasurerCache } from 'react-virtualized';
import { VscCopy, VscEye, VscEyeClosed, VscSourceControl, VscScreenFull, VscTerminal } from 'react-icons/vsc';

import * as dayjs from "dayjs"
import YAML from 'js-yaml'

import DirektivEditor from '../../components/editor';
import WorkflowDiagram from '../../components/diagram';

import Modal, { ButtonDefinition } from '../../components/modal';
import Alert from '../../components/alert';
import Loader from '../../components/loader';

function InstancePageWrapper(props) {

    let {namespace} = props;
    if (!namespace) {
        return <></>
    }

    return <InstancePage namespace={namespace} />

}

export default InstancePageWrapper;

function TabbedButtons(props) {

    let {tabBtn, setTabBtn, setSearchParams} = props;

    let tabBtns = [];
    let tabBtnLabels = ["Flow Graph", "Child Instances"];

    for (let i = 0; i < tabBtnLabels.length; i++) {
        let key = GenerateRandomKey();
        let classes = "tab-btn";
        if (i === tabBtn) {
            classes += " active-tab-btn"
        }

        tabBtns.push(<FlexBox key={key} className={classes}>
            <div onClick={() => {
                setTabBtn(i)
                setSearchParams({
                    tab: i
                })
            }}>
                {tabBtnLabels[i]}
            </div>
        </FlexBox>)
    }

    return(
            <FlexBox className="tabbed-btns-container" style={{flexShrink:"1", flexGrow: "0"}}>
                <FlexBox className="tabbed-btns" >
                    {tabBtns}
                </FlexBox>
            </FlexBox>
    )
}

function InstancePage(props) {

    let {namespace} = props;
    const { id } = useParams();
    const [searchParams, setSearchParams] = useSearchParams()
    const navigate = useNavigate()

    const [load, setLoad] = useState(true)
    const [wfpath, setWFPath] = useState("")
    const [ref, setRef] = useState("")
    const [rev, setRev] = useState(null)
    const [follow, setFollow] = useState(true)
    const [width,] = useState(window.innerWidth);
    const [clipData, setClipData] = useState(null)
    const [instanceID, setInstanceID] = useState(id)
    const [tabBtn, setTabBtn] = useState(searchParams.get('tab') !== null ? parseInt(searchParams.get('tab')): 0);


    // let instanceID = params["id"];
    React.useEffect(() => {
        setLoad(true)
        setInstanceID(id)
    }, [id]);

    let {data, err, workflow, latestRevision, getInput, getOutput, cancelInstance} = useInstance(Config.url, true, namespace, instanceID, localStorage.getItem("apikey"));

    useEffect(()=>{
        if(load && data !== null && workflow != null && latestRevision != null) {
            let split = data.as.split(":")
            setWFPath(split[0])
            if (workflow.revision === latestRevision) {
                setRef("latest")
            } else if(split[1] !== "latest"){
                setRef(split[1])
            }
            setRev(workflow.revision)
            setLoad(false)
        }

    },[load, data, workflow, latestRevision])

    useEffect(()=>{      
        if(err === "Not Found" || (err !== null && err.indexOf("invalid UUID") >= 0)) {
            navigate(`/not-found`)
        }
    },[err, navigate])

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
    } else if (data.status === "cancelled" || data.errorCode === "direktiv.cancels.api") {
        label = <CancelledState />
    } else if (data.status === "failed" || data.status === "crashed") {
        label = <FailState />
    }  else  if (data.status === "running") {
        label = <RunningState />
    }

    let wfName = data.as.split(":")[0]

    return (
    <Loader load={load} timer={3000}>
        <FlexBox className="col gap" style={{paddingRight: "8px"}}>
            <FlexBox className="gap wrap" style={{minHeight: "50%", flex: "1"}}>
                <FlexBox style={{minWidth: "340px", flex: "5", }}>
                    <ContentPanel style={{width: "100%", minHeight: "40vh"}}>
                        <ContentPanelTitle>
                            <ContentPanelTitleIcon>
                                <VscTerminal />
                            </ContentPanelTitleIcon>
                            <FlexBox className="gap" style={{alignItems:"center"}}>
                                <div>
                                    Instance Details
                                </div>
                                {label} 
                                <FlexBox style={{flex: "auto", justifyContent: "right", paddingRight: "6px", alignItems: "center"}}>
                                    { data.status === "running" || data.status === "pending" ? 
                                    <Button className="small light" style={{marginRight: "8px"}} onClick={() => {
                                        cancelInstance()
                                        setLoad(true)
                                    }}>
                                        <span className="red-text">
                                            Cancel
                                        </span>
                                    </Button>
                                    :<></>}
                                    {rev === null || rev === ""
                                    ?
                                    <>
                                    </>
                                    :
                                    <Link to={`/n/${namespace}/explorer/${wfName}?${ref==="latest" ? `tab=2` : `tab=1&revision=${rev}&revtab=0` }`}>
                                        <Button className="small light">
                                            <span className="hide-on-small">View</span> Workflow
                                        </Button>
                                    </Link>
                                    }
                                    <Modal
                                    escapeToCancel
                                    activeOverlay
                                    maximised
                                    noPadding
                                    title="Instance Details"
                                    titleIcon={
                                        <VscTerminal />
                                    }
                                    style={{
                                        maxWidth: "50px"
                                    }}
                                    modalStyle={{
                                        overflow: "hidden",
                                        padding: "0px"
                                    }}
                                    button={(
                                        <ContentPanelHeaderButton hackyStyle={{ marginBottom: "8px", height: "29px" }}>
                                            <ContentPanelHeaderButtonIcon>
                                                <VscScreenFull />
                                            </ContentPanelHeaderButtonIcon>
                                        </ContentPanelHeaderButton>
                                    )}
                                    actionButtons={[
                                        ButtonDefinition("Close", () => {}, "small light", ()=>{}, true, false)
                                    ]}
                                >
                                    <InstanceLogs clipData={clipData} noPadding namespace={namespace} instanceID={instanceID} follow={follow} setFollow={setFollow} width={width}/>
                                </Modal>
                                </FlexBox>
                            </FlexBox>
                        </ContentPanelTitle>
                        <InstanceLogs setClipData={setClipData} clipData={clipData} namespace={namespace} instanceID={instanceID} follow={follow} setFollow={setFollow} width={width} />
                    </ContentPanel>
                </FlexBox>
                <FlexBox className="gap wrap" style={{minIoCopyHeight: "40%", minWidth: "300px", flex: "2", flexWrap: "wrap-reverse"}}>
                    <FlexBox style={{minWidth: "300px"}}>
                        <ContentPanel style={{width: "100%", minHeight: "40vh"}}>
                        <ContentPanelTitle>
                            <ContentPanelTitleIcon>
                                <VscTerminal />
                            </ContentPanelTitleIcon>
                            <FlexBox className="gap">
                                <div>
                                Input
                                </div>
                            </FlexBox>
                            <Modal
                                escapeToCancel
                                activeOverlay                                
                                maximised
                                noPadding
                                title="Input"
                                titleIcon={
                                    <VscTerminal />
                                }
                                modalStyle={{
                                    overflow: "hidden",
                                    padding: "0px"
                                }}
                                button={(
                                    <ContentPanelHeaderButton>
                                        <ContentPanelHeaderButtonIcon>
                                            <VscScreenFull />
                                        </ContentPanelHeaderButtonIcon>
                                    </ContentPanelHeaderButton>
                                )}
                                actionButtons={[
                                    ButtonDefinition("Close", () => {}, "small light", ()=>{}, true, false)
                                ]}
                            >
                                <Input getInput={getInput}/>
                            </Modal>
                        </ContentPanelTitle>
                        <ContentPanelBody>
                            <Input getInput={getInput}/>
                        </ContentPanelBody>
                    </ContentPanel>
                    </FlexBox>
                </FlexBox>
            </FlexBox>
            <FlexBox className="gap wrap" style={{minHeight: "40%", flex: "1"}}>
                <FlexBox style={{minWidth: "300px", flex: "5"}}>
                    <ContentPanel style={{width: "100%", minHeight: "40vh"}}>
                        <ContentPanelTitle>
                            <ContentPanelTitleIcon>
                                <VscSourceControl />
                            </ContentPanelTitleIcon>
                            <FlexBox className="gap" style={{alignItems:"center"}}>
                                <div style={{flex: "1", whiteSpace: "nowrap"}}>
                                    {`${tabBtn === 0 ? "Flow Graph" : "Child Instances"}`}
                                </div>
                                {tabBtn === 1 && data.invoker.startsWith("instance:") ?
                                    <Link to={`/n/${namespace}/instances/${data.invoker.replace("instance:", "")}`} reloadDocument>
                                    <Button className="small light">
                                        <span className="hide-on-small">View</span> Parent
                                    </Button>
                                    </Link>
                                    :
                                    <></>
                                }
                                <TabbedButtons setSearchParams={setSearchParams} searchParams={searchParams} tabBtn={tabBtn} setTabBtn={setTabBtn} />
                            </FlexBox>
                        </ContentPanelTitle>
                        {tabBtn === 0 ?<ContentPanelBody><InstanceDiagram status={data.status} namespace={namespace} wfpath={wfpath} rev={rev} instRef={ref} flow={data.flow}/></ContentPanelBody>:<></>}
                        {tabBtn === 1 ?<InstancesTable placeholder={"No child instances have executed from this instance. Child instances will appear here."} namespace={namespace} mini={true} hideTitle={true} panelStyle={{border: "unset"}} filter={[`filter.field=TRIGGER&filter.type=MATCH&filter.val=instance:${instanceID}`]}/>:<></>}
                    </ContentPanel>
                </FlexBox>
                <FlexBox style={{minWidth: "300px", flex: "2"}}>
                    <ContentPanel style={{width: "100%", minHeight: "40vh"}}>
                        <ContentPanelTitle>
                            <ContentPanelTitleIcon>
                                <VscTerminal />
                            </ContentPanelTitleIcon>
                            <FlexBox className="gap">
                                <div>
                                Output
                                </div>
                            </FlexBox>
                            <Modal
                                escapeToCancel
                                activeOverlay
                                maximised
                                noPadding
                                title="Output"
                                titleIcon={
                                    <VscTerminal />
                                }
                                modalStyle={{
                                    overflow: "hidden",
                                    padding: "0px"
                                }}
                                button={(
                                    <ContentPanelHeaderButton>
                                        <ContentPanelHeaderButtonIcon>
                                            <VscScreenFull />
                                        </ContentPanelHeaderButtonIcon>
                                    </ContentPanelHeaderButton>
                                )}
                                actionButtons={[
                                    ButtonDefinition("Close", () => {}, "small light", ()=>{}, true, false)
                                ]}
                            >
                                <Output getOutput={getOutput} status={data.status}/>
                            </Modal>
                        </ContentPanelTitle>
                        <ContentPanelBody>
                            <Output getOutput={getOutput} status={data.status}/>
                        </ContentPanelBody>
                    </ContentPanel>
                </FlexBox>
            </FlexBox>
        </FlexBox>
    </Loader>)
}

function InstanceLogs(props) {

    let {noPadding, namespace, setClipData, instanceID, follow, setFollow, width, clipData} = props;
    let paddingStyle = { padding: "12px" }
    if (noPadding) {
        paddingStyle = { padding: "0px" }
    }

    return (
        <>
            <FlexBox className="col" style={{...paddingStyle}}>
                <FlexBox className={"logs"}>
                    <Logs clipData={clipData} setClipData={setClipData} namespace={namespace} instanceID={instanceID} follow={follow} setFollow={setFollow} />
                </FlexBox>
                <div className={"logs-footer"} style={{  alignItems:'center', borderRadius: " 0px 0px 8px 8px", overflow: "hidden" }}>
                    <FlexBox className="gap" style={{width: "100%", flexDirection: "row-reverse", height: "100%", alignItems: "center"}}>
                        <TerminalButton onClick={()=>{
                            copyTextToClipboard(clipData)
                        }}>
                                <VscCopy/> Copy {width > 999 ? <span>to Clipboard</span>:""}
                        </TerminalButton>
                        {follow ?
                            <TerminalButton onClick={(e)=>setFollow(!follow)} className={"btn-terminal"}>
                                <VscEyeClosed/> Stop {width > 999 ? <span>watching</span>: ""}
                            </TerminalButton>
                            :
                            <TerminalButton onClick={(e)=>setFollow(!follow)} className={"btn-terminal"} >
                                <VscEye/> <div>Follow {width > 999 ? <span>logs</span>: ""}</div>
                            </TerminalButton>
                        }
                    </FlexBox>
                </div>
            </FlexBox>
        </>
    )
}

export function TerminalButton(props) {

    let {children, onClick, className} = props;
    return (
        <div onClick={onClick} className={`btn-terminal ${className}`} style={{
            maxHeight: "22px"
        }}>
            <FlexBox className="gap" style={{ alignItems: "center", userSelect: "none" }}>
                {children}
            </FlexBox>
        </div>
    )

}

function InstanceDiagram(props) {
    const {wfpath, rev, instRef, flow, namespace, status} = props

    const [load, setLoad] = useState(true)
    const [workflowMissing, setWorkflowMissing] = useState(false)
    const [wfdata, setWFData] = useState("")

    const {getWorkflowRevisionData} = useWorkflow(Config.url, false, namespace, wfpath, localStorage.getItem("apikey"))

    useEffect(()=>{
        async function getwf() {
            if(wfpath !== "" && instRef !== "" && rev !== null && rev !== "" && load){
                let refWF = await getWorkflowRevisionData(instRef === "latest" ? instRef : rev)
                setWFData(atob(refWF.revision.source))
                setLoad(false)
            } else if(rev === ""){
                setWorkflowMissing(true)
            }
        }
        
        getwf()
    },[wfpath, rev, load, instRef, getWorkflowRevisionData])

    if (workflowMissing) {
        return  (
            <div className='container-alert'>
                Workflow revision that executed instance no longer exists
            </div>
            )
    }

    if(load){
        return <></>
    }
    
    return(
        <WorkflowDiagram instanceStatus={status} disabled={true} flow={flow} workflow={YAML.load(wfdata)}/>
    )
}

function Input(props) {
    const {getInput} = props
    
    const [input, setInput] = useState(null)
    const [load, setLoad] = useState(true)

    useEffect(()=>{
        async function get() {
            let data = await getInput()
            try {
                let x = JSON.stringify(JSON.parse(data),null,2)
                setInput(x)
            } catch(e) {
                setInput(data)
            }
        }

        if (load && input === null && getInput) {
            setLoad(false)
            get()
        }
    },[input, getInput, load])

    return(
        <FlexBox style={{flexDirection:"column"}}>
            {!input ? 
            <Alert className="instance-input-banner">No input data was provided</Alert> : null}
            <FlexBox style={{overflow: "hidden"}}>
                {/* <div style={{width: "100%", height: "100%"}}> */}
                    <AutoSizer>
                        {({height, width})=>(
                            <DirektivEditor height={height} width={width} dlang="json" value={input} readonly={true}/>
                        )}
                    </AutoSizer>
                {/* </div> */}
            </FlexBox>
        </FlexBox>
    )
}

function Output(props){
    const {getOutput, status} = props

    const [load, setLoad] = useState(true)
    const [output, setOutput] = useState(null)

    useEffect(()=>{
        async function get() {
            if (load && status !== "pending" && output === null && getOutput){
                setLoad(false)
                try {
                    let data = await getOutput()
                    let x = JSON.stringify(JSON.parse(data),null,2)
                    setOutput(x)
                } catch(e) {
                    console.log(e);
                }
            }
        }
        get()
    },[output, load, getOutput, status])

    return(
        <FlexBox style={{flexDirection:"column"}}>
        {!output ? 
            <Alert className="instance-input-banner">No output data was resolved</Alert> : null}
        <FlexBox style={{padding: "0px", overflow: "hidden"}}>
            <AutoSizer>
                {({height, width})=>(
                    <DirektivEditor disableCursor height={height} width={width} dlang="json" value={output} readonly={true}/>
                )}
            </AutoSizer>
        </FlexBox>
        </FlexBox>
    )
}


function Logs(props){ 
    const cache = new CellMeasurerCache({
        fixedWidth: true,
        fixedHeight: false
    })

    let {namespace, instanceID, follow, setClipData, clipData} = props;
    const [logLength, setLogLength] = useState(0)
    let {data, err} = useInstanceLogs(Config.url, true, namespace, instanceID, localStorage.getItem("apikey"))
    useEffect(()=>{
        if (!setClipData) {
            // Skip ClipData if unset
            return 
        }

        if(data !== null) {
            if(clipData === null || logLength === 0) {

                let cd = ""
                for(let i=0; i < data.length; i++) {
                    cd += `[${dayjs.utc(data[i].node.t).local().format("HH:mm:ss.SSS")}] ${data[i].node.msg}\n`
                }
                setClipData(cd)
                setLogLength(data.length)
            } else if (data.length !== logLength) {
                let cd = clipData
                for(let i=logLength-1; i < data.length; i++) {
                    cd += `[${dayjs.utc(data[i].node.t).local().format("HH:mm:ss.SSS")}] ${data[i].node.msg}\n`

                }
                setClipData(cd)
                setLogLength(data.length)
            }
        }
    },[data, clipData, setClipData, logLength])


    if (!data) {
        return <></>
    }

    if (err) {
        return <></> // TODO 
    }

    function rowRenderer({index, parent, key, style}) {
        if(!data[index]){
            return ""
        }

        return (
        <CellMeasurer
            key={key}
            cache={cache}
            parent={parent}
            columnIndex={0}
            rowIndex={index}
        >
          <div style={{...style, minWidth:"800px", width:"800px"}}>
            <div style={{display:"inline-block",minWidth:"112px", color:"#b5b5b5"}}>
                <div className="log-timestamp">
                    <div>[</div>
                        <div style={{display: "flex", flex: "auto", justifyContent: "center"}}>{dayjs.utc(data[index].node.t).local().format("HH:mm:ss.SSS")}</div>
                    <div>]</div>
                </div>
            </div> 
            <span style={{marginLeft:"5px", whiteSpace:"pre-wrap"}}>
                {data[index].node.msg}
            </span>
            <div style={{height: `fit-content`}}></div>
          </div>
          </CellMeasurer>
        );
    }
      

    return (
        <div style={{ flex: "1 1 auto", lineHeight: "20px" }}>
            <AutoSizer>
                {({ height, width }) => (
                    <div style={{ height: "100%", minHeight: "100%" }}>
                        <List
                            width={width}
                            height={height}
                            rowRenderer={rowRenderer}
                            deferredMeasurementCache={cache}
                            scrollToIndex={follow ? data.length - 1 : 0}
                            rowCount={data.length}
                            rowHeight={cache.rowHeight}
                            scrollToAlignment={"start"}
                        />
                    </div>
                )}
            </AutoSizer>
        </div>
    )
}