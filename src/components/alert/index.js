import React from 'react';
import './style.css';
import FlexBox from '../flexbox';
import { VscWarning } from 'react-icons/vsc';

function Alert(props){

    let {children, style, className} = props;

    if (!className) {
        className = ""
    }
    className = "alert " + className

    return(
        <FlexBox className={className} style={{...style}}>
            <FlexBox className="alert-opacity">
                <FlexBox className="alert-icon auto-margin" style={{maxWidth: "60px"}}>
                    <VscWarning className="auto-margin" />
                </FlexBox>
                <FlexBox className="alert-body auto-margin">
                    {children}
                </FlexBox>
            </FlexBox>
        </FlexBox>
    );
}

export default Alert;