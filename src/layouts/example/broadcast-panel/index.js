import React from 'react';
import './style.css';
import ContentPanel, {ContentPanelTitle, ContentPanelTitleIcon, ContentPanelBody } from '../../../components/content-panel';
import { BsSliders } from 'react-icons/bs';
import FlexBox from '../../../components/flexbox';

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
                <BroadcastOptions />
            </ContentPanelBody>
        </ContentPanel>
    )
}

export default BroadcastConfigurationsPanel;

function BroadcastOptions(props) {

    return(
        <FlexBox>
            <FlexBox className="col gap">
                <FlexBox className="options-row">
                <BroadcastOptionsRow 
                    title="Directory" 
                    options={[{
                        label: "Create",
                        value: false,
                        onClick: () => {}
                    },{
                        label: "Delete",
                        value: true,
                        onClick: () => {}
                    }]}
                />
                <BroadcastOptionsRow></BroadcastOptionsRow>
                </FlexBox>
                <FlexBox className="options-row">
                    <BroadcastOptionsRow 
                        title="Instance"
                        options={[{
                            label: "Success",
                            value: true,
                            onClick: () => {}
                        },{
                            label: "Started",
                            value: false,
                            onClick: () => {}
                        },{
                            label: "Failed",
                            value: false,
                            onClick: () => {}
                        },]}
                    />
                    <BroadcastOptionsRow 
                        title="Instance Variable"
                        options={[{
                            label: "Create",
                            value: false,
                            onClick: () => {}
                        },{
                            label: "Update",
                            value: true,
                            onClick: () => {}
                        },{
                            label: "Delete",
                            value: true,
                            onClick: () => {}
                        },]}
                    />
                </FlexBox>
                <FlexBox className="options-row">
                    <BroadcastOptionsRow 
                        title="Namespace Variable"
                        options={[{
                            label: "Create",
                            value: true,
                            onClick: () => {}
                        },{
                            label: "Update",
                            value: false,
                            onClick: () => {}
                        },{
                            label: "Delete",
                            value: false,
                            onClick: () => {}
                        },]}
                    />
                    <BroadcastOptionsRow/>
                </FlexBox>
                <FlexBox>
                    <BroadcastOptionsRow 
                        title="Workflow"
                        options={[{
                            label: "Create",
                            value: false,
                            onClick: () => {}
                        },{
                            label: "Update",
                            value: true,
                            onClick: () => {}
                        },{
                            label: "Delete",
                            value: false,
                            onClick: () => {}
                        },]}
                    />
                    <BroadcastOptionsRow 
                        title="Workflow Variable"
                        options={[{
                            label: "Create",
                            value: false,
                            onClick: () => {}
                        },{
                            label: "Update",
                            value: false,
                            onClick: () => {}
                        },{
                            label: "Delete",
                            value: true,
                            onClick: () => {}
                        },]}
                    />
                </FlexBox>
            </FlexBox>
        </FlexBox>
    );
}

function BroadcastOptionsRow(props) {

    let {title, options} = props;
    let opts = [];

    for (let i = 0; i < 3; i++) {

        if ((!options) || (i >= options.length)) {
            opts.push(
                <FlexBox className="col gap">
                    <FlexBox></FlexBox>
                    <FlexBox key={"broadcast-opts-"+title+"-"+i}>
                        <label className="switch" style={{visibility: "hidden"}}>
                            <input type="checkbox" />
                            <span className="slider"></span>
                        </label>
                    </FlexBox>
                </FlexBox>
            )
        } else {
            opts.push(
                <FlexBox className="col gap broadcast-option">
                    <FlexBox>{options[i].label}</FlexBox>
                    <FlexBox key={"broadcast-opts-"+title+"-"+i}>
                        <label className="switch">
                            <input checked={options[i].value} type="checkbox" />
                            <span className="slider"></span>
                        </label>
                    </FlexBox>
                </FlexBox>
            )
        }

    }

    return(
        <FlexBox className="col broadcast-options-panel gap">
            <div className="broadcast-options-header">
                {title}
            </div>
            <FlexBox className="broadcast-options-inputs gap"> 
                {opts}
            </FlexBox>
        </FlexBox>
    )
}