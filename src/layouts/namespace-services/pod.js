import * as dayjs from 'dayjs';
import { useNamespaceServiceRevision, usePodLogs } from "direktiv-react-hooks";
import { useEffect, useState } from "react";
import { VscLayers, VscServerEnvironment, VscTerminal } from 'react-icons/vsc';
import { useParams } from "react-router";
import { useSearchParams } from "react-router-dom";
import { AutoSizer, List } from 'react-virtualized';
import 'react-virtualized/styles.css'; // only needs to be imported once
import { ServiceStatus } from ".";
import ContentPanel, { ContentPanelBody, ContentPanelTitle, ContentPanelTitleIcon } from "../../components/content-panel";
import FlexBox from "../../components/flexbox";
import { LogFooterButtons } from "../../components/logs/logs";
import { Config } from "../../util";

export default function PodPanel(props) {
    const {namespace} = props
    const {service, revision} = useParams()

    if(!namespace) {
        return <></>
    }

    return(
        <FlexBox gap wrap style={{paddingRight:"8px"}}>
            <NamespaceRevisionDetails service={service} namespace={namespace} revision={revision}/>
        </FlexBox>
    )
}


function NamespaceRevisionDetails(props){
    const {service, namespace, revision} = props
    const {revisionDetails, pods, err} = useNamespaceServiceRevision(Config.url, namespace, service, revision, localStorage.getItem("apikey"))

    if(err) {
        console.log(err, "listing pods")
    }
    
    if(revisionDetails === null){
        return <></>
    }

    let size = "small"
    if(revisionDetails.size === 1) {
        size = "medium"
    } else if(revisionDetails.size === 2) {
        size ="large"
    }
    

    return(
        <FlexBox col gap>
            <div >
                <ContentPanel style={{width:"100%"}}>
                    <ContentPanelTitle>
                        <ContentPanelTitleIcon>
                            <VscLayers/>
                        </ContentPanelTitleIcon>
                        <FlexBox>
                            Details for {revision}
                        </FlexBox>
                    </ContentPanelTitle>
                        <ContentPanelBody className="secrets-panel" style={{fontSize:"11pt"}}>
                            <FlexBox className="wrap gap" style={{padding:"10px"}}>
                                <FlexBox col gap style={{minWidth: "200px"}}>
                                    <div>
                                        <span style={{fontWeight:"bold"}}>Created:</span> 
                                        <span style={{marginLeft:"5px"}}>{dayjs.unix(revisionDetails.created).format("HH:mmA, DD/MM/YYYY")}</span>
                                    </div>
                                    <div>
                                        <span style={{fontWeight:"bold"}}>Size:</span> 
                                        <span style={{marginLeft:"5px"}}>{size}</span>
                                    </div>
                                    <div>
                                        <span style={{fontWeight:"bold"}}>Generation:</span> 
                                        <span style={{marginLeft:"5px"}}>{revisionDetails.generation}</span>
                                    </div>
                                    {pods !== null && pods.length > 0 ?
                                    <div>
                                        <span style={{fontWeight:"bold"}}>Pods:</span> 
                                        <ul style={{marginTop:"0px", listStyle:"none", paddingLeft:'10px'}}>
                                            {pods.map((obj)=>{
                                                return(
                                                    <li style={{display:"flex", alignItems:'center', gap:"5px"}}>
                                                        <ServiceStatus status={obj.status}/>
                                                        {obj.name}
                                                    </li>
                                                )
                                            })}
                                        </ul>
                                    </div>:""}
                                </FlexBox>
                                <FlexBox col gap style={{minWidth: "200px"}}>
                                    <div>
                                        <span style={{fontWeight:"bold"}}>Image:</span>
                                        <span style={{marginLeft:"5px"}}>{revisionDetails.image}</span>
                                    </div>
                                    <div>
                                        <span style={{fontWeight:"bold"}}>Scale:</span>
                                        <span style={{marginLeft:"5px"}}>{revisionDetails.minScale}</span>
                                    </div>
                                    <div>
                                        <span style={{fontWeight:"bold"}}>Actual Replicas:</span>
                                        <span style={{marginLeft:"5px"}}>{revisionDetails.actualReplicas}</span>
                                    </div>
                                    <div>
                                        <span style={{fontWeight:"bold"}}>Desired Replicas:</span>
                                        <span style={{marginLeft:"5px"}}>{revisionDetails.desiredReplicas}</span>
                                    </div>
                                </FlexBox>
                                <FlexBox col gap style={{minWidth: "200px"}}>
                                    <span style={{fontWeight:"bold"}}>Conditions:</span>
                                    <ul style={{marginTop:"0px", listStyle:"none", paddingLeft:'10px'}}>
                                            {revisionDetails.conditions.map((obj)=>{
                                                if(obj.name === 'Active' && obj.reason === 'NoTraffic' && obj.message === "The target is not receiving traffic."){
                                                    return(
                                                        <li style={{display:"flex", alignItems:'center', gap:"5px"}}>
                                                            <ServiceStatus status={"True"}/>
                                                            {obj.name}
                                                        </li>
                                                    )
                                                }
                                                return(
                                                    <li style={{display:"flex", alignItems:'center', gap:"5px"}}>
                                                        <ServiceStatus status={obj.status}/>
                                                        {obj.name}
                                                    </li>
                                                )
                                            })}
                                        </ul>
                                </FlexBox>
                            </FlexBox>
                        </ContentPanelBody>
                </ContentPanel>
            </div>
            {pods !== null && pods.length > 0 ?
            <FlexBox>
                <PodLogs namespace={namespace} service={service} revision={revision} pods={pods} />
            </FlexBox>:""}
        </FlexBox>
    )
}

export function PodLogs(props){
    const {namespace, service, revision, pods} = props

    const [follow, setFollow] = useState(true)
    const [tab, setTab] = useState(pods[0] ? pods[0].name: "")
    const [clipData, setClipData] = useState(null)
    const [searchParams] = useSearchParams() // removed 'setSearchParams' from square brackets (this should not affect anything: search 'destructuring assignment')

    let version = searchParams.get('version')

    return (
        <ContentPanel style={{width:"100%"}}>
            <ContentPanelTitle>
                <ContentPanelTitleIcon>
                    <VscTerminal/>
                </ContentPanelTitleIcon>
                <FlexBox>
                    Pods
                </FlexBox>
            </ContentPanelTitle>
                <ContentPanelBody className="secrets-panel" style={{color:"white"}}>
                    <FlexBox col style={{backgroundColor:"#223848"}}>
                        <FlexBox style={{maxHeight:"30px"}}>
                            {pods.map((obj)=>{
                                let name = `global-${service}-${revision}-deployment-`
                                if(namespace){
                                    name = `namespace-${namespace}-${service}-${revision}-deployment-`
                                }
                                if(version){
                                    name = `workflow-${version}-${service}-${revision}-deployment-`
                                }

                                return(
                                    <div onClick={()=>setTab(obj.name)} style={{color: tab === obj.name ? "white": "#b5b5b5", display:"flex", alignItems:"center", cursor:"pointer", backgroundColor: tab === obj.name ? "#223848":"#355166", padding:"5px", maxWidth:"150px", gap:"3px"}}>
                                        <VscServerEnvironment style={{fill: tab === obj.name ? "white": "#b5b5b5"}}/> 
                                        <span style={{textOverflow: "ellipsis", whiteSpace:"nowrap", overflow:"hidden"}}>{obj.name.split(name)[1]}</span>
                                    </div>
                                )
                            })}
                        </FlexBox>
                        <FlexBox style={{flexGrow:1}}>
                            <Logs setClipData={setClipData} clipData={clipData} follow={follow} pod={tab} setFollow={setFollow}/>
                        </FlexBox>
                        <FlexBox style={{height:"40px", maxHeight:"40px", paddingRight:"10px", paddingLeft:"10px", boxShadow:"0px 0px 3px 0px #fcfdfe", alignItems:'center'}}>
                            <div title={tab} style={{whiteSpace:"nowrap",textOverflow:"ellipsis", overflow:"hidden", maxWidth:"300px"}}>
                                {version ? 
                                tab.split(`workflow-${version}-${service}-${revision}-deployment-`)
                                :""
                                }
                                {namespace ? 
                                 tab.split(`${namespace}-${service}-${revision}-deployment-`)[1]
                                :
                                 tab.split(`global-${service}-${revision}-deployment-`)
                                }
                            </div>
                            <FlexBox gap style={{justifyContent:"flex-end"}}>
                                <LogFooterButtons follow={follow} setFollow={setFollow} clipData={clipData}/>
                            </FlexBox>
                        </FlexBox>
                    </FlexBox>
                </ContentPanelBody>
        </ContentPanel>
    )
}

function Logs(props) {
    const {pod, follow, clipData, setClipData} = props

    const {data} = usePodLogs(Config.url, pod, localStorage.getItem("apikey"))

    useEffect(()=>{
        if(data !== null) {
            if(clipData === null) {
                setClipData(data.data)
            }
            if(clipData !== data){
                setClipData(data.data)
            }
        }
    },[data, clipData, setClipData])

    const renderRow = ({index, key, style}) => (
        <div key={key} style={style}>
            {data.data.split("\n")[index]}
        </div>
    )

    if (data === null || pod === "") {
        return <></>
    }
    
    return(
        <div style={{flex:"1 1 auto", paddingLeft:'10px'}}>
            <AutoSizer>
                {({height, width})=>(
                    <List
                        width={width}
                        height={height}
                        rowRenderer={renderRow}
                        scrollToIndex={follow ? data.data.split("\n").length - 1: undefined}
                        rowCount={data.data.split("\n").length}
                        rowHeight={20}
                    />
                )}
            </AutoSizer>
        </div>

    )
}

