import React, { useEffect, useState } from 'react';
import './style.css';
import FlexBox from '../../../components/flexbox';
import {Link} from 'react-router-dom'
import ContentPanel, { ContentPanelBody, ContentPanelHeaderButton, ContentPanelTitle, ContentPanelTitleIcon } from '../../../components/content-panel';
import {BsCodeSquare} from 'react-icons/bs'
import { useWorkflow, useWorkflowServices, useWorkflowVariables } from 'direktiv-react-hooks';
import { Config } from '../../../util';
import { useParams } from 'react-router';

import * as dayjs from "dayjs"
import relativeTime from "dayjs/plugin/relativeTime";
import utc from "dayjs/plugin/utc"
import { InstanceRow } from '../../instances';
import { IoMdLock } from 'react-icons/io';
import { ServiceStatus } from '../../namespace-services';
import Modal, { ButtonDefinition } from '../../../components/modal';
import AddValueButton from '../../../components/add-button';
import DirektivEditor from '../../../components/editor';
import { VariableFilePicker } from '../../settings/variables-panel';
import Tabs from '../../../components/tabs';
dayjs.extend(utc)
dayjs.extend(relativeTime);



function WorkflowPage(props) {
    const {namespace} = props
    const params = useParams()
    let filepath = "/"

    if(!namespace) {
        return ""
    }

    if(params["*"] !== undefined){
        filepath = `/${params["*"]}`
    }

    return(
        <InitialWorkflowHook namespace={namespace} filepath={filepath}/>
    )
}

function InitialWorkflowHook(props){
    const {namespace, filepath} = props

    const [activeTab, setActiveTab] = useState(0)

    const {data, err, getInstancesForWorkflow, getRevisions} = useWorkflow(Config.url, true, namespace, filepath)
    
    console.log(data, err, "test")
    
    return(
        <>
            <FlexBox id="workflow-page" className="gap col" style={{paddingRight: "8px"}}>
                <TabBar activeTab={activeTab} setActiveTab={setActiveTab} />
                <FlexBox className="col gap">
                    { activeTab === 0 ? 
                        <OverviewTab namespace={namespace} getInstancesForWorkflow={getInstancesForWorkflow} filepath={filepath}/>
                    :<></>}
                    { activeTab === 1 ?
                        <RevisionSelectorTab />
                    :<></>}
                    { activeTab === 4 ?
                        <SettingsTab namespace={namespace} workflow={filepath} />
                    :<></>}
                </FlexBox>
            </FlexBox>
        </>
    )
}

export default WorkflowPage;

function TabBar(props) {

    let {activeTab, setActiveTab} = props;
    let tabLabels = [
        "Overview",
        "Revisions",
        "Working Revisions",
        "Dependency Graph", 
        "Settings"
    ]

    let tabDOMs = [];
    for (let i = 0; i < 5; i++) {

        let className = "tab-bar-item"
        if (i === activeTab) {
            className += " active"
        }

        tabDOMs.push(
            <FlexBox className={className} onClick={() => {
                setActiveTab(i)
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
                <input type="checkbox" />
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
            <div>
        {
            instances !== null && instances.length === 0 ? <div style={{paddingLeft:"10px", fontSize:"10pt"}}>No instances have been recently executed. Recent instances will appear here.</div>:
            <table className="instances-table">

     <>       <thead>
                <tr>
                    <th>
                        State
                    </th>
                    <th>
                        Name
                    </th>
                    <th>
                        Started <span className="hide-on-small">at</span>
                    </th>
                    <th>
                        <span className="hide-on-small">Last</span> Updated
                    </th>
                </tr>
            </thead>
            <tbody>
                {instances !== null ? 
                <>
                    <>
                    {instances.map((obj)=>{
                    return(
                        <InstanceRow
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
            </tbody></>
        </table>}
            </div>
            </>
        </ContentPanelBody>
    )
}

function OverviewTab(props) {
    const {getInstancesForWorkflow,  namespace, filepath} = props

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
                    console.log(resp)
                } else {
                    setErr(resp)
                }

            }
            setLoad(false)
        }
        listData()
    },[load, getInstancesForWorkflow])

    console.log(err, "FETCHING INSTANCES OR REVISIONS")

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

function WorkflowServices(props) {
    const {namespace, filepath} = props

    const {data, err} = useWorkflowServices(Config.url, true, namespace, filepath.substring(1))
    console.log(data, err)
    if (data === null) {
        return ""
    }

    return(
        <ContentPanelBody>
            <ul style={{listStyle:"none", margin:0, paddingLeft:"10px"}}>
                {data.map((obj)=>{
                    console.log(obj)
                    return(
                        <Link to={`/n/${namespace}/explorer/${filepath.substring(1)}?function=${obj.info.name}&version=${obj.info.revision}`}>
                            <li style={{display:"flex", alignItems:'center', gap :"10px"}}>
                                <ServiceStatus status={obj.status}/>
                                {obj.info.name}({obj.info.image})
                            </li>
                        </Link>
                    )
                })}
            </ul>
        </ContentPanelBody>
    )
}
 

function RevisionSelectorTab(props) {
    return(
        <>
            <FlexBox>
                <ContentPanel style={{ width: "100%", minWidth: "300px"}}>
                    <ContentPanelTitle>
                        <ContentPanelTitleIcon>
                            <BsCodeSquare />
                        </ContentPanelTitleIcon>
                        <div>
                            Revision name 001
                        </div>
                        <FlexBox style={{justifyContent: "end", paddingRight: "8px"}}>
                            <div>
                                <FlexBox className="revision-panel-btn-bar">
                                    <div>Editor</div>
                                    <div>Diagram</div>
                                    <div>Sankey</div>
                                </FlexBox>
                            </div>
                        </FlexBox>
                    </ContentPanelTitle>
                </ContentPanel>
            </FlexBox>
        </>
    )
}

function SettingsTab(props) {

    const {namespace, workflow} = props
    const [keyValue, setKeyValue] = useState("")
    const [dValue, setDValue] = useState("")
    const [file, setFile] = useState(null)
    const [uploading, setUploading] = useState(false)
    const [mimeType, setMimeType] = useState("application/json")
    const {data, err, setWorkflowVariable, getWorkflowVariable, deleteWorkflowVariable} = useWorkflowVariables(Config.url, true, namespace, workflow, localStorage.getItem("apikey"))

    
    let uploadingBtn = "small green"
    if (uploading) {
        uploadingBtn += " btn-loading"
    }
    
    return (
        <>
            <FlexBox className="gap wrap col">
                <div style={{width: "100%", minHeight: "200px"}}>
                    <ContentPanel style={{width: "100%", height: "100%"}}>
                        <ContentPanelTitle>
                            <ContentPanelTitleIcon>
                                <IoMdLock/>
                            </ContentPanelTitleIcon>
                            Variables
                                <Modal title="New variable" 
                                    escapeToCancel
                                    button={(
                                        <AddValueButton label=" " />
                                    )}  
                                    onClose={()=>{
                                        setKeyValue("")
                                        setDValue("")
                                        setFile(null)
                                        setUploading(false)
                                        setMimeType("application/json")
                                    }}
                                    actionButtons={[
                                        ButtonDefinition("Add", async () => {
                                            if(document.getElementById("file-picker")){
                                                setUploading(true)
                                                if(keyValue === "") {
                                                    setUploading(false)
                                                    return "Variable key name needs to be provided."
                                                }
                                                let err = await setWorkflowVariable(keyValue, file, mimeType)
                                                if (err) {
                                                    setUploading(false)
                                                    return err
                                                }
                                            } else {
                                                if(keyValue === "") {
                                                    setUploading(false)
                                                    return "Variable key name needs to be provided."
                                                }
                                                let err = await setWorkflowVariable(keyValue, dValue, mimeType)
                                                if (err) return err
                                            }
                                        }, uploadingBtn, true, false),
                                        ButtonDefinition("Cancel", () => {
                                        }, "small light", true, false)
                                    ]}
                                >
                                    <AddVariablePanel mimeType={mimeType} setMimeType={setMimeType} file={file} setFile={setFile} setKeyValue={setKeyValue} keyValue={keyValue} dValue={dValue} setDValue={setDValue}/>
                                </Modal>
                        </ContentPanelTitle>
                        <ContentPanelBody>

                        </ContentPanelBody>
                    </ContentPanel>
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
                                <ContentPanelBody>
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
                                </ContentPanelBody>
                            </ContentPanel>
                        </div>
                    </FlexBox>
                </FlexBox>
            </FlexBox>
        </>
    )

}

function AddVariablePanel(props) {
    const {keyValue, setKeyValue, dValue, setDValue, file, setFile, mimeType, setMimeType} = props

    let lang = ""

    switch(mimeType){
    case "application/json":
        lang = "json"
        break
    case "application/x-sh":
        lang = "shell"
        break
    case "text/html":
        lang = "html"
        break
    case "text/css":
        lang = "css"
        break
    case "application/yaml":
        lang = "yaml"
        break
    default:
        lang = "plain"
    }

    return(
        <Tabs
            style={{minHeight: "400px", minWidth: "90%"}}
            headers={["Manual", "Upload"]}
            tabs={[(
                <FlexBox id="written" className="col gap" style={{fontSize: "12px"}}>
                    <div style={{width: "100%", paddingRight: "12px", display: "flex"}}>
                        <input value={keyValue} onChange={(e)=>setKeyValue(e.target.value)} autoFocus placeholder="Enter variable key name" />
                    </div>
                    <div style={{width: "100%", paddingRight: "12px", display: "flex"}}>
                        <select style={{width:"100%"}} defaultValue={mimeType} onChange={(e)=>setMimeType(e.target.value)}>
                            <option value="">Choose a mimetype</option>
                            <option value="application/json">json</option>
                            <option value="application/yaml">yaml</option>
                            <option value="application/x-sh">shell</option>
                            <option value="text/plain">plaintext</option>
                            <option value="text/html">html</option>
                            <option value="text/css">css</option>
                        </select>
                    </div>
                    <FlexBox className="gap" style={{maxHeight: "600px"}}>
                        <FlexBox style={{overflow:"hidden"}}>
                            <DirektivEditor dlang={lang} width={"600px"} dvalue={dValue} setDValue={setDValue} height={"600px"}/>
                        </FlexBox>
                    </FlexBox>
                </FlexBox>
            ),(
                <FlexBox id="file-picker" className="col gap" style={{fontSize: "12px"}}>
                    <div style={{width: "100%", paddingRight: "12px", display: "flex"}}>
                        <input value={keyValue} onChange={(e)=>setKeyValue(e.target.value)} autoFocus placeholder="Enter variable key name" />
                    </div>
                    <FlexBox className="gap">
                        <VariableFilePicker setKeyValue={setKeyValue} setMimeType={setMimeType} mimeType={mimeType} file={file} setFile={setFile} id="add-variable-panel" />
                    </FlexBox>
                </FlexBox>
            )]}
        />
    )
}

