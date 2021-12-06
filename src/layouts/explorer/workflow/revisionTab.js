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
import { RiDeleteBin2Line } from 'react-icons/ri';
import SankeyDiagram from '../../../components/sankey';

function RevisionTab(props) {

    const {searchParams, setSearchParams, revision, setRevision, getWorkflowRevisionData, getWorkflowSankeyMetrics, executeWorkflow} = props
    const [load, setLoad] = useState(true)
    const [workflow, setWorkflowData] = useState(null)
    const [tabBtn, setTabBtn] = useState(searchParams.get('revtab') !== null ? parseInt(searchParams.get('revtab')): 0);

    useEffect(()=>{
        if(searchParams.get('revtab') === null) {
            setSearchParams({
                tab: searchParams.get('tab'),
                revision: revision,
                revtab: 0
            })
        }
    },[searchParams])

    useEffect(()=>{
        async function getRevWorkflow() {
            if(load && searchParams.get('revtab') !== null) {
                let wfdata = await getWorkflowRevisionData(revision)
                setWorkflowData(atob(wfdata.revision.source))
                setLoad(false)
            }
        }
        getRevWorkflow()
    },[load, searchParams])

    console.log(workflow)

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
                                    <DirektivEditor value={workflow} readonly={true} dlang="yaml" />
                                </FlexBox>
                                <FlexBox className="gap" style={{backgroundColor:"#223848", color:"white", height:"40px", maxHeight:"40px", paddingLeft:"10px", minHeight:"40px", borderTop:"1px solid white", alignItems:'center'}}>
                                    <div style={{display:"flex", flex:1 }}>
                                    </div>
                                    <div style={{display:"flex", flex:1, justifyContent:"center"}}>
                                        <div onClick={async ()=>{
                                            let id = await executeWorkflow("", revision)
                                            console.log(id, "ID")
                                        }} style={{alignItems:"center", gap:"3px",backgroundColor:"#355166", paddingTop:"3px", paddingBottom:"3px", paddingLeft:"6px", paddingRight:"6px", cursor:"pointer", borderRadius:"3px"}}>
                                            Run
                                        </div>
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

    console.log(tabBtn);

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
    const {getRevisions, deleteRevision, getWorkflowSankeyMetrics, executeWorkflow, searchParams, setSearchParams, getWorkflowRevisionData} = props
    const [load, setLoad] = useState(true)
    const [revisions, setRevisions] = useState([])
    const [revision, setRevision] = useState(null)
    const [err, setErr] = useState(null)

    // fetch revisions using the workflow hook from above
    useEffect(()=>{
        async function listData() {
            if(load){
                // get the instances
                let resp = await getRevisions()
                if(Array.isArray(resp)){
                    setRevisions(resp)
                } else {
                    setErr(resp)
                }

            }
            setLoad(false)
        }
        listData()
    },[load, getRevisions])

    useEffect(()=>{
        if(searchParams.get('revision') !== null) {
            setRevision(searchParams.get('revision'))
        }
    },[searchParams])

    if(revision !== null) {
        return(
            <RevisionTab getWorkflowSankeyMetrics={getWorkflowSankeyMetrics} executeWorkflow={executeWorkflow} setRevision={setRevision} getWorkflowRevisionData={getWorkflowRevisionData}  searchParams={searchParams} setSearchParams={setSearchParams} revision={revision}/>
        )
    }

    return (
        <FlexBox className="col">
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
                        return (
                            <FlexBox className="gap wrap" style={{
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
                                <FlexBox style={{
                                    flex: "1",
                                    minWidth: "300px"
                                }}>
                                    TODO: Traffic Component
                                </FlexBox>
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
                                        <Button className="small light bold">
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
        </FlexBox>
    )

    return(
        <>
            <FlexBox className="gap col wrap" style={{height:"100%"}}>
                <ContentPanel style={{ width: "100%", minWidth: "300px"}}>
                    <ContentPanelTitle>
                        <ContentPanelTitleIcon>
                            <BsCodeSquare />
                        </ContentPanelTitleIcon>
                        <div>
                            All Revisions
                        </div>
                    </ContentPanelTitle>
                    <ContentPanelBody>
                        <table>
                            <tbody>
                                {
                                    revisions.map((obj)=>{
                                        return(
                                            <tr>
                                                <td>
                                                    {obj.node.name}
                                                </td>
                                                <td>
                                                <Modal
                                                    escapeToCancel
                                                    style={{
                                                        flexDirection: "row-reverse",
                                                    }}
                                                    title="Delete a revision" 
                                                    button={(
                                                        <div className="secrets-delete-btn grey-text auto-margin red-text" style={{display: "flex", alignItems: "center", height: "100%"}}>
                                                        <RiDeleteBin2Line className="auto-margin"/>
                                                    </div>
                                                    )}
                                                    actionButtons={
                                                        [
                                                            ButtonDefinition("Delete", async () => {
                                                                let err = await deleteRevision(obj.node.name)
                                                                if (err) return err
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
                                                </td>
                                                <td>
                                                    set working rev
                                                </td>
                                                <td onClick={()=>{
                                                    setSearchParams({tab: 1, revision: obj.node.name})
                                                }}>
                                                    open revision
                                                </td>
                                            </tr>
                                        )
                                    })
                                }
                            </tbody>
                        </table>
                    </ContentPanelBody>
                </ContentPanel>
                <ContentPanel style={{ width: "100%", minWidth: "300px", minHeight:"200px"}}>
                    <ContentPanelTitle>
                        <ContentPanelTitleIcon>
                            <BsCodeSquare />
                        </ContentPanelTitleIcon>
                        <div>
                            Revision Traffic Shaping
                        </div>
                    </ContentPanelTitle>
                    <ContentPanelBody>
                        testing
                    </ContentPanelBody>
                </ContentPanel>
            </FlexBox>
        </>
    )
}
