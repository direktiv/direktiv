import React, { useCallback, useEffect, useRef, useState } from 'react';
import './style.css';
import FlexBox from '../../../components/flexbox';
import {useSearchParams} from 'react-router-dom'
import ContentPanel, { ContentPanelBody, ContentPanelTitle, ContentPanelTitleIcon } from '../../../components/content-panel';
import {BsCodeSquare} from 'react-icons/bs'
import { useNamespaceDependencies, useWorkflow, useWorkflowServices } from 'direktiv-react-hooks';
import { Config } from '../../../util';
import { useNavigate, useParams } from 'react-router';
import {  GenerateRandomKey } from '../../../util';

import * as dayjs from "dayjs"
import relativeTime from "dayjs/plugin/relativeTime";
import utc from "dayjs/plugin/utc"
import { InstanceRow } from '../../instances';
import { IoMdLock } from 'react-icons/io';
import {IoCloseCircleOutline, IoCheckmarkCircleOutline} from 'react-icons/io5'
import {VscChevronDown, VscChevronUp} from 'react-icons/vsc'
import { Service } from '../../namespace-services';
import DirektivEditor from '../../../components/editor';
import AddWorkflowVariablePanel from './variables';
import { RevisionSelectorTab, TabbedButtons } from './revisionTab';
import DependencyDiagram from '../../../components/dependency-diagram';
import YAML from 'js-yaml'
import WorkflowDiagram from '../../../components/diagram';

import Slider from 'rc-slider';
import 'rc-slider/assets/index.css';
import Button from '../../../components/button';
import Modal, { ButtonDefinition } from '../../../components/modal';

import SankeyDiagram from '../../../components/sankey';
import {PieChart} from 'react-minimal-pie-chart'
import HelpIcon from "../../../components/help";
import Loader from '../../../components/loader';
import Alert from '../../../components/alert';
import {AutoSizer} from "react-virtualized";

dayjs.extend(utc)
dayjs.extend(relativeTime);


function WorkflowPage(props) {
    const {namespace} = props
    const [searchParams, setSearchParams] = useSearchParams()
    const params = useParams()
    const [load, setLoad] = useState(true);

    // set tab query param on load
    useEffect(()=>{
        if(searchParams.get('tab') === null) {
            setSearchParams({tab: 0}, {replace:true})
        }

        setLoad(false)
    },[searchParams, setSearchParams, setLoad])
    
    let filepath = "/"

    if(!namespace) {
        return <> </>
    }

    if(params["*"] !== undefined){
        filepath = `/${params["*"]}`
    }

    return(
        <Loader load={load} timer={3000}>
            <InitialWorkflowHook setSearchParams={setSearchParams} searchParams={searchParams} namespace={namespace} filepath={filepath}/>
        </Loader>
    )
}

function InitialWorkflowHook(props){
    const {namespace, filepath, searchParams, setSearchParams} = props

    const [activeTab, setActiveTab] = useState(searchParams.get("tab") !== null ? parseInt(searchParams.get('tab')): 0)

    useEffect(()=>{
        setActiveTab(searchParams.get("tab") !== null ? parseInt(searchParams.get('tab')): 0)
    }, [searchParams])
    // todo handle err from hook below
    const {data,  getSuccessFailedMetrics, tagWorkflow, addAttributes, deleteAttributes, setWorkflowLogToEvent, editWorkflowRouter, getWorkflowSankeyMetrics, getWorkflowRevisionData, getWorkflowRouter, toggleWorkflow, executeWorkflow, getInstancesForWorkflow, getRevisions, getTags, deleteRevision, saveWorkflow, updateWorkflow, discardWorkflow, removeTag} = useWorkflow(Config.url, true, namespace, filepath.substring(1), localStorage.getItem("apikey"))
    const [router, setRouter] = useState(null)


    const [revisions, setRevisions] = useState(null)
    // todo handle revsErr
    const [, setRevsErr] = useState("")

    // fetch revisions using the workflow hook from above
    useEffect(()=>{
        async function listData() {
            if(revisions === null){
                // get workflow revisions
                let resp = await getRevisions()
                if(Array.isArray(resp.edges)){
                    setRevisions(resp.edges)
                } else {
                    setRevsErr(resp)
                }
            }
        }
        listData()
    },[getRevisions, revisions])

    useEffect(()=>{
        async function getD() {
            if(data !== null && router === null) {
                setRouter(await getWorkflowRouter())
            }
        }
        getD()
    },[router, data, getWorkflowRouter])

    if(data === null || router === null) {
        return <></>
    }
    return(
        <>
            <FlexBox id="workflow-page" className="gap col" style={{paddingRight: "8px"}}>
                <TabBar setRouter={setRouter} router={router} getWorkflowRouter={getWorkflowRouter} toggleWorkflow={toggleWorkflow}  setSearchParams={setSearchParams} activeTab={activeTab} setActiveTab={setActiveTab} />
                <FlexBox className="col gap">
                    { activeTab === 0 ? 
                        <OverviewTab getSuccessFailedMetrics={getSuccessFailedMetrics} router={router} namespace={namespace} getInstancesForWorkflow={getInstancesForWorkflow} filepath={filepath}/>
                    :<></>}
                    { activeTab === 1 ?
                        <>
                        <RevisionSelectorTab 
                        workflowName={filepath.substring(1)}
                        tagWorkflow={tagWorkflow}
                         namespace={namespace}
                          filepath={filepath} updateWorkflow={updateWorkflow} setRouter={setRouter} editWorkflowRouter={editWorkflowRouter} getWorkflowRouter={getWorkflowRouter} setRevisions={setRevisions} revisions={revisions} router={router} getWorkflowSankeyMetrics={getWorkflowSankeyMetrics} executeWorkflow={executeWorkflow} getWorkflowRevisionData={getWorkflowRevisionData} searchParams={searchParams} setSearchParams={setSearchParams} deleteRevision={deleteRevision}  getRevisions={getRevisions} getTags={getTags} removeTag={removeTag}  />
                        </>
                    :<></>}
                    { activeTab === 2 ?
                        <WorkingRevision 
                            getWorkflowSankeyMetrics={getWorkflowSankeyMetrics}
                            searchParams={searchParams}
                            setSearchParams={setSearchParams}
                            namespace={namespace}
                            executeWorkflow={executeWorkflow}
                            saveWorkflow={saveWorkflow} 
                            updateWorkflow={updateWorkflow} 
                            discardWorkflow={discardWorkflow} 
                            updateRevisions={() => {
                                setRevisions(null)
                            }}
                            wf={atob(data.revision.source)} 
                        />
                    :<></>}
                    { activeTab === 3 ?
                        <WorkflowDependencies namespace={namespace} workflow={filepath} />
                    :
                    <></>
                    }
                    { activeTab === 4 ?
                        <SettingsTab addAttributes={addAttributes} deleteAttributes={deleteAttributes} workflowData={data} setWorkflowLogToEvent={setWorkflowLogToEvent} namespace={namespace} workflow={filepath} />
                    :<></>}
                </FlexBox>
            </FlexBox>
        </>
    )
}

export default WorkflowPage;


function WorkflowDependencies(props) {
    const {workflow, namespace} = props
    const [load, setLoad] = useState(true)
    const [dependencies, setDependencies] = useState(null)
    const {data, getWorkflows} = useNamespaceDependencies(Config.url, namespace, localStorage.getItem('apikey'))

    useEffect(()=>{
        async function getDependencies() {
            if(load && data !== null){
                let wfo = await getWorkflows()
                let arr = Object.keys(wfo)
                for(let i=0 ; i < arr.length; i++) {
                    if(arr[i] === workflow) {
                        setDependencies(wfo[workflow])
                        break
                    }
                }
                setLoad(false)
            }
        }
        getDependencies()
    },[load, data, getWorkflows, workflow])

    return(
        <FlexBox style={{width:"100%"}}>
            <ContentPanel style={{width:"100%"}}>
                <ContentPanelTitle>
                    <ContentPanelTitleIcon>
                        <BsCodeSquare />
                    </ContentPanelTitleIcon>
                    <FlexBox style={{display:"flex", alignItems:"center"}} className="gap">
                        <div>
                            Dependency Graph
                        </div>
                        <HelpIcon msg={"Shows the dependencies the workflow requires in a graph format."} />
                    </FlexBox>
                </ContentPanelTitle>
                <ContentPanelBody>
                    <DependencyDiagram dependencies={dependencies} workflow={workflow} type={"workflow"}/>
                </ContentPanelBody>
            </ContentPanel>
        </FlexBox>
    )
}

function WorkingRevisionErrorBar(props) {
    const { errors, showErrors } = props

    return (
        <div className={`editor-drawer ${showErrors ? "expanded" : ""}`}>
            <FlexBox className="col">
                <FlexBox className="row" style={{ justifyContent: "flex-start", alignItems: "center", borderBottom: "1px solid #536470", padding: "6px 10px 6px 10px" }}>
                    <div style={{ paddingRight: "6px" }}>
                        Problem
                    </div>
                    <div style={{ display: "flex", justifyContent: "center", alignItems: "center", borderRadius: "50%", backgroundColor: "#384c5a", width: "18px", height: "18px", fontWeight: "bold", textAlign: "center" }}>
                        <pre style={{ margin: "0px", fontSize: "medium" }}>{errors.length}</pre>
                    </div>
                </FlexBox>
                <FlexBox className="col" style={{ padding: "6px 10px 6px 10px", overflowY: "scroll" }}>
                    {errors.length > 0 ?
                        <>
                            {errors.map((err) => {
                                return (
                                    <FlexBox className="row" style={{ justifyContent: "flex-start", alignItems: "center", paddingBottom: "4px" }}>
                                        <IoCloseCircleOutline style={{ paddingRight: "6px", color: "#ec4f79" }} />
                                        <div>
                                            {err}
                                        </div>
                                    </FlexBox>)

                            })}

                        </>
                        :
                        <FlexBox className="row" style={{ justifyContent: "flex-start", alignItems: "center" }}>
                            <IoCheckmarkCircleOutline style={{ paddingRight: "6px", color: "#28a745" }} />
                            <div>
                                No Errors
                            </div>
                        </FlexBox>
                    }
                </FlexBox>
            </FlexBox>
        </div>
    )
}

function WorkingRevision(props) {
    const {updateRevisions, searchParams, setSearchParams, getWorkflowSankeyMetrics, wf, updateWorkflow, discardWorkflow, saveWorkflow, executeWorkflow,namespace} = props

    const navigate = useNavigate()
    const [load, setLoad] = useState(true)
    const [oldWf, setOldWf] = useState("")
    const [workflow, setWorkflow] = useState("")
    const [input, setInput] = useState("{\n\t\n}")

    const [tabBtn, setTabBtn] = useState(searchParams.get('revtab') !== null ? parseInt(searchParams.get('revtab')): 0);

    useEffect(()=>{
        setTabBtn(searchParams.get('revtab') !== null ? parseInt(searchParams.get('revtab')): 0)
    }, [searchParams])

    // Error States
    const [errors, setErrors] = useState([])
    const [showErrors, setShowErrors] = useState(false)

    // Loading States
    // Tracks if a button tied to a operation is currently executing.
    const [opLoadingStates, setOpLoadingStates] = useState({
        "IsLoading": false,
        "Save": false,
        "Update": false,
        "Undo": false
    })

    // Push a operation loading state to a target.
    const pushOpLoadingState = useCallback((target, value)=>{
        let old = opLoadingStates
        old[target] = value

        // If any operation is executing, this is set to ture
        old["IsLoading"] = (opLoadingStates["Save"] || opLoadingStates["Update"] || opLoadingStates["Undo"])
        setOpLoadingStates({...old})
    },[opLoadingStates])

    useEffect(()=>{
        if(wf !== workflow && load) {
            setLoad(false)
            setWorkflow(wf)
        }
    },[wf, workflow, load])
   
    useEffect(()=>{
        if (oldWf !== wf) {
            setWorkflow(wf)
            setOldWf(wf)
            pushOpLoadingState("Save", false)
        }
    },[oldWf, wf, pushOpLoadingState])

    let saveFn = (newWf, oldWf) => {

        return () => {
            if (newWf === oldWf) {
                setErrors(["Can't save - no changes have been made."])
                setShowErrors(true)
                pushOpLoadingState("Save", false)
                return
            }
            setErrors([])
            pushOpLoadingState("Save", true)
            updateWorkflow(newWf).then(()=>{
                setShowErrors(false)
            }).catch((opError) => {
                setErrors([opError.message])
                setShowErrors(true)
                pushOpLoadingState("Save", false)
            })
        }
    }

    return(
        <FlexBox style={{width:"100%"}}>
            <ContentPanel style={{width:"100%"}}>
                <ContentPanelTitle>
                    <ContentPanelTitleIcon>
                        <BsCodeSquare />
                    </ContentPanelTitleIcon>
                    <FlexBox style={{display:"flex", alignItems:"center"}} className="gap">
                        <div>
                            Active Revision
                        </div>
                        <HelpIcon msg={"Latest revision where you can edit and create new revisions."} />
                        <TabbedButtons revision={"latest"} setSearchParams={setSearchParams} searchParams={searchParams} tabBtn={tabBtn} setTabBtn={setTabBtn} />
                    </FlexBox>
                </ContentPanelTitle>
                <ContentPanelBody style={{padding: "0px"}}>
                {tabBtn === 0 ?
                    <FlexBox className="col" style={{ overflow: "hidden" }}>
                        <FlexBox>
                            <DirektivEditor saveFn={saveFn(workflow, oldWf)} style={{borderRadius: "0px"}} dlang="yaml" value={workflow} dvalue={oldWf} setDValue={setWorkflow} disableBottomRadius={true} />
                        </FlexBox>
                        <FlexBox className="gap" style={{ backgroundColor: "#223848", color: "white", height: "44px", maxHeight: "44px", paddingLeft: "10px", minHeight: "44px", alignItems: 'center', position: "relative", borderRadius: "0px 0px 8px 8px" }}>
                            <WorkingRevisionErrorBar errors={errors} showErrors={showErrors}/>
                            <div style={{ display: "flex", flex: 1 }}>
                                <div onClick={async () => {
                                    setErrors([])
                                    await discardWorkflow()
                                    setShowErrors(false)
                                }} className={`btn-terminal ${opLoadingStates["IsLoading"] ? "terminal-disabled" : ""}`}>
                                    Undo
                                </div>
                            </div>
                            <div style={{display:"flex", flex:1, justifyContent:"center"}}>
                                <Modal 
                                    style={{ justifyContent: "center" }}
                                    className="run-workflow-modal"
                                    modalStyle={{color: "black"}}
                                    title="Run Workflow"
                                    buttonDisabled={opLoadingStates["IsLoading"]}
                                    onClose={()=>{
                                        setInput("{\n\t\n}")
                                    }}
                                    actionButtons={[
                                        ButtonDefinition("Run", async () => {
                                            let r = ""
                                            if(input === "{\n\t\n}"){
                                                r = await executeWorkflow()
                                            } else {
                                                r = await executeWorkflow(input)
                                            }
                                            if(r.includes("execute workflow")){
                                                // is an error
                                                throw new Error(r)
                                            } else {
                                                navigate(`/n/${namespace}/instances/${r}`)
                                            }
                                        }, "small blue", ()=>{}, true, false),
                                        ButtonDefinition("Cancel", async () => {
                                        }, "small light", ()=>{}, true, false)
                                    ]}
                                    button={(
                                        <>
                                            { workflow !== oldWf ?
                                            <div className='btn-terminal disabled' >
                                                Run (requires save)
                                            </div>
                                            :
                                            <div className={`btn-terminal ${opLoadingStates["IsLoading"] ? "terminal-disabled" : ""}`}>
                                                Run
                                            </div>
                                            }                                        
                                        </>
                                    )}
                                >
                                    <FlexBox style={{height: "40vh", width: "30vw", minWidth: "250px", minHeight: "200px"}}>
                                        <FlexBox style={{overflow:"hidden"}}>
                                            <AutoSizer>
                                                {({height, width})=>(
                                                    <DirektivEditor height={height} width={width} dlang="json" dvalue={input} setDValue={setInput}/>
                                                )}
                                            </AutoSizer>
                                        </FlexBox>
                                    </FlexBox>
                                </Modal>
                            </div>
                            <div style={{ display: "flex", flex: 1, gap: "3px", justifyContent: "flex-end", paddingRight: "10px"}}>
                                <div className={`btn-terminal ${opLoadingStates["Save"] ? "terminal-loading" : ""} ${workflow === oldWf ? "terminal-disabled" : ""}`} title={"Save workflow to latest"} onClick={async () => {
                                    setErrors([])
                                    pushOpLoadingState("Save", true)
                                    updateWorkflow(workflow).then(()=>{
                                        setShowErrors(false)
                                    }).catch((opError) => {
                                        setErrors([opError.message])
                                        setShowErrors(true)
                                        pushOpLoadingState("Save", false)
                                    })
                                }}>
                                    Save
                                </div>
                                <div className={`btn-terminal ${opLoadingStates["IsLoading"] ? "terminal-disabled" : ""}`} title={"Save latest workflow as new revision"} onClick={async () => {
                                    setErrors([])
                                    try{
                                        const result = await saveWorkflow()
                                        if(result?.node?.name)
                                        {
                                            updateRevisions()
                                            setShowErrors(false)
                                            navigate(`/n/${namespace}/explorer/${result.node.name}?tab=1&revision=${result.revision.name}&revtab=0`)
                                        }else{
                                            setErrors("Something went wrong")
                                            setShowErrors(true)
                                        }
                                    }catch(e){
                                        setErrors(e.toString())
                                        setShowErrors(true)
                                    }
                                    
                                }}>
                                    Make Revision
                                </div>
                                <div className={"btn-terminal editor-info"} title={`${showErrors ? "Hide Problems": "Show Problems"}`} onClick={async () => {
                                    setShowErrors(!showErrors)
                                }}>
                                    {showErrors?
                                    <VscChevronDown style={{ width: "80%", height: "80%" }} />
                                    :
                                    <VscChevronUp style={{ width: "80%", height: "80%" }} />
                                    }
                                </div>
                            </div>
                        </FlexBox>
                    </FlexBox>:""}
                    {tabBtn === 1 ? <WorkflowDiagram disabled={true} workflow={YAML.load(workflow)}/>:""}
                    {tabBtn === 2 ? <SankeyDiagram revision={"latest"} getWorkflowSankeyMetrics={getWorkflowSankeyMetrics} />:""}
                </ContentPanelBody>
            </ContentPanel>
        </FlexBox>
    )
}

function TabBar(props) {

    let {activeTab, setActiveTab, setSearchParams, toggleWorkflow, getWorkflowRouter, router, setRouter} = props;
    let tabLabels = [
        "Overview",
        "Revisions",
        "Active Revision",
        "Dependency Graph", 
        "Settings"
    ]


    let tabDOMs = [];
    for (let i = 0; i < 5; i++) {

        let className = "tab-bar-item"
        if (i === activeTab) {
            className += " active"
        }

        let key = GenerateRandomKey("tab-item-")
        tabDOMs.push(
            <FlexBox key={key} className={className} onClick={() => {
                setActiveTab(i)
                setSearchParams({tab: i}, {replace:true})
            }}>
                {tabLabels[i]}
            </FlexBox>
        )
    }

    return (
        <FlexBox className="tab-bar">
            {tabDOMs}
            <FlexBox className="tab-bar-item gap uninteractive">
            <label className="switch">
                <input onChange={async()=>{
                    await toggleWorkflow(!router.live)
                    setRouter(await getWorkflowRouter())
                }} type="checkbox" checked={router ? router.live : false}/>
                <span className="slider-broadcast"></span>
            </label>
            <div className="rev-toggle-label hide-on-small">
                {!router.live ? 
                    "Disabled":
                    "Enabled"}
            </div>
            </FlexBox>
        </FlexBox>
    )
}

function WorkflowInstances(props) {
    const {instances, namespace} = props
    return(
        <ContentPanelBody>
            <>
            <div style={{width: "100%"}}>
        {
            instances !== null && instances.length === 0 ? <div style={{paddingLeft:"10px", fontSize:"10pt"}}>No instances have been recently executed. Recent instances will appear here.</div>:
            <table className="instances-table" style={{width: "100%"}}>
            <thead>
                <tr>
                    <th className="center-align" style={{maxWidth: "120px", minWidth: "120px", width: "120px"}}>
                        State
                    </th>
                    {/* <th className="center-align">
                        Name
                    </th> */}
                    <th className="center-align">
                        Revision
                    </th>
                    <th className="center-align">
                        Started <span className="hide-on-small">at</span>
                    </th>
                    <th className="center-align">
                        <span className="hide-on-small">Last</span> Updated
                    </th>
                </tr>
            </thead>
            <tbody>
                {instances !== null ? 
                <>
                    <>
                    {instances.map((obj)=>{
                    let key = GenerateRandomKey("instance-")
                    return(
                        <InstanceRow
                            wf={true}
                            key={key}
                            namespace={namespace}
                            state={obj.node.status} 
                            name={obj.node.as} 
                            id={obj.node.id}
                            started={dayjs.utc(obj.node.createdAt).local().format("HH:mm:ss a")} 
                            startedFrom={dayjs.utc(obj.node.createdAt).local().fromNow()}
                            finished={dayjs.utc(obj.node.updatedAt).local().format("HH:mm:ss a")}
                            finishedFrom={dayjs.utc(obj.node.updatedAt).local().fromNow()}
                        />
                    )
                    })}</>
                </>
                :<></>}
            </tbody>
        </table>}
            </div>
            </>
        </ContentPanelBody>
    )
}

function OverviewTab(props) {
    const {getInstancesForWorkflow,  namespace, filepath, router, getSuccessFailedMetrics} = props
    const [load, setLoad] = useState(true)
    const [instances, setInstances] = useState([])
    const [err, setErr] = useState(null)

    // fetch instances using the workflow hook from above
    useEffect(()=>{
        async function listData() {
            if(load){
                // get the instances
                try {
                    let resp = await getInstancesForWorkflow()
                    if (resp.instances.edges) {
                        setInstances(resp.instances.edges)
                    }
                } catch (e) {
                    setErr(e)
                }
            }
            setLoad(false)
        }
        listData()
    },[load, getInstancesForWorkflow])

    if (err) {
        // TODO report error
    }

    return(
        <>
            <div className="gap">
                <FlexBox className="gap wrap">
                    <FlexBox style={{ minWidth: "370px", width:"60%", maxHeight: "342px"}}>
                        <ContentPanel style={{ width: "100%", minWidth: "300px"}}>
                            <ContentPanelTitle>
                                <ContentPanelTitleIcon>
                                    <BsCodeSquare />
                                </ContentPanelTitleIcon>
                                <FlexBox style={{display:"flex", alignItems:"center"}} className="gap">
                                    <div>
                                        Instances
                                    </div>
                                    <HelpIcon msg={"List of instances for this workflow."} />
                                </FlexBox>
                            </ContentPanelTitle>
                            <WorkflowInstances instances={instances} namespace={namespace} />
                        </ContentPanel>
                    </FlexBox>
                    <FlexBox style={{ minWidth: "370px", maxHeight: "342px" }}>
                        <ContentPanel style={{ width: "100%", minWidth: "300px"}}>
                            <ContentPanelTitle>
                                <ContentPanelTitleIcon>
                                    <BsCodeSquare />
                                </ContentPanelTitleIcon>
                                <FlexBox style={{display:"flex", alignItems:"center"}} className="gap">
                                    <div>
                                        Success/Failure Rate
                                    </div>
                                    <HelpIcon msg={"Success and failure of the workflow being run."} />
                                </FlexBox>
                            </ContentPanelTitle>
                            <ContentPanelBody>
                                <SuccessFailureGraph getSuccessFailedMetrics={getSuccessFailedMetrics} />
                            </ContentPanelBody>
                        </ContentPanel>
                    </FlexBox>
                </FlexBox>
            </div>
            <FlexBox style={{maxHeight: "140px", minHeight:"140px"}}>
                <ContentPanel style={{ width: "100%", minWidth: "300px" }}>
                    <ContentPanelTitle>
                        <ContentPanelTitleIcon>
                            <BsCodeSquare />
                        </ContentPanelTitleIcon>
                        <FlexBox style={{display:"flex", alignItems:"center"}} className="gap">
                            <div>
                                Traffic Distribution
                            </div>
                            <HelpIcon msg={"Distributed traffic between different workflow revisions."} />
                        </FlexBox>
                    </ContentPanelTitle>
                    <TrafficDistribution routes={router.routes}/>
                </ContentPanel>
            </FlexBox>
            <FlexBox>
                <ContentPanel style={{ width: "100%", minWidth: "300px"}}>
                    <ContentPanelTitle>
                        <ContentPanelTitleIcon>
                            <BsCodeSquare />
                        </ContentPanelTitleIcon>
                        <FlexBox style={{display:"flex", alignItems:"center"}} className="gap">
                            <div>
                                Workflow Services
                            </div>
                            <HelpIcon msg={"A List of services for this workflow."} />
                        </FlexBox>
                    </ContentPanelTitle>
                    <WorkflowServices namespace={namespace} filepath={filepath} />
                </ContentPanel>
            </FlexBox>
        </>
    )
}

function SuccessFailureGraph(props){

    const {getSuccessFailedMetrics} = props
    const [metrics, setMetrics] = useState([
        {title: 'Success', value: 0, color: "url(#success)", percentage: 0},
        {title: 'Failed', value: 0, color: "url(#failed)", percentage: 0}
    ])
    const [total, setTotal] = useState(0)
    const [load, setLoad] = useState(true)
    const [err, setErr] = useState("")

    useEffect(()=>{
        async function get() {
            try {
                if(load){
                    let ms = metrics
                    let mets = await getSuccessFailedMetrics()
                    let t = 0
                    if(mets.success && mets.failure) {
                        if(mets.success[0]){
                            ms[0].value = mets.success[0].value[1]
                            t = t + parseInt(mets.success[0].value[1])
                        }
                        if(mets.failure[0]){
                            ms[1].value = mets.failure[0].value[1]
                            t = t + parseInt(mets.failure[0].value[1])
                        }
    
                        if(mets.success[0]) {
                            ms[0].percentage = (ms[0].value / t * 100).toFixed(2)
                        }
                        if(mets.failure[0]){
                            ms[1].percentage = (ms[1].value / t * 100).toFixed(2)
                        }
    
                        if(t > 0) {
                            setMetrics(ms)
                            setTotal(t)
                        } else {
                            setErr("No metrics have been found.")
                        }
                        
                    } else {
                        setErr(mets)
                    }
                    setLoad(false)
                }
            } catch(e){
                setErr(e.message)
                setLoad(false)
            }
        }
        get()
    },[load, getSuccessFailedMetrics, metrics])

    if(load){
        return ""
    }
    if(err !== "") {
        return(
            <FlexBox style={{justifyContent:"center", alignItems:'center', color:"red", fontSize:"10pt"}}>
                <div className="error-message-metrics">{err.replace("get failed metrics:", "")}</div>
            </FlexBox>
        )
    }

    return(
        <FlexBox className="col" style={{maxHeight:"250px", marginTop:"20px"}}>
            <PieChart
                totalValue={total}
                label=""
                labelStyle={{
                    fontSize:"6pt",
                    fontWeight: "bold",
                    fill: "white"
                }}
                data ={metrics}
            >
            <defs>
                <radialGradient id="failed" gradientUnits="userSpaceOnUse">
                    <stop offset="0%" stopColor="#F3537E" />
                    <stop offset="80%" stopColor="#DE184D" />
                </radialGradient>
                <radialGradient id="success" gradientUnits="userSpaceOnUse">
                    <stop offset="0%" stopColor="#1FEAC5" />
                    <stop offset="100%" stopColor="#25B49A" />
                </radialGradient>
            </defs>
            </PieChart>
            <FlexBox style={{marginTop:"10px", gap:"50px"}}>
                <FlexBox className="col">
                    <FlexBox style={{justifyContent:"center", alignItems:"center", gap:"5px", fontWeight:"bold", fontSize:"12pt"}}>
                        <div style={{height:"8px", width:"8px", background:"linear-gradient(180deg, #1FEAC5 0%, #25B49A 100%)"}} />
                        Success
                    </FlexBox>
                    <FlexBox style={{justifyContent:"center"}}>
                        {metrics[0].percentage}%
                    </FlexBox>
                </FlexBox>
                <FlexBox className="col">
                    <FlexBox style={{justifyContent:"center", alignItems:"center", gap:"5px", fontWeight:"bold", fontSize:"12pt"}}>
                        <div style={{height:"8px", width:"8px", background:"linear-gradient(180deg, #F3537E 0%, #DE184D 100%)"}} />
                        Failure
                    </FlexBox>
                    <FlexBox style={{justifyContent:"center"}}>
                        {metrics[1].percentage}%
                    </FlexBox>
                </FlexBox>
            </FlexBox>
        </FlexBox>
        
    )
}

function TrafficDistribution(props) {
    const {routes} = props

    if (!routes) {
        return <></>
    }

    // using latest for traffic
    if (routes.length === 0) {
        return (
            <ContentPanelBody>
                <FlexBox className="col gap" style={{justifyContent:"center"}}>
                    <Slider className="traffic-distribution" disabled={true}/>
                    <FlexBox className="col" style={{fontSize:"10pt", marginTop:"5px", maxHeight:"20px", color: "#C1C5C8"}}>
                        latest<span style={{fontSize:"8pt"}}>100%</span>
                    </FlexBox>
                </FlexBox>
            </ContentPanelBody>
        )
    }


    return(
        <ContentPanelBody>
            <FlexBox className="col gap" style={{justifyContent:'center'}}>
                <FlexBox style={{fontSize:"10pt", marginTop:"5px", maxHeight:"20px", color: "#C1C5C8"}}>
                    {routes[0] ? 
                    <FlexBox className="col">
                        <span title={routes[0].ref}>{routes[0].ref.substr(0, 8)}</span>
                    </FlexBox>:""}
                    {routes[1] ? 
                    <FlexBox className="col" style={{ textAlign:'right'}}>
                        <span title={routes[1].ref}>{routes[1].ref.substr(0,8)}</span>
                    </FlexBox>:""}
                </FlexBox>
                <Slider value={routes[0] ? routes.length === 2 ? `${routes[0].weight}`: `100` : 0} className="traffic-distribution" disabled={true}/>
                <FlexBox style={{fontSize:"10pt", marginTop:"5px", maxHeight:"50px", color: "#C1C5C8"}}>
                    {routes[0] ? 
                    <FlexBox className="col">
                        <span>{routes.length === 2 ? `${routes[0].weight}%`: `100%`}</span>
                    </FlexBox>:""}
                    {routes[1] ? 
                    <FlexBox className="col" style={{ textAlign:'right'}}>
                        <span>{routes[1].weight}%</span>
                    </FlexBox>:""}
                </FlexBox>
            </FlexBox>
        </ContentPanelBody>
    )
}

function WorkflowServices(props) {
    const {namespace, filepath} = props
    const {data, err} = useWorkflowServices(Config.url, true, namespace, filepath.substring(1), localStorage.getItem("apikey"))

    if (data === null) {
        return     <div className="col">
        <FlexBox style={{ height:"40px", }}>
                <FlexBox className="gap" style={{alignItems:"center", paddingLeft:"8px"}}>
                    <div style={{fontSize:"10pt", }}>
                        No services have been created.
                    </div>
                </FlexBox>
        </FlexBox>
    </div>
    }

    if (err) {
        // TODO report error
    }

    return(
        <ContentPanelBody>
            <FlexBox className="col gap">
                {data.length === 0 ? 
                       <div className="col">
                       <FlexBox style={{ height:"40px", }}>
                               <FlexBox className="gap" style={{alignItems:"center", paddingLeft:"8px"}}>
                                   <div style={{fontSize:"10pt", }}>
                                       No services have been created.
                                   </div>
                               </FlexBox>
                       </FlexBox>
                   </div>
                :""}
                {data.map((obj)=>{
                    return(
                        <Service
                            dontDelete={true}
                            url={`/n/${namespace}/explorer/${filepath.substring(1)}?function=${obj.info.name}&version=${obj.info.revision}`}
                            name={obj.info.name}
                            status={obj.status}
                            image={obj.info.image}
                            conditions={obj.conditions}
                        />
      
                    )
                })}
            </FlexBox>
        </ContentPanelBody>
    )
}

function WorkflowAttributes(props) {
    const {attributes, deleteAttributes, addAttributes} = props


    const [load, setLoad] = useState(true)
    const [attris, setAttris] = useState(attributes)
    const tagInput = useRef()

    useEffect(()=>{
        if(load){
            setAttris(attributes)
            setLoad(false)
        }
    },[attributes,load])

    const removeTag = async(i) => {
        // await deleteAttribute(attris[i])
        await deleteAttributes([attris[i]])
        const newTags = [...attris]
        newTags.splice(i,1)
        setAttris(newTags)
    }

    const inputKeyDown = async (e) => {
        const val = e.target.value
        if((e.key === " " || e.key === "Enter") && val) {
            if(attris.find(tag => tag.toLowerCase() === val.toLowerCase())){
                return;
            }
            try {
                await addAttributes([val])
                setAttris([...attris, val])
                tagInput.current.value = null
            } catch(e) {
                
            }
        } else if (e.key === "Backspace" && !val) {
            removeTag(attris.length - 1)
        }
    }

    return(
            // <FlexBox>
                <div className="input-tag" style={{width: "100%", padding:"2px"}}>
                    <ul className="input-tag__tags">
                        {attris.map((tag, i) => (
                            <li key={tag}>
                                {tag}
                                <button type="button" onClick={() => { removeTag(i); }}>+</button>
                            </li>
                        ))}
                        <li className="input-tag__tags__input"><input placeholder="Enter attribute" type="text" onKeyDown={inputKeyDown} ref={tagInput} /></li>
                    </ul>
                </div>
            // </FlexBox>
    )
}

function SettingsTab(props) {

    const {namespace, workflow, addAttributes, deleteAttributes, workflowData, setWorkflowLogToEvent} = props
    const [logToEvent, setLogToEvent] = useState(workflowData.eventLogging)

    const [lteStatus, setLTEStatus] = useState(null);
    const [lteStatusMessage, setLTEStatusMessage] = useState(null);

    return (
        <>
            <FlexBox className="gap wrap col">
                <div style={{width: "100%"}}>
                    <AddWorkflowVariablePanel namespace={namespace} workflow={workflow} />
                </div>
                <FlexBox className="gap">
                    <FlexBox  style={{flex:1, maxHeight: "156px", minWidth:"300px"}}>          
                        <div style={{width: "100%", minHeight: "144px"}}>
                            <ContentPanel style={{width: "100%", height: "100%"}}>
                                <ContentPanelTitle>
                                    <ContentPanelTitleIcon>
                                        <IoMdLock/>
                                    </ContentPanelTitleIcon>
                                    <FlexBox style={{display:"flex", alignItems:"center"}} className="gap">
                                        <div>
                                            Log to Event
                                        </div>
                                        <HelpIcon msg={"Ability to trigger cloud event logging for that workflow."} />
                                    </FlexBox>
                                </ContentPanelTitle>
                                <ContentPanelBody className="col" style={{
                                    alignItems: "center"
                                }}>
                                    <FlexBox className="gap" style={{flexDirection: "column", alignItems: "center"}}>
                                        <FlexBox style={{width:"100%"}}>
                                            <input value={logToEvent} onChange={(e)=>setLogToEvent(e.target.value)} type="text" placeholder="Enter the 'event' type to send logs to" />
                                        </FlexBox>
                                        <div style={{width:"99.5%", margin:"auto", background: "#E9ECEF", height:"1px"}}/>
                                        <FlexBox className="gap" style={{justifyContent:"flex-end", width:"100%"}}>
                                            { lteStatus ? <Alert className={`${lteStatus} small`}>{lteStatusMessage}</Alert> : <></> }
                                            <Button onClick={async()=>{
                                                try { 
                                                    await setWorkflowLogToEvent(logToEvent)
                                                } catch(err) {
                                                    setLTEStatus("failed")
                                                    setLTEStatusMessage(err.message)
                                                    return err
                                                }

                                                setLTEStatus("success")
                                                setLTEStatusMessage("'Log to Event' value set!")
                                            }} className="small">
                                                Save
                                            </Button>
                                        </FlexBox>
                                    </FlexBox>
                                </ContentPanelBody>
                            </ContentPanel>
                        </div>
                    </FlexBox>

                    <FlexBox style={{flex: 4,maxWidth:"1200px", height:"fit-content", minHeight: "156px"}}>
                        {/* <div style={{width: "100%", minHeight: "200px"}}> */}
                            <ContentPanel style={{width: "100%"}}>
                                <ContentPanelTitle>
                                    <ContentPanelTitleIcon>
                                        <IoMdLock/>
                                    </ContentPanelTitleIcon>
                                    <FlexBox style={{display:"flex", alignItems:"center"}} className="gap">
                                        <div>
                                            Add Attributes
                                        </div>
                                        <HelpIcon msg={"Attributes to define the workflow."} />
                                    </FlexBox>
                                    {/* <ContentPanelHeaderButton>
                                        + Add
                                    </ContentPanelHeaderButton> */}
                                </ContentPanelTitle>
                                <ContentPanelBody>
                                    <WorkflowAttributes attributes={workflowData.node.attributes} deleteAttributes={deleteAttributes} addAttributes={addAttributes}/>
                                </ContentPanelBody>
                            </ContentPanel>
                        {/* </div> */}
                    </FlexBox>
                </FlexBox>

            </FlexBox>
        </>
    )

}

