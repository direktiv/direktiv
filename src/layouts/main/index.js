import React from 'react';
import './style.css';
import Breadcrumbs from '../../components/breadcrumbs';
import Settings from '../settings';
import FlexBox from '../../components/flexbox';
import NavBar from '../../components/navbar';

import { BrowserRouter, Routes, Route, Link } from 'react-router-dom'

function MainLayout(props) {
    let {onClick, style, className} = props;

    return(
        <div id="main-layout" onClick={onClick} style={style} className={className}>
            <FlexBox className="row gap tall" style={{minHeight: "100vh"}}>
                {/* 
                    Left col: navigation
                    Right col: page contents 
                */}

                <FlexBox className="navigation-col">
                    <NavBar />
                </FlexBox>

                <BrowserRouter>
                    <FlexBox className="content-col col">
                        <FlexBox className="breadcrumbs-row">
                            <Breadcrumbs/>
                        </FlexBox>
                        <FlexBox className="col" style={{paddingBottom: "8px"}}>
                            <Routes>
                                <Route path="/" element={<div>index route:)</div>} />
                                <Route path="/settings" element={<Settings/>} />
                            </Routes>
                        </FlexBox>
                    </FlexBox>
                </BrowserRouter>

            </FlexBox>
        </div>
    );
}

export default MainLayout;