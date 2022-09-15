import { useEffect, useState } from "react"
import { VscLayers } from 'react-icons/vsc';
import { useNavigate, useParams } from "react-router"
import { Service } from "../namespace-services"
import { RevisionCreatePanel, UpdateTraffic } from "../namespace-services/revisions"
import ContentPanel, { ContentPanelBody, ContentPanelTitle, ContentPanelTitleIcon } from "../../components/content-panel"
import FlexBox from "../../components/flexbox"
import Modal  from "../../components/modal"
import { Config } from "../../util"
import { useGlobalService } from "direktiv-react-hooks"

import { VscAdd } from 'react-icons/vsc';


export default function GlobalRevisionsPanel(props){
    const {service} = useParams()
    const navigate = useNavigate()
    const {revisions, config, traffic, createGlobalServiceRevision, deleteGlobalServiceRevision, setGlobalServiceRevisionTraffic, getServiceConfig} = useGlobalService(Config.url, service, navigate, localStorage.getItem("apikey"))

    const [load, setLoad] = useState(true)
    const [image, setImage] = useState("")
    const [scale, setScale] = useState(0)
    const [size, setSize] = useState(0)
    const [trafficPercent, setTrafficPercent] = useState(100)
    const [cmd, setCmd] = useState("")
    const [maxScale, setMaxScale] = useState(0)

    useEffect(()=>{
        if(revisions !== null && revisions.length > 0) {
            setScale(revisions[0].minScale)
            setSize(revisions[0].size)
            setImage(revisions[0].image)
            setCmd(revisions[0].cmd)
        }
    },[revisions])

    useEffect(()=>{
        async function cfgGet() {
            try {
                await getServiceConfig().then(response => setMaxScale(response.maxscale));
            } catch(e) {
                if(e.message === "get global service: not found"){
                    navigate(`/not-found`)
                }
            }
        }
        if(load && config === null) {
            cfgGet()
            setLoad(false)
        }
    },[config, getServiceConfig, load, navigate])

    if(revisions === null) {
        return <></>
    }

    return (
        <FlexBox gap wrap style={{paddingRight:"8px"}}>
            <FlexBox  gap>
                    <FlexBox>
                        <ContentPanel style={{width:"100%"}}>
                        <ContentPanelTitle>
                            <ContentPanelTitleIcon>
                                <VscLayers/>
                            </ContentPanelTitleIcon>
                            <FlexBox>
                                Service '{service}' Revisions
                            </FlexBox>
                            <div>
                            <Modal title={`New '${service}' revision`} 
                                escapeToCancel
                                modalStyle={{
                                    maxWidth: "300px"
                                }}
                                onOpen={() => {
                                }}
                                onClose={()=>{
                                }}
                                button={(
                                    <VscAdd/>
                                )}
                                buttonProps={{
                                    auto: true,
                                }}
                                keyDownActions={[
                                    {
                                        code: "Enter",

                                        fn: async () => {
                                        },

                                        errFunc: ()=>{},
                                        closeModal: true
                                    }
                                ]}
                                requiredFields={[
                                    {tip: "image is required", value: image}
                                ]}
                                actionButtons={[
                                    {
                                        label: "Add",

                                        onClick: async () => {
                                            await createGlobalServiceRevision(image, parseInt(scale), parseInt(size), cmd, parseInt(trafficPercent))
                                        },

                                        buttonProps: {variant: "contained", color: "primary"},
                                        errFunc: ()=>{},
                                        closesModal: true,
                                        validate: true
                                    },
                                    {
                                        label: "Cancel",

                                        onClick: () => {
                                        },

                                        buttonProps: {},
                                        errFunc: ()=>{},
                                        closesModal: true
                                    }
                                ]}
                            >
                                {config !== null ? 
                                <RevisionCreatePanel 
                                    image={image} setImage={setImage}
                                    scale={scale} setScale={setScale}
                                    size={size} setSize={setSize}
                                    cmd={cmd} setCmd={setCmd}
                                    traffic={trafficPercent} setTraffic={setTrafficPercent}
                                    maxScale={maxScale}
                                />:""}
                            </Modal>
                        </div>
                        </ContentPanelTitle>
                            <ContentPanelBody className="secrets-panel">
                                <FlexBox col gap>
                                    <FlexBox col gap>
                                        {
                                            revisions.sort((a, b)=> (a.created > b.created) ? -1 : 1).map((obj, key)=>{
                                            let dontDelete = false
                                            if(revisions.length === 1) {
                                                dontDelete = true
                                            }
                                            let t = 0
                                            if(traffic && typeof traffic == typeof [])
                                                for(var i=0; i < traffic.length; i++) {
                                                    if(traffic[i].revisionName === obj.name){
                                                        dontDelete= true
                                                        t= traffic[i].traffic
                                                        break
                                                    }
                                                }
                                            return(
                                                <Service 
                                                    latest={key===0}
                                                    traffic={t}
                                                    key={key}
                                                    dontDelete={dontDelete && key !== 0}
                                                    revision={obj.rev}
                                                    deleteService={deleteGlobalServiceRevision}
                                                    url={`/g/services/${service}/${obj.rev}`}
                                                    conditions={obj.conditions}
                                                    name={obj.name}
                                                    status={obj.status}
                                                />
                                            )
                                        })}
                                    </FlexBox>
                                </FlexBox>
                            </ContentPanelBody>
                        </ContentPanel>
                    </FlexBox>
                    {
                        traffic &&
                        <UpdateTraffic setNamespaceServiceRevisionTraffic={setGlobalServiceRevisionTraffic} service={service} revisions={revisions} traffic={traffic}/>
                    }
                    </FlexBox>
        </FlexBox>
    );
}