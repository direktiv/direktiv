import React from 'react';
import './style.css';
import { IoAdd, IoFolderOpen, IoSearch } from 'react-icons/io5';
import ContentPanel, { ContentPanelBody, ContentPanelHeaderButton, ContentPanelHeaderButtonIcon, ContentPanelTitle, ContentPanelTitleIcon } from '../../components/content-panel';
import FlexBox from '../../components/flexbox';
import { VscTriangleDown } from 'react-icons/vsc';
import { GenerateRandomKey } from '../../util';
import { FiEdit, FiFolder } from 'react-icons/fi';
import { FcWorkflow } from 'react-icons/fc';
import { HiOutlineTrash } from 'react-icons/hi';
import { BsCodeSlash } from 'react-icons/bs';
import Button from '../../components/button';
import Pagination from '../../components/pagination';

function Explorer(props) {
    return(
        <>
            <SearchBar />
            <FlexBox className="col gap" style={{ paddingRight: "8px" }}>
                <ContentPanel>
                    <ContentPanelTitle>
                        <ContentPanelTitleIcon>
                            <IoFolderOpen/>
                        </ContentPanelTitleIcon>
                        <FlexBox>
                            Explorer
                        </FlexBox>
                        <div className="explorer-sort-by">
                            <div className="esb-label inline" style={{marginRight: "8px"}}>
                                Sort by:
                            </div>
                            <div className="esb-field inline">
                                <FlexBox className="gap">
                                    <div className="inline">
                                        Name
                                    </div>
                                    <VscTriangleDown className="auto-margin"/>
                                </FlexBox>
                            </div>
                        </div>
                        <ContentPanelHeaderButton style={{ maxWidth: "170px", width: "170px", minWidth: "170px" }}>
                            <ContentPanelHeaderButtonIcon>
                                <IoAdd/>
                            </ContentPanelHeaderButtonIcon>
                            New Folder/Workflow
                        </ContentPanelHeaderButton>
                    </ContentPanelTitle>
                    <ContentPanelBody>
                        <ExplorerList />
                    </ContentPanelBody>
                </ContentPanel>
                <FlexBox style={{maxHeight: "32px"}}>
                    <FlexBox>
                        <Button className="small light" style={{ display: "flex" }}>
                            <ContentPanelHeaderButtonIcon>
                                <BsCodeSlash style={{ maxHeight: "12px", marginRight: "4px" }} />
                            </ContentPanelHeaderButtonIcon>
                            Open API Commands
                        </Button>
                    </FlexBox>
                    <Pagination max={10} currentIndex={1} />
                </FlexBox>
            </FlexBox>
        </>
    )
}

export default Explorer;

function SearchBar(props) {
    return(
        <div className="explorer-searchbar">
            <FlexBox className="">
                <IoSearch className="auto-margin" />
                <input placeholder={"Search items"}></input>
            </FlexBox>
        </div>
    );
}

function ExplorerList(props) {

    let tmp = [{
        "type": "dir",
        "name": "important"
    },{
        "type": "wf",
        "name": "example"
    }]

    return(
        <FlexBox className="explorer-list col">
            {tmp.map((obj) => {
                console.log(obj);
                if (obj.type === "dir") {
                    return (<DirListItem key={GenerateRandomKey("explorer-item-")} name={obj.name} />)
                } else if (obj.type === "wf") {
                    return (<WorkflowListItem key={GenerateRandomKey("explorer-item-")} name={obj.name} />)
                }
            })}
        </FlexBox>
    )
}

function DirListItem(props) {

    let {name} = props;

    return(
        <div className="explorer-item">
            <FlexBox className="explorer-item-container">
                <FlexBox className="explorer-item-icon">
                    <FiFolder className="auto-margin" />
                </FlexBox>
                <FlexBox className="explorer-item-name">
                    {name}
                </FlexBox>
                <FlexBox className="explorer-item-actions gap">
                    <FlexBox>
                        <FiEdit className="auto-margin" />
                    </FlexBox>
                    <FlexBox>
                        <HiOutlineTrash className="auto-margin red-text" />
                    </FlexBox>
                </FlexBox>
            </FlexBox>
        </div>
    )
}

function WorkflowListItem(props) {

    let {name} = props;

    return(
        <div className="explorer-item">
            <FlexBox className="explorer-item-container">
                <FlexBox className="explorer-item-icon">
                    <FcWorkflow className="auto-margin" />
                </FlexBox>
                <FlexBox className="explorer-item-name">
                    {name}
                </FlexBox>
                <FlexBox className="explorer-item-actions gap">
                    <FlexBox>
                        <FiEdit className="auto-margin" />
                    </FlexBox>
                    <FlexBox>
                        <HiOutlineTrash className="auto-margin red-text" />
                    </FlexBox>
                </FlexBox>
            </FlexBox>
        </div>
    )
}
