import React from 'react';
import './style.css';
import { IoAdd, IoChevronDown, IoChevronDownSharp, IoFolderOpen, IoSearch } from 'react-icons/io5';
import ContentPanel, { ContentPanelBody, ContentPanelHeaderButton, ContentPanelHeaderButtonIcon, ContentPanelTitle, ContentPanelTitleIcon } from '../../components/content-panel';
import FlexBox from '../../components/flexbox';
import { BsChevronBarDown } from 'react-icons/bs';
import { FaChevronDown } from 'react-icons/fa';
import { VscTriangleDown } from 'react-icons/vsc';

function Explorer(props) {
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
                        <ContentPanelHeaderButton style={{ maxWidth: "150px", width: "150px", minWidth: "150px" }}>
                            <ContentPanelHeaderButtonIcon>
                                <IoAdd/>
                            </ContentPanelHeaderButtonIcon>
                            Create workflow
                        </ContentPanelHeaderButton>
                    </ContentPanelTitle>
                    <ContentPanelBody>
                        Hello world (Explorer)
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