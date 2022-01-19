import React, {useEffect, useState} from 'react';
import './style.css';

import ContentPanel, { ContentPanelBody, ContentPanelHeaderButton, ContentPanelHeaderButtonIcon, ContentPanelTitle, ContentPanelTitleIcon } from '../../components/content-panel';
import FlexBox from '../../components/flexbox';
import { VscAdd, VscClose,  VscSearch, VscEdit, VscTrash, VscFolderOpened } from 'react-icons/vsc';
import { Config, GenerateRandomKey } from '../../util';
import { FiFolder } from 'react-icons/fi';
import { FcWorkflow } from 'react-icons/fc';
import { HiOutlineTrash } from 'react-icons/hi';
import { useNodes } from 'direktiv-react-hooks';
import { useNavigate, useParams } from 'react-router';
import Modal, {ButtonDefinition, KeyDownDefinition} from '../../components/modal'
import DirektivEditor from '../../components/editor';
import { BsCodeSlash } from 'react-icons/bs';
import Button from '../../components/button';
import HelpIcon from "../../components/help"
import Loader from '../../components/loader';
import WorkflowPage from './workflow';
import { useSearchParams } from 'react-router-dom';
import WorkflowRevisions from './workflow/revision';
import WorkflowPod from './workflow/pod'
import { AutoSizer } from 'react-virtualized';


function Explorer(props) {
    const params = useParams()
    const [searchParams] = useSearchParams() // removed 'setSearchParams' from square brackets (this should not affect anything: search 'destructuring assignment')

    const {namespace}  = props
    let filepath = `/`
    if(!namespace){
        return <></>
    }
    if(params["*"] !== undefined){
        filepath = `/${params["*"]}`
    }

    // pod revisions
    if (searchParams.get('function') && searchParams.get('version') && searchParams.get('revision')){
        return (
            <WorkflowPod filepath={filepath} namespace={namespace} service={searchParams.get('function')} version={searchParams.get('version')} revision={searchParams.get('revision')}/>
        )
    }
    // service revisions
    if (searchParams.get('function') && searchParams.get('version')){
        return(
            <WorkflowRevisions filepath={filepath} namespace={namespace} service={searchParams.get('function')} version={searchParams.get('version')}/>
        )
    }

    return(
        <>
            <ExplorerList  namespace={namespace} path={filepath}/>
        </>
    )
}

export default Explorer;

function SearchBar(props) {
    return(
        <div className="explorer-searchbar">
            <FlexBox className="" style={{height: "29px"}}>
                <VscSearch className="auto-margin" />
                <input placeholder={"Search items"} style={{ boxSizing: "border-box" }}></input>
            </FlexBox>
        </div>
    );
}

const orderFieldDictionary = {
    "Name": "NAME", // Default
    "Created": "CREATED"
}

const orderFieldKeys = Object.keys(orderFieldDictionary)

function ExplorerList(props) {
    const {namespace, path} = props
    const navigate= useNavigate()
    
    const [currPath, setCurrPath] = useState("")
    
    const [name, setName] = useState("")
    const [load, setLoad] = useState(true)

    const [orderFieldKey, setOrderFieldKey] = useState(orderFieldKeys[0])

    const [wfData, setWfData] = useState("")
    const [wfTemplate, setWfTemplate] = useState("")
    // const [pageNo, setPageNo] = useState(1);

    const {data, err, templates, pageInfo, createNode, deleteNode, renameNode } = useNodes(Config.url, true, namespace, path, localStorage.getItem("apikey"), `order.field=${orderFieldDictionary[orderFieldKey]}`)

    // control loading icon todo work out how to display this error
    useEffect(()=>{
        if(data !== null || err !== null) {
            setLoad(false)
        }
    },[data, err])

    useEffect(()=>{
        if(path !== currPath) {
            setCurrPath(path)
            setLoad(true)
        }
    },[path, currPath])

    if(data !== null) {
        if(data.node.type === "workflow") {
            return <WorkflowPage namespace={namespace}/>
        }
    }
    

    return(
        <FlexBox className="col gap"  style={{paddingRight: "8px"}}>
        <Loader load={load} timer={1000}>
        <FlexBox className="gap" style={{maxHeight: "32px"}}>
            <FlexBox>
                <Button className="small light" style={{ display: "flex", minWidth: "120px" }}>
                    <ContentPanelHeaderButtonIcon>
                        <BsCodeSlash style={{ maxHeight: "12px", marginRight: "4px" }} />
                    </ContentPanelHeaderButtonIcon>
                    API Commands
                </Button>
            </FlexBox>
            <FlexBox style={{flexDirection: "row-reverse"}}>
                <SearchBar />
            </FlexBox>
        </FlexBox>
        <ContentPanel>
            <ContentPanelTitle>
                <ContentPanelTitleIcon>
                    <VscFolderOpened/>
                </ContentPanelTitleIcon>
                <FlexBox style={{display:"flex", alignItems:"center"}} className="gap">
                    <div>
                        Explorer
                    </div>
                    <HelpIcon msg={"Directory/workflow browser."} />
                </FlexBox>
                <FlexBox className="gap" style={{flexDirection: "row-reverse"}}>
                    <ContentPanelHeaderButton className="explorer-action-btn">
                        <Modal title="New Workflow" 
                            modalStyle={{height: "90vh"}}
                            escapeToCancel
                            button={(
                                <div style={{display:"flex"}}>
                                    <ContentPanelHeaderButtonIcon>
                                        <VscAdd/>
                                    </ContentPanelHeaderButtonIcon>
                                    <span className="hide-on-small">Workflow</span>
                                    <span className="hide-on-medium-and-up">WF</span>
                                </div>
                            )}  
                            onClose={()=>{
                                setWfData("")
                                setWfTemplate("")
                                setName("")
                            }}
                            actionButtons={[
                                ButtonDefinition("Add", async () => {
                                    const result = await createNode(name, "workflow", wfData)
                                    if(result.node && result.namespace){
                                        navigate(`/n/${result.namespace}/explorer/${result.node.path.substring(1)}`)
                                    }
                                }, `small blue ${(name.trim() && wfTemplate) ? "" : "disabled"}`, ()=>{}, true, false),
                                ButtonDefinition("Cancel", () => {
                                }, "small light", ()=>{}, true, false)
                            ]}

                            keyDownActions={[
                                KeyDownDefinition("Enter", async () => {
                                    if(name.trim() && wfTemplate) {
                                        await createNode(name, "workflow", wfData)
                                    } else {
                                        throw new Error("Please fill in name and choose template")
                                    }
                                }, ()=>{}, true, "workflow-name")
                            ]}
                        >
                            <FlexBox className="col gap" style={{fontSize: "12px", minHeight: "300px", minWidth: "550px"}}>
                                <div style={{width: "100%", paddingRight: "12px", display: "flex"}}>
                                    <input id={"workflow-name"} value={name} onChange={(e)=>setName(e.target.value)} autoFocus placeholder="Enter workflow name"/>
                                </div>
                                <select value={wfTemplate} onChange={(e)=>{
                                    setWfTemplate(e.target.value)
                                    // todo set wfdata to template on change
                                    setWfData(templates[e.target.value])
                                }}>
                                    <option value="" >Choose a workflow template...</option>
                                    {Object.keys(templates).map((obj)=>{
                                        let key = GenerateRandomKey("")
                                        return(
                                            <option key={key} value={obj}>{obj}</option>
                                        )
                                    })}
                                </select>
                                <FlexBox className="gap">
                                    <FlexBox style={{overflow:"hidden"}}>
                                    <AutoSizer>
                                        {({height, width})=>(
                                        <DirektivEditor dlang={"yaml"} width={width} value={wfData} setDValue={setWfData} height={height}/>
                                        )}
                                    </AutoSizer>
                                    </FlexBox>
                                </FlexBox>
                            </FlexBox>
                        </Modal>
                    </ContentPanelHeaderButton>
                    <ContentPanelHeaderButton className="explorer-action-btn">
                        <div>
                            <Modal title="New Directory" 
                                escapeToCancel
                                button={(
                                    <div style={{display:"flex"}}>
                                        <ContentPanelHeaderButtonIcon>
                                            <VscAdd/>
                                        </ContentPanelHeaderButtonIcon>
                                        <span className="hide-on-small">Directory</span>
                                        <span className="hide-on-medium-and-up">Dir</span>
                                    </div>
                                )}  
                                onClose={()=>{
                                    setName("")
                                
                                }}
                                actionButtons={[
                                    ButtonDefinition("Add", async () => {
                                        await createNode(name, "directory")
                                    }, `small blue ${name.trim() ? "" : "disabled"}`, ()=>{}, true, false),
                                    ButtonDefinition("Cancel", () => {
                                    }, "small light", ()=>{}, true, false)
                                ]}

                                keyDownActions={[
                                    KeyDownDefinition("Enter", async () => {
                                        if(name.trim()) {
                                            await createNode(name, "directory")
                                        } else {
                                            throw new Error("Please enter directory name")
                                        }
                                        setName("")
                                    }, ()=>{}, true)
                                ]}

                            >
                                <FlexBox  className="col gap" style={{fontSize: "12px"}}>
                                    <div style={{width: "100%", paddingRight: "12px", display: "flex"}}>
                                        <input value={name} onChange={(e)=>setName(e.target.value)} autoFocus placeholder="Enter a directory name" />
                                    </div>
                                </FlexBox>
                            </Modal>
                        </div>
                    </ContentPanelHeaderButton>
                    <div className="explorer-sort-by explorer-action-btn hide-on-small">
                    <FlexBox className="gap" style={{marginRight: "8px"}}>
                        <FlexBox className="center">
                            Sort by:
                        </FlexBox>
                        <FlexBox className="center">
                            <select onChange={(e)=>{
                                setOrderFieldKey(e.target.value)
                                }} value={orderFieldKey} className="dropdown-select" style={{paddingBottom: "0px", paddingTop: "0px", height:"27px"}}>
                                <option value="">{orderFieldKey}</option>
                                {orderFieldKeys.map((key)=>{
                                    if(key === orderFieldKey){
                                        return ""
                                    }
                                    return(
                                        <option key={GenerateRandomKey()} value={key}>{key}</option>
                                    )
                                })}
                            </select>
                        </FlexBox>
                        </FlexBox>
                    </div>
                </FlexBox>
            </ContentPanelTitle>
            <ContentPanelBody style={{height:"100%"}}>
                    <FlexBox className="col">
                        {data !== null ? <>
                        {data.children.edges.length === 0 ? 
                                <div className="explorer-item">
                                    <FlexBox className="explorer-item-container">
                                        <FlexBox style={{display:"flex", alignItems:"center"}} className="explorer-item-icon">
                                            <VscSearch />
                                        </FlexBox>
                                        <FlexBox style={{fontSize:"10pt"}} className="explorer-item-name">
                                            No results found under '{path}'.
                                        </FlexBox>
                                        <FlexBox className="explorer-item-actions gap">
                        
                                        </FlexBox>
                                    </FlexBox>
                                </div>
                        :
                        <>
                        {data.children.edges.map((obj) => {
                            if (obj.node.type === "directory") {
                                return (<DirListItem namespace={namespace} renameNode={renameNode} deleteNode={deleteNode} path={obj.node.path} key={GenerateRandomKey("explorer-item-")} name={obj.node.name} />)
                            } else if (obj.node.type === "workflow") {
                                return (<WorkflowListItem namespace={namespace} renameNode={renameNode} deleteNode={deleteNode} path={obj.node.path} key={GenerateRandomKey("explorer-item-")} name={obj.node.name} />)
                            }
                            return <></>
                        })}</>}</>: <></>}
                    </FlexBox>
            </ContentPanelBody>
        </ContentPanel>
    {/* <FlexBox>
        <Pagination max={10} currentIndex={pageNo} pageNoSetter={setPageNo} />
    </FlexBox> */}
    </Loader>
  
    </FlexBox>
    )
}

function DirListItem(props) {

    let {name, path, deleteNode, renameNode, namespace} = props;

    const navigate = useNavigate()
    const [renameValue, setRenameValue] = useState(path)
    const [rename, setRename] = useState(false)
    const [err, setErr] = useState("")


    return(
        <div style={{cursor:"pointer"}} onClick={(e)=>{
            navigate(`/n/${namespace}/explorer/${path.substring(1)}`)
        }} className="explorer-item">
            <FlexBox className="explorer-item-container">
                <FlexBox className="explorer-item-icon">
                    <FiFolder className="auto-margin" />
                </FlexBox>
                {
                    rename ? 
                    <FlexBox className="explorer-item-name">
                        <input onClick={(ev)=>ev.stopPropagation()} type="text" value={renameValue} onKeyPress={async (e)=>{
                            if(e.key === "Enter"){
                                try { 
                                    await renameNode("/", path, renameValue)
                                    setRename(!rename)
                                } catch(err) {
                                    setErr(err.message)
                                }
                            }
                        }} onChange={(e)=>setRenameValue(e.target.value)} autoFocus style={{maxWidth:"300px", height:"38px"}}/>
                        {err !== "" ? 
                        <span>{err}</span>
                        :""
                        }
                     </FlexBox>
                    :
                    <FlexBox className="explorer-item-name">
                        {name}
                    </FlexBox>
                }
                
                <FlexBox className="explorer-item-actions gap">
                {rename ? 
                    <FlexBox onClick={(ev)=>{
                        setRename(!rename)
                        setErr("")
                        ev.stopPropagation()
                    }}>
                        <VscClose className="auto-margin" />
                    </FlexBox>
                    :
                    <FlexBox onClick={(ev)=>{
                        setRename(!rename)
                        setErr("")
                        ev.stopPropagation()
                    }}>
                        <VscEdit className="auto-margin" />
                    </FlexBox>}
                    <FlexBox onClick={(ev)=>ev.stopPropagation()}>

                    <Modal
                            escapeToCancel
                            style={{
                                flexDirection: "row-reverse",
                            }}
                            title="Delete a directory" 
                            button={(
                                <FlexBox>
                                    <VscTrash className="auto-margin red-text" />
                                </FlexBox>
                            )}
                            actionButtons={
                                [
                                    ButtonDefinition("Delete", async () => {
                                        let p = path.split('/', -1);
                                        let pLast = p[p.length-1];
                                        await deleteNode(pLast)
                                    }, "small red", ()=>{}, true, false),
                                    ButtonDefinition("Cancel", () => {
                                    }, "small light", ()=>{}, true, false)
                                ]
                            } 
                        >
                                <FlexBox className="col gap">
                            <FlexBox >
                                Are you sure you want to delete '{name}'?
                                <br/>
                                This action cannot be undone.
                            </FlexBox>
                        </FlexBox>
                    </Modal>
                    </FlexBox>

                </FlexBox>
            </FlexBox>
        </div>
    )
}

function WorkflowListItem(props) {

    let {name, path, deleteNode, renameNode, namespace} = props;

    const navigate= useNavigate()
    const [renameValue, setRenameValue] = useState(path)
    const [rename, setRename] = useState(false)
    const [err, setErr] = useState("")

    return(
        <div style={{cursor:"pointer"}} onClick={()=>{
            navigate(`/n/${namespace}/explorer/${path.substring(1)}`)
        }} className="explorer-item">
            <FlexBox className="explorer-item-container">
                <FlexBox className="explorer-item-icon">
                    <FcWorkflow className="auto-margin" />
                </FlexBox>
                {
                    rename ? 
                    <FlexBox className="explorer-item-name">
                        <input onClick={(ev)=>ev.stopPropagation()} type="text" value={renameValue} onKeyPress={async (e)=>{
                            if(e.key === "Enter"){
                                try { 
                                    await renameNode("/", path, renameValue)
                                    setRename(!rename)
                                } catch(err) {
                                    setErr(err.message)
                                }
                            }
                        }} onChange={(e)=>setRenameValue(e.target.value)} autoFocus style={{maxWidth:"300px", height:"38px"}}/>
                        {err !== "" ? 
                        <span>{err}</span>
                        :""
                        }
                     </FlexBox>
                    :
                    <FlexBox className="explorer-item-name">
                        {name}
                    </FlexBox>
                }
                
                <FlexBox className="explorer-item-actions gap">
                    {rename ? 
                    <FlexBox onClick={(ev)=>{
                        setRename(!rename)
                        setErr("")
                        ev.stopPropagation()
                    }}>
                        <VscClose className="auto-margin" />
                    </FlexBox>
                    :
                    <FlexBox onClick={(ev)=>{
                        setRename(!rename)
                        setErr("")
                        ev.stopPropagation()
                    }}>
                        <VscEdit className="auto-margin" />
                    </FlexBox>}
                    <FlexBox onClick={(ev)=>ev.stopPropagation()}>

                        <Modal
                                escapeToCancel
                                style={{
                                    flexDirection: "row-reverse",
                                }}
                                title="Delete a workflow" 
                                button={(
                                    <FlexBox style={{alignItems:'center'}}>
                                        <HiOutlineTrash className="auto-margin red-text" />
                                    </FlexBox>
                                )}
                                actionButtons={
                                    [
                                        ButtonDefinition("Delete", async () => {
                                            let p = path.split('/', -1);
                                            let pLast = p[p.length-1];
                                            await deleteNode(pLast)
                                        }, "small red", ()=>{}, true, false),
                                        ButtonDefinition("Cancel", () => {
                                        }, "small light", ()=>{}, true, false)
                                    ]
                                } 
                            >
                                    <FlexBox className="col gap">
                                <FlexBox >
                                    Are you sure you want to delete '{name}'?
                                    <br/>
                                    This action cannot be undone.
                                </FlexBox>
                            </FlexBox>
                            </Modal>
                </FlexBox>
                </FlexBox>
            </FlexBox>
        </div>
    )
}
