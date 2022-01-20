import { useGlobalServices } from "direktiv-react-hooks"
import { Service, ServiceCreatePanel } from "../namespace-services"
import {useEffect, useState} from "react"
import ContentPanel, { ContentPanelBody, ContentPanelTitle, ContentPanelTitleIcon } from "../../components/content-panel";
import FlexBox from "../../components/flexbox";
import { Config, GenerateRandomKey } from "../../util";
import Modal, { ButtonDefinition, KeyDownDefinition } from "../../components/modal";
import AddValueButton from "../../components/add-button";
import { IoPlay } from "react-icons/io5";
import HelpIcon from "../../components/help"

export default function GlobalServicesPanel(props) {
    const {data, err, config, createGlobalService, getConfig, getGlobalServices, deleteGlobalService} = useGlobalServices(Config.url, true, localStorage.getItem("apikey"))
    const [load, setLoad] = useState(true)

    const [serviceName, setServiceName] = useState("")
    const [image, setImage] = useState("")
    const [scale, setScale] = useState(0)
    const [size, setSize] = useState(0)
    const [cmd, setCmd] = useState("")
    const [maxScale, setMaxScale] = useState(0)

    useEffect(()=>{
        async function getcfg() {
            await getConfig().then(response => setMaxScale(response.maxscale));
            await getGlobalServices();
        }
        if(load && config === null && data === null) {
            getcfg()
            setLoad(false)
        }
    },[config, getConfig, load, data, getGlobalServices])

    if (err !== null) {
        // error happened with listing services
    }

    if(data === null) {
        return <></>
    }

    return(
        <FlexBox className="gap wrap" style={{paddingRight:"8px"}}>
            <ContentPanel style={{width:"100%"}}>
                <ContentPanelTitle>
                    <ContentPanelTitleIcon>
                        <IoPlay/>
                    </ContentPanelTitleIcon>
                    <FlexBox style={{display:"flex", alignItems:"center"}} className="gap">
                        <div>
                            Services 
                        </div>
                        <HelpIcon msg={"Create a global service that can be referenced from any workflow."} />
                    </FlexBox>
                    <div>
                    <Modal title="New global service" 
                        escapeToCancel
                        modalStyle={{
                            maxWidth: "300px"
                        }}
                        onOpen={() => {
                        }}
                        onClose={()=>{
                            setServiceName("")
                            setImage("")
                            setScale(0)
                            setSize(0)
                            setCmd("")
                        }}
                        button={(
                            <AddValueButton  label=" " />
                        )}  
                        keyDownActions={[
                            KeyDownDefinition("Enter", async () => {
                            }, ()=>{}, true)
                        ]}
                        actionButtons={[
                            ButtonDefinition("Add", async () => {
                                    await createGlobalService(serviceName, image, parseInt(scale), parseInt(size), cmd)
                            }, "small blue", ()=>{}, true, false),
                            ButtonDefinition("Cancel", () => {
                            }, "small light", ()=>{}, true, false)
                        ]}
                    >
                        {config !== null ? 
                            <ServiceCreatePanel maxScale={maxScale} cmd={cmd} setCmd={setCmd} size={size} setSize={setSize} name={serviceName} setName={setServiceName} image={image} setImage={setImage} scale={scale} setScale={setScale} />
                            :
                            ""
                        }
                    </Modal>
                </div>
                </ContentPanelTitle>
                <ContentPanelBody className="secrets-panel">
                    <FlexBox className="gap col">
                        <FlexBox className="col gap">
                        {data.length === 0 ?
                     <div className="col">
                     <FlexBox style={{ height:"40px", }}>
                             <FlexBox className="gap" style={{alignItems:"center", paddingLeft:"8px"}}>
                                 <div style={{fontSize:"10pt", }}>
                                     No services have been created.
                                 </div>
                             </FlexBox>
                     </FlexBox>
                 </div>
                    :
                    <>
                            {
                                data.map((obj)=>{
                                    return(
                                        <Service 
                                            key={GenerateRandomKey('global-service-')}
                                            url={`/g/services/${obj.info.name}`} 
                                            deleteService={deleteGlobalService}
                                            conditions={obj.conditions} 
                                            name={obj.info.name} 
                                            status={obj.status} 
                                            image={obj.info.image} 
                                        />
                                    )
                                })
                            }</>}
                        </FlexBox>
                    </FlexBox>
                </ContentPanelBody>
            </ContentPanel>

        </FlexBox>
    )
}