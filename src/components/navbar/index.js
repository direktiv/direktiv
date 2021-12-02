import React, {useState} from 'react';
import './style.css';
import Logo from '../../assets/nav-logo.png'
import FlexBox from '../flexbox';
import NamespaceSelector from '../namespace-selector';

import Modal, { KeyDownDefinition } from '../modal';
import { ButtonDefinition } from '../modal';
import {BsSpeedometer, BsFolder2Open, BsSliders, BsCodeSquare} from 'react-icons/bs';
import {IoLockClosedOutline, IoCubeOutline, IoExtensionPuzzleOutline, IoGlobeOutline, IoLogOutOutline} from 'react-icons/io5';
import {GrFormAdd} from 'react-icons/gr'
import { Link, matchPath, useLocation, useNavigate } from 'react-router-dom';

function NavBar(props) {

    let {onClick, style, className, createNamespace, namespace, namespaces, createErr, toggleResponsive, setToggleResponsive} = props;

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
                            <NamespaceSelector toggleResponsive={setToggleResponsive} namespace={namespace} namespaces={namespaces}/>
                        </FlexBox>
                        <FlexBox>
                            <NewNamespaceBtn createErr={createErr} createNamespace={createNamespace} />
                        </FlexBox>
                        <NavItems toggleResponsive={setToggleResponsive} namespace={namespace} style={{ marginTop: "12px" }} />
                    </div>

                    <div className="navbar-panel shadow col">
                        <GlobalNavItems />
                    </div>

                    <FlexBox>
                        <FlexBox className="nav-items" style={{ paddingLeft: "10px" }}>
                            <ul style={{ marginTop: "0px" }}>
                                <li>
                                    <NavItem className="red-text" label="Log Out">
                                        <IoLogOutOutline/>
                                    </NavItem>
                                </li>
                            </ul>
                        </FlexBox>
                    </FlexBox>

                    <div>
                        <FlexBox className="col navbar-userinfo">
                            <FlexBox className="navbar-username">
                                UserName007
                            </FlexBox>
                            <FlexBox className="navbar-version">
                                Version: 0.5.8 (abdgdj)
                            </FlexBox>
                        </FlexBox>
                    </div>

                </FlexBox>
            </FlexBox>
        </>
    );
}

export default NavBar;

function NewNamespaceBtn(props) {
    const {createNamespace} = props

    // createErr is filled when someone tries to create namespace but proceeded to error out


    const [ns, setNs] = useState("")
    const navigate = useNavigate()

    return (
        <Modal title="New namespace" 
            escapeToCancel
            button={(
                <FlexBox className="new-namespace-btn">
                    <div className="auto-margin">
                        <FlexBox className="row" style={{ gap: "8px" }}>
                            <FlexBox>
                                +
                            </FlexBox>
                            <FlexBox>
                                New namespace
                            </FlexBox>
                        </FlexBox>
                    </div>
                </FlexBox>
            )} 

            titleIcon={<GrFormAdd />}

            onClose={ () => {setNs("")}}

            keyDownActions={[
                KeyDownDefinition("Enter", async () => {
                    let err = await createNamespace(ns)
                    if(err) return err
                    setTimeout(()=>{
                        navigate(`/n/${ns}`)
                    },200)
                    setNs("")
                }, true)
            ]}

            actionButtons={[
                ButtonDefinition("Add", async () => {
                    let err = await createNamespace(ns)
                    if(err) return err
                    setTimeout(()=>{
                        navigate(`/n/${ns}`)
                    },200)
                    setNs("")
                }, "small blue", true, false),
                ButtonDefinition("Cancel", () => {
                    setNs("")
                }, "small light", true, false)
            ]}
        >
            <FlexBox>
                <input autoFocus value={ns} onChange={(e)=>setNs(e.target.value)} placeholder="Enter namespace name" />
            </FlexBox>
        </Modal>
    );
}

function NavItems(props) {

    let {style, namespace, toggleResponsive} = props;

    const {pathname} = useLocation()

    let explorer = matchPath("/n/:namespace", pathname)
    let monitoring = matchPath("/n/:namespace/monitoring", pathname)
    // let builder = matchPath("/n/:namespace/builder", pathname)
    let events = matchPath("/n/:namespace/events", pathname)

    // instance path matching
    let instances = matchPath("/n/:namespace/instances", pathname)
    let instanceid = matchPath("/n/:namespace/instances/:id", pathname)
    
    let permissions = matchPath("/n/:namespaces/permissions", pathname)

    // services pathname matching
    let services = matchPath("/n/:namespace/services", pathname)
    let service = matchPath("/n/:namespace/services/:service", pathname)
    let revision = matchPath("/n/:namespace/services/:service/:revision", pathname)

    let settings = matchPath("/n/:namespace/settings", pathname)


    return (
        <FlexBox style={{...style}} className="nav-items">
            <ul>
                <li>
                    <Link to={`/n/${namespace}`} onClick={() => {
                        toggleResponsive(false)
                    }}>
                        <NavItem className={explorer ? "active":""} label="Explorer">
                            <BsFolder2Open/>
                        </NavItem>
                    </Link>
                </li>
                <li>
                    <Link to={`/n/${namespace}/monitoring`} onClick={() => {
                        toggleResponsive(false)
                    }}>
                        <NavItem className={monitoring ? "active":""} label="Monitoring">
                            <BsSpeedometer/>
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
                <li>
                    <Link to={`/n/${namespace}/instances`} onClick={() => {
                        toggleResponsive(false)
                    }}>
                        <NavItem className={instances || instanceid ? "active":""} label="Instances">
                            <BsCodeSquare/>
                        </NavItem>
                    </Link>
                </li>
                <li>
                    <Link to={`/n/${namespace}/events`} onClick={() => {
                        toggleResponsive(false)
                    }}>
                        <NavItem className={events ? "active":""} label="Events">
                            <BsCodeSquare/>
                        </NavItem>
                    </Link>
                </li>
                <li>
                    <Link to={`/n/${namespace}/permissions`} onClick={() => {
                        toggleResponsive(false)
                    }}>
                        <NavItem className={permissions ? "active":""} label="Permissions">
                            <IoLockClosedOutline/>
                        </NavItem>
                    </Link>
                </li>
                <li>
                    <Link to={`/n/${namespace}/services`} onClick={() => {
                        toggleResponsive(false)
                    }}>
                        <NavItem className={services || service || revision ? "active":""} label="Services">
                            <IoCubeOutline/>
                        </NavItem>
                    </Link>
                </li>
                <li>
                    <Link to={`/n/${namespace}/settings`} onClick={() => {
                        toggleResponsive(false)
                    }}>
                        <NavItem className={settings ? "active":""} label="Settings">
                            <BsSliders/>
                        </NavItem>
                    </Link>
                </li>
            </ul>
        </FlexBox>
    );
}

function GlobalNavItems(props) {

    const {pathname} = useLocation()

    let jq = matchPath("/jq", pathname)
    let gs = matchPath("/g/services", pathname)
    let gservice = matchPath("/g/services/:service", pathname)
    let grevision = matchPath("/g/services/:service/:revision", pathname)

    let gr = matchPath("/g/registries", pathname)

    return (
        <FlexBox className="nav-items">
            <ul>
                <li style={{marginTop: "0px"}}>
                    <Link  to={"/jq"}>
                        <NavItem className={jq ? "active":""} label="jq Playground">
                            <IoExtensionPuzzleOutline/>
                        </NavItem>
                    </Link>
                </li>
                <li>
                    <Link to={"/g/services"}>
                        <NavItem className={gs || gservice || grevision ? "active":""} label="Global Services">
                            <IoGlobeOutline/>
                        </NavItem>
                    </Link>
                </li>
                <li>
                    <Link to={"/g/registries"}>
                        <NavItem className={gr ? "active":""} label="Global Registries">
                            <IoGlobeOutline/>
                        </NavItem>
                    </Link>
                </li>
            </ul>
        </FlexBox>
    );
}

function NavItem(props) {

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