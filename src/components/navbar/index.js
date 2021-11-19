import React, {useState} from 'react';
import './style.css';
import Logo from '../../assets/nav-logo.png'
import FlexBox from '../flexbox';
import NamespaceSelector from '../namespace-selector';

import Modal from '../modal';
import { ButtonDefinition } from '../modal';
import {BsSpeedometer, BsFolder2Open, BsSliders, BsCodeSquare} from 'react-icons/bs';
import {IoGitNetworkOutline, IoLockClosedOutline, IoCubeOutline, IoExtensionPuzzleOutline, IoGlobeOutline, IoLogOutOutline} from 'react-icons/io5';
import { Link } from 'react-router-dom';

function NavBar(props) {

    let {onClick, style, className, createNamespace, namespace, namespaces} = props;
    
    if (!className) {
        className = ""
    }

    return (
        <FlexBox onClick={onClick} style={{...style}} className={className}>
            <FlexBox className="col tall" style={{ gap: "12px" }}>
                <FlexBox className="navbar-logo">
                    <img alt="logo" src={Logo} />
                </FlexBox>

                {/* <div className="navbar-panel shadow col">
                    <FlexBox>
                        <NamespaceSelector/>
                    </FlexBox>
                    <FlexBox>
                        <NewNamespaceBtn />
                    </FlexBox>
                </div>

                <div className="navbar-panel shadow col">
                    <NavItems />
                </div> */}


                <div className="navbar-panel shadow col">
                    <FlexBox>
                        <NamespaceSelector namespace={namespace} namespaces={namespaces}/>
                    </FlexBox>
                    <FlexBox>
                        <NewNamespaceBtn createNamespace={createNamespace} />
                    </FlexBox>
                    <NavItems namespace={namespace} style={{ marginTop: "12px" }} />
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

                <FlexBox className="col navbar-userinfo">
                    <FlexBox className="navbar-username">
                        UserName007
                    </FlexBox>
                    <FlexBox className="navbar-version">
                        Version: 0.5.8 (abdgdj)
                    </FlexBox>
                </FlexBox>

            </FlexBox>
        </FlexBox>
    );
}

export default NavBar;

function NewNamespaceBtn(props) {
    const {createNamespace} = props

    const [ns, setNs] = useState("")

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
            actionButtons={[
                ButtonDefinition("Add", () => {
                    createNamespace(ns)
                    setNs("")
                }, "small blue", true, false),
                ButtonDefinition("Cancel", () => {
                    console.log("close modal");
                    setNs("")
                }, "small light", true, false)
            ]}
        >
            <FlexBox>
                <input value={ns} onChange={(e)=>setNs(e.target.value)} placeholder="Enter namespace name" />
            </FlexBox>
        </Modal>
    );
}

function NavItems(props) {

    let {style, namespace} = props;

    return (
        <FlexBox style={{...style}} className="nav-items">
            <ul>
                <li>
                    <Link to={`/n/${namespace}`}>
                        <NavItem label="Explorer">
                            <BsFolder2Open/>
                        </NavItem>
                    </Link>
                </li>
                <li>
                    <Link to={`/n/${namespace}/monitoring`}>
                        <NavItem label="Monitoring">
                            <BsSpeedometer/>
                        </NavItem>
                    </Link>
                </li>
                <li>
                    <Link to={`/n/${namespace}/builder`}>
                        <NavItem label="Workflow Builder">
                            <IoGitNetworkOutline/>
                        </NavItem>
                    </Link>
                </li>
                <li>
                    <Link to={`/n/${namespace}/instances`}>
                        <NavItem label="Instances">
                            <BsCodeSquare/>
                        </NavItem>
                    </Link>
                </li>
                <li>
                    <Link to={`/n/${namespace}/permissions`}>
                        <NavItem label="Permissions">
                            <IoLockClosedOutline/>
                        </NavItem>
                    </Link>
                </li>
                <li>
                    <Link to={`/n/${namespace}/services`}>
                        <NavItem label="Services">
                            <IoCubeOutline/>
                        </NavItem>
                    </Link>
                </li>
                <li>
                    <Link to={`/n/${namespace}/settings`}>
                        <NavItem label="Settings">
                            <BsSliders/>
                        </NavItem>
                    </Link>
                </li>
            </ul>
        </FlexBox>
    );
}

function GlobalNavItems(props) {
    return (
        <FlexBox className="nav-items">
            <ul>
                <li style={{marginTop: "0px"}}>
                    <NavItem label="jq Playground">
                        <IoExtensionPuzzleOutline/>
                    </NavItem>
                </li>
                <li>
                    <NavItem label="Global Services">
                        <IoGlobeOutline/>
                    </NavItem>
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