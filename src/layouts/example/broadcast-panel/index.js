import React from 'react';
import './style.css';
import ContentPanel, {ContentPanelTitle, ContentPanelTitleIcon, ContentPanelBody } from '../../../components/content-panel';
import { BsSliders } from 'react-icons/bs';

function BroadcastConfigurationsPanel(props){
    return (
        <ContentPanel className="broadcast-panel">
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