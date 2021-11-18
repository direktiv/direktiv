import React from 'react';
import './style.css';
import AddValueButton from '../../../components/add-button';
import ContentPanel, {ContentPanelTitle, ContentPanelTitleIcon, ContentPanelBody } from '../../../components/content-panel';
import { IoCloseCircleSharp, IoLockClosedOutline } from 'react-icons/io5';
import Modal, {ButtonDefinition} from '../../../components/modal';
import FlexBox from '../../../components/flexbox';
import Alert from '../../../components/alert';

function SecretsPanel(props){
    return (
        <ContentPanel style={{width: "100%"}}>
            <ContentPanelTitle>
                <ContentPanelTitleIcon>
                    <IoLockClosedOutline />
                </ContentPanelTitleIcon>
                <FlexBox>
                    Secrets   
                </FlexBox>
                <div>
                    <Modal title="New secret" 
                        escapeToCancel
                        button={(
                            <AddValueButton label="Add" />
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
                        <AddSecretPanel />
                    </Modal>
                </div>
            </ContentPanelTitle>
            <ContentPanelBody className="secrets-panel">
                <FlexBox className="gap col">
                    <FlexBox>
                        <Alert>Once a secret is removed, it can never be restored.</Alert>
                    </FlexBox>
                    <FlexBox className="secrets-list">
                        <Secrets />
                    </FlexBox>
                </FlexBox>
            </ContentPanelBody>
        </ContentPanel>
    )
}

export default SecretsPanel;

function Secrets(props) {

    return(
        <>
            <FlexBox className="col gap">
                <FlexBox className="secret-tuple">
                    <FlexBox className="key">randomKey</FlexBox>
                    <FlexBox className="val"><span>******</span></FlexBox>
                    <FlexBox className="actions">
                        <Modal 
                            escapeToCancel
                            style={{
                                flexDirection: "row-reverse",
                                marginRight: "8px"
                            }}
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
                                    }, "small light", true, false)
                                ]
                            }   
                        >
                            <FlexBox className="col gap">
                                <FlexBox >
                                    Are you sure you want to delete 'SECRET_NAME_HERE'?
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

export function SecretsDeleteButton(props) {
    return (
        <div className="secrets-delete-btn red-text">
            <IoCloseCircleSharp/>
        </div>
    )
}

function AddSecretPanel(props) {

    return (
        <FlexBox className="col gap" style={{fontSize: "12px"}}>
            <FlexBox className="gap">
                <FlexBox>
                    <input placeholder="Enter key" />
                </FlexBox>
            </FlexBox>
            <FlexBox className="gap">
                <FlexBox><input placeholder="Enter value" /></FlexBox>
            </FlexBox>
        </FlexBox>
    );
}