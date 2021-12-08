import React, { useEffect, useRef, useState } from 'react';
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
import {IoInformation} from 'react-icons/io5'
import { Service } from '../../namespace-services';
import DirektivEditor from '../../../components/editor';
import AddWorkflowVariablePanel from './variables';
import { RevisionSelectorTab } from './revisionTab';
import DependencyDiagram from '../../../components/dependency-diagram';

import Slider from 'rc-slider';
import 'rc-slider/assets/index.css';
import Button from '../../../components/button';
import Modal, { ButtonDefinition } from '../../../components/modal';


dayjs.extend(utc)
dayjs.extend(relativeTime);


function WorkflowPage(props) {
    const {namespace} = props
    const [searchParams, setSearchParams] = useSearchParams()
    const params = useParams()

    // set tab query param on load
    useEffect(()=>{
        if(searchParams.get('tab') === null) {
            setSearchParams({tab: 0}, {replace:true})
        }
    },[searchParams, setSearchParams])
    
    let filepath = "/"

    if(!namespace) {
        return <></>
    }

    if(params["*"] !== undefined){
        filepath = `/${params["*"]}`
    }

    return(
        <InitialWorkflowHook setSearchParams={setSearchParams} searchParams={searchParams} namespace={namespace} filepath={filepath}/>
    )
}

function InitialWorkflowHook(props){
    const {namespace, filepath, searchParams, setSearchParams} = props

    const [activeTab, setActiveTab] = useState(searchParams.get("tab") !== null ? parseInt(searchParams.get('tab')): 0)

    const {data, err, tagWorkflow, addAttributes, deleteAttributes, setWorkflowLogToEvent, editWorkflowRouter, getWorkflowSankeyMetrics, getWorkflowRevisionData, getWorkflowRouter, toggleWorkflow, executeWorkflow, getInstancesForWorkflow, getRevisions, deleteRevision, saveWorkflow, updateWorkflow, discardWorkflow} = useWorkflow(Config.url, true, namespace, filepath.substring(1))
    const [router, setRouter] = useState(null)

    const [revisions, setRevisions] = useState(null)
    const [revsErr, setRevsErr] = useState("")

    // fetch revisions using the workflow hook from above
    useEffect(()=>{
        async function listData() {
            if(revisions === null){
                // get the instances
                let resp = await getRevisions()
                if(Array.isArray(resp)){
                    setRevisions(resp)
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
                        <OverviewTab router={router} namespace={namespace} getInstancesForWorkflow={getInstancesForWorkflow} filepath={filepath}/>
                    :<></>}
                    { activeTab === 1 ?
                        <>
                        <RevisionSelectorTab tagWorkflow={tagWorkflow} namespace={namespace} filepath={filepath} updateWorkflow={updateWorkflow} setRouter={setRouter} editWorkflowRouter={editWorkflowRouter} getWorkflowRouter={getWorkflowRouter} setRevisions={setRevisions} revisions={revisions} router={router} getWorkflowSankeyMetrics={getWorkflowSankeyMetrics} executeWorkflow={executeWorkflow} getWorkflowRevisionData={getWorkflowRevisionData} searchParams={searchParams} setSearchParams={setSearchParams} deleteRevision={deleteRevision} namespace={namespace} getRevisions={getRevisions} filepath={filepath} />
                        </>
                    :<></>}
                    { activeTab === 2 ?
                        <WorkingRevision 
                            namespace={namespace}
                            executeWorkflow={executeWorkflow}
                            saveWorkflow={saveWorkflow} 
                            updateWorkflow={updateWorkflow} 
                            discardWorkflow={discardWorkflow} 
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
                    <div>
                        Dependency Graph
                    </div>
                </ContentPanelTitle>
                <ContentPanelBody>
                    <DependencyDiagram dependencies={dependencies} workflow={workflow} type={"workflow"}/>
                </ContentPanelBody>
            </ContentPanel>
        </FlexBox>
    )
}

function WorkingRevision(props) {
    const {wf, updateWorkflow, discardWorkflow, saveWorkflow, executeWorkflow,namespace} = props

    const navigate = useNavigate()
    const [load, setLoad] = useState(true)
    const [oldWf, setOldWf] = useState("")
    const [workflow, setWorkflow] = useState("")
    const [input, setInput] = useState("{\n\t\n}")
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
        }
    },[oldWf, wf])

    return(
        <FlexBox style={{width:"100%"}}>
            <ContentPanel style={{width:"100%"}}>
                <ContentPanelTitle>
                    <ContentPanelTitleIcon>
                        <BsCodeSquare />
                    </ContentPanelTitleIcon>
                    <div>
                        Active Revision
                    </div>
                </ContentPanelTitle>
                <ContentPanelBody>
                    <FlexBox className="col" style={{overflow:"hidden"}}>
                        <FlexBox >
                            <DirektivEditor dlang="yaml" value={workflow} dvalue={oldWf} setDValue={setWorkflow} />
                        </FlexBox>
                        <FlexBox className="gap zest" style={{backgroundColor:"#223848", color:"white", height:"40px", maxHeight:"40px", paddingLeft:"10px", minHeight:"40px", borderTop:"1px solid white", alignItems:'center', position: "relative"}}>
                            <div style={{display:"flex", flex:1 }}>
                                <div onClick={async ()=> {
                                    // console.log("oldWf =", oldWf)
                                    // console.log("workflow =", workflow)
                                    await discardWorkflow()
                                }} className={"btn-terminal"}>
                                    Undo
                                </div>
                            </div>
                            <div style={{display:"flex", flex:1, justifyContent:"center"}}>
                                <Modal 
                                    style={{ justifyContent: "center" }}
                                    className="run-workflow-modal"
                                    modalStyle={{color: "black"}}
                                    title="Run Workflow"
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
                                                return r
                                            } else {
                                                navigate(`/n/${namespace}/instances/${r}`)
                                            }
                                        }, "small blue", true, false),
                                        ButtonDefinition("Cancel", async () => {
                                        }, "small light", true, false)
                                    ]}
                                    button={(
                                        <div className={"btn-terminal"}>
                                            Run
                                        </div>
                                    )}
                                >
                                    <FlexBox style={{overflow:"hidden"}}>
                                        <DirektivEditor height="200px" width="300px" dlang="json" dvalue={input} setDValue={setInput}/>
                                    </FlexBox>
                                </Modal>
                            </div>
                            <div style={{display:"flex", flex:1, gap :"3px", justifyContent:"flex-end", paddingRight:"10px"}}>
                                <div className={`btn-terminal ${workflow === oldWf ? "terminal-disabled" : ""}`} onClick={async()=>{
                                    await updateWorkflow(workflow)
                                }}>
                                    Save
                                </div>
                                <div className={"btn-terminal"} onClick={async()=>{
                                    await saveWorkflow()
                                }}>
                                    Save as new revision
                                </div>
                                {/* <div className={"btn-terminal editor-info"} onClick={async()=>{
                                }}>
                                    <IoInformation style={{width: "80%", height: "80%"}}/>
                                </div> */}
                            </div>
                        </FlexBox>
                    </FlexBox>
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
                Enabled
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
                    <th className="center-align">
                        Name
                    </th>
                    <th className="center-align">
                        Revision ID
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
                            key={key}
                            namespace={namespace}
                            state={obj.node.status} 
                            name={obj.node.as} 
                            id={obj.node.id}
                            started={dayjs.utc(obj.node.createdAt).local().format("HH:mm a")} 
                            startedFrom={dayjs.utc(obj.node.createdAt).local().fromNow()}
                            finished={dayjs.utc(obj.node.updatedAt).local().format("HH:mm a")}
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
    const {getInstancesForWorkflow,  namespace, filepath, router} = props
    const [load, setLoad] = useState(true)
    const [instances, setInstances] = useState([])
    const [err, setErr] = useState(null)

    // fetch instances using the workflow hook from above
    useEffect(()=>{
        async function listData() {
            if(load){
                // get the instances
                let resp = await getInstancesForWorkflow()
                if(Array.isArray(resp)){
                    setInstances(resp)
                } else {
                    setErr(resp)
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
                                <div>
                                    Instances
                                </div>
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
                                <div>
                                    Success/Failure Rate
                                </div>
                            </ContentPanelTitle>
                        </ContentPanel>
                    </FlexBox>
                </FlexBox>
            </div>
            <FlexBox style={{maxHeight: "140px"}}>
                <ContentPanel style={{ width: "100%", minWidth: "300px" }}>
                    <ContentPanelTitle>
                        <ContentPanelTitleIcon>
                            <BsCodeSquare />
                        </ContentPanelTitleIcon>
                        <div>
                            Traffic Distribution
                        </div>
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
                        <div>
                            Workflow Services
                        </div>
                    </ContentPanelTitle>
                    <WorkflowServices namespace={namespace} filepath={filepath} />
                </ContentPanel>
            </FlexBox>
        </>
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
                <Slider value={routes[0] ? routes[0].weight : 0} className="traffic-distribution" disabled={true}/>
                <FlexBox style={{fontSize:"10pt", marginTop:"5px", maxHeight:"50px", color: "#C1C5C8"}}>
                    {routes[0] ? 
                    <FlexBox className="col">
                        <span>{routes[0].weight}%</span>
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

    const {data, err} = useWorkflowServices(Config.url, true, namespace, filepath.substring(1))
    if (data === null) {
        return <></>
    }

    if (err) {
        // TODO report error
    }

    return(
        <ContentPanelBody>
            <FlexBox className="col gap">
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
                <div className="input-tag" style={{width: "100%"}} >
                    <ul className="input-tag__tags">
                        {attris.map((tag, i) => (
                            <li key={tag}>
                                {tag}
                                <button type="button" onClick={() => { removeTag(i); }}>+</button>
                            </li>
                        ))}
                        <li className="input-tag__tags__input"><input type="text" onKeyDown={inputKeyDown} ref={tagInput} /></li>
                    </ul>
                </div>
            // </FlexBox>
    )
}

function SettingsTab(props) {

    const {namespace, workflow, addAttributes, deleteAttributes, workflowData, setWorkflowLogToEvent} = props

    const [logToEvent, setLogToEvent] = useState(workflowData.eventLogging)

    return (
        <>
            <FlexBox className="gap wrap col">
                <div style={{width: "100%"}}>
                    <AddWorkflowVariablePanel namespace={namespace} workflow={workflow} />
                </div>
                <FlexBox className="gap wrap" style={{maxHeight: "144px"}}>
                    <FlexBox style={{flexGrow: "1"}}>          
                        <div style={{width: "100%", minHeight: "144px"}}>
                            <ContentPanel style={{width: "100%", height: "100%"}}>
                                <ContentPanelTitle>
                                    <ContentPanelTitleIcon>
                                        <IoMdLock/>
                                    </ContentPanelTitleIcon>
                                    Log to Event
                                </ContentPanelTitle>
                                <ContentPanelBody className="col" style={{
                                    alignItems: "center"
                                }}>
                                    <FlexBox className="gap" style={{flexDirection: "column", alignItems: "center"}}>
                                        <FlexBox style={{width:"100%"}}>
                                            <input value={logToEvent} onChange={(e)=>setLogToEvent(e.target.value)} type="text" placeholder="Event to log to" />
                                        </FlexBox>
                                        <FlexBox style={{justifyContent:"flex-end", width:"100%"}}>
                                            <Button onClick={async()=>{
                                                let err = await setWorkflowLogToEvent(logToEvent)
                                                // todo err
                                                console.log(err, "NOTIFY IF ERR")
                                            }} className="small">
                                                Save
                                            </Button>
                                        </FlexBox>
                                    </FlexBox>
                                </ContentPanelBody>
                            </ContentPanel>
                        </div>
                    </FlexBox>
                    <FlexBox style={{ flexGrow: "5" }}>
                        {/* <div style={{width: "100%", minHeight: "200px"}}> */}
                            <ContentPanel style={{width: "100%", height: "100%"}}>
                                <ContentPanelTitle>
                                    <ContentPanelTitleIcon>
                                        <IoMdLock/>
                                    </ContentPanelTitleIcon>
                                    Add Attributes
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

