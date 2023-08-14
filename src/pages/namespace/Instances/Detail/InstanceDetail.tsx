import Diagram from "./Main/Diagram";
import Header from "./Header";
import InputOutput from "./Main/InputOutput";
import Logs from "./Main/Logs";
import WorkspaceLayout from "./Main";
import { useInstanceDetails } from "~/api/instances/query/details";
import { useInstanceId } from "./store/instanceContext";
import { useLogsPreferencesMaximizedPanel } from "~/util/store/logs";

const InstancesDetail = () => {
  const instanceId = useInstanceId();
  const { data } = useInstanceDetails({ instanceId });
  const preferedLayout = useLogsPreferencesMaximizedPanel();

  if (!data) return null;
  return (
    <div className="grid grow grid-rows-[auto_1fr]">
      <Header />
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
