import React, { useCallback, useEffect, useState } from 'react';
import Button from '../../../components/button';
import {HiOutlineTrash} from 'react-icons/hi';
import ContentPanel, { ContentPanelBody, ContentPanelHeaderButtonIcon, ContentPanelTitle, ContentPanelTitleIcon } from '../../../components/content-panel';
import FlexBox from '../../../components/flexbox';
import {GenerateRandomKey} from '../../../util';
import {BiChevronLeft} from 'react-icons/bi';
import DirektivEditor from '../../../components/editor';
import WorkflowDiagram from '../../../components/diagram';
import YAML from 'js-yaml'
import Modal, { ButtonDefinition, ModalHeadless } from '../../../components/modal';
import SankeyDiagram from '../../../components/sankey';
import { VscVersions, VscTypeHierarchySub } from 'react-icons/vsc'
import Slider from 'rc-slider';
import 'rc-slider/assets/index.css';
import { useNavigate } from 'react-router';
import HelpIcon from "../../../components/help";
import { AutoSizer } from 'react-virtualized';
import { VscCode, VscDebugStepBack } from 'react-icons/vsc';
import { ApiFragment } from '..';

import Form from "@rjsf/core";
import  Tabs  from '../../../components/tabs';
import Tippy from '@tippyjs/react';
import { windowBlocker } from '../../../components/diagram-editor/usePrompt';

function RevisionTab(props) {

    const navigate = useNavigate()
    const {searchParams, setSearchParams, revision, setRevision, getWorkflowRevisionData, getWorkflowSankeyMetrics, executeWorkflow, namespace} = props
    const [load, setLoad] = useState(true)
    const [workflow, setWorkflowData] = useState(null)
    const [revisionID, setRevisionID] = useState(null)
    const [tabBtn, setTabBtn] = useState(searchParams.get('revtab') !== null ? parseInt(searchParams.get('revtab')): 0);
    const [input, setInput] = useState("{\n\t\n}")

    const [tabIndex, setTabIndex] = useState(0)
    const [workflowJSONSchema, setWorkflowJSONSchema] = useState(null)
    const [inputFormSubmitRef, setInputFormSubmitRef] = useState(null)

    useEffect(()=>{
        if(searchParams.get('revtab') === null) {
            setSearchParams({
                tab: searchParams.get('tab'),
                revision: revision,
                revtab: 0
            })
        }
    },[searchParams, revision, setSearchParams])

    useEffect(()=>{
        async function getRevWorkflow() {
            if(load && searchParams.get('revtab') !== null) {
                let wfdata = await getWorkflowRevisionData(revision)
                setWorkflowData(atob(wfdata.revision.source))
                setRevisionID(wfdata.revision.name)
                setLoad(false)
            }
        }
        getRevWorkflow()
    },[load, searchParams, getWorkflowRevisionData, revision])

    return(
        <FlexBox>
            <FlexBox className="col gap">
                <FlexBox  style={{maxHeight:"32px"}}>
                    <Button onClick={()=>{
                        setRevision(null)
                        setSearchParams({
                            tab: searchParams.get('tab')
                        })
                    }} className="small light" style={{ minWidth: "160px", maxWidth: "160px" }}>
                        <FlexBox className="gap" style={{ alignItems: "center", justifyContent: "center" }}>
                            <BiChevronLeft style={{ fontSize: "16px" }} />
                            <div>Back to All Revisions</div>
                        </FlexBox>
                    </Button>
                </FlexBox>
                <FlexBox>
                <ContentPanel style={{ width: "100%", minWidth: "300px", flex: 1}}>
                    <ContentPanelTitle>
                        <ContentPanelTitleIcon>
                            <VscVersions />
                        </ContentPanelTitleIcon>
                        <div>
                        {revision === revisionID ? revision : `${revision} => ${revisionID}`}
                        </div>
                        <TabbedButtons revision={revision} setSearchParams={setSearchParams} searchParams={searchParams} tabBtn={tabBtn} setTabBtn={setTabBtn} />
                    </ContentPanelTitle>
                    <ContentPanelBody style={{padding: "0px"}}>
                        {tabBtn === 0 ?
                            <FlexBox className="col" style={{overflow:"hidden"}}>
                                <FlexBox >
                                    <DirektivEditor style={{borderRadius: "0px"}} value={workflow} readonly={true} disableBottomRadius={true} dlang="yaml" />
                                </FlexBox>
                                <FlexBox className="gap editor-footer" style={{borderTop:"1px solid white", overflow: "hidden"}}>
                                    <div style={{display:"flex", flex:1 }}>
                                    </div>
                                    <div style={{display:"flex", flex:1, justifyContent:"center"}}>
                                        <Modal 
                                            style={{ justifyContent: "center" }}
                                            className="run-workflow-modal"
                                            modalStyle={{color: "black", width: "600px", minWidth:"30vw"}}
                                            title="Run Workflow"
                                            onClose={()=>{
                                                setInput("{\n\t\n}")
                                                setTabIndex(0)
                                                setWorkflowJSONSchema(null)
                                            }}
                                            actionButtons={[
                                                ButtonDefinition(`${tabIndex === 0 ? "Run": "Generate JSON"}`, async () => {
                                                    if (tabIndex === 1) {
                                                        inputFormSubmitRef.click()
                                                        return
                                                    }
                                                    let r = ""
                                                    r = await executeWorkflow(input, revision)
                                                    if(r.includes("execute workflow")){
                                                        // is an error
                                                        throw new Error(r)
                                                    } else {
                                                        navigate(`/n/${namespace}/instances/${r}`)
                                                    }
                                                }, `small ${tabIndex === 1 && workflowJSONSchema === null ? "disabled" : ""}`, ()=>{}, tabIndex === 0, false),
                                                ButtonDefinition("Cancel", async () => {
                                                }, "small light", ()=>{}, true, false)
                                            ]}
                                            onOpen={()=>{
                                                let wfObj =  YAML.load(workflow)
                                                if (wfObj && wfObj.states && wfObj.states.length > 0 && wfObj.states[0].type === "validate") {
                                                    setWorkflowJSONSchema(  wfObj.states[0].schema)
                                                    setTabIndex(1)
                                                }
                                            }}
                                            button={(
                                                <div style={{alignItems:"center", gap:"3px",backgroundColor:"#355166", paddingTop:"3px", paddingBottom:"3px", paddingLeft:"6px", paddingRight:"6px", cursor:"pointer", borderRadius:"3px"}}>
                                                    Run
                                                </div>
                                            )}
                                        >
                                        <FlexBox style={{ height: "45vh", minWidth: "250px", minHeight: "160px", overflow:"hidden" }}>
                                            <Tabs
                                                id={"wf-execute-input"}
                                                key={"inputForm"}
                                                callback={setTabIndex}
                                                tabIndex={tabIndex}
                                                style={tabIndex === 1 ? { overflowY: "auto", paddingTop: "1px" } : { paddingTop: "1px" }}
                                                headers={["JSON", "Form"]}
                                                tabs={[(
                                                    <FlexBox>
                                                        <AutoSizer>
                                                            {({ height, width }) => (
                                                                <DirektivEditor height={height} width={width} dlang="json" dvalue={input} setDValue={setInput} />
                                                            )}
                                                        </AutoSizer>
                                                    </FlexBox>
                                                ), (<FlexBox className="col" style={{ overflow: "hidden" }}>
                                                    {workflowJSONSchema === null ?
                                                        <div className='container-alert' style={{ textAlign: "center", height: "80%" }}>
                                                            Workflow first state must be a validate state to generate form.
                                                        </div> : <></>
                                                    }
                                                    <div className="formContainer">
                                                        <Form onSubmit={(form) => {
                                                            setInput(JSON.stringify(form.formData, null, 2))
                                                            setTabIndex(0)
                                                        }}
                                                            schema={workflowJSONSchema ? workflowJSONSchema : {}} >
                                                            <button ref={setInputFormSubmitRef} style={{ display: "none" }} />
                                                        </Form>
                                                    </div>
                                                </FlexBox>)]} />
                                        </FlexBox>
                                        </Modal>
                                    </div>
                                    <div style={{display:"flex", flex:1, gap :"3px", justifyContent:"flex-end", paddingRight:"10px"}}>
                                    </div>
                                </FlexBox>
                            </FlexBox>
                            :
                            ""
                        }
                        {tabBtn === 1 ? <WorkflowDiagram disabled={true} workflow={workflow}/>:""}
                        {tabBtn === 2 ? <SankeyDiagram revision={revision} getWorkflowSankeyMetrics={getWorkflowSankeyMetrics} />:""}
                    </ContentPanelBody>
                </ContentPanel>
                </FlexBox>
            </FlexBox>
        </FlexBox>
    )

}

export default RevisionTab;
export function TabbedButtons(props) {

    let {tabBtn, setTabBtn, searchParams, setSearchParams, revision, enableDiagramEditor, setBlock, block, blockMsg} = props;

    let tabBtns = [];
    let tabBtnLabels = ["YAML", "Diagram", "Sankey"];
    if (enableDiagramEditor !== undefined) {
        tabBtnLabels = ["YAML", "Editor", "Diagram", "Sankey"];
    }

    for (let i = 0; i < tabBtnLabels.length; i++) {
        let key = GenerateRandomKey();
        let classes = "tab-btn";
        if (i === tabBtn) {
            classes += " active-tab-btn"
        }

        if (tabBtnLabels[i] === "Editor" && !enableDiagramEditor) {
            classes += " disable"
            tabBtns.push(
                <FlexBox key={key} className={classes}>
                    <Tippy content={"Unsaved changes in Workflow"} trigger={'mouseenter focus click'} zIndex={10}>
                        <div>
                            {tabBtnLabels[i]}
                        </div>
                    </Tippy>
                </FlexBox>
            )
            continue

        }

        tabBtns.push(
            <FlexBox key={key} className={classes} onClick={(e) => {
                if (block) {
                    e.stopPropagation();
                    if (!windowBlocker(blockMsg)) {
                        return
                    }

                    setBlock(false)
                }

                setTabBtn(i)
                setSearchParams({
                    tab: searchParams.get('tab'),
                    revision: revision,
                    revtab: i
                })
            }}>
                <div>
                    {tabBtnLabels[i]}
                </div>
            </FlexBox>
        )
    }

    return(
            <FlexBox className="tabbed-btns-container">
                <FlexBox className="tabbed-btns" >
                    {tabBtns}
                </FlexBox>
            </FlexBox>
    )
}

const apiHelps = (namespace, workflow) => {
    let url = window.location.origin
    return(
        [
            {
                method: "GET",
                url: `${url}/api/namespaces/${namespace}/tree/${workflow}?op=wait`,
                description: `Execute a Workflow`,
            },
            {
                method: "POST",
                description: `Execute a Workflow With Body`,
                url: `${url}/api/namespaces/${namespace}/tree/${workflow}?op=wait`,
                body: `{}`,
                type: "json"
            },
            {
                method: "POST",
                description: `Update a workflow `,
                url: `${url}/api/namespaces/${namespace}/tree/${workflow}?op=update-workflow`,
                body: `description: A simple 'no-op' state that returns 'Hello world!'
states:
- id: helloworld
type: noop
transform:
    result: Hello world!`,
                type: "yaml"
            },
            {
                method: "POST",
                description: "Execute a workflow",
                url: `${url}/api/namespaces/${namespace}/tree/${workflow}?op=execute`,
                body:`{}`,
                type: "json"
            }
        ]
    )
}

export function RevisionSelectorTab(props) {
    const {workflowName, setRouter, namespace, tagWorkflow, filepath, updateWorkflow, editWorkflowRouter, getWorkflowRouter, getRevisions, setRevisions, err, revisions, router, deleteRevision, getWorkflowSankeyMetrics, executeWorkflow, executeWorkflowRouter, searchParams, setSearchParams, getWorkflowRevisionData, getTags, removeTag} = props
    
    const navigate = useNavigate()
    // const [load, setLoad] = useState(true)
    const [revision, setRevision] = useState(null)
    const [rev1, setRev1] = useState(router.routes.length === 0 ? "latest": "")
    const [rev2, setRev2] = useState("")

    const [tags, setTags] = useState(null)

    function updateTags(newTags){
        let processedTags = {}
        newTags.forEach(edge => {
            if (edge.name !== "latest") {
                processedTags[edge.name] = true
            }
        });
        setTags(processedTags)
    }

    useEffect(()=>{
        async function listData() {
            if(tags === null){
                // get workflow tags
                let resp = await getTags()
                if(Array.isArray(resp.results)){
                    updateTags(resp.results)
                } else {
                    // FIXME: find location for this error
                    console.error("could not retrive tags", resp)
                }
            }
        }
        return listData()
    },[getTags, tags])


    useEffect(()=>{
        if(searchParams.get('revision') !== null) {
            setRevision(searchParams.get('revision'))
        }
    },[searchParams])

    if (err) {
        // TODO report err
    }

    if(revision !== null) {
        return(
            <RevisionTab namespace={namespace} getWorkflowSankeyMetrics={getWorkflowSankeyMetrics} executeWorkflow={executeWorkflow} setRevision={setRevision} getWorkflowRevisionData={getWorkflowRevisionData}  searchParams={searchParams} setSearchParams={setSearchParams} revision={revision}/>
        )
    }

  
    if(!revisions) return null

    return (
        <FlexBox className="col gap">
            <div style={{maxWidth: "142px"}}>
                <Modal
                    titleIcon={<VscCode/>}
                    button={
                        <Button className="small light" style={{ display: "flex", minWidth: "120px" }}>
                            <ContentPanelHeaderButtonIcon>
                                <VscCode style={{ maxHeight: "12px", marginRight: "4px" }} />
                            </ContentPanelHeaderButtonIcon>
                            API Commands
                        </Button>
                    }
                    escapeToCancel
                    withCloseButton
                    maximised
                    title={"Namespace API Interactions"}
                >
                    {
                        apiHelps(namespace, workflowName).map((help)=>(
                            <ApiFragment
                                key={`key-${help.type}`}
                                description={help.description}
                                url={help.url}
                                method={help.method}
                                body={help.body}
                                type={help.type}
                            />
                        ))
                    }
                </Modal>
            </div>
            <div>
                <RevisionTrafficShaper rev1={rev1} rev2={rev2} setRev1={setRev1} setRev2={setRev2} setRouter={setRouter} revisions={revisions}  router={router} editWorkflowRouter={editWorkflowRouter} getWorkflowRouter={getWorkflowRouter} namespace={namespace} executeWorkflowRouter={executeWorkflowRouter}/>
            </div>
            <div>   
                <ContentPanel style={{width: "100%", minWidth: "330px"}}>
                    <ContentPanelTitle>
                        <ContentPanelTitleIcon>
                            <VscVersions/>
                        </ContentPanelTitleIcon>
                        <FlexBox style={{display:"flex", alignItems:"center"}} className="gap">
                            <div>
                                All Revisions
                            </div>
                            <HelpIcon msg={"A list of all revisions for that workflow."} />
                        </FlexBox>
                    </ContentPanelTitle>
                    <ContentPanelBody style={{flexDirection: "column"}}>
                        {revisions.map((obj) => {
                            for(var i=0; i < router.routes.length; i++) {
                                if(obj.name === router.routes[i].ref){}
                            }
                            return (
                                <FlexBox key={GenerateRandomKey()} className="gap wrap" style={{
                                    alignItems: "center"
                                }}>
                                    <FlexBox className="wrap gap" style={{
                                        flex: "4",
                                        minWidth: "300px"
                                    }}>
                                        <div>
                                            <FlexBox className="col revision-label-tuple">
                                                <div>
                                                    ID
                                                </div>
                                                <div>
                                                    {obj.name}
                                                </div>
                                            </FlexBox>
                                        </div>
                                    </FlexBox>
                                    <RevertTrafficAmount revisionName={obj.name} routes={router.routes}/>
                                    <div style={obj.name === "latest" ? {visibility: "hidden"} : null}>
                                        <FlexBox className="gap">
                                            {tags !== null && tags[obj.name] ? 
                                                <Modal
                                                    modalStyle={{width: "400px"}}
                                                    escapeToCancel
                                                    style={{
                                                        flexDirection: "row-reverse",
                                                    }}
                                                    title="Remove a Tag"
                                                    button={(
                                                        <Button className="small light bold" title="Remove Tag" >
                                                            <HiOutlineTrash className="red-text" style={{ fontSize: "16px" }} />
                                                        </Button>
                                                    )}
                                                    actionButtons={
                                                        [
                                                            ButtonDefinition("Remove", async () => {
                                                                await removeTag(obj.name)
                                                                let tagsResp = await getTags()
                                                                let revResp = await getRevisions()
                                                                setRevisions(revResp.results)
                                                                updateTags(tagsResp.results)
                                                            }, "small red", ()=>{}, true, false),
                                                            ButtonDefinition("Cancel", () => {
                                                            }, "small light", ()=>{}, true, false)
                                                        ]
                                                    }
                                                >
                                                    <FlexBox className="col gap">
                                                        <FlexBox >
                                                            Are you sure you want to remove the tag '{obj.name}'?
                                                        </FlexBox>
                                                    </FlexBox>
                                                </Modal>
                                                :
                                                <Modal
                                                    escapeToCancel
                                                    style={{
                                                        flexDirection: "row-reverse",
                                                    }}
                                                    modalStyle={{width: "400px"}}
                                                    title="Delete a revision"
                                                    button={(
                                                        <Button className="small light bold" tip="Remove Tag">
                                                            <HiOutlineTrash className="red-text" style={{ fontSize: "16px" }} />
                                                        </Button>
                                                    )}
                                                    actionButtons={
                                                        [
                                                            ButtonDefinition("Delete", async () => {
                                                                    await deleteRevision(obj.name)
                                                                    let x = await getRevisions()
                                                                    setRevisions(x.results)
                                                                    setRouter(await getWorkflowRouter())
                                                            }, "small red", ()=>{}, true, false),
                                                            ButtonDefinition("Cancel", () => {
                                                            }, "small light", ()=>{}, true, false)
                                                        ]
                                                    }
                                                >
                                                    <FlexBox className="col gap">
                                                        <FlexBox >
                                                            Are you sure you want to delete '{obj.name}'?
                                                            <br />
                                                            This action cannot be undone.
                                                        </FlexBox>
                                                    </FlexBox>
                                                </Modal>
                                            }
                                            {obj.name !== "latest" ? 
                                            <>
                                            <TagRevisionBtn tagWorkflow={tagWorkflow} obj={obj} setRevisions={setRevisions} getRevisions={getRevisions} isTag={(tags !== null && tags[obj.name])} updateTags={updateTags} getTags={getTags} />
                                            <Modal
                                                    escapeToCancel
                                                    style={{
                                                        flexDirection: "row-reverse",
                                                    }}
                                                    modalStyle={{width: "400px"}}
                                                    title={`Revert to ${obj.name}`}
                                                    button={(
                                                        <Button className="small light bold" tip="Revert revision to latest">
                                                            <VscDebugStepBack className='show-700'/>
                                                            <span className="hide-700">Revert{" "}</span>
                                                            <span className="hide-900">To</span>
                                                        </Button>
                                                    )}
                                                    actionButtons={
                                                        [
                                                            ButtonDefinition("Revert", async () => {
                                                                let data = await getWorkflowRevisionData(obj.name)
                                                                await updateWorkflow(atob(data.revision.source))
                                                                navigate(`/n/${namespace}/explorer/${filepath.substring(1)}?tab=2`)
                                                            }, "small red", ()=>{}, true, false),
                                                            ButtonDefinition("Cancel", () => {
                                                            }, "small light", ()=>{}, true, false)
                                                        ]
                                                    }
                                                >
                                                    <FlexBox className="col gap">
                                                        <FlexBox >
                                                            Are you sure you want to revert to '{obj.name}'?
                                                        </FlexBox>
                                                    </FlexBox>
                                            </Modal>
                                            <Button className="small light bold" onClick={()=>{
                                                setSearchParams({tab: 1, revision: obj.name})
                                            }}>
                                                Open{" "}<span className="hide-900">Revision</span>
                                            </Button></>
                                            : 
                                            <>
                                                {/* Hidden buttons to retain same spacing on latest */}
                                                <div style={{visibility:"hidden"}}>
                                                <Button className="small light bold" onClick={async()=>{
                                                }}>
                                                    Tag
                                                </Button>
                                                </div>
                                                <div style={{visibility:"hidden"}}>
                                                <Button className="small light bold" onClick={async()=>{
                                                }}>
                                                    Revert{" "}<span className="hide-900">To</span>
                                                </Button>
                                                </div>
                                                <div>
                                                <Button className="small light bold" onClick={()=>{
                                                }}>
                                                    Open{" "}<span className="hide-900">Revision</span>
                                                </Button></div>
                                            </>
                                            }
                                        </FlexBox>
                                    </div>
                                </FlexBox>
                            )
                        })}
                    </ContentPanelBody>
                </ContentPanel>
            </div>
    
        </FlexBox>
    )
}

function RevertTrafficAmount(props) {
    const { routes, revisionName } = props
    const TrafficAmount = useCallback(() => {
        const routeIndex = routes.findIndex(function (r) {
            return r.ref === revisionName
        })

        // Return empty element if revisionName does not exist in routes
        if (routeIndex === -1) {
            return (
                <></>
            )
        }

        const sliderClass = routeIndex === 0  ? "traffic-mini-distribution" : "traffic-mini2-distribution"

        return (
            <FlexBox className="col revision-label-tuple">
                <div>
                    Traffic amount
                </div>
                <div style={{ width: '100%' }}>
                    <Slider defaultValue={routes[routeIndex].weight} className={sliderClass} disabled={true} />
                    <div>
                        {`${routes[routeIndex].weight}%`}
                    </div>
                </div>
            </FlexBox>
        )
    }, [routes, revisionName])

    return (
        <FlexBox style={{
            flex: "1",
            maxWidth: "150px",
            minWidth: "90px"
        }}>
            <TrafficAmount/>
        </FlexBox>
        
    )
}

function TagRevisionBtn(props) {

    let {tagWorkflow, obj, getRevisions, setRevisions, updateTags, getTags} = props;
    const [tag, setTag] = useState("")

    return(
        <Modal
            escapeToCancel
            style={{
                flexDirection: "row-reverse",
            }}
            modalStyle={{width: "400px"}}
            title="Tag" 
            onClose={()=>{
                setTag("")
            }}
            button={(
                <Button className="light small bold" tip="Tag Revision">
                    <FlexBox className="gap">
                        <div>
                            Tag
                        </div>
                    </FlexBox>
                </Button>
            )}
            actionButtons={
                [
                    ButtonDefinition("Tag", async () => {
                            await tagWorkflow(obj.name, tag)
                            let tagsResp = await getTags()
                            let revResp = await getRevisions()
                            setRevisions(revResp.results)
                            updateTags(tagsResp.results)
                    }, "small", ()=>{}, true, false, true),
                    ButtonDefinition("Cancel", () => {
                    }, "small light", ()=>{}, true, false)
                ]
            } 

            requiredFields={[
                {tip: "tag is required", value: tag}
            ]}
        >
            <FlexBox>
                <input autoFocus value={tag} onChange={(e)=>setTag(e.target.value)} placeholder="Enter Tag" />
            </FlexBox>
        </Modal>
    )
}

export function RevisionTrafficShaper(props) {
    const {editWorkflowRouter, rev1, rev2, setRev1, setRev2, setRouter, getWorkflowRouter, router, revisions, namespace, executeWorkflowRouter} = props
    const navigate = useNavigate()

    const [load, setLoad] = useState(true)
    const [input, setInput] = useState("{\n\t\n}")
    const [showRunModal, setShowRunModal] = useState(false)

    const [traffic, setTraffic] = useState(router.routes.length === 0 ? 100 : 0)

    useEffect(()=>{

        if (router.routes[0]){
            setRev1(router.routes[0].ref)
            setTraffic(router.routes[1] ? router.routes[0].weight: 100)
        }

        if(router.routes[1]){
            setRev2(router.routes[1].ref)
        } else {setRev2("")}
    },[router, setRev1, setRev2])

    useEffect(()=>{
        if(load){
            if (router.routes[0]){
                setRev1(router.routes[0].ref)
                setTraffic(router.routes[1] ? router.routes[0].weight: 100)
            }

            if(router.routes[1]){
                setRev2(router.routes[1].ref)
            }
            setLoad(false)
        }
    },[load, router.routes,rev1, rev2, setRev1, setRev2])

    useEffect(()=>{
        if(!load){
            if(rev1 === "") {
                setTraffic(0)
            }
            if(rev2 === "") {
                setTraffic(100)
            }
        }
    },[rev1, rev2, load])

    return(
        <>
        <ContentPanel>
            <ContentPanelTitle>
                <ContentPanelTitleIcon>
                    <VscTypeHierarchySub />
                </ContentPanelTitleIcon>
                <FlexBox style={{display:"flex", alignItems:"center"}} className="gap">
                    <div>
                        Traffic Shaping
                    </div>
                    <HelpIcon msg={"Change the way the traffic is distributed for revisions of this workflow."} />
                </FlexBox>
            </ContentPanelTitle>
            <ContentPanelBody style={{flexDirection:"column"}}>
                <FlexBox className="gap wrap" style={{justifyContent: "space-between"}}>
                    <FlexBox style={{maxWidth:"350px"}}>
                        <FlexBox className="col">
                            <div>
                                <b>Revision One</b>
                            </div>
                            <FlexBox style={{ alignItems:"center", marginTop:"10px"}}>
                                <select onChange={(e)=>setRev1(e.target.value)} value={rev1} className="dropdown-select">
                                    <option value="">Select a revision</option>
                                    {revisions.map((obj)=>{
                                        if(rev2 === obj.name){
                                            return ""
                                        }
                                        return(
                                            <option key={GenerateRandomKey()} value={obj.name}>{obj.name}</option>
                                        )
                                    })}
                                </select>
                            </FlexBox>
                        </FlexBox>
                    </FlexBox>
                    <FlexBox style={{maxWidth:"350px"}}>
                        <FlexBox className="col">
                            <div>
                                <b>Revision Two</b>
                            </div>
                            <FlexBox style={{alignItems:"center", marginTop:"10px"}}>
                                <select onChange={(e)=>setRev2(e.target.value)} value={rev2} className="dropdown-select">
                                    <option value="">Select a revision</option>
                                    {revisions.map((obj)=>{
                                        if(rev1 === obj.name){
                                            return ""
                                        }
                                        return(
                                            <option key={GenerateRandomKey()} value={obj.name}>{obj.name}</option>
                                        )
                                    })}
                                </select>
                            </FlexBox>
                        </FlexBox>
                    </FlexBox>
                    <FlexBox style={{maxWidth: "350px", justifyContent: "center", paddingRight:"15px"}}>
                        <FlexBox className="col">
                            <div>
                                <b>Traffic Distribution</b>
                            </div>
                            <FlexBox style={{fontSize:"10pt", marginTop:"5px", maxHeight:"20px", color: "#C1C5C8"}}>
                                {rev1 ? 
                                <FlexBox className="col">
                                    <span title={rev1}>{rev1.substr(0, 8)}</span>
                                </FlexBox>:""}
                                {rev2 ? 
                                <FlexBox className="col" style={{ textAlign:'right'}}>
                                    <span title={rev2}>{rev2.substr(0,8)}</span>
                                </FlexBox>:""}
                            </FlexBox>
                            <Slider disabled={rev1 !== "" && rev2 !== "" ? false: true} className="red-green" value={traffic} onChange={(e)=>{setTraffic(e)}}/>
                            <FlexBox style={{marginTop:"15px", fontSize:"10pt", color: "#C1C5C8"}}>
                                {rev1 !== "" ? 
                                <FlexBox className="col">
                                    <span>{traffic}%</span>
                                </FlexBox>: ""}
                                {rev2 !== "" ?
                                <FlexBox  className="col" style={{justifyContent:'flex-end', textAlign:"right"}}>
                                    <span>{100-traffic}%</span>
                                </FlexBox>:""}
                            </FlexBox>
                        </FlexBox>
                    </FlexBox>
                    <div style={{width:"99.5%", margin:"auto", background: "#E9ECEF", height:"1px"}}/>
                </FlexBox>
                <FlexBox className={"row gap"} style={{ marginTop: "10px", justifyContent: "flex-end" }}>
                    <Button onClick={async () => {
                        let arr = []
                        if (rev1 !== "" && rev2 !== "") {
                            arr.push({
                                ref: rev1,
                                weight: parseInt(traffic)
                            })
                            arr.push({
                                ref: rev2,
                                weight: parseInt(100 - traffic)
                            })
                        } else if (rev1 !== "") {
                            arr.push({
                                ref: rev1,
                                weight: 100
                            })
                        } else if (rev2 !== "") {
                            arr.push({
                                ref: rev2,
                                weight: 100
                            })
                        }
                        await editWorkflowRouter(arr, router.live)
                        setRouter(await getWorkflowRouter())
                    }} className={`small ${rev2 && rev1 ? "" : "disabled"}`}>
                        Save
                    </Button>
                    <ModalHeadless
                        setVisible={setShowRunModal}
                        visible={showRunModal}
                        style={{ justifyContent: "center" }}
                        className="run-workflow-modal"
                        modalStyle={{ color: "black", width: "600px", minWidth: "30vw" }}
                        title="Run Workflow Router"
                        onClose={() => {
                            setInput("{\n\t\n}")
                        }}
                        actionButtons={[
                            ButtonDefinition(`Run`, async () => {
                                let r = ""
                                r = await executeWorkflowRouter(input)
                                if (r.includes("execute workflow")) {
                                    // is an error
                                    throw new Error(r)
                                } else {
                                    navigate(`/n/${namespace}/instances/${r}`)
                                }
                            }, `small`, () => { }, true, false),
                            ButtonDefinition("Cancel", async () => {
                            }, "small light", () => { }, true, false)
                        ]}
                    >
                        <FlexBox style={{ height: "45vh", minWidth: "250px", minHeight: "160px", overflow: "hidden" }}>
                            <FlexBox>
                                <AutoSizer>
                                    {({ height, width }) => (
                                        <DirektivEditor height={height} width={width} dlang="json" dvalue={input} setDValue={setInput} />
                                    )}
                                </AutoSizer>
                            </FlexBox>
                        </FlexBox>
                    </ModalHeadless>
                    <Button className="small" onClick={() =>{setShowRunModal(true)}} tip="Run workflow with router traffic">
                        Run
                    </Button>
                </FlexBox>
            </ContentPanelBody>
        </ContentPanel>
        </>
    )
}
