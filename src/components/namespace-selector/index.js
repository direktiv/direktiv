import React, {useState} from 'react';
import './style.css';
import FlexBox from '../flexbox';
import {IoChevronDown} from 'react-icons/io5';

function NamespaceSelector(props) {

    let {data, style, className} = props;
    if (!className) {
        className = ""
    }
    className += " border"
    
    const [showSelector, setShowSelector] = useState(false);
    let selectorClass = "selector-section hidden";
    let selectorBorderClass = "selector-border hidden"
    let chevronClass = "chevron-icon"
    let namespaceSelectorClass = "namespace-selector"

    if (showSelector) {
        selectorBorderClass = "selector-border"
        selectorClass = "selector-section"
        chevronClass = "chevron-icon spin"
    }

    let loading = false;
    if (!data) {
        // If data is null/undefined, the request to get namespaces has not yet succeeded.
        // Show a loader.
        loading = true;
    } else if (data === []) {
        // Else if data is an empty array, the request succeeded but no namespaces are listed.
        // In this case, prompt with a 'create namespace' modal.
    }

    if (loading) {
        namespaceSelectorClass += " loading"
        chevronClass += " hidden"
    }

    return (
        <>
            <FlexBox className="col gap">
                <FlexBox onClick={() => {
                    setShowSelector(!showSelector)
                }} style={{...style, maxHeight: "64px"}} className={className}>
                    <FlexBox className={namespaceSelectorClass}>
                        <NamespaceListItem namespace="example" label="ACTIVE NAMESPACE" loading={loading} />
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

    let {namespace, loading, label} = props;
    let className = "namespace-list-item";

    if (!namespace) {
        namespace = "Undefined"
    }

    if (!label) {
        label = "Namespace"
    }

    if (loading === true) {
        className += " loading"
        label = "Loading..."
        namespace = "..."
    }


    return (
        <FlexBox className={className} style={{height: "45px", minHeight: "45px", maxHeight: "45px"}}>
            <FlexBox className="">
                <FlexBox className="namespace-selector-logo">
                    <div className="auto-margin">
                        
                    </div>
                </FlexBox>
                <FlexBox className="col">
                    <div className="auto-margin" style={{marginLeft: "8px"}}>
                        <FlexBox className="namespace-selector-label-header">
                            {label}
                        </FlexBox>
                        <FlexBox className="namespace-selector-label-value">
                            {namespace}
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