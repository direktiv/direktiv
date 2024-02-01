import { Card } from "~/design/Card";
import Diagram from "./Main/Diagram";
import Header from "./Header";
import InputOutput from "./Main/InputOutput";
import Logs from "./Main/Logs";
import { NoPermissions } from "~/design/Table";
import WorkspaceLayout from "./Main";
import { useInstanceDetails } from "~/api/instances/query/details";
import { useInstanceId } from "./store/instanceContext";
import { useLogsPreferencesMaximizedPanel } from "~/util/store/logs";

const InstancesDetail = () => {
  const instanceId = useInstanceId();
  const { data, isFetched, isAllowed, noPermissionMessage } =
    useInstanceDetails({ instanceId });
  const preferedLayout = useLogsPreferencesMaximizedPanel();

  if (!isFetched) return null;
  if (!isAllowed)
    return (
      <Card className="m-5 flex grow flex-col p-4">
        <NoPermissions>{noPermissionMessage}</NoPermissions>
      </Card>
    );

  if (!data) return null;

  return (
    <div className="grid grow grid-rows-[auto_1fr]">
      <Header />
      <WorkspaceLayout
        layout={preferedLayout}
        logComponent={<Logs />}
        diagramComponent={
          <Diagram
            workflowPath={data.instance.as}
            flow={data.flow}
            status={data.instance.status}
          />
        }
        inputOutputComponent={<InputOutput />}
      />
    </div>
  );
};

export default InstancesDetail;
