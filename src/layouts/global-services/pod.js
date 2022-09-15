import { useGlobalServiceRevision } from "direktiv-react-hooks";
import FlexBox from "../../components/flexbox";
import { Config } from "../../util";
import { useParams } from "react-router"
import ContentPanel, { ContentPanelBody, ContentPanelTitle, ContentPanelTitleIcon } from "../../components/content-panel"
import { VscLayers } from 'react-icons/vsc';
import 'react-virtualized/styles.css'; // only needs to be imported once
import { ServiceStatus } from "../namespace-services"
import { PodLogs } from "../namespace-services/pod"
import * as dayjs from 'dayjs'

export default function GlobalPodPanel(props) {
    const {service, revision} = useParams()

    const {revisionDetails, pods, err} = useGlobalServiceRevision(Config.url, service, revision, localStorage.getItem("apikey"))

    if (err) {
        console.log(err, "listing pods")
    }

    if (revisionDetails === null) {
        return <></>
    }

    let size = "small"
    if(revisionDetails.size === 1) {
        size = "medium"
    } else if(revisionDetails.size === 2) {
        size ="large"
    }

    return (
        <FlexBox gap wrap style={{paddingRight:"8px"}}>
            <FlexBox col gap>
                        <FlexBox >
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
                                        <FlexBox style={{padding:"10px"}}>
                                            <FlexBox col gap>
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
                                            <FlexBox col gap>
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
                                            <FlexBox col gap>
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
                            <PodLogs  service={service} revision={revision} pods={pods} />
                        </FlexBox>:""}
                    </FlexBox>
        </FlexBox>
    )
}