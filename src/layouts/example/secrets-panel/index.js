import React from 'react';
import './style.css';
import AddValueButton from '../../../components/add-button';
import ContentPanel, {ContentPanelTitle, ContentPanelTitleIcon, ContentPanelBody } from '../../../components/content-panel';
import { IoCloseCircleSharp, IoLockClosedOutline } from 'react-icons/io5';
import Modal from '../../../components/modal';
import FlexBox from '../../../components/flexbox';
import Alert from '../../../components/alert';
import Button from '../../../components/button';

function SecretsPanel(props){
    return (
        <ContentPanel style={{width: "100%"}}>
            <ContentPanelTitle>
                <ContentPanelTitleIcon>
                    <IoLockClosedOutline />
                </ContentPanelTitleIcon>
                Secrets   
                <Modal title="New secret" 
                    button={(
                         <AddValueButton label="Add" />
                    )} 
                    withCloseButton activeOverlay
                >
                </Modal>
            </ContentPanelTitle>
            <ContentPanelBody className="secrets-panel">
                <FlexBox className="gap col">
                    <FlexBox>
                        <Alert className="info">Once a value if removed, it can never be restored.</Alert>
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
                            withCloseButton 
                            title="Remove secret" 
                            button={(
                                <SecretsDeleteButton/>
                            )}    
                            actionButtonLabel="Delete"
                            actionButtonFunc={() => {
                                // do logic here
                                console.log(":)");
                            }}
                        >
                            <FlexBox className="col gap">
                                <FlexBox >
                                    Are you sure you want to delete 'SECRET_NAME_HERE'?
                                    This action cannot be undone.
                                </FlexBox>
                            </FlexBox>
                        </Modal>
                    </FlexBox>
                </FlexBox>
                <FlexBox className="secret-tuple">
                    <FlexBox className="key">GCP_CREDENTIALS</FlexBox>
                    <FlexBox className="val"><span>******</span></FlexBox>
                    <FlexBox className="actions">
                        <Modal 
                            withCloseButton 
                            title="Remove secret" 
                            button={(
                                <SecretsDeleteButton/>
                            )}    
                        >
                        </Modal>
                    </FlexBox>
                </FlexBox>
                <FlexBox className="secret-tuple">
                    <FlexBox className="key">GCP_BUCKET</FlexBox>
                    <FlexBox className="val"><span>******</span></FlexBox>
                    <FlexBox className="actions">
                        <Modal 
                            withCloseButton 
                            title="Remove secret" 
                            button={(
                                <SecretsDeleteButton/>
                            )}    
                        >
                        </Modal>
                    </FlexBox>
                </FlexBox>
                <FlexBox className="secret-tuple">
                    <FlexBox className="key">DOCKER_TOKEN</FlexBox>
                    <FlexBox className="val"><span>******</span></FlexBox>
                    <FlexBox className="actions">
                        <Modal 
                            withCloseButton 
                            title="Remove secret" 
                            button={(
                                <SecretsDeleteButton/>
                            )}    
                        >
                        </Modal>
                    </FlexBox>
                </FlexBox>
            </FlexBox>
        </>
    );
}

function SecretsDeleteButton(props) {
    return (
        <div className="secrets-delete-btn red-text">
            <IoCloseCircleSharp/>
        </div>
    )
}