import React from 'react'
import FlexBox from '../flexbox';
import './style.css';

function ContentPanel(props) {

    let {style, children, className} = props;
    if (!className) {
        className = ""
    }

    className = "content-panel-parent opaque " + className

    return(
        <div style={{...style}} className={className} >
            {children}
        </div>
    );
}

export default ContentPanel;

export function ContentPanelTitle(props) {
    
    let {children, className} = props;
    if (!className) {
        className = ""
    }

    className = "content-panel-title " + className

    return(
        <FlexBox className={className}>
            {children}
        </FlexBox>
    );
}

export function ContentPanelTitleIcon(props) {
    
    let {children, className} = props;
    if (!className) {
        className = ""
    }

    className = "content-panel-title-icon " + className

    return(
        <div className={className}>
            {children}
        </div>
    );
}

export function ContentPanelBody(props) {
    
    let {children, className} = props;
    if (!className) {
        className = ""
    }

    className = "content-panel-body " + className

    return(
        <div className={className}>
            {children}
        </div>
    );
}

export function ContentPanelHeaderButton(props) {

    let {children, onClick} = props;

    return(
        <FlexBox style={{ flexDirection: "row-reverse" }}>
            <div onClick={onClick} className="control-panel-header-button">
                <FlexBox className="shadow">
                    {children}
                </FlexBox>
            </div>
        </FlexBox>
    );
}

export function ContentPanelHeaderButtonIcon(props) {

    let {children} = props;

    return(
        <FlexBox className="control-panel-header-button-icon">
            {children}
        </FlexBox>
    );
}