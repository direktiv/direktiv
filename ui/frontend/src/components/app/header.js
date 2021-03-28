import React from 'react';

import Navbar from 'react-bootstrap/Navbar'
import NavbarDropdown from 'react-bootstrap/NavDropdown'
import {useHistory} from 'react-router-dom'

import BannerLogo from 'img/banner-logo.png'


export default function TopNavbar(props) {
    const history = useHistory()

    let dropdownTitle = (
        <span>
                </span>
    )

    return (
        <div className="nav-top">

            <Navbar id="primary-nav" bg="dark" variant="dark">
                <Navbar.Brand style={{cursor: 'pointer'}} onClick={() => history.push('/')}>
                    <img
                        alt=""
                        src={BannerLogo}
                        height="40"
                        className="d-inline-block align-top"
                    />{' '}
                </Navbar.Brand>
                <NavbarDropdown alignRight className="custom-navbar-btn" menualign="right" title={dropdownTitle}
                                variant="success">
                    <NavbarDropdown.Item
                        onClick={() => window.open("https://docs.direktiv.io/", "_blank")}>Documentation</NavbarDropdown.Item>
                </NavbarDropdown>
            </Navbar>
        </div>
    );
}