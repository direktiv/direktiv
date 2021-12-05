import React, { useState } from 'react';
import Button from '../../../components/button';
import { BsCodeSquare } from 'react-icons/bs';
import ContentPanel, { ContentPanelBody, ContentPanelTitle, ContentPanelTitleIcon } from '../../../components/content-panel';
import FlexBox from '../../../components/flexbox';
import {GenerateRandomKey} from '../../../util';
import {BiChevronLeft} from 'react-icons/bi';

function RevisionTab(props) {

    const {searchParams, setSearchParams, revision} = props
    const [tabBtn, setTabBtn] = useState(0);

    return(
        <FlexBox>
        <FlexBox className="col gap" style={{maxHeight:"100px"}}>
            <FlexBox>
                <Button className="small light" style={{ minWidth: "160px", maxWidth: "160px" }}>
                    <FlexBox className="gap" style={{ alignItems: "center", justifyContent: "center" }}>
                        <BiChevronLeft style={{ fontSize: "16px" }} />
                        <div>Back to All Revisions</div>
                    </FlexBox>
                </Button>
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
                    <TabbedButtons tabBtn={tabBtn} setTabBtn={setTabBtn} />
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

    let {tabBtn, setTabBtn} = props;

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