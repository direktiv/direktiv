import { useParams } from "react-router"
import FlexBox from "../../components/flexbox"
import ContentPanel, { ContentPanelBody, ContentPanelTitle, ContentPanelTitleIcon, ContentPanelFooter } from "../../components/content-panel"
import { IoPlay } from "react-icons/io5"
import { useNamespaceServiceRevision, usePodLogs } from "direktiv-react-hooks"
import { Config } from "../../util"
import { useEffect, useRef, useState } from "react"
import { AutoSizer, List } from 'react-virtualized'
import 'react-virtualized/styles.css'; // only needs to be imported once
import { ServiceStatus } from "."
import { copyTextToClipboard} from '../../util'

export default function PodPanel(props) {
    const {namespace} = props
    const {service, revision} = useParams()

    if(!namespace) {
        return ""
    }

    return(
        <FlexBox className="gap wrap" style={{paddingRight:"8px"}}>
            <NamespaceRevisionDetails service={service} namespace={namespace} revision={revision}/>
        </FlexBox>
    )
}


function NamespaceRevisionDetails(props){
    const {service, namespace, revision} = props
    const {revisionDetails, pods, err} = useNamespaceServiceRevision(Config.url, namespace, service, revision)

    console.log(revisionDetails)

    if(err) {
        console.log(err, "listing pods")
    }
    
    if(revisionDetails === null){
        return ""
    }

    let size = "small"
    if(revisionDetails.size === 1) {
        size = "medium"
    } else if(revisionDetails.size === 2) {
        size ="large"
    }
    

    return(
        <FlexBox className="col gap">
            <FlexBox >
                <ContentPanel style={{width:"100%"}}>
                    <ContentPanelTitle>
                        <ContentPanelTitleIcon>
                            <IoPlay/>
                        </ContentPanelTitleIcon>
                        <FlexBox>
                            Details for {revision}
                        </FlexBox>
                    </ContentPanelTitle>
                        <ContentPanelBody className="secrets-panel" style={{fontSize:"11pt"}}>
                            <FlexBox style={{padding:"10px"}}>
                                <FlexBox className="col">
                                    <div>
                                        <span style={{fontWeight:"bold"}}>Created:</span> 
                                        <span style={{marginLeft:"3px"}}>2:12pm, 22/11/2021</span>
                                    </div>
                                    <div>
                                        <span style={{fontWeight:"bold"}}>Size:</span> 
                                        <span style={{marginLeft:"3px"}}>{size}</span>
                                    </div>
                                    <div>
                                        <span style={{fontWeight:"bold"}}>Generation:</span> 
                                        <span style={{marginLeft:"3px"}}>{revisionDetails.generation}</span>
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
                                <FlexBox className="col">
                                    <div>
                                        <span style={{fontWeight:"bold"}}>Image:</span>
                                        <span style={{marginLeft:"3px"}}>{revisionDetails.image}</span>
                                    </div>
                                    <div>
                                        <span style={{fontWeight:"bold"}}>Scale:</span>
                                        <span style={{marginLeft:"3px"}}>{revisionDetails.minScale}</span>
                                    </div>
                                    <div>
                                        <span style={{fontWeight:"bold"}}>Actual Replicas:</span>
                                        <span style={{marginLeft:"3px"}}>{revisionDetails.actualReplicas}</span>
                                    </div>
                                    <div>
                                        <span style={{fontWeight:"bold"}}>Desired Replicas:</span>
                                        <span style={{marginLeft:"3px"}}>{revisionDetails.desiredReplicas}</span>
                                    </div>
                                </FlexBox>
                                <FlexBox className="col">
                                    <span style={{fontWeight:"bold"}}>Conditions:</span>
                                    <ul style={{marginTop:"0px", listStyle:"none", paddingLeft:'10px'}}>
                                            {revisionDetails.conditions.map((obj)=>{
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
            </FlexBox>
            {pods !== null && pods.length > 0 ?
            <FlexBox>
                <PodLogs namespace={namespace} service={service} revision={revision} pods={pods} />
            </FlexBox>:""}
        </FlexBox>
    )
}

function PodLogs(props){
    const {namespace, service, revision, pods} = props

    const [follow, setFollow] = useState(true)

    const [tab, setTab] = useState(pods[0] ? pods[0].name: "")
    const [clipData, setClipData] = useState(null)


    return (
        <ContentPanel style={{width:"100%"}}>
            <ContentPanelTitle>
                <ContentPanelTitleIcon>
                    <IoPlay/>
                </ContentPanelTitleIcon>
                <FlexBox>
                    Pods
                </FlexBox>
            </ContentPanelTitle>
                <ContentPanelBody className="secrets-panel" style={{color:"white"}}>
                    <FlexBox className="col" style={{backgroundColor:"#223848"}}>
                        <FlexBox style={{maxHeight:"30px"}}>
                            {pods.map((obj)=>{
                                return(
                                    <div onClick={()=>setTab(obj.name)} style={{cursor:"pointer", backgroundColor: tab === obj.name ? "#223848":"#355166", padding:"5px", maxWidth:"150px"}}>
                                        {obj.name.split(`namespace-${namespace}-${service}-${revision}-deployment-`)[1]}
                                    </div>
                                )
                            })}
                        </FlexBox>
                        <FlexBox style={{flexGrow:1}}>
                            <Logs setClipData={setClipData} clipData={clipData} follow={follow} pod={tab} setFollow={setFollow}/>
                        </FlexBox>
                        <FlexBox style={{maxHeight:"40px", paddingRight:"10px", paddingLeft:"10px", boxShadow:"0px 0px 3px 0px #fcfdfe", alignItems:'center'}}>
                            <FlexBox>
                                {tab}
                            </FlexBox>
                            <FlexBox className="gap" style={{justifyContent:"flex-end"}}>
                                {follow ? 
                                    <div onClick={(e)=>setFollow(!follow)} style={{backgroundColor:"#355166", paddingTop:"3px", paddingBottom:"3px", paddingLeft:"6px", paddingRight:"6px", cursor:"pointer", borderRadius:"3px"}}>
                                        Stop Watching
                                    </div>
                                    :
                                    <div onClick={(e)=>setFollow(!follow)} style={{backgroundColor:"#355166", paddingTop:"3px", paddingBottom:"3px", paddingLeft:"6px", paddingRight:"6px", cursor:"pointer", borderRadius:"3px"}}>
                                        Follow Logs
                                    </div>
                                }
                                <div onClick={()=>{
                                    copyTextToClipboard(clipData)
                                }} style={{backgroundColor:"#355166",paddingTop:"3px", paddingBottom:"3px",  paddingLeft:"6px", paddingRight:"6px", cursor:"pointer", borderRadius:"3px"}}>
                                    Copy to Clipboard
                                </div>
                            </FlexBox>
                        </FlexBox>
                    </FlexBox>
                </ContentPanelBody>
        </ContentPanel>
    )
}

function Logs(props) {
    const {pod, follow, clipData, setClipData} = props

    const {data, err} = usePodLogs(Config.url, pod)

    useEffect(()=>{
        if(data !== null) {
            if(clipData === null) {
                setClipData(data.data)
            }
            if(clipData !== data){
                setClipData(data.data)
            }
        }
    },[data])

    const renderRow = ({index, key, style}) => (
        <div key={key} style={style}>
            {data.data.split("\n")[index]}
        </div>
    )

    if (data === null || pod === "") {
        return ""
    }
    
    return(
        <div style={{flex:"1 1 auto", paddingLeft:'10px'}}>
            <AutoSizer>
                {({height, width})=>(
                    <List
                        width={width}
                        height={height}
                        rowRenderer={renderRow}
                        scrollToIndex={follow ? data.data.split("\n").length - 1: 0}
                        rowCount={data.data.split("\n").length}
                        rowHeight={20}
                    />
                )}
            </AutoSizer>
        </div>

    )
}

