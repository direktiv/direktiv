import { IoPlay } from "react-icons/io5";
import ContentPanel, { ContentPanelBody, ContentPanelTitle, ContentPanelTitleIcon } from "../../components/content-panel";
import FlexBox from "../../components/flexbox";

export default function WorkflowBuilder(props) {

    return(
        <FlexBox gap wrap style={{paddingRight:"8px"}}>
            <ContentPanel style={{width:"100%"}}>
                <ContentPanelTitle>
                    <ContentPanelTitleIcon>
                        <IoPlay/>
                    </ContentPanelTitleIcon>
                    <FlexBox>
                        Workflow Builder
                    </FlexBox>
                </ContentPanelTitle>
                <ContentPanelBody className="secrets-panel">
                    <FlexBox col gap>
                        wf builder
                    </FlexBox>
                </ContentPanelBody>
            </ContentPanel>
        </FlexBox>
    )
}