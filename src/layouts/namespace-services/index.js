import { useNamespaceServices } from "direktiv-react-hooks";
import { IoPlay } from "react-icons/io5";
import ContentPanel, { ContentPanelBody, ContentPanelTitle, ContentPanelTitleIcon } from "../../components/content-panel";
import FlexBox from "../../components/flexbox";
import { Config } from "../../util";


export default function ServicesPanel(props) {
    const {namespace} = props

    if(namespace === null) {
        return ""
    }

    return(
        <FlexBox className="gap wrap" style={{paddingRight:"8px"}}>
            <ContentPanel style={{width:"100%"}}>
                <ContentPanelTitle>
                    <ContentPanelTitleIcon>
                        <IoPlay/>
                    </ContentPanelTitleIcon>
                    <FlexBox>
                        Namespace Services
                    </FlexBox>
                </ContentPanelTitle>
                <ContentPanelBody className="secrets-panel">
                    <FlexBox className="gap col">
                        <NamespaceServices namespace={namespace}/>
                    </FlexBox>
                </ContentPanelBody>
            </ContentPanel>
        </FlexBox>
    )
}

function NamespaceServices(props) {
    const {namespace} = props
    const {data, err} = useNamespaceServices(Config.url, true, namespace)

    return(
        <FlexBox>
            <FlexBox className="col">
                <FlexBox style={{maxHeight:"40px", border:"1px solid #f4f4f4"}}>
                    header
                </FlexBox>
                <FlexBox style={{border:"1px solid #f4f4f4", borderTop:"none"}}>
                    details
                </FlexBox>
            </FlexBox>
        </FlexBox>
    )

}