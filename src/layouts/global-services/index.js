import { useGlobalServices } from "direktiv-react-hooks"
import { Service, ServiceCreatePanel } from "../namespace-services"
import {useEffect, useState} from "react"
import ContentPanel, { ContentPanelBody, ContentPanelTitle, ContentPanelTitleIcon } from "../../components/content-panel";
import FlexBox from "../../components/flexbox";
import { Config } from "../../util";
import Modal, { ButtonDefinition, KeyDownDefinition } from "../../components/modal";
import AddValueButton from "../../components/add-button";
import { IoPlay, IoWarning } from "react-icons/io5";
// import Slider, { SliderTooltip, Handle } from 'rc-slider';

export default function GlobalServicesPanel(props) {
    const {data, err, config, createGlobalService, getConfig, getGlobalServices, deleteGlobalService} = useGlobalServices(Config.url, true)
    const [load, setLoad] = useState(true)

    const [serviceName, setServiceName] = useState("")
    const [image, setImage] = useState("")
    const [scale, setScale] = useState(0)
    const [size, setSize] = useState(0)
    const [cmd, setCmd] = useState("")

    console.log(data, err, config)

    useEffect(()=>{
        async function getcfg() {
            await getConfig()
            await getGlobalServices()
        }
        if(load && config === null && data === null) {
            getcfg()
            setLoad(false)
        }
    },[config, getConfig, load])

    if (err !== null) {
        // error happened with listing services
        console.log(err)
    }

    if(data === null) {
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
                        Services
                    </FlexBox>
                    <div>
                    <Modal title="New global service" 
                        escapeToCancel
                        modalStyle={{
                            maxWidth: "300px"
                        }}
                        onOpen={() => {
                            console.log("ON OPEN");
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
                            }, true)
                        ]}
                        actionButtons={[
                            ButtonDefinition("Add", async () => {
                                let err = await createGlobalService(serviceName, image, parseInt(scale), parseInt(size), cmd)
                                if(err) return err
                            }, "small blue", true, false),
                            ButtonDefinition("Cancel", () => {
                            }, "small light", true, false)
                        ]}
                    >
                        {config !== null ? 
                            <ServiceCreatePanel cmd={cmd} setCmd={setCmd} size={size} setSize={setSize} name={serviceName} setName={setServiceName} image={image} setImage={setImage} scale={scale} setScale={setScale} maxscale={config.maxscale} />
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
                                 <IoWarning/>
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