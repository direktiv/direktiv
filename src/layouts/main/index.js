import React from 'react';
import './style.css';
import Breadcrumbs from '../../components/breadcrumbs';
import FlexBox from '../../components/flexbox';
import NavBar from '../../components/navbar';

function MainLayout(props) {

    let {onClick, style, className} = props;
    return(
        <div id="main-layout" onClick={onClick} style={style} className={className}>
            <FlexBox className="row">
                {/* 
                    Left col: navigation
                    Right col: page contents 
                */}

                <FlexBox className="navigation-col">
                    <NavBar />
                </FlexBox>

                <FlexBox className="col">
                    <FlexBox className="breadcrumbs-row">
                        <Breadcrumbs/>
                    </FlexBox>
                    <FlexBox>
                    </FlexBox>
                </FlexBox>

            </FlexBox>
        </div>
    );
}

export default MainLayout;