import Diagram from "./Diagram";
import Header from "./Header";
import InputOutput from "./InputOutput";
import Logs from "./Logs";
import WorkspaceLayout from "./WorkspaceLayout";
import { useInstanceDetails } from "~/api/instances/query/details";
import { useInstanceId } from "./state/instanceContext";
import { useLogsPreferencesMaximizedPanel } from "~/util/store/logs";

const InstancesDetail = () => {
  const instanceId = useInstanceId();
  const { data } = useInstanceDetails({ instanceId });
  const preferedLayout = useLogsPreferencesMaximizedPanel();

  if (!data) return null;
  return (
    <div className="grid grow grid-rows-[auto_1fr]">
      <Header instanceId={instanceId} />
      <WorkspaceLayout
        layout={preferedLayout}
        logComponent={<Logs />}
        diagramComponent={
          <Diagram workflowPath={data.instance.as} flow={data.flow} />
        }
        inputOutputComponent={<InputOutput />}
      />
    </div>
  );
};

export default InstancesDetail;
