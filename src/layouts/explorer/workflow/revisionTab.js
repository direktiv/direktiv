import React, { useEffect, useState } from 'react';
import { BsCodeSquare } from 'react-icons/bs';
import ContentPanel, { ContentPanelBody, ContentPanelTitle, ContentPanelTitleIcon } from '../../../components/content-panel';
import FlexBox from '../../../components/flexbox';
import {GenerateRandomKey} from '../../../util';

function RevisionTab(props) {

    const {searchParams, setSearchParams, revision} = props
    const [tabBtn, setTabBtn] = useState(searchParams.get('revtab') !== null ? parseInt(searchParams.get('revtab')): 0);

    useEffect(()=>{
        if(searchParams.get('revtab') === null) {
            setSearchParams({
                tab: searchParams.get('tab'),
                revision: revision,
                revtab: 0
            }, {replace: true})
        }
    },[searchParams])

    return(
        <FlexBox>
        <FlexBox className="col gap" style={{maxHeight:"100px"}}>
            <FlexBox>
                Back to All Revisions
            </FlexBox>
            <FlexBox>
            <ContentPanel style={{ width: "100%", minWidth: "300px"}}>
                <ContentPanelTitle>
                    <ContentPanelTitleIcon>
                        <BsCodeSquare />
                    </ContentPanelTitleIcon>
                    <div>
                       {revision}
                    </div>
                    <TabbedButtons revision={revision} setSearchParams={setSearchParams} searchParams={searchParams} tabBtn={tabBtn} setTabBtn={setTabBtn} />
                    {/* <FlexBox style={{maxWidth:"150px"}}>
                        <FlexBox>
                            <Button className="reveal-btn small shadow">
                                <FlexBox className="gap">
                                    <div>
                                       YAML
                                    </div>
                                </FlexBox>
                            </Button>
                        </FlexBox>
                        <FlexBox>
                            <Button className="reveal-btn small shadow">
                                <FlexBox className="gap">
                                    <div>
                                       Diagram
                                    </div>
                                </FlexBox>
                            </Button>
                        </FlexBox>
                        <FlexBox>
                            <Button className="reveal-btn small shadow">
                                <FlexBox className="gap">
                                    <div>
                                       Sankey
                                    </div>
                                </FlexBox>
                            </Button>
                        </FlexBox>
                    </FlexBox> */}
                </ContentPanelTitle>
                <ContentPanelBody>
                    
                </ContentPanelBody>
            </ContentPanel>
            </FlexBox>
        </FlexBox>
    </FlexBox>
    )

}

export default RevisionTab;

function TabbedButtons(props) {

    let {tabBtn, setTabBtn, searchParams, setSearchParams, revision} = props;

    let tabBtns = [];
    let tabBtnLabels = ["YAML", "Diagram", "Sankey"];

    console.log(tabBtn);

    for (let i = 0; i < tabBtnLabels.length; i++) {
        let key = GenerateRandomKey();
        let classes = "tab-btn";
        if (i === tabBtn) {
            classes += " active-tab-btn"
        }

        tabBtns.push(<FlexBox key={key} className={classes}>
            <div onClick={() => {
                setTabBtn(i)
                setSearchParams({
                    tab: searchParams.get('tab'),
                    revision: revision,
                    revtab: i
                }, {replace: true})
            }}>
                {tabBtnLabels[i]}
            </div>
        </FlexBox>)
    }

    return(
            <FlexBox className="tabbed-btns-container">
                <FlexBox className="tabbed-btns" >
                    {tabBtns}
                </FlexBox>
            </FlexBox>
    )
}