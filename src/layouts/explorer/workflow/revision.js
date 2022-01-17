import { useWorkflowService } from "direktiv-react-hooks"
import { Config } from "../../../util"
import FlexBox from "../../../components/flexbox"
import ContentPanel, { ContentPanelBody, ContentPanelTitle, ContentPanelTitleIcon } from "../../../components/content-panel"
import { IoPlay } from "react-icons/io5"
import { Service } from "../../namespace-services"
import { useNavigate } from "react-router"

export default function WorkflowRevisions(props) {
    const {namespace, service, version, filepath} = props
    const navigate = useNavigate()
    const {revisions, err} = useWorkflowService(Config.url, namespace, filepath, service, version, navigate, localStorage.getItem("apikey"))

    if(revisions === null) {
        return <></>
    }

    if (err) {
        // TODO report error
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
                    {/* <div>
                        <Modal title={`New '${service}' revision`} 
                            escapeToCancel
                            modalStyle={{
                                maxWidth: "300px"
                            }}
                            onOpen={() => {
                            }}
                            onClose={()=>{
                            }}
                            button={(
                                <AddValueButton  label=" " />
                            )}  
                            keyDownActions={[
                                KeyDownDefinition("Enter", async () => {
                                }, true)
                            ]}
                            actionButtons={[
                                ButtonDefinition("Add", async () => {
                                    try { await createNamespaceServiceRevision(image, parseInt(scale), parseInt(size), cmd, parseInt(trafficPercent))
                                    if (err) return err
                                }, "small blue", true, false),
                                ButtonDefinition("Cancel", () => {
                                }, "small light", true, false)
                            ]}
                        >
                            {config !== null ? 
                            <RevisionCreatePanel 
                                image={image} setImage={setImage}
                                scale={scale} setScale={setScale}
                                size={size} setSize={setSize}
                                cmd={cmd} setCmd={setCmd}
                                traffic={trafficPercent} setTraffic={setTrafficPercent}
                                maxscale={config.maxscale}
                            />:""}
                        </Modal>
                    </div> */}
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