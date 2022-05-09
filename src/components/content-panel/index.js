import React from 'react'
import FlexBox from '../flexbox';
import './style.css';

function ContentPanel(props) {

    let {style, children, className, id} = props;
    if (!className) {
        className = ""
    }

    className = "content-panel-parent opaque " + className

    return(
        <div id={id} style={{...style, display: "flex", flexDirection: "column"}} className={className} >
            {children}
        </div>
    );
}

export default ContentPanel;

export function ContentPanelTitle(props) {
    
    let {style, children, className} = props;
    if (!className) {
        className = ""
    }

    className = "content-panel-title " + className

    return(
        <FlexBox style={{...style}} className={className}>
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
    
    let {children, className, style} = props;
    if (!className) {
        className = ""
    }

    className = "content-panel-body " + className

    return(
        <div style={{...style}} className={className}>
            {children}
        </div>
    );
}

export function ContentPanelFooter(props) {

    let {children, className, style} = props;
    if (!className) {
        className = ""
    }

    className = "content-panel-footer " + className

    return (
        <div style={{...style}} className={className}>
            {children}
        </div>
    )
}

export function ContentPanelHeaderButton(props) {

    let {children, onClick, className, style, hackyStyle} = props;
    if (!className) {
        className=""
    }

    return(
        <FlexBox className={className} style={{ ...style, flexDirection: "row-reverse" }}>
            <div onClick={onClick} className="control-panel-header-button">
                <FlexBox className="shadow" style={{...hackyStyle}}>
                    {children}
                </FlexBox>
            </div>
        </FlexBox>
    );
}

export function ContentPanelHeaderButtonIcon(props) {

    let {children, style} = props;

    return(
        <FlexBox style={{...style}} className="control-panel-header-button-icon">
            {children}
        </FlexBox>
    );
}