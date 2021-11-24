import { IoPlay } from "react-icons/io5";
import { RiDeleteBin2Line } from "react-icons/ri";
import ContentPanel, { ContentPanelBody, ContentPanelTitle, ContentPanelTitleIcon } from "../../components/content-panel";
import FlexBox from "../../components/flexbox";


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
    // const {namespace} = props
    // const {data, err} = useNamespaceServices(Config.url, true, namespace)

    return(
        <FlexBox>
            <FlexBox className="col">
                <FlexBox style={{maxHeight:"40px", border:"1px solid #f4f4f4"}}>
                    <FlexBox className="gap" style={{alignItems:"center", paddingLeft:"8px"}}>
                        <div>
                            s
                        </div>
                        <div>
                            name
                        </div>
                        <div>
                            direktiv
                        </div>
                        <div>
                            x
                        </div>
                    </FlexBox>
                    <FlexBox>
                        <ServicesDeleteButton/>
                    </FlexBox>
                </FlexBox>
                <FlexBox style={{border:"1px solid #f4f4f4", borderTop:"none"}}>
                    details
                </FlexBox>
            </FlexBox>
        </FlexBox>
    )

}

function ServicesDeleteButton() {
    return (
        <FlexBox className="col  auto-margin red-text" style={{display: "flex", alignItems:"flex-end", width:"100%", height: "100%"}}>
            <div className="secrets-delete-btn" style={{height: "100%", display: "flex", alignItems: "center", paddingRight: "8px" }}>
                <RiDeleteBin2Line className="auto-margin" />
            </div>
        </FlexBox>
    )
}