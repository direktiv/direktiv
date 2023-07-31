import { Dispatch, FC, SetStateAction } from "react";

import { Card } from "~/design/Card";
import Diagram from "./Diagram";
import { FiltersObj } from "~/api/logs/query/get";
import Header from "./Header";
import Input from "./Input";
import Logs from "./Logs";
import Output from "./Output";
import { twMergeClsx } from "~/util/helpers";
import { useInstanceDetails } from "~/api/instances/query/details";

const InstancesDetail: FC<{
  instanceId: string;
  query: FiltersObj;
  setQuery: Dispatch<SetStateAction<FiltersObj>>;
}> = ({ instanceId, query, setQuery }) => {
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
          "grid-cols-[1fr_400px]"
        )}
      >
        <Card className="relative grid grid-rows-[auto,1fr] p-5">
          <Logs instanceId={instanceId} query={query} setQuery={setQuery} />
        </Card>
        <Card className="p-5">
          <Input instanceId={instanceId} />
        </Card>
        <Card>
          <Diagram workflowPath={data.workflow.path} flow={data.flow} />
        </Card>
        <Card className="p-5">
          <Output
            instanceId={instanceId}
            instanceIsFinished={instanceIsFinished}
          />
        </Card>
      </div>
    </div>
  );
};

export default InstancesDetail;
