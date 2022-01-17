import React from 'react';
import './style.css';
import ContentPanel, {ContentPanelTitle, ContentPanelTitleIcon, ContentPanelBody } from '../../../components/content-panel';
import FlexBox from '../../../components/flexbox';
import {Config, GenerateRandomKey} from '../../../util';
import { useBroadcastConfiguration } from 'direktiv-react-hooks';
import HelpIcon from '../../../components/help';
import { VscSettings } from 'react-icons/vsc';

function BroadcastConfigurationsPanel(props){
    const {namespace} = props
    const {data, setBroadcastConfiguration, getBroadcastConfiguration} = useBroadcastConfiguration(Config.url, namespace, localStorage.getItem("apikey"))

    return (
        <ContentPanel className="broadcast-panel">
            <ContentPanelTitle>
                <ContentPanelTitleIcon>
                    <VscSettings />
                </ContentPanelTitleIcon>
                <FlexBox style={{display:"flex", alignItems:"center"}} className="gap">
                    <div>
                        Broadcast Configurations   
                    </div>
                    <HelpIcon msg={"Toggle which Direktiv system events will cause a Cloud Event to be sent to the current namespace."} />
                </FlexBox>
            </ContentPanelTitle>
            <ContentPanelBody >
                {data !== null ?
                    <BroadcastOptions getBroadcastConfiguration={getBroadcastConfiguration} setBroadcastConfiguration={setBroadcastConfiguration} config={data} />
                    :
                    ""
                }
            </ContentPanelBody>
        </ContentPanel>
    )
}

export default BroadcastConfigurationsPanel;

function BroadcastOptions(props) {
    const {config, setBroadcastConfiguration, getBroadcastConfiguration} = props
    return(
        <FlexBox>
            <FlexBox className="col gap">
                <FlexBox className="options-row">
                <BroadcastOptionsRow 
                    title="Directory" 
                    options={[{
                        label: "Create",
                        value: config.broadcast["directory.create"],
                        onClick: async () => {
                            let cc = config
                            cc.broadcast["directory.create"] = !config.broadcast["directory.create"]
                            await setBroadcastConfiguration(JSON.stringify(cc))
                            await getBroadcastConfiguration()
                        }
                    },{
                        label: "Delete",
                        value: config.broadcast["directory.delete"],
                        onClick: async () => {
                            let cc = config
                            cc.broadcast["directory.delete"] = !config.broadcast["directory.delete"]
                            await setBroadcastConfiguration(JSON.stringify(cc))
                            await getBroadcastConfiguration()
                        }
                    }]}
                />
                <BroadcastOptionsRow></BroadcastOptionsRow>
                </FlexBox>
                <FlexBox className="options-row">
                    <BroadcastOptionsRow 
                        title="Instance"
                        options={[{
                            label: "Success",
                            value: config.broadcast["instance.success"],
                            onClick: async () => {
                                let cc = config
                                cc.broadcast["instance.success"] = !config.broadcast["instance.success"]
                                await setBroadcastConfiguration(JSON.stringify(cc))
                                await getBroadcastConfiguration()
                            }
                        },{
                            label: "Started",
                            value: config.broadcast["instance.started"],
                            onClick: async () => {
                                let cc = config
                                cc.broadcast["instance.started"] = !config.broadcast["instance.started"]
                                await setBroadcastConfiguration(JSON.stringify(cc))
                                await getBroadcastConfiguration()
                            }
                        },{
                            label: "Failed",
                            value: config.broadcast["instance.failed"],
                            onClick: async () => {
                                let cc = config
                                cc.broadcast["instance.failed"] = !config.broadcast["instance.failed"]
                                await setBroadcastConfiguration(JSON.stringify(cc))
                                await getBroadcastConfiguration()
                            }
                        },]}
                    />
                    <BroadcastOptionsRow 
                        title="Instance Variable"
                        options={[{
                            label: "Create",
                            value: config.broadcast["instance.variable.create"],
                            onClick: async () => {
                                let cc = config
                                cc.broadcast["instance.variable.create"] = !config.broadcast["instance.variable.create"]
                                await setBroadcastConfiguration(JSON.stringify(cc))
                                await getBroadcastConfiguration()
                            }
                        },{
                            label: "Update",
                            value: config.broadcast["instance.variable.update"],
                            onClick: async () => {
                                let cc = config
                                cc.broadcast["instance.variable.update"] = !config.broadcast["instance.variable.update"]
                                await setBroadcastConfiguration(JSON.stringify(cc))
                                await getBroadcastConfiguration()
                            }
                        },{
                            label: "Delete",
                            value: config.broadcast["instance.variable.delete"],
                            onClick: async () => {
                                let cc = config
                                cc.broadcast["instance.variable.delete"] = !config.broadcast["instance.variable.delete"]
                                await setBroadcastConfiguration(JSON.stringify(cc))
                                await getBroadcastConfiguration()
                            }
                        },]}
                    />
                </FlexBox>
                <FlexBox className="options-row">
                    <BroadcastOptionsRow 
                        title="Namespace Variable"
                        options={[{
                            label: "Create",
                            value: config.broadcast["namespace.variable.create"],
                            onClick: async () => {
                                let cc = config
                                cc.broadcast["namespace.variable.create"] = !config.broadcast["namespace.variable.create"]
                                await setBroadcastConfiguration(JSON.stringify(cc))
                                await getBroadcastConfiguration()
                            }
                        },{
                            label: "Update",
                            value: config.broadcast["namespace.variable.update"],
                            onClick: async () => {
                                let cc = config
                                cc.broadcast["namespace.variable.update"] = !config.broadcast["namespace.variable.update"]
                                await setBroadcastConfiguration(JSON.stringify(cc))
                                await getBroadcastConfiguration()
                            }
                        },{
                            label: "Delete",
                            value: config.broadcast["namespace.variable.delete"],
                            onClick: async () => {
                                let cc = config
                                cc.broadcast["namespace.variable.delete"] = !config.broadcast["namespace.variable.delete"]
                                await setBroadcastConfiguration(JSON.stringify(cc))
                                await getBroadcastConfiguration()
                            }
                        },]}
                    />
                    <BroadcastOptionsRow/>
                </FlexBox>
                <FlexBox>
                    <BroadcastOptionsRow 
                        title="Workflow"
                        options={[{
                            label: "Create",
                            value: config.broadcast["workflow.create"],
                            onClick: async () => {
                                let cc = config
                                cc.broadcast["workflow.create"] = !config.broadcast["workflow.create"]
                                await setBroadcastConfiguration(JSON.stringify(cc))
                                await getBroadcastConfiguration()
                            }
                        },{
                            label: "Update",
                            value: config.broadcast["workflow.update"],
                            onClick: async () => {
                                let cc = config
                                cc.broadcast["workflow.update"] = !config.broadcast["workflow.update"]
                                await setBroadcastConfiguration(JSON.stringify(cc))
                                await getBroadcastConfiguration()
                            }
                        },{
                            label: "Delete",
                            value: config.broadcast["workflow.delete"],
                            onClick: async () => {
                                let cc = config
                                cc.broadcast["workflow.delete"] = !config.broadcast["workflow.delete"]
                                await setBroadcastConfiguration(JSON.stringify(cc))
                                await getBroadcastConfiguration()
                            }
                        },]}
                    />
                    <BroadcastOptionsRow 
                        title="Workflow Variable"
                        options={[{
                            label: "Create",
                            value: config.broadcast["workflow.variable.create"],
                            onClick: async () => {
                                let cc = config
                                cc.broadcast["workflow.variable.create"] = !config.broadcast["workflow.variable.create"]
                                await setBroadcastConfiguration(JSON.stringify(cc))
                                await getBroadcastConfiguration()
                            }
                        },{
                            label: "Update",
                            value: config.broadcast["workflow.variable.update"],
                            onClick: async () => {
                                let cc = config
                                cc.broadcast["workflow.variable.update"] = !config.broadcast["workflow.variable.update"]
                                await setBroadcastConfiguration(JSON.stringify(cc))
                                await getBroadcastConfiguration()
                            }
                        },{
                            label: "Delete",
                            value: config.broadcast["workflow.variable.delete"],
                            onClick: async () => {
                                let cc = config
                                cc.broadcast["workflow.variable.delete"] = !config.broadcast["workflow.variable.delete"]
                                await setBroadcastConfiguration(JSON.stringify(cc))
                                await getBroadcastConfiguration()
                            }
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

        let key = GenerateRandomKey("broadcast-opt-");

        if ((!options) || (i >= options.length)) {
            opts.push(
                <FlexBox id={key} key={key} className="col gap">
                    <FlexBox></FlexBox>
                    <FlexBox key={"broadcast-opts-"+title+"-"+i}>
                        <label className="switch" style={{visibility: "hidden"}}>
                            <input type="checkbox" />
                            <span className="slider-broadcast"></span>
                        </label>
                    </FlexBox>
                </FlexBox>
            )
        } else {
            opts.push(
                <FlexBox  id={key} key={key} className="col gap broadcast-option">
                    <FlexBox>{options[i].label}</FlexBox>
                    <FlexBox key={"broadcast-opts-"+title+"-"+i}>
                        <label className="switch">
                        <input onClick={()=>{ options[i].onClick()}} defaultChecked={options[i].value} type="checkbox" />
                            <span className="slider-broadcast"></span>
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