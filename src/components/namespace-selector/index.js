import React, {useState} from 'react';
import './style.css';
import FlexBox from '../flexbox';
import {IoChevronDown} from 'react-icons/io5';

function NamespaceSelector(props) {

    let {style, className} = props;
    if (!className) {
        className = ""
    }
    className += " border"
    
    const [showSelector, setShowSelector] = useState(false);
    let selectorClass = "selector-section hidden";
    let selectorBorderClass = "selector-border hidden"
    let chevronClass = "chevron-icon"

    if (showSelector) {
        selectorBorderClass = "selector-border"
        selectorClass = "selector-section"
        chevronClass = "chevron-icon spin"
    }

    return (
        <>
            <FlexBox className="col gap">
                <FlexBox onClick={() => {
                    setShowSelector(!showSelector)
                }} style={{...style, maxHeight: "64px"}} className={className}>
                    <FlexBox className="namespace-selector">
                        <NamespaceListItem/>
                        <FlexBox className="tall">
                            <div className="auto-margin grey-text">
                                <IoChevronDown className={chevronClass} style={{ marginTop: "8px" }} />
                            </div>
                        </FlexBox>
                    </FlexBox>
                </FlexBox>
                <FlexBox className={selectorBorderClass}>
                    <FlexBox className={selectorClass}>
                        <NamespaceList />
                    </FlexBox>
                </FlexBox>
            </FlexBox>
        </>
    );
}

export default NamespaceSelector;

function NamespaceListItem(props) {
    return (
        <FlexBox className="namespace-list-item" style={{height: "45px", minHeight: "45px", maxHeight: "45px"}}>
            <FlexBox className="">
                <FlexBox className="namespace-selector-logo">
                    <div className="auto-margin">
                        IMG
                    </div>
                </FlexBox>
                <FlexBox className="col">
                    <div className="auto-margin" style={{marginLeft: "8px"}}>
                        <FlexBox className="namespace-selector-label-header">
                            LOGGED IN
                        </FlexBox>
                        <FlexBox className="namespace-selector-label-value">
                            Namespace Inc.
                        </FlexBox>
                    </div>
                </FlexBox>
            </FlexBox>
        </FlexBox>
    );
}

function NamespaceList(props){
    return (
        <FlexBox className="namespaces-list gap col">
            <NamespaceListItem/>
            <NamespaceListItem/>
            <NamespaceListItem/>
        </FlexBox>
    );
}