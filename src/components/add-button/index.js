import React from 'react';
import {ContentPanelHeaderButton, ContentPanelHeaderButtonIcon} from '../content-panel';
import {IoAddOutline} from 'react-icons/io5';

function AddValueButton(props) {

    let {onClick, label} = props;
    if (!label) {
        label = "Add value"
    }

    return (
        <ContentPanelHeaderButton onClick={onClick}>
            <ContentPanelHeaderButtonIcon>
                <IoAddOutline/>
            </ContentPanelHeaderButtonIcon>
            {label}
        </ContentPanelHeaderButton>
    )
}

export default AddValueButton;