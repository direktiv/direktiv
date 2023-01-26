import { useWorkflowService } from "../../../hooks";
import { Config } from "../../../util";
import FlexBox from "../../../components/flexbox";
import ContentPanel, {
  ContentPanelBody,
  ContentPanelTitle,
  ContentPanelTitleIcon,
} from "../../../components/content-panel";
import { VscLayers } from "react-icons/vsc";

import { Service } from "../../namespace-services";
import { useNavigate } from "react-router";
import { useApiKey } from "../../../util/apiKeyProvider";

export default function WorkflowRevisions(props) {
  const { namespace, service, version, filepath } = props;
  const navigate = useNavigate();
  const [apiKey] = useApiKey();
  const { revisions, err } = useWorkflowService(
    Config.url,
    namespace,
    filepath,
    service,
    version,
    navigate,
    apiKey
  );

  if (revisions === null) {
    return <></>;
  }

  if (err) {
    // TODO report error
  }

  return (
    <FlexBox gap wrap style={{ paddingRight: "8px" }}>
      <FlexBox style={{ flex: 6 }}>
        <ContentPanel style={{ width: "100%" }}>
          <ContentPanelTitle>
            <ContentPanelTitleIcon>
              <VscLayers />
            </ContentPanelTitleIcon>
            <FlexBox>Service '{service}' Revisions</FlexBox>
          </ContentPanelTitle>
          <ContentPanelBody>
            <FlexBox col gap>
              <FlexBox col gap>
                {revisions.map((obj, i) => {
                  const dontDelete = true;
                  return (
                    <Service
                      key={i}
                      dontDelete={dontDelete}
                      revision={obj.rev}
                      url={`/n/${namespace}/explorer/${filepath.substring(
                        1
                      )}?revision=${
                        obj.rev
                      }&function=${service}&version=${version}`}
                      conditions={obj.conditions}
                      name={obj.name}
                      status={obj.status}
                    />
                  );
                })}
              </FlexBox>
            </FlexBox>
          </ContentPanelBody>
        </ContentPanel>
      </FlexBox>
      {/* <UpdateTraffic setNamespaceServiceRevisionTraffic={setNamespaceServiceRevisionTraffic} service={service} revisions={revisions} traffic={traffic}/> */}
    </FlexBox>
  );
}
