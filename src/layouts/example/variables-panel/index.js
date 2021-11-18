import React from 'react';
import ContentPanel, {ContentPanelTitle, ContentPanelTitleIcon, ContentPanelBody } from '../../../components/content-panel';
import { IoLockClosedOutline } from 'react-icons/io5';
import FlexBox from '../../../components/flexbox';

function VariablesPanel(props){
    return (
        <ContentPanel style={{width: "100%"}}>
            <ContentPanelTitle>
                <ContentPanelTitleIcon>
                    <IoLockClosedOutline />
                </ContentPanelTitleIcon>
                Variables   
            </ContentPanelTitle>
            <ContentPanelBody >
                <Variables />
            </ContentPanelBody>
        </ContentPanel>
    )
}

export default VariablesPanel;

function Variables(props) {

    return(
        <FlexBox>
            <table>
                <tr>
                    <th>
                        Name
                    </th>
                    <th>
                        Value
                    </th>
                    <th>
                        Size
                    </th>
                    <th>
                        Action
                    </th>
                </tr>
            </table>
        </FlexBox>
    );
}