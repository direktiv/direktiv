import { useNamespaceServices } from "direktiv-react-hooks";
import { IoPlay } from "react-icons/io5";
import "./style.css"
import { RiDeleteBin2Line } from "react-icons/ri";
import { FaCircle} from "react-icons/fa"
import ContentPanel, { ContentPanelBody, ContentPanelTitle, ContentPanelTitleIcon } from "../../components/content-panel";
import FlexBox from "../../components/flexbox";
import { Config } from "../../util";
import Modal, { ButtonDefinition, KeyDownDefinition } from "../../components/modal";
import AddValueButton from "../../components/add-button";


export default function ServicesPanel(props) {
    const {namespace} = props

    if(namespace === null) {
        return ""
    }
    return(
        <FlexBox className="gap wrap" style={{paddingRight:"8px"}}>
            <ContentPanel style={{width:"100%"}}>
                <ContentPanelTitle>
                    <ContentPanelTitleIcon>
                        <IoPlay/>
                    </ContentPanelTitleIcon>
                    <FlexBox>
                        Namespace Services
                    </FlexBox>
                    <div>
                    <Modal title="New namespace service" 
                        escapeToCancel

                        modalStyle={{
                            maxWidth: "300px"
                        }}

                        onOpen={() => {
                            console.log("ON OPEN");
                        }}

                        onClose={()=>{
                        }}
                        
                        button={(
                            <AddValueButton label=" " />
                        )}  
                        
                        keyDownActions={[
                            KeyDownDefinition("Enter", async () => {
                            }, true)
                        ]}
                        
                        actionButtons={[
                            ButtonDefinition("Add", async () => {
                            }, "small blue", true, false),
                            ButtonDefinition("Cancel", () => {
                            }, "small light", true, false)
                        ]}
                    >
                        create service here :)
                    </Modal>
                </div>
                </ContentPanelTitle>
                <ContentPanelBody className="secrets-panel">
                    <FlexBox className="gap col">
                        <NamespaceServices namespace={namespace}/>
                    </FlexBox>
                </ContentPanelBody>
            </ContentPanel>
        </FlexBox>
    )
}

function NamespaceServices(props) {
    const {namespace} = props
    const {data, err} = useNamespaceServices(Config.url, true, namespace)
    console.log(data, err)
    if(data === null) {
        return ""
    }

    return(
        <FlexBox className="col gap">
            {
                data.map((obj)=>{
                    console.log(obj)
                    return(
                        <Service conditions={obj.conditions} name={obj.info.name} status={obj.status} image={obj.info.image} />
                    )
                })
            }
        </FlexBox>
    )
}

export function Service(props) {
    const {name, image, status, conditions} = props
    
    return(
        <div className="col">
            <FlexBox style={{height:"40px", border:"1px solid #f4f4f4", backgroundColor:"#fcfdfe"}}>
                <FlexBox className="gap" style={{alignItems:"center", paddingLeft:"8px"}}>
                    <ServiceStatus status={status} />
                    <div style={{fontWeight:"bold"}}>
                        {name}
                    </div>
                    <div style={{fontStyle:"italic"}} className="grey-text">
                        {image}
                    </div>
                    {/* 
                     // Todo add contextually what is using this service
                    <div>
                        x
                    </div> */}
                </FlexBox>
                <FlexBox>
                    <ServicesDeleteButton/>
                </FlexBox>
            </FlexBox>
            <FlexBox style={{border:"1px solid #f4f4f4", borderTop:"none"}}>
                <ServiceDetails conditions={conditions} />
            </FlexBox>
        </div>
    )
}

function ServiceDetails(props) {
    const {conditions} = props

    return(
        <ul style={{listStyle:"none", paddingLeft:"25px", paddingRight:"40px", width:"100%"}}>
            {conditions.map((obj)=>{
                return(
                    <li style={{display:"flex", gap:"10px"}}>
                        <div>
                            <ServiceStatus status={obj.status}/>
                        </div>
                        <FlexBox className="col gap" style={{marginBottom:"10px"}}>
                            <div>
                                {obj.name}
                            </div>
                            {obj.status === 'Unknown' ?
                            <>
                            {obj.reason !== ""  ? 
                            <div className="grey-text" style={{fontSize:"10pt", fontStyle:"italic"}}>
                                {obj.reason}
                            </div>:""}
                            {obj.message !== "" ? 
                            <div className="wait-message" >
                                {obj.message}
                            </div>:""}
                            </>:""
                            }
                            {obj.status === 'False' ? 
                            <>
                            <div className="grey-text" style={{fontSize:"10pt", fontStyle:"italic"}}>
                                {obj.reason}
                            </div>
                            <div className="fail-message" >
                                {obj.message}
                            </div></>:""}
                        </FlexBox>
                    </li>
                )
            })}

        </ul>
    )
}

export function ServiceStatus(props) {
    const {status} = props

    let color = "#66DE93"
    if (status === "False") {
        color = "#FF616D"
    }

    if (status === "Unknown") {
        color = "#082032"
    }

    return(
        <div>   
            <FaCircle style={{fontSize:"6pt", fill: color}} />
        </div>
    )
}

function ServicesDeleteButton() {
    return (
        <FlexBox className="col  auto-margin red-text" style={{display: "flex", alignItems:"flex-end", width:"100%", height: "100%"}}>
            <div className="secrets-delete-btn" style={{height: "100%", display: "flex", alignItems: "center", paddingRight: "8px" }}>
                <RiDeleteBin2Line className="auto-margin" />
            </div>
        </FlexBox>
    )
}