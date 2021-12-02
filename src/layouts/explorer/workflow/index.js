import React, { useState } from 'react';
import './style.css';
import FlexBox from '../../../components/flexbox';
import ContentPanel, { ContentPanelHeaderButton, ContentPanelTitle, ContentPanelTitleIcon } from '../../../components/content-panel';
import {BsCodeSquare} from 'react-icons/bs'

function WorkflowPage(props) {

    const [activeTab, setActiveTab] = useState(0)

    return(
        <>
            <FlexBox id="workflow-page" className="gap col" style={{paddingRight: "8px"}}>
                <TabBar activeTab={activeTab} setActiveTab={setActiveTab} />
                <FlexBox className="col gap">
                    { activeTab === 0 ? 
                        <OverviewTab />
                    :<></>}
                    { activeTab === 1 ?
                        <RevisionSelectorTab />
                    :<></>}
                </FlexBox>
            </FlexBox>
        </>
    )
}

export default WorkflowPage;

function TabBar(props) {

    let {activeTab, setActiveTab} = props;
    let tabLabels = [
        "Overview",
        "Revisions",
        "Working Revisions",
        "Dependency Graph", 
        "Settings"
    ]

    let tabDOMs = [];
    for (let i = 0; i < 5; i++) {

        let className = "tab-bar-item"
        if (i === activeTab) {
            className += " active"
        }

        tabDOMs.push(
            <FlexBox className={className} onClick={() => {
                setActiveTab(i)
            }}>
                {tabLabels[i]}
            </FlexBox>
        )
    }

    return (
        <FlexBox className="tab-bar">
            {tabDOMs}
            <FlexBox className="tab-bar-item gap">
            <label className="switch">
                <input type="checkbox" />
                <span className="slider-broadcast"></span>
            </label>
            <div className="rev-toggle-label hide-on-small">
                Enabled
            </div>
            </FlexBox>
        </FlexBox>
    )
}

function OverviewTab(props) {
    return(
        <>
            <FlexBox className="gap wrap">
                <FlexBox>
                    <ContentPanel style={{ width: "100%", minWidth: "300px"}}>
                        <ContentPanelTitle>
                            <ContentPanelTitleIcon>
                                <BsCodeSquare />
                            </ContentPanelTitleIcon>
                            <div>
                                Instances
                            </div>
                        </ContentPanelTitle>
                    </ContentPanel>
                </FlexBox>
                <FlexBox>
                    <ContentPanel style={{ width: "100%", minWidth: "300px"}}>
                        <ContentPanelTitle>
                            <ContentPanelTitleIcon>
                                <BsCodeSquare />
                            </ContentPanelTitleIcon>
                            <div>
                                Success/Failure Rate
                            </div>
                        </ContentPanelTitle>
                    </ContentPanel>
                </FlexBox>
            </FlexBox>
            <FlexBox style={{maxHeight: "140px"}}>
                <ContentPanel style={{ width: "100%", minWidth: "300px" }}>
                    <ContentPanelTitle>
                        <ContentPanelTitleIcon>
                            <BsCodeSquare />
                        </ContentPanelTitleIcon>
                        <div>
                            Traffic Distribution
                        </div>
                    </ContentPanelTitle>
                </ContentPanel>
            </FlexBox>
            <FlexBox>
                <ContentPanel style={{ width: "100%", minWidth: "300px"}}>
                    <ContentPanelTitle>
                        <ContentPanelTitleIcon>
                            <BsCodeSquare />
                        </ContentPanelTitleIcon>
                        <div>
                            All Revisions
                        </div>
                    </ContentPanelTitle>
                </ContentPanel>
            </FlexBox>
        </>
    )
}

function RevisionSelectorTab(props) {
    return(
        <>
            <FlexBox>
                <ContentPanel style={{ width: "100%", minWidth: "300px"}}>
                    <ContentPanelTitle>
                        <ContentPanelTitleIcon>
                            <BsCodeSquare />
                        </ContentPanelTitleIcon>
                        <div>
                            Revision name 001
                        </div>
                        <FlexBox style={{justifyContent: "end", paddingRight: "8px"}}>
                            <div>
                                <FlexBox className="revision-panel-btn-bar">
                                    <div>Editor</div>
                                    <div>Diagram</div>
                                    <div>Sankey</div>
                                </FlexBox>
                            </div>
                        </FlexBox>
                    </ContentPanelTitle>
                </ContentPanel>
            </FlexBox>
        </>
    )
}