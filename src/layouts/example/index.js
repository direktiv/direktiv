import React from 'react';
import Button from '../../components/button';
import ContentPanel, {ContentPanelTitle, ContentPanelBody, ContentPanelTitleIcon} from '../../components/content-panel';
import FlexBox from '../../components/flexbox';
import { IoLockClosedOutline } from 'react-icons/io5';

function ExamplePage(props) {
    return(
        <FlexBox className="row gap" style={{ paddingRight: "8px" }}>
            <FlexBox className="col">
                <ContentPanel>
                    <ContentPanelTitle>
                        <ContentPanelTitleIcon>
                            <IoLockClosedOutline />
                        </ContentPanelTitleIcon>
                        Hello world!   
                    </ContentPanelTitle>
                    <ContentPanelBody >
                        <FlexBox>
                            <Button className="auto-margin">Click me!</Button>
                        </FlexBox>
                    </ContentPanelBody>
                </ContentPanel>
            </FlexBox>
            <FlexBox className="col">
                <ContentPanel>
                    <ContentPanelTitle>
                        <ContentPanelTitleIcon>
                            <IoLockClosedOutline />
                        </ContentPanelTitleIcon>
                        This    
                    </ContentPanelTitle>
                    <ContentPanelBody>
                        Is me
                    </ContentPanelBody>
                </ContentPanel>
            </FlexBox>
        </FlexBox>
    )
}

export default ExamplePage;