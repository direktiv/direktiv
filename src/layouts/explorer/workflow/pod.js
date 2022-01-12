import { useWorkflowServiceRevision } from "direktiv-react-hooks"
import { IoPlay } from "react-icons/io5"
import ContentPanel, { ContentPanelBody, ContentPanelTitle, ContentPanelTitleIcon } from "../../../components/content-panel"
import FlexBox from "../../../components/flexbox"
import { Config } from "../../../util"
import { ServiceStatus } from "../../namespace-services"
import * as dayjs from 'dayjs'
import { PodLogs } from "../../namespace-services/pod"

export default function WorkflowPod(props) {
    const {namespace, service, version, filepath, revision} = props

    const {revisionDetails, pods, err} = useWorkflowServiceRevision(Config.url, namespace, filepath, service, version, revision)

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
    
    return (
        <FlexBox className="col gap">
        <div >
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
                        <FlexBox className="wrap gap" style={{padding:"10px"}}>
                            <FlexBox className="col gap" style={{minWidth: "200px"}}>
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
                                </div>:<></>}
                            </FlexBox>
                            <FlexBox className="col gap" style={{minWidth: "200px"}}>
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
                            <FlexBox className="col gap" style={{minWidth: "200px"}}>
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
            <PodLogs  namespace={namespace} service={service} revision={revision} pods={pods} />
        </FlexBox>:<></>}
    </FlexBox>
    )
}