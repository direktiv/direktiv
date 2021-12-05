import { useWorkflowService } from "direktiv-react-hooks"
import { Config } from "../../../util"
import FlexBox from "../../../components/flexbox"
import ContentPanel, { ContentPanelBody, ContentPanelTitle, ContentPanelTitleIcon, ContentPanelFooter } from "../../../components/content-panel"
import { IoPlay } from "react-icons/io5"
import { Service } from "../../namespace-services"

export default function WorkflowRevisions(props) {
    const {namespace, service, version, filepath} = props
    const {revisions, err} = useWorkflowService(Config.url, namespace, filepath, service, version)

    if(revisions === null) {
        return <></>
    }

    return (
        <FlexBox className="gap wrap" style={{paddingRight: "8px"}}>
        <FlexBox style={{flex: 6}}>
            <ContentPanel style={{width: "100%"}}>
                <ContentPanelTitle>
                    <ContentPanelTitleIcon>
                        <IoPlay/>
                    </ContentPanelTitleIcon>
                    <FlexBox>
                        Service '{service}' Revisions
                    </FlexBox>
                </ContentPanelTitle>
                <ContentPanelBody>

                    <FlexBox className="gap col">
                        <FlexBox className="gap col">
                            {revisions.map((obj) => {
                                let dontDelete = true

                                return (
                                    <Service 
                                        dontDelete={dontDelete}
                                        revision={obj.rev}
                                        url={`/n/${namespace}/explorer/${filepath.substring(1)}?revision=${obj.rev}&function=${service}&version=${version}`}
                                        conditions={obj.conditions}
                                        name={obj.name}
                                        status={obj.status}
                                    />
                                )
                            })}
                        </FlexBox>
                    </FlexBox>

                </ContentPanelBody>
            </ContentPanel>
        </FlexBox>
        {/* <UpdateTraffic setNamespaceServiceRevisionTraffic={setNamespaceServiceRevisionTraffic} service={service} revisions={revisions} traffic={traffic}/> */}
    </FlexBox>
    )
}