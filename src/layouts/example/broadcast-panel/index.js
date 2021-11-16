import React from 'react';
import ContentPanel, {ContentPanelTitle, ContentPanelTitleIcon, ContentPanelBody } from '../../../components/content-panel';
import { BsSliders } from 'react-icons/bs';

function BroadcastConfigurationsPanel(props){
    return (
        <ContentPanel style={{width: "100%"}}>
            <ContentPanelTitle>
                <ContentPanelTitleIcon>
                    <BsSliders />
                </ContentPanelTitleIcon>
                Broadcast Configurations   
            </ContentPanelTitle>
            <ContentPanelBody >
                
            </ContentPanelBody>
        </ContentPanel>
    )
}

export default BroadcastConfigurationsPanel;