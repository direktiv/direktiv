import { useNamespaceServices } from "direktiv-react-hooks";
import {VscLayers, VscChevronDown, VscChevronRight, VscRefresh} from 'react-icons/vsc';

import "./style.css"
import {useEffect, useState, useMemo} from "react"
import { VscTrash, VscCircleLargeFilled } from 'react-icons/vsc';

import ContentPanel, { ContentPanelBody, ContentPanelTitle, ContentPanelTitleIcon } from "../../components/content-panel";
import FlexBox from "../../components/flexbox";
import { Config, GenerateRandomKey } from "../../util";
import Modal  from "../../components/modal";
import {Link} from 'react-router-dom'
import HelpIcon from "../../components/help"
import Tippy from '@tippyjs/react';
import 'tippy.js/dist/tippy.css';

import { VscAdd } from 'react-icons/vsc';


export default function ServicesPanel(props) {
    const {namespace} = props

    if(!namespace) {
        return <></>
    }
    return(
        <FlexBox gap wrap style={{paddingRight:"8px"}}>
            <NamespaceServices namespace={namespace}/>
        </FlexBox>
    )
}

export function ServiceCreatePanel(props) {
    const {name, setName, image, setImage, scale, setScale, size, setSize, cmd, setCmd, maxScale} = props

    return(
        <FlexBox col gap style={{fontSize: "12px"}}>
                <FlexBox col gap>
                    <FlexBox col style={{paddingRight:"10px"}}>
                        Name
                        <input value={name} onChange={(e)=>setName(e.target.value)} placeholder="Enter name for service" />
                    </FlexBox>
                    <FlexBox col style={{paddingRight:"10px"}}>
                        Image
                        <input value={image} onChange={(e)=>setImage(e.target.value)} placeholder="Enter an image name" />
                    </FlexBox>
                    <FlexBox col style={{paddingRight:"10px"}}>
                        Scale
                        <Tippy content={scale} trigger={"mouseenter click"}>
                            <input type="range" style={{paddingLeft:"0px"}} min={"0"} max={maxScale.toString()} value={scale.toString()} onChange={(e)=>setScale(e.target.value)} />
                        </Tippy>
                        <datalist style={{display:"flex", alignItems:'center'}} id="sizeMarks">
                            <option style={{flex:"auto", textAlign:"left", lineHeight:"10px", paddingLeft:"8px"}} value="0" label="0"/>
                            <option style={{flex:"auto", textAlign:"right", lineHeight:"10px", paddingRight:"5px" }} value={maxScale} label={maxScale}/>
                        </datalist>
                    </FlexBox>
                    <FlexBox col style={{paddingRight:"10px"}}>
                        Size
                        <input list="sizeMarks" style={{paddingLeft:"0px"}} type="range" min={"0"} value={size.toString()}  max={"2"} onChange={(e)=>setSize(e.target.value)}/>
                        <datalist style={{display:"flex", alignItems:'center'}} id="sizeMarks">
                            <option style={{flex:"auto", textAlign:"left", lineHeight:"10px"}} value="0" label="small"/>
                            <option style={{flex:"auto", textAlign:"center" , lineHeight:"10px"}} value="1" label="medium"/>
                            <option style={{flex:"auto", textAlign:"right", lineHeight:"10px" }} value="2" label="large"/>
                        </datalist>
                    </FlexBox>
                    <FlexBox col style={{paddingRight:"10px"}}>
                        CMD
                        <input value={cmd} onChange={(e)=>setCmd(e.target.value)} placeholder="Enter the CMD for a service" />
                    </FlexBox>
                </FlexBox>
        </FlexBox>
    )
}

function NamespaceServices(props) {
    const {namespace} = props

    const [load, setLoad] = useState(true)
    const [serviceName, setServiceName] = useState("")
    const [image, setImage] = useState("")
    const [scale, setScale] = useState(0)
    const [size, setSize] = useState(0)
    const [cmd, setCmd] = useState("")
    const [maxScale, setMaxScale] = useState(0)

    const {data, err, config, getNamespaceConfig, getNamespaceServices, createNamespaceService, deleteNamespaceService} = useNamespaceServices(Config.url, true, namespace, localStorage.getItem("apikey"))

    useEffect(()=>{
        async function getcfg() {
            await getNamespaceConfig().then(response => setMaxScale(response.maxscale));
            await getNamespaceServices();
        }
        if(load && config === null && data === null) {
            getcfg()
            setLoad(false)
        }
    },[config, getNamespaceConfig, data, getNamespaceServices, load])



    if (err !== null) {
        // error happened with listing services
        console.log(err)
    }

    if(data === null) {
        return <></>
    }

    return (
        <ContentPanel style={{width:"100%", minWidth: "300px"}}>
            <ContentPanelTitle>
                <ContentPanelTitleIcon>
                    <VscLayers/>
                </ContentPanelTitleIcon>
                <FlexBox style={{display:"flex", alignItems:"center"}} gap>
                            <div>
                                Services 
                            </div>
                            <HelpIcon msg={"Services that are available to be used by workflows in the same namespace."} />
                        </FlexBox>
                <div>
                <Modal title="New namespace service" 
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
                    actionButtons={[
                        {
                            label: "Add",

                            onClick: async () => {
                                await createNamespaceService(serviceName, image, parseInt(scale), parseInt(size), cmd)
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
                    requiredFields={[
                        {tip: "service name is required", value: serviceName},
                        {tip: "image is required", value: image}
                    ]}
                >
                    {config !== null ? 
                        <ServiceCreatePanel cmd={cmd} setCmd={setCmd} size={size} setSize={setSize} name={serviceName} setName={setServiceName} image={image} setImage={setImage} scale={scale} setScale={setScale} maxScale={maxScale} />
                        :
                        ""
                    }
                </Modal>
            </div>
            </ContentPanelTitle>
            <ContentPanelBody className="secrets-panel">
                <FlexBox col gap>
                    <FlexBox col gap>
                        {data.length === 0 ?
                        <div className="col">
                        <FlexBox style={{ height:"40px", }}>
                                <FlexBox gap style={{alignItems:"center", paddingLeft:"8px"}}>
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
                                        url={`/n/${namespace}/services/${obj.info.name}`} 
                                        deleteService={deleteNamespaceService} 
                                        conditions={obj.conditions} 
                                        name={obj.info.name} 
                                        status={obj.status} 
                                        image={obj.info.image} 
                                    />
                                )
                            })
                        }
                        </>}
                    </FlexBox>
                </FlexBox>
            </ContentPanelBody>
        </ContentPanel>
    );
}

export function Service(props) {
    const {allowRedeploy, name, image, status, conditions, deleteService, url, revision, dontDelete, traffic, latest} = props
    return (
        <div className="col" style={{minWidth: "300px"}}>
            <FlexBox style={{ height:"40px", border:"1px solid #f4f4f4", backgroundColor:"#fcfdfe"}}>
                <Link to={url} style={{ width: "100%", display: "flex", alignItems: "center" }}>
                    <FlexBox gap style={{alignItems:"center", paddingLeft:"8px"}}>
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
                </Link>
                {!dontDelete && !traffic ? 
                <>
                        {latest ? 
                         <div style={{height: "100%", display: "flex", paddingRight: "26px" }}>
                         <HelpIcon msg={"Unable to delete latest revision"} />

                         </div>
                    :
                <div style={{paddingRight:"25px", maxWidth:"20px", margin: "auto"}}>
                    <Modal  title="Delete namespace service" 

                        escapeToCancel
                        modalStyle={{
                            maxWidth: "400px",
                            width: "400px"
                        }}
                        onOpen={() => {
                        }}
                        onClose={()=>{
                        }}
                        button={(
                            <ServicesDeleteButton />
                        )}
                        buttonProps={{
                            color: "info"
                        }} 
                        actionButtons={[
                            {
                                label: "Delete",

                                onClick: async () => {
                                    if(revision !== undefined) {
                                        await deleteService(revision)
                                    }else {
                                        await deleteService(name)
                                    }
                                 
                                },

                                buttonProps: {variant: "contained", color: "error"},
                                errFunc: ()=>{},
                                closesModal: true
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
                        <FlexBox col gap>
                            <FlexBox >
                                Are you sure you want to delete '{name}'?
                                <br/>
                                This action cannot be undone.
                            </FlexBox>
                        </FlexBox>
                    </Modal>
                </div>}
                </>
                : 
                    <>
                    {traffic ?     
                    <div style={{paddingRight:"25px", maxWidth:"20px", margin: "auto", fontSize:"10pt", fontWeight:"bold"}}>
                        {traffic}%
                    </div>:""}
                    </>
                }
                { allowRedeploy ? 
                <div>
                    <FlexBox style={{ alignItems: "center", justifyContent: "center", height: "100%", paddingRight: "6px" }}>
                        <Modal
                            title="Redeploy service"
                            modalStyle={{width: "400px"}}
                            titleIcon={(
                                <VscRefresh />
                            )}
                            button={(
                                <VscRefresh className="grey-text" style={{ fontSize: "16px" }} />
                            )}
                            buttonProps={{
                                auto: true,
                            }}
                            actionButtons={[
                                {
                                    label: "Yes",

                                    onClick: async () => {
                                        await deleteService(name, revision)
                                    },

                                    buttonProps: {variant: "contained", color: "primary"},

                                    errFunc: () => {
                                        console.log("err func")
                                    },

                                    closesModal: true
                                },
                                {
                                    label: "Cancel",
                                    onClick: () => {},
                                    buttonProps: {},
                                    errFunc: () => {},
                                    closesModal: true
                                }
                            ]}
                        >
                            <div style={{ textAlign: "center" }}>
                                <div>
                                    This will delete the pods running to support this service.
                                </div>
                                <div>
                                    The pods will be recreated the next time an action is executed that requires this service.
                                </div>
                                <br/>
                                <div>
                                    Do you wish to continue?
                                </div>
                            </div>
                        </Modal>
                    </FlexBox>
                </div>
                :<></>}
            </FlexBox>
            <FlexBox style={{border:"1px solid #f4f4f4", borderTop:"none"}}>
                <ServiceDetails conditions={conditions} />
            </FlexBox>
        </div>
    );
}

function ServiceDetails(props) {
    const {conditions} = props

    return(
        <ul className="condition-list" style={{listStyle:"none", paddingLeft:"25px", paddingRight:"40px", width:"100%"}}>
            {conditions.map((obj)=>{
                if(obj.name === 'Active' && obj.reason === 'NoTraffic' && obj.message === "The target is not receiving traffic."){
                    return(
                        <Condition key={GenerateRandomKey('service-condition-')} status={"True"} name={obj.name} reason={""} message={""} />
                    )
                }
                return(
                    <Condition key={GenerateRandomKey('service-condition-')} status={obj.status} name={obj.name} reason={obj.reason} message={obj.message}/>
                )
            })}

        </ul>
    )
}

function Condition(props){
    const {status, name, reason, message} = props

    const [showDetails, setShowDetails] = useState(false)

    let waitMsgClasses = "wait-message "
    let failMsgClasses = "fail-message "

    if (showDetails) {
        waitMsgClasses += "visible"
        failMsgClasses += "visible"
    }

    return (
        <li style={{ display: "flex", gap: "8px" }}>
            <FlexBox col>
                <FlexBox gap>
                    <div>
                        <ServiceStatus status={status}/>
                    </div>
                    <FlexBox gap>
                        <FlexBox>
                            {name}
                        </FlexBox>
                        {status !== 'True' && reason !== "" && message !== "" ?
                        <FlexBox style={{
                            maxWidth: "120px"
                        }}>
                            <div className="toggle-details" onClick={() => {setShowDetails(!showDetails)}} style={{
                                color:"#dbd9d9", 
                                display:"flex", 
                                alignItems:"center", 
                                fontSize:"10pt", 
                                cursor:"pointer"
                            }}>
                            {showDetails ?
                            <>
                                <VscChevronDown />
                                <div>Hide Details</div>
                            </>
                            :   
                            <>
                                <VscChevronRight />
                                <div>Show Details</div>
                            </>}
                            </div>
                        </FlexBox>
                        :""}
                    </FlexBox>
                </FlexBox>
                <FlexBox>
                {status === 'Unknown' ?
                    <div className={waitMsgClasses} style={{marginLeft:"14px"}}>
                {reason !== ""  ? 
                    <div className="grey-text" style={{fontSize:"10pt", fontStyle:"italic"}}>
                        {reason}
                    </div>:""}
                {message !== "" ? 
                    <div>
                        <div className="msg-box">
                            {message}
                        </div>
                    </div>
                :""}
                    </div>
                :""}

                { status === 'False' ? 
                    <div className={failMsgClasses} style={{marginLeft:"14px"}}>
                        <div className="grey-text" style={{fontSize: "10pt", fontStyle: "italic"}}>
                            {reason}
                        </div>
                        <div className="msg-box">
                            {message}
                        </div>
                    </div>
                :""}
                </FlexBox>
            </FlexBox>
        </li>
    )
}

export function ServiceStatus(props) {
    const color =  useMemo(()=>{
        if (props.status) {
            switch (props.status) {
                case "False":
                    return "#FF616D"
                case "Unknown":
                    return "#082032"
                default:
                    break;
            }
        }

        // default status color
        return "#66DE93"

    },[props])

    return(
        <div>   
            <VscCircleLargeFilled style={{fontSize:"6pt", fill: color}} />
        </div>
    )
}

function ServicesDeleteButton() {
    return (
        <div className="red-text" style={{ display: "flex", alignItems: "center", height: "100%" }}>
            <VscTrash className="auto-margin" />
        </div>
    )
}
