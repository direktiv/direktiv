import React from 'react';
import {ContentPanelHeaderButton, ContentPanelHeaderButtonIcon} from '../content-panel';
import {IoAddOutline} from 'react-icons/io5';
import FlexBox from '../flexbox';

function AddValueButton(props) {

    let {onClick, label} = props;
    if (!label) {
        label = "Add value"
    }

    return (
        <FlexBox>
            <ContentPanelHeaderButton className="add-panel-btn" onClick={onClick}>
                <ContentPanelHeaderButtonIcon>
                    <IoAddOutline/>
                </ContentPanelHeaderButtonIcon>
                {label}
            </ContentPanelHeaderButton>
        </FlexBox>
    )
}

export default AddValueButton;