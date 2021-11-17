import React from 'react';
import ContentPanel, {ContentPanelTitle, ContentPanelTitleIcon, ContentPanelBody } from '../../../components/content-panel';
import Modal, { ButtonDefinition } from '../../../components/modal';
import { IoLogoDocker } from 'react-icons/io5';
import AddValueButton from '../../../components/add-button';
import FlexBox from '../../../components/flexbox';
import {SecretsDeleteButton} from '../secrets-panel';
import Alert from '../../../components/alert';

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
                        withCloseButton 
                        actionButtons={[
                            ButtonDefinition("Cancel", () => {
                                console.log("close modal");
                            }, "small red", true, false),
                            ButtonDefinition("Add", () => {
                                console.log("add registry func");
                            }, "small blue", true, false)
                        ]}
                    >
                        <AddRegistryPanel/>    
                    </Modal> 
                </div>
            </ContentPanelTitle>
            <ContentPanelBody>
                <FlexBox className="gap col">
                    <FlexBox>
                        <Alert className="info">Once a registry is removed, it can never be restored.</Alert>
                    </FlexBox>
                    <FlexBox>
                        <Registries/>
                    </FlexBox>
                </FlexBox>
            </ContentPanelBody>
        </ContentPanel>
    )
}

export default RegistriesPanel;

function AddRegistryPanel(props) {

    return (
        <FlexBox className="col gap" style={{fontSize: "12px"}}>
            <FlexBox className="gap">
                <FlexBox style={{width: "40px"}}>URL:</FlexBox>
                <FlexBox>
                    <input placeholder="Enter URL" />
                </FlexBox>
            </FlexBox>
            <FlexBox className="gap">
                <FlexBox style={{width: "40px"}}>User:</FlexBox>
                <FlexBox><input placeholder="Enter username" /></FlexBox>
            </FlexBox>
            <FlexBox className="gap">
                <FlexBox style={{width: "40px"}}>Token:</FlexBox>
                <FlexBox><input type="password" placeholder="Enter token" /></FlexBox>
            </FlexBox>
        </FlexBox>
    );
}

function Registries(props) {

    return(
        <>
            <FlexBox className="col gap">
                <FlexBox className="secret-tuple">
                    <FlexBox className="key">randomKey</FlexBox>
                    <FlexBox className="val"><span>******</span></FlexBox>
                    <FlexBox className="val"><span>******</span></FlexBox>
                    <FlexBox className="actions">
                        <Modal 
                            style={{
                                flexDirection: "row-reverse",
                                marginRight: "8px"
                            }}
                            withCloseButton 
                            title="Remove secret" 
                            button={(
                                <SecretsDeleteButton/>
                            )} 
                            actionButtons={
                                [
                                    // label, onClick, classList, closesModal, async
                                    ButtonDefinition("Delete", () => {
                                        console.log("DELETE FUNC");
                                    }, "small red", true, false),
                                    ButtonDefinition("Cancel", () => {
                                        console.log("DONT DELETE");
                                    }, "small blue", true, false)
                                ]
                            }   
                        >
                            <FlexBox className="col gap">
                                <FlexBox>
                                    Are you sure you want to remove 'REGISTRY_NAME_HERE'?
                                    <br/>
                                    This action cannot be undone.
                                </FlexBox>
                            </FlexBox>
                        </Modal>
                    </FlexBox>
                </FlexBox>
            </FlexBox>
        </>
    );
}