import React from 'react';
import './style.css';
import ContentPanel, {ContentPanelTitle, ContentPanelTitleIcon, ContentPanelBody} from '../../../components/content-panel';
import FlexBox from '../../../components/flexbox';
import { IoLockClosedOutline } from 'react-icons/io5';
import Alert from '../../../components/alert';

function ScarySettings(props) {
    return (<>
        <ContentPanel className="scary-panel">
            <ContentPanelTitle>
                <ContentPanelTitleIcon>
                    <IoLockClosedOutline className="red-text" />
                </ContentPanelTitleIcon>
                <FlexBox className="red-text">
                    Important Settings   
                </FlexBox>
            </ContentPanelTitle>
            <ContentPanelBody className="secrets-panel">
                <FlexBox className="gap col">
                    <FlexBox className="secrets-list"> 

                    </FlexBox>
                    <FlexBox>
                        <Alert className="critical">The following settings are super dangerous! Use at your own risk!</Alert>
                    </FlexBox>
                </FlexBox>
            </ContentPanelBody>
        </ContentPanel>
    </>)
}

export default ScarySettings;