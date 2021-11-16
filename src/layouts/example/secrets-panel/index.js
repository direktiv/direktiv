import React from 'react';
import ContentPanel, {ContentPanelTitle, ContentPanelTitleIcon, ContentPanelBody } from '../../../components/content-panel';
import AddValueButton from '../../../components/add-button';
import { IoLockClosedOutline } from 'react-icons/io5';

function SecretsPanel(props){
    return (
        <ContentPanel style={{width: "100%"}}>
            <ContentPanelTitle>
                <ContentPanelTitleIcon>
                    <IoLockClosedOutline />
                </ContentPanelTitleIcon>
                Secrets   
                <AddValueButton onClick={() => { 
                    console.log(":D");
                }} />
            </ContentPanelTitle>
            <ContentPanelBody >
                
            </ContentPanelBody>
        </ContentPanel>
    )
}

export default SecretsPanel;