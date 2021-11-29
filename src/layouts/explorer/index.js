import React from 'react';
import './style.css';
import { IoAdd, IoFolderOpen, IoSearch } from 'react-icons/io5';
import ContentPanel, { ContentPanelBody, ContentPanelHeaderButton, ContentPanelHeaderButtonIcon, ContentPanelTitle, ContentPanelTitleIcon } from '../../components/content-panel';
import FlexBox from '../../components/flexbox';
import { VscTriangleDown } from 'react-icons/vsc';
import { Config, GenerateRandomKey } from '../../util';
import { FiEdit, FiFolder } from 'react-icons/fi';
import { FcWorkflow } from 'react-icons/fc';
import { HiOutlineTrash } from 'react-icons/hi';
import { useNodes } from 'direktiv-react-hooks';
import { useParams } from 'react-router';

function Explorer(props) {
    const {path} = useParams()
    const {namespace}  = props

    let filepath = "/"
    if(!namespace){
        return ""
    }

    if(path !== undefined){
        filepath = path
    }

    return(
        <>
            <SearchBar />
            <FlexBox className="col" style={{ paddingRight: "8px" }}>
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
                        <ExplorerList namespace={namespace} path={filepath}/>
                    </ContentPanelBody>
                </ContentPanel>
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
    const {namespace, path} = props
    let tmp = [{
        "type": "dir",
        "name": "important"
    },{
        "type": "wf",
        "name": "example"
    }]


    const {data, err} = useNodes(Config.url, true, namespace, path)
    console.log(data, err)

    return(
        <FlexBox className="col">
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
                <FlexBox className="explorer-item-actions">
                    <FlexBox>
                        <FiEdit />
                    </FlexBox>
                    <FlexBox>
                        <HiOutlineTrash />
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
                <FlexBox className="explorer-item-actions">
                    <FlexBox>
                        <FiEdit />
                    </FlexBox>
                    <FlexBox>
                        <HiOutlineTrash />
                    </FlexBox>
                </FlexBox>
            </FlexBox>
        </div>
    )
}
