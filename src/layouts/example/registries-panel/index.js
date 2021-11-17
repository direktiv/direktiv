import React from 'react';
import ContentPanel, {ContentPanelTitle, ContentPanelTitleIcon, ContentPanelBody } from '../../../components/content-panel';
import Modal from '../../../components/modal';
import { IoLogoDocker } from 'react-icons/io5';
import AddValueButton from '../../../components/add-button';

function RegistriesPanel(props){
    return (
        <ContentPanel style={{width: "100%", minHeight: "180px"}}>
            <ContentPanelTitle>
                <ContentPanelTitleIcon>
                    <IoLogoDocker />
                </ContentPanelTitleIcon>
                Container Registries  
                <Modal title="New secret" 
                    button={(
                         <AddValueButton label="Add" />
                    )} 
                    withCloseButton activeOverlay
                ></Modal> 
            </ContentPanelTitle>
            <ContentPanelBody >
                
            </ContentPanelBody>
        </ContentPanel>
    )
}

export default RegistriesPanel;