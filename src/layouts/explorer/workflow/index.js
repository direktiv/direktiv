import React, { useEffect, useState } from 'react';
import './style.css';
import FlexBox from '../../../components/flexbox';
import {useSearchParams} from 'react-router-dom'
import ContentPanel, { ContentPanelBody, ContentPanelHeaderButton, ContentPanelTitle, ContentPanelTitleIcon } from '../../../components/content-panel';
import {BsCodeSquare} from 'react-icons/bs'
import { useNamespaceDependencies, useWorkflow, useWorkflowServices } from 'direktiv-react-hooks';
import { Config } from '../../../util';
import { useParams } from 'react-router';
import {  GenerateRandomKey } from '../../../util';

import * as dayjs from "dayjs"
import relativeTime from "dayjs/plugin/relativeTime";
import utc from "dayjs/plugin/utc"
import { InstanceRow } from '../../instances';
import { IoMdLock } from 'react-icons/io';
import { Service } from '../../namespace-services';
import DirektivEditor from '../../../components/editor';
import AddWorkflowVariablePanel from './variables';
import { RevisionSelectorTab } from './revisionTab';
import DependencyDiagram from '../../../components/dependency-diagram';

import Slider from 'rc-slider';
import 'rc-slider/assets/index.css';


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

    const {data, err, editWorkflowRouter, getWorkflowSankeyMetrics, getWorkflowRevisionData, getWorkflowRouter, toggleWorkflow, executeWorkflow, getInstancesForWorkflow, getRevisions, deleteRevision, saveWorkflow, updateWorkflow, discardWorkflow} = useWorkflow(Config.url, true, namespace, filepath.substring(1))
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
                        <RevisionSelectorTab setRouter={setRouter} editWorkflowRouter={editWorkflowRouter} getWorkflowRouter={getWorkflowRouter} setRevisions={setRevisions} revisions={revisions} router={router} getWorkflowSankeyMetrics={getWorkflowSankeyMetrics} executeWorkflow={executeWorkflow} getWorkflowRevisionData={getWorkflowRevisionData} searchParams={searchParams} setSearchParams={setSearchParams} deleteRevision={deleteRevision} namespace={namespace} getRevisions={getRevisions} filepath={filepath} />
                        </>
                    :<></>}
                    { activeTab === 2 ?
                        <WorkingRevision 
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
                        <SettingsTab namespace={namespace} workflow={filepath} />
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
    const {wf, updateWorkflow, discardWorkflow, saveWorkflow, executeWorkflow} = props

    const [load, setLoad] = useState(true)
    const [oldWf, setOldWf] = useState("")
    const [workflow, setWorkflow] = useState("")

    useEffect(()=>{
        if(wf !== workflow && load) {
            setLoad(false)
            setWorkflow(wf)
            setOldWf(wf)
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
                            <DirektivEditor dlang="yaml" value={oldWf} dvalue={workflow} setDValue={setWorkflow} />
                        </FlexBox>
                        <FlexBox className="gap" style={{backgroundColor:"#223848", color:"white", height:"40px", maxHeight:"40px", paddingLeft:"10px", minHeight:"40px", borderTop:"1px solid white", alignItems:'center'}}>
                            <div style={{display:"flex", flex:1 }}>
                                <div onClick={async ()=> {
                                    await discardWorkflow()
                                }} style={{alignItems:"center", gap:"3px",backgroundColor:"#355166", paddingTop:"3px", paddingBottom:"3px", paddingLeft:"6px", paddingRight:"6px", cursor:"pointer", borderRadius:"3px"}}>
                                    Undo
                                </div>
                            </div>
                            <div style={{display:"flex", flex:1, justifyContent:"center"}}>
                                <div onClick={async ()=>{
                                    await executeWorkflow()
                                }} style={{alignItems:"center", gap:"3px",backgroundColor:"#355166", paddingTop:"3px", paddingBottom:"3px", paddingLeft:"6px", paddingRight:"6px", cursor:"pointer", borderRadius:"3px"}}>
                                    Run
                                </div>
                            </div>
                            <div style={{display:"flex", flex:1, gap :"3px", justifyContent:"flex-end", paddingRight:"10px"}}>
                                <div onClick={async()=>{
                                    await updateWorkflow(workflow)
                                }} style={{alignItems:"center", gap:"3px",backgroundColor:"#355166", paddingTop:"3px", paddingBottom:"3px", paddingLeft:"6px", paddingRight:"6px", cursor:"pointer", borderRadius:"3px"}}>
                                    Save
                                </div>
                                <div onClick={async()=>{
                                    await saveWorkflow()
                                }} style={{alignItems:"center", gap:"3px",backgroundColor:"#355166", paddingTop:"3px", paddingBottom:"3px", paddingLeft:"6px", paddingRight:"6px", cursor:"pointer", borderRadius:"3px"}}>
                                    Save as new revision
                                </div>
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

function SettingsTab(props) {

    const {namespace, workflow} = props

    return (
        <>
            <FlexBox className="gap wrap col">
                <div style={{width: "100%", minHeight: "200px"}}>
                    <AddWorkflowVariablePanel namespace={namespace} workflow={workflow} />
                </div>
                <FlexBox className="gap wrap">
                    <FlexBox style={{maxHeight: "200px", flexGrow: "5"}}>
                        <div style={{width: "100%", minHeight: "200px"}}>
                            <ContentPanel style={{width: "100%", height: "100%"}}>
                                <ContentPanelTitle>
                                    <ContentPanelTitleIcon>
                                        <IoMdLock/>
                                    </ContentPanelTitleIcon>
                                    Add Attributes
                                    <ContentPanelHeaderButton>
                                        + Add
                                    </ContentPanelHeaderButton>
                                </ContentPanelTitle>
                                <ContentPanelBody> 

                                </ContentPanelBody>
                            </ContentPanel>
                        </div>
                    </FlexBox>
                    <FlexBox style={{maxHeight: "200px", flexGrow: "1"}}>          
                        <div style={{width: "100%", minHeight: "200px"}}>
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
                                        <FlexBox className="gap" style={{
                                            alignItems: "center",
                                            justifyContent: "center"
                                        }}>
                                            <label className="switch">
                                                <input type="checkbox" />
                                                <span className="slider-broadcast"></span>
                                            </label>
                                            <div className="rev-toggle-label hide-on-small">
                                                Enabled
                                            </div>                          
                                        </FlexBox>
                                        <input style={{width: "auto", maxWidth: "200px"}} type="text" placeholder="Message" />
                                    </FlexBox>
                                </ContentPanelBody>
                            </ContentPanel>
                        </div>
                    </FlexBox>
                </FlexBox>
            </FlexBox>
        </>
    )

}

