import React from 'react';
import './style.css';
import Logo from '../../assets/nav-logo.png'
import FlexBox from '../flexbox';
import NamespaceSelector from '../namespace-selector';

import {BsSpeedometer, BsFolder2Open, BsSliders, BsCodeSquare} from 'react-icons/bs';
import {IoGitNetworkOutline, IoLockClosedOutline, IoCubeOutline, IoExtensionPuzzleOutline, IoGlobeOutline, IoLogOutOutline} from 'react-icons/io5';

function NavBar(props) {

    let {onClick, style, className} = props;
    
    if (!className) {
        className = ""
    }

    return (
        <FlexBox onClick={onClick} style={{...style}} className={className}>
            <FlexBox className="col" style={{ gap: "12px" }}>
                <FlexBox className="navbar-logo">
                    <img alt="logo" src={Logo} />
                </FlexBox>

                <FlexBox className="navbar-panel shadow col">
                    <FlexBox>
                        <NamespaceSelector/>
                    </FlexBox>
                    <FlexBox>
                        <NewNamespaceBtn />
                    </FlexBox>
                    <NavItems style={{ marginTop: "12px" }} />
                </FlexBox>

                <FlexBox className="navbar-panel shadow col">
                    <GlobalNavItems />
                </FlexBox>

                <FlexBox>
                    <FlexBox className="nav-items" style={{ paddingLeft: "10px" }}>
                        <ul style={{ marginTop: "0px" }}>
                            <li>
                                <NavItem className="alert" label="Log Out">
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

    return (
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
    );
}

function NavItems(props) {

    let {style} = props;

    return (
        <FlexBox style={{...style}} className="nav-items">
            <ul>
                <li>
                    <NavItem label="Dashboard">
                        <BsSpeedometer/>
                    </NavItem>
                </li>
                <li>
                    <NavItem label="Explorer">
                        <BsFolder2Open/>
                    </NavItem>
                </li>
                <li>
                    <NavItem label="Workflow Builder">
                        <IoGitNetworkOutline/>
                    </NavItem>
                </li>
                <li>
                    <NavItem label="Instances">
                        <BsCodeSquare/>
                    </NavItem>
                </li>
                <li>
                    <NavItem label="Permissions">
                        <IoLockClosedOutline/>
                    </NavItem>
                </li>
                <li>
                    <NavItem label="Services">
                        <IoCubeOutline/>
                    </NavItem>
                </li>
                <li>
                    <NavItem label="Settings">
                        <BsSliders/>
                    </NavItem>
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