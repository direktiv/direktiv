import React from 'react';
import './style.css';
import ContentPanel, {ContentPanelTitle, ContentPanelTitleIcon, ContentPanelBody } from '../../../components/content-panel';
import { IoLockClosedOutline } from 'react-icons/io5';
import FlexBox from '../../../components/flexbox';
import Modal, { ButtonDefinition } from '../../../components/modal';
import AddValueButton from '../../../components/add-button';

function VariablesPanel(props){
    return (
        <ContentPanel style={{width: "100%"}}>
            <ContentPanelTitle>
                <ContentPanelTitleIcon>
                    <IoLockClosedOutline />
                </ContentPanelTitleIcon>
                <FlexBox>
                    Variables   
                </FlexBox>
                <div>
                    <Modal title="New variable" 
                        escapeToCancel
                        button={(
                            <AddValueButton label=" " />
                        )}  
                        actionButtons={[
                            ButtonDefinition("Add", () => {
                                console.log("add namespace");
                            }, "small blue", true, false),
                            ButtonDefinition("Cancel", () => {
                                console.log("close modal");
                            }, "small light", true, false)
                        ]}
                    >
                    </Modal>
                </div>
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
            <table className="variables-table">
                <thead>
                    <tr className="header-row">
                        <th className="left-align" style={{ width: "180px", maxWidth: "180px" }}>
                            Name
                        </th>
                        <th>
                            Value
                        </th>
                        <th className="left-align" style={{ width: "80px", maxWidth: "80px" }}>
                            Size
                        </th>
                        <th className="center-align" style={{ width: "120px", maxWidth: "120px" }}>
                            Action
                        </th>
                    </tr>
                </thead>
                <tbody>
                    <tr>
                        <td>IMG_1923.jpeg</td>
                        <td className="muted-text">Cannot Show Binary Variable</td>
                        <td>168917 B</td>
                        <td></td>
                    </tr>
                    <tr>    
                        <td>
                            <div className="editor-var-name">
                                Var1
                            </div>
                        </td>
                        <td style={{padding: "8px", paddingLeft: "0px"}}>
                            <FlexBox className="editor-placeholder">
                                <div style={{marginLeft: "8px"}}>
                                    placeholder
                                </div>
                            </FlexBox>
                        </td>
                        <td>168917 B</td>
                        <td></td>
                    </tr>
                </tbody>
            </table>
        </FlexBox>
    );
}