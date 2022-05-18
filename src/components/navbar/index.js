import React, {useCallback, useState} from 'react';
import './style.css';
import Logo from '../../assets/nav-logo.png'
import FlexBox from '../flexbox';
import NamespaceSelector from '../namespace-selector';

import Modal, { KeyDownDefinition } from '../modal';
import { ButtonDefinition } from '../modal';
import {VscAdd,  VscFolderOpened, VscGraph, VscLayers, VscServer,  VscSettingsGear,  VscSymbolEvent, VscVmRunning, VscPlayCircle, VscCloudUpload} from 'react-icons/vsc';

import { Link, matchPath, useLocation, useNavigate } from 'react-router-dom';
import Tabs from '../tabs';
import { mirrorSettingInfoMetaInfo } from '../../layouts/mirror/info';
import HelpIcon from '../help';
import Tippy from '@tippyjs/react';
import {useDropzone} from 'react-dropzone';


function NavBar(props) {

    let {onClick, style, footer,  className, createNamespace, createMirrorNamespace, namespace, namespaces, createErr, toggleResponsive, setToggleResponsive, extraNavigation} = props;
    if (!className) {
        className = ""
    }

    className = "navigation-master " + className

    if (!namespace) {
        className += " loading"
    }

    if (toggleResponsive) {
        className += " toggled"
    }

    const {pathname} = useLocation()

    return (
        <>
            <ResponsiveNavbar toggled={toggleResponsive} setToggled={setToggleResponsive} />
            <FlexBox onClick={onClick} style={{...style}} className={className}>
                <FlexBox className="col tall" style={{ gap: "12px" }}>
                    <FlexBox className="navbar-logo">
                        <img alt="logo" src={Logo} />
                    </FlexBox>
                    <div className="navbar-panel shadow col">
                        <FlexBox>
                            <NamespaceSelector pathname={pathname} toggleResponsive={setToggleResponsive} namespace={namespace} namespaces={namespaces}/>
                        </FlexBox>
                        <FlexBox>
                            <NewNamespaceBtn createErr={createErr} createNamespace={createNamespace} createMirrorNamespace={createMirrorNamespace}/>
                        </FlexBox>
                        <NavItems extraNavigation={extraNavigation} pathname={pathname} toggleResponsive={setToggleResponsive} namespace={namespace} style={{ marginTop: "12px" }} />
                    </div>

                    <div className="navbar-panel shadow col">
                        <GlobalNavItems namespace={namespace}/>
                    </div>

                    {footer}
                </FlexBox>
            </FlexBox>
        </>
    );
}

export default NavBar;

export function ClientFileUpload(props) {
    const { setError, setFile, maxSize } = props

    const onDropAccepted = useCallback((acceptedFiles) => {
        acceptedFiles.forEach((file) => {
            const reader = new FileReader()

            reader.onabort = () => {
                setError("Failed to load file.")
            }
            reader.onerror = () => {
                setError("Failed to read file.")
            }
            reader.onload = () => {
                try {
                    const binaryStr = reader.result
                    var enc = new TextDecoder("utf-8");
                    let fileData = enc.decode(binaryStr)
                    setFile(fileData)
                } catch (e) {
                    setError("Failed to decode file: " + e.message)
                }
            }
            reader.readAsArrayBuffer(file)
        })

    }, [setError, setFile])

    const onDropRejected = useCallback((rejectedFiles) => {
        setError("Failed to upload file: " + rejectedFiles?.[0]?.errors?.[0]?.message)
    }, [setError])
    const { getRootProps, getInputProps } = useDropzone({ onDropAccepted, onDropRejected, multiple: false, maxSize: maxSize })

    return (
        <div {...getRootProps()} className='input-title-button'>
            <input {...getInputProps()} />
            {props.children}
        </div>
    )

}

function NewNamespaceBtn(props) {
    const {createNamespace, createMirrorNamespace} = props
    
    const [tabIndex, setTabIndex] = useState(0)

    const [mirrorSettings, setMirrorSettings] = useState({
        "url": "",
        "ref": "",
        "cron": "",
        "passphrase": "",
        "publicKey": "",
        "privateKey": "",
    })

    const [mirrorErrors, setMirrorErrors] = useState({
        "publicKey": null,
        "privateKey": null,
    })

    const [ns, setNs] = useState("")
    const navigate = useNavigate()


    

    return (
        <Modal title="New namespace"
               escapeToCancel
               modalStyle={{width: "240px"}}
               button={(
                   <FlexBox className="new-namespace-btn">
                       <div className="auto-margin">
                           <FlexBox className="row" style={{ gap: "8px", alignItems:"center" }}>
                               <FlexBox>
                                   <VscAdd />
                               </FlexBox>
                               <FlexBox>
                                   New namespace
                               </FlexBox>
                           </FlexBox>
                       </div>
                   </FlexBox>
               )}

               titleIcon={<VscAdd/>}

                onClose={() => {
                    setNs("")
                    setMirrorSettings({
                        "url": "",
                        "ref": "",
                        "cron": "",
                        "passphrase": "",
                        "publicKey": "",
                        "privateKey": "",
                    })
                    setTabIndex(0)
                }}

               keyDownActions={[
                   KeyDownDefinition("Enter", async () => {
                        await createNamespace(ns)
                        setTimeout(()=>{
                            navigate(`/n/${ns}`)
                        },200)
                        setNs("")
                   }, ()=>{}, true)
               ]}

               actionButtons={[
                   ButtonDefinition("Add", async () => {
                          if (tabIndex === 0) {
                            await createNamespace(ns)
                          } else {
                            await createMirrorNamespace(ns, mirrorSettings)
                          }
                          setTimeout(()=>{
                            navigate(`/n/${ns}`)
                          },200)
                          setNs("")
                   }, "small blue", ()=>{}, true, false, true),
                   ButtonDefinition("Cancel", () => {
                       setNs("")
                   }, "small light", ()=>{}, true, false)
               ]}

               requiredFields={[
                   {tip: "namespace is required", value: ns},
                   {tip: "mirror url is required", value: tabIndex === 0 ? true : mirrorSettings.url},
                   {tip: "mirror ref is required", value: tabIndex === 0 ? true : mirrorSettings.ref},
                   {tip: "public key is required", value: tabIndex === 0 ? true : mirrorSettings.publicKey},
                   {tip: "private key is required", value: tabIndex === 0 ? true : mirrorSettings.privateKey}
               ]}
        >
            <Tabs
                // TODO: change wf-execute-input => tabbed-form
                id={"wf-execute-input"}
                key={"inputForm"}
                callback={setTabIndex}
                tabIndex={tabIndex}
                style={{minWidth: "280px"}}
                headers={["Standard", "Mirror"]}
                tabs={[(
                    <FlexBox className="col gap-md" style={{ paddingRight: "12px" }}>
                        <FlexBox className="row gap-sm" style={{ justifyContent: "flex-start" }}>
                            <span className={`input-title`}>Namespace*</span>
                        </FlexBox>
                        <input autoFocus value={ns} onChange={(e) => setNs(e.target.value)} placeholder="Enter namespace name" />
                    </FlexBox>), (
                    <FlexBox className="col gap">
                        <FlexBox className="col gap-md" style={{ paddingRight: "12px" }}>
                            <FlexBox className="row gap-sm" style={{ justifyContent: "flex-start" }}>
                                <span className={`input-title`}>Namespace*</span>
                            </FlexBox>
                            <input autoFocus value={ns} onChange={(e) => setNs(e.target.value)} placeholder="Enter namespace name" />
                        </FlexBox>
                        {Object.entries(mirrorSettings).map(([key, value]) => {
                            return (
                                <FlexBox key={`input-new-ns-${key}`} className="col gap-md" style={{ paddingRight: "12px" }}>
                                    <FlexBox className="row" style={{ justifyContent: "space-between" }}>
                                        <FlexBox className="row gap-sm" style={{ justifyContent: "flex-start" }}>
                                            <span className={`input-title`}>{mirrorSettingInfoMetaInfo[key].plainName}{mirrorSettingInfoMetaInfo[key].required ? "*" : ""}</span>
                                            {
                                                mirrorSettingInfoMetaInfo[key].info ?
                                                    <HelpIcon msg={mirrorSettingInfoMetaInfo[key].info}/> : <></>
                                            }
                                        </FlexBox>
                                        {key === "publicKey" || key === "privateKey" ?
                                            <ClientFileUpload 
                                            setFile={(fileData)=>{
                                                let newSettings = mirrorSettings
                                                newSettings[key] = fileData
                                                setMirrorSettings({ ...newSettings })
                                            }}
                                            setError={(errorMsg)=>{
                                                let newErrors = mirrorErrors
                                                newErrors[key] = errorMsg
                                                setMirrorErrors({ ...newErrors })
                                            }}
                                            maxSize={40960}
                                            >
                                                <Tippy content={mirrorErrors[key] ? mirrorErrors[key] : `Upload key plaintext file content to ${mirrorSettingInfoMetaInfo[key].plainName} input. Warning: this will replace current ${mirrorSettingInfoMetaInfo[key].plainName} content`} trigger={'click mouseenter focus'} onHide={()=>{
                                                    let newErrors = mirrorErrors
                                                    newErrors[key] = null
                                                    setMirrorErrors({ ...newErrors })
                                                }}
                                                zIndex={10}>
                                                    <div className='input-title-button'>
                                                        <FlexBox className="row gap-sm center" style={{ justifyContent: "flex-end", marginRight: "-6px" }}>
                                                            <span  onClick={(e) => {
                                                            }}>Upload</span>
                                                            <VscCloudUpload/>
                                                        </FlexBox>
                                                    </div>
                                                </Tippy>
                                            </ClientFileUpload>
                                        :<></>}
                                    </FlexBox>
                                    {key === "publicKey" || key === "privateKey" ?
                                        <textarea style={{ width: "100%", resize: "none" }} rows={5} value={value} spellcheck="false" onChange={(e) => {
                                            let newSettings = mirrorSettings
                                            newSettings[key] = e.target.value
                                            setMirrorSettings({ ...newSettings })
                                        }} autoFocus placeholder={mirrorSettingInfoMetaInfo[key].placeholder} />
                                        :
                                        <input type={key === "passphrase" ? "password" : "text"} style={{ width: "100%" }} value={value} spellcheck="false" onChange={(e) => {
                                            let newSettings = mirrorSettings
                                            newSettings[key] = e.target.value
                                            setMirrorSettings({ ...newSettings })
                                        }} autoFocus placeholder={mirrorSettingInfoMetaInfo[key].placeholder} />
                                    }

                                </FlexBox>
                            )
                        })}
                    </FlexBox>
                )]}
            />
        </Modal>
    );
}

function NavItems(props) {

    let {pathname, style, namespace, toggleResponsive, extraNavigation} = props;

    let explorer = matchPath("/n/:namespace", pathname)
    let wfexplorer = matchPath("/n/:namespace/explorer/*", pathname)
    let monitoring = matchPath("/n/:namespace/monitoring", pathname)
    // let builder = matchPath("/n/:namespace/builder", pathname)
    let events = matchPath("/n/:namespace/events", pathname)

    // instance path matching
    let instances = matchPath("/n/:namespace/instances", pathname)
    let instanceid = matchPath("/n/:namespace/instances/:id", pathname)

    let navItemMap = {}
    if(namespace !== null && namespace !== "") {
        if(extraNavigation) {
            for(let i=0; i < extraNavigation.length; i++) {
                navItemMap[extraNavigation[i].path(namespace)] = matchPath(extraNavigation[i].path(namespace), pathname)
            }
        }
    }
    // let permissions = matchPath("/n/:namespace/permissions", pathname)

    // services pathname matching
    let services = matchPath("/n/:namespace/services", pathname)
    let service = matchPath("/n/:namespace/services/:service", pathname)
    let revision = matchPath("/n/:namespace/services/:service/:revision", pathname)

    let settings = matchPath("/n/:namespace/settings", pathname)


    return (
        <FlexBox style={{...style}} className="nav-items">
            <ul>
                <li className={`${!namespace ? "disabled-nav-item":""}`}>
                    <Link to={!!namespace && `/n/${namespace}`} onClick={() => {
                        !!namespace && toggleResponsive(false)
                    }}>
                        <NavItem className={explorer || wfexplorer ? "active":""} label="Explorer">
                            <VscFolderOpened/>
                        </NavItem>
                    </Link>
                </li>
                <li className={`${!namespace ? "disabled-nav-item":""}`}>
                    <Link to={!!namespace && `/n/${namespace}/monitoring`} onClick={() => {
                        !!namespace && toggleResponsive(false)
                    }}>
                        <NavItem className={monitoring ? "active":""} label="Monitoring">
                            <VscGraph />
                        </NavItem>
                    </Link>
                </li>
                {/* <li>
                    <Link to={`/n/${namespace}/builder`}>
                        <NavItem className={builder ? "active":""} label="Workflow Builder">
                            <IoGitNetworkOutline/>
                        </NavItem>
                    </Link>
                </li> */}
                <li className={`${!namespace ? "disabled-nav-item":""}`}>
                    <Link to={!!namespace && `/n/${namespace}/instances`} onClick={() => {
                        !!namespace && toggleResponsive(false)
                    }}>
                        <NavItem className={instances || instanceid ? "active":""} label="Instances">
                            <VscVmRunning/>
                        </NavItem>
                    </Link>
                </li>
                <li className={`${!namespace ? "disabled-nav-item":""}`}>
                    <Link to={!!namespace && `/n/${namespace}/events`} onClick={() => {
                        !!namespace && toggleResponsive(false)
                    }}>
                        <NavItem className={events ? "active":""} label="Events">
                            <VscSymbolEvent/>
                        </NavItem>
                    </Link>
                </li>
                {namespace !== null && namespace !== "" ?
                    extraNavigation?.map((obj)=>{
                        if(obj.hreflink){
                            return (
                                <li key={obj.title}>
                                    <a href={obj.path(namespace)}>
                                        <NavItem className={navItemMap[obj.path(namespace)] !== null ? "active": ""} label={obj.title}>
                                            {obj.icon}
                                        </NavItem>
                                    </a>
                                </li>
                            )
                        } else {
                            return (
                                <li key={obj.title}>
                                    <Link to={obj.path(namespace)} onClick={() => {
                                        toggleResponsive(false)
                                    }}>
                                        <NavItem className={navItemMap[obj.path(namespace)] !== null ? "active":""} label={obj.title}>
                                            {obj.icon}
                                        </NavItem>
                                    </Link>
                                </li>
                            )
                        }
                    }):""}
                {/* <li>
                    <Link to={`/n/${namespace}/permissions`} onClick={() => {
                        toggleResponsive(false)
                    }}>
                        <NavItem className={permissions ? "active":""} label="Permissions">
                            <VscLock/>
                        </NavItem>
                    </Link>
                </li> */}
                <li className={`${!namespace ? "disabled-nav-item":""}`}>
                    <Link to={!!namespace && `/n/${namespace}/services`} onClick={() => {
                        !!namespace && toggleResponsive(false)
                    }}>
                        <NavItem className={services || service || revision ? "active":""} label="Services">
                            <VscLayers/>
                        </NavItem>
                    </Link>
                </li>
                <li className={`${!namespace ? "disabled-nav-item":""}`}>
                    <Link to={!!namespace && `/n/${namespace}/settings`} onClick={() => {
                        !!namespace && toggleResponsive(false)
                    }}>
                        <NavItem className={settings ? "active":""} label="Settings">
                            <VscSettingsGear/>
                        </NavItem>
                    </Link>
                </li>

            </ul>
        </FlexBox>
    );
}

function GlobalNavItems({namespace}) {

    const {pathname} = useLocation()

    let jq = matchPath("/jq", pathname)
    let gs = matchPath("/g/services", pathname)
    let gservice = matchPath("/g/services/:service", pathname)
    let grevision = matchPath("/g/services/:service/:revision", pathname)

    let gr = matchPath("/g/registries", pathname)

    return (
        <FlexBox className="nav-items">
            <ul>
                <li className={`${!namespace ? "disabled-nav-item":""}`} style={{marginTop: "0px"}}>
                    <Link to={!!namespace && "/jq"}>
                        <NavItem className={jq ? "active":""} label="jq Playground">
                            <VscPlayCircle/>
                        </NavItem>
                    </Link>
                </li>
                <li className={`${!namespace ? "disabled-nav-item":""}`}>
                    <Link to={!!namespace && "/g/services"}>
                        <NavItem className={gs || gservice || grevision ? "active":""} label="Global Services">
                            <VscLayers />
                        </NavItem>
                    </Link>
                </li>
                <li className={`${!namespace ? "disabled-nav-item":""}`}>
                    <Link to={!!namespace && "/g/registries"}>
                        <NavItem className={gr ? "active":""} label="Global Registries">
                            <VscServer/>
                        </NavItem>
                    </Link>
                </li>
            </ul>
        </FlexBox>
    );
}

export function NavItem(props) {

    let {children, label, className} = props;
    if (!className) {
        className = ""
    }

    return (
        <FlexBox className={"nav-item " + className} style={{ gap: "8px" }}>
            <FlexBox style={{ maxWidth: "30px", width: "30px", margin: "auto" }}>
                {children}
            </FlexBox>
            <FlexBox style={{ textAlign: "left" }}>
                {label}
            </FlexBox>
        </FlexBox>
    );
}

function ResponsiveNavbar(props) {

    let {toggled, setToggled} = props;
    let panelClasses = "panel";
    let responsiveNavClasses = "responsive-nav hide-on-large";
    let responsiveNavOverlayClasses = "responsive-nav-overlay hide-on-large";

    if (toggled) {
        panelClasses += " toggled"
        responsiveNavClasses += " toggled"
        responsiveNavOverlayClasses += " toggled"
    } else {
        panelClasses += " disabled"
        responsiveNavClasses += " disabled"
        responsiveNavOverlayClasses += " disabled"
    }

    return(
        <>
            <div className={responsiveNavOverlayClasses}>

            </div>
            <FlexBox id="clickme" className={responsiveNavClasses} onClick={(e) => {
                setToggled(false)
                e.stopPropagation()
            }}>
                <div className={panelClasses}>

                </div>
            </FlexBox>
        </>
    )
}