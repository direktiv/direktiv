import React from 'react';
import ContentPanel, {ContentPanelTitle, ContentPanelTitleIcon, ContentPanelBody } from '../../../components/content-panel';
import AddValueButton from '../../../components/add-button';
import { IoLogoDocker } from 'react-icons/io5';

function RegistriesPanel(props){
    return (
        <ContentPanel style={{width: "100%"}}>
            <ContentPanelTitle>
                <ContentPanelTitleIcon>
                    <IoLogoDocker />
                </ContentPanelTitleIcon>
                Container Registries   
                <AddValueButton label="Add registry" onClick={() => { 
                    console.log(":D");
                }} />
            </ContentPanelTitle>
            <ContentPanelBody >
                
            </ContentPanelBody>
        </ContentPanel>
    )
}

export default RegistriesPanel;