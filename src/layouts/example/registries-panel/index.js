import React from 'react';
import ContentPanel, {ContentPanelTitle, ContentPanelTitleIcon, ContentPanelBody } from '../../../components/content-panel';
import Modal from '../../../components/modal';
import { IoLogoDocker } from 'react-icons/io5';
import AddValueButton from '../../../components/add-button';
import FlexBox from '../../../components/flexbox';

function RegistriesPanel(props){
    return (
        <ContentPanel style={{width: "100%", minHeight: "180px"}}>
            <ContentPanelTitle>
                <ContentPanelTitleIcon>
                    <IoLogoDocker />
                </ContentPanelTitleIcon>
                <FlexBox>
                    Container Registries  
                </FlexBox>
                <div>
                    <Modal title="New registry" 
                        button={(
                            <AddValueButton label="Add" />
                        )} 
                        withCloseButton activeOverlay
                    ></Modal> 
                </div>
            </ContentPanelTitle>
            <ContentPanelBody >
                
            </ContentPanelBody>
        </ContentPanel>
    )
}

export default RegistriesPanel;