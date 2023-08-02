import { Card } from "~/design/Card";
import Diagram from "./Diagram";
import Header from "./Header";
import Logs from "./Logs";
import Output from "./InputOutput/Output";
import { twMergeClsx } from "~/util/helpers";
import { useInstanceDetails } from "~/api/instances/query/details";
import { useInstanceId } from "./state/instanceContext";

const InstancesDetail = () => {
  const instanceId = useInstanceId();
  const { data } = useInstanceDetails({ instanceId });

  if (!data) return null;

  const instanceIsFinished = data.instance.status !== "pending";

  return (
    <div className="grid grow grid-rows-[auto_1fr]">
      <Header instanceId={instanceId} />
      <div
        className={twMergeClsx(
          "grid grow gap-5 p-5",
          "grid-rows-[minmax(300px,50vh)_1fr]",
          "grid-cols-[1fr_400px]",
          "grid-template"
        )}
      >
        <Card className="relative col-span-2 grid grid-rows-[auto,1fr] p-5">
          <Logs />
        </Card>
        <Card>
          <Diagram workflowPath={data.workflow.path} flow={data.flow} />
        </Card>
        <Card className="p-5">
          <Output instanceIsFinished={instanceIsFinished} />
        </Card>
      </div>
    </div>
  );
};

export default InstancesDetail;
