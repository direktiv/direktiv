import React, { useEffect, useState } from 'react';
import Button from '../../../components/button';
import { BsCodeSquare } from 'react-icons/bs';
import {HiOutlineTrash} from 'react-icons/hi';
import ContentPanel, { ContentPanelBody, ContentPanelTitle, ContentPanelTitleIcon } from '../../../components/content-panel';
import FlexBox from '../../../components/flexbox';
import {GenerateRandomKey} from '../../../util';
import {BiChevronLeft} from 'react-icons/bi';
import DirektivEditor from '../../../components/editor';
import WorkflowDiagram from '../../../components/diagram';
import YAML from 'js-yaml'
import Modal, { ButtonDefinition } from '../../../components/modal';
import SankeyDiagram from '../../../components/sankey';
import { IoSettings } from 'react-icons/io5';

import Slider from 'rc-slider';
import 'rc-slider/assets/index.css';
import { useNavigate } from 'react-router';
function RevisionTab(props) {

    const navigate = useNavigate()
    const {searchParams, setSearchParams, revision, setRevision, getWorkflowRevisionData, getWorkflowSankeyMetrics, executeWorkflow, namespace} = props
    const [load, setLoad] = useState(true)
    const [workflow, setWorkflowData] = useState(null)
    const [tabBtn, setTabBtn] = useState(searchParams.get('revtab') !== null ? parseInt(searchParams.get('revtab')): 0);
    const [input, setInput] = useState("{\n\t\n}")

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
                            <BsCodeSquare />
                        </ContentPanelTitleIcon>
                        <div>
                        {revision}
                        </div>
                        <TabbedButtons revision={revision} setSearchParams={setSearchParams} searchParams={searchParams} tabBtn={tabBtn} setTabBtn={setTabBtn} />
                    </ContentPanelTitle>
                    <ContentPanelBody >
                        {tabBtn === 0 ? 
                            <FlexBox className="col" style={{overflow:"hidden"}}>
                                <FlexBox >
                                    <DirektivEditor value={workflow} readonly={true} disableBottomRadius={true} dlang="yaml" />
                                </FlexBox>
                                <FlexBox className="gap" style={{backgroundColor:"#223848", color:"white", height:"44px", maxHeight:"44px", paddingLeft:"10px", minHeight:"44px", borderTop:"1px solid white", alignItems:'center', borderRadius:"0px 0px 8px 8px", overflow: "hidden"}}>
                                    <div style={{display:"flex", flex:1 }}>
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
                                                        r = await executeWorkflow("", revision)
                                                    } else {
                                                        r = await executeWorkflow(input, revision)
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
                                                <div style={{alignItems:"center", gap:"3px",backgroundColor:"#355166", paddingTop:"3px", paddingBottom:"3px", paddingLeft:"6px", paddingRight:"6px", cursor:"pointer", borderRadius:"3px"}}>
                                                    Run
                                                </div>
                                            )}
                                        >
                                            <FlexBox style={{overflow:"hidden"}}>
                                                <DirektivEditor height="200" width="300" dlang="json" dvalue={input} setDValue={setInput}/>
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
                        {tabBtn === 1 ? <WorkflowDiagram disabled={true} workflow={YAML.load(workflow)}/>:""}
                        {tabBtn === 2 ? <SankeyDiagram revision={revision} getWorkflowSankeyMetrics={getWorkflowSankeyMetrics} />:""}
                    </ContentPanelBody>
                </ContentPanel>
                </FlexBox>
            </FlexBox>
        </FlexBox>
    )

}

export default RevisionTab;

function TabbedButtons(props) {

    let {tabBtn, setTabBtn, searchParams, setSearchParams, revision} = props;

    let tabBtns = [];
    let tabBtnLabels = ["YAML", "Diagram", "Sankey"];

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
                    tab: searchParams.get('tab'),
                    revision: revision,
                    revtab: i
                })
            }}>
                {tabBtnLabels[i]}
            </div>
        </FlexBox>)
    }

    return(
            <FlexBox className="tabbed-btns-container">
                <FlexBox className="tabbed-btns" >
                    {tabBtns}
                </FlexBox>
            </FlexBox>
    )
}


export function RevisionSelectorTab(props) {
    const {setRouter, namespace, tagWorkflow, filepath, updateWorkflow, editWorkflowRouter, getWorkflowRouter, getRevisions, setRevisions, err, revisions, router, deleteRevision, getWorkflowSankeyMetrics, executeWorkflow, searchParams, setSearchParams, getWorkflowRevisionData} = props
    
    const navigate = useNavigate()
    // const [load, setLoad] = useState(true)
    const [revision, setRevision] = useState(null)
    const [rev1, setRev1] = useState(router.routes.length === 0 ? "latest": "")
    const [rev2, setRev2] = useState("")
    const [tag, setTag] = useState("")

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
            <RevisionTab getWorkflowSankeyMetrics={getWorkflowSankeyMetrics} executeWorkflow={executeWorkflow} setRevision={setRevision} getWorkflowRevisionData={getWorkflowRevisionData}  searchParams={searchParams} setSearchParams={setSearchParams} revision={revision}/>
        )
    }

  

    return (
        <FlexBox className="col gap">
            <div>
                <ContentPanel style={{width: "100%", minWidth: "300px"}}>
                    <ContentPanelTitle>
                        <ContentPanelTitleIcon>
                            <BsCodeSquare/>
                        </ContentPanelTitleIcon>
                        <div>
                            All Revisions
                        </div>
                    </ContentPanelTitle>
                    <ContentPanelBody style={{flexDirection: "column"}}>
                        {revisions.map((obj) => {
                            let ref1 = false
                            let ref2 = false
                            if(router.routes[0]){
                                if(router.routes[0].ref === obj.node.name){
                                    ref1= true
                                }
                            }
                            if(router.routes[1]){
                                if(router.routes[1].ref === obj.node.name){
                                    ref2 = true
                                }
                            }

                            for(var i=0; i < router.routes.length; i++) {
                                if(obj.node.name === router.routes[i].ref){}
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
                                                    {obj.node.name}
                                                </div>
                                            </FlexBox>
                                        </div>
                                    </FlexBox>
                                    {obj.node.name !== "latest" ? 
                                        <FlexBox style={{
                                            flex: "1",
                                            maxWidth: "150px"
                                        }}>
                                            <FlexBox className="col revision-label-tuple">
                                                <Modal
                                                    escapeToCancel
                                                    style={{
                                                        flexDirection: "row-reverse",
                                                        marginRight: "8px"
                                                    }}
                                                    title="Tag" 
                                                    onClose={()=>{
                                                        setTag("")
                                                    }}
                                                    button={(
                                                        <Button className="reveal-btn small shadow">
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
                                                                let err = await tagWorkflow(obj.node.name, tag)
                                                                if(err) return err
                                                                setRevisions(await getRevisions())
                                                            }, "small blue", true, false),
                                                            ButtonDefinition("Cancel", () => {
                                                            }, "small light", true, false)
                                                        ]
                                                    } 
                                                >
                                                    <FlexBox>
                                                        <input autoFocus value={tag} onChange={(e)=>setTag(e.target.value)} placeholder="Enter Tag or leave blank to untag" />
                                                    </FlexBox>
                                                </Modal>
                                            </FlexBox>
                                        </FlexBox>
                                    :<FlexBox style={{
                                        flex: "1",
                                        maxWidth: "150px"
                                    }}></FlexBox>}
                                    {router.routes.length > 0 ? 
                                    <>
                                        {router.routes[0] && router.routes[0].ref === obj.node.name  ? 
                                            <FlexBox style={{
                                                flex: "1",
                                                maxWidth: "150px"
                                            }}>
                                            <FlexBox className="col revision-label-tuple">
                                                    <div>
                                                        Traffic amount
                                                    </div>
                                                    <div style={{width:'100%'}}>
                                                        <Slider defaultValue={router.routes[0].weight} className="traffic-mini-distribution" disabled={true}/>
                                                        <div>
                                                           {router.routes[0].weight}%
                                                        </div>
                                                    </div>
                                                </FlexBox>
                                            </FlexBox>
                                        :""}
                                        {router.routes[1]  && router.routes[1].ref === obj.node.name  ? 
                                            <FlexBox style={{
                                                flex: "1",
                                                maxWidth: "150px"
                                            }}>
                                            <FlexBox className="col revision-label-tuple">
                                                    <div>
                                                        Traffic amount
                                                    </div>
                                                    <div style={{width:'100%'}}>
                                                        <Slider defaultValue={router.routes[1].weight} className="traffic-mini2-distribution" disabled={true}/>
                                                        <div>
                                                            {router.routes[1].weight}%
                                                        </div>
                                                    </div>
                                                </FlexBox>
                                            </FlexBox>
                                        :""}
                                        {!ref1 && !ref2 ? 
                                          <FlexBox style={{
                                            flex: "1",
                                            maxWidth: "150px"
                                        }}></FlexBox>
                                        :""
                                        }
                                    </>
                                    : <>
                                        {obj.node.name === "latest" ? 
                                        <FlexBox style={{
                                            flex: "1",
                                            maxWidth: "150px"
                                        }}>
                                            <FlexBox className="col revision-label-tuple">
                                                <div>
                                                    Traffic amount
                                                </div>
                                                <div style={{width:'100%'}}>
                                                    <Slider defaultValue={100} className="traffic-mini-distribution" disabled={true}/>
                                                    <div>
                                                        100%
                                                    </div>
                                                </div>
                                            </FlexBox>
                                        </FlexBox>
                                        :""}
                                    </>}
                                    {/* <FlexBox style={{
                                        flex: "1",
                                        minWidth: "300px"
                                    }}>
                                        
                                    </FlexBox> */}
                                    <div>
                                        <FlexBox className="gap">
                                                <Modal
                                                        escapeToCancel
                                                        style={{
                                                            flexDirection: "row-reverse",
                                                        }}
                                                        title="Delete a revision" 
                                                        button={(
                                                            <Button className="small light bold">
                                                                <HiOutlineTrash className="red-text" style={{fontSize: "16px"}} />
                                                            </Button>
                                                        )}
                                                        actionButtons={
                                                            [
                                                                ButtonDefinition("Delete", async () => {
                                                                    let err = await deleteRevision(obj.node.name)
                                                                    if (err) return err
                                                                    setRevisions(await getRevisions())
                                                                }, "small red", true, false),
                                                                ButtonDefinition("Cancel", () => {
                                                                }, "small light", true, false)
                                                            ]
                                                        } 
                                                    >
                                                            <FlexBox className="col gap">
                                                        <FlexBox >
                                                            Are you sure you want to delete '{obj.node.name}'?
                                                            <br/>
                                                            This action cannot be undone.
                                                        </FlexBox>
                                                    </FlexBox>
                                                    </Modal>
                                            <Button className="small light bold" onClick={async()=>{
                                                let data = await getWorkflowRevisionData(obj.node.name)
                                                await updateWorkflow(atob(data.revision.source))
                                                navigate(`/n/${namespace}/explorer/${filepath.substring(1)}?tab=2`)
                                            }}>
                                                Use Revision
                                            </Button>
                                            <Button className="small light bold" onClick={()=>{
                                                setSearchParams({tab: 1, revision: obj.node.name})
                                            }}>
                                                Open Revision
                                            </Button>
                                        </FlexBox>
                                    </div>
                                </FlexBox>
                            )
                        })}
                    </ContentPanelBody>
                </ContentPanel>
            </div>
            <div>
                <RevisionTrafficShaper rev1={rev1} rev2={rev2} setRev1={setRev1} setRev2={setRev2} setRouter={setRouter} revisions={revisions}  router={router} editWorkflowRouter={editWorkflowRouter} getWorkflowRouter={getWorkflowRouter} />
            </div>
        </FlexBox>
    )

}

export function RevisionTrafficShaper(props) {
    const {editWorkflowRouter, rev1, rev2, setRev1, setRev2, setRouter, getWorkflowRouter, router, revisions} = props

    const [load, setLoad] = useState(true)

    const [traffic, setTraffic] = useState(router.routes.length === 0 ? 100 : 0)

    useEffect(()=>{
        if(load){
            if (router.routes[0]){
                setRev1(router.routes[0].ref)
                setTraffic(router.routes[0].weight)
            }

            if(router.routes[1]){
                setRev2(router.routes[1].ref)
            }
            setLoad(false)
        }
    },[load, router.routes,rev1, rev2])

    console.log(traffic)
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
                    <IoSettings />
                </ContentPanelTitleIcon>
                <div>
                    Traffic Shaping
                </div>
            </ContentPanelTitle>
            <ContentPanelBody style={{flexDirection:"column"}}>
                <FlexBox className="gap wrap" style={{justifyContent: "space-between"}}>
                    <FlexBox style={{maxWidth: "300px", justifyContent: "center"}}>
                        <FlexBox className="gap col">
                            <div>
                                <b>Revision 1</b>
                            </div>
                            <FlexBox style={{alignItems:"center"}}>
                                <select onChange={(e)=>setRev1(e.target.value)} value={rev1}>
                                    <option value="">Select a workflow revision</option>
                                    {revisions.map((obj)=>{
                                        if(rev2 === obj.node.name){
                                            return ""
                                        }
                                        return(
                                            <option key={GenerateRandomKey()} value={obj.node.name}>{obj.node.name}</option>
                                        )
                                    })}
                                </select>
                            </FlexBox>
                 
                        </FlexBox>
                    </FlexBox>
                    <FlexBox style={{ maxWidth: "300px", justifyContent: "center"}}>
                        <FlexBox className="gap col">
                            <div>
                                <b>Revision 2</b>
                            </div>
                            <FlexBox style={{alignItems:"center"}}>
                                <select onChange={(e)=>setRev2(e.target.value)} value={rev2}>
                                    <option value="">Select a workflow revision</option>
                                    {revisions.map((obj)=>{
                                        if(rev1 === obj.node.name){
                                            return ""
                                        }
                                        return(
                                            <option key={GenerateRandomKey()} value={obj.node.name}>{obj.node.name}</option>
                                        )
                                    })}
                                </select>
                            </FlexBox>
                            {/* <input style={{width: "auto"}}></input> */}
                        </FlexBox>
                    </FlexBox>
                    <FlexBox style={{maxWidth: "300px", justifyContent: "center", paddingRight:"15px"}}>
                        <FlexBox className="gap col">
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
                            <FlexBox style={{marginTop:"10px", fontSize:"10pt", color: "#C1C5C8"}}>
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
                </FlexBox>
                <FlexBox style={{marginTop:"10px", justifyContent:"flex-end"}}>
                    <Button onClick={async()=>{
                        let arr = []
                        if(rev1 !== "" && rev2 !== "") {
                            arr.push({
                                ref: rev1,
                                weight: parseInt(traffic)
                            })
                            arr.push({
                                ref: rev2,
                                weight: parseInt(100-traffic)
                            })
                        } else if(rev1 !== "") {
                            arr.push({
                                ref: rev1,
                                weight: 100
                            })
                        } else if(rev2 !== "") {
                            arr.push({
                                ref: rev2,
                                weight: 100
                            })
                        }
                        await editWorkflowRouter(arr, router.live)
                        setRouter(await getWorkflowRouter())
                    }} className="small">
                        Save
                    </Button>
                </FlexBox>
            </ContentPanelBody>
        </ContentPanel>
        </>
    )
}