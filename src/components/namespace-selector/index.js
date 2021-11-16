import React from 'react';
import './style.css';
import FlexBox from '../flexbox';
import {IoChevronDown} from 'react-icons/io5';

function NamespaceSelector(props) {

    let {onClick, style, className} = props;
    if (!className) {
        className = ""
    }

    className += " border"

    return (
        <FlexBox onClick={onClick} style={{...style}} className={className}>
            <FlexBox className="namespace-selector">
                <FlexBox className="namespace-selector-logo">
                    <div className="auto-margin">
                        IMG
                    </div>
                </FlexBox>
                <FlexBox className="col">
                    <div className="auto-margin">
                        <FlexBox className="namespace-selector-label-header">
                            LOGGED IN
                        </FlexBox>
                        <FlexBox className="namespace-selector-label-value">
                            Namespace Inc.
                        </FlexBox>
                    </div>
                </FlexBox>
                <FlexBox className="tall">
                    <div className="auto-margin grey-text">
                        <IoChevronDown style={{ marginTop: "8px" }} />
                    </div>
                </FlexBox>
            </FlexBox>
        </FlexBox>
    );
}

export default NamespaceSelector;