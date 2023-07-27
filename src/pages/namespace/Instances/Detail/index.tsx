import { Card } from "~/design/Card";
import Diagram from "./Diagram";
import { FC } from "react";
import Header from "./Header";
import Input from "./Input";
import Logs from "./Logs";
import Output from "./Output";
import { pages } from "~/util/router/pages";
import { useInstanceDetails } from "~/api/instances/query/details";

const InstancesDetailPage = () => {
  const { instance } = pages.instances.useParams();
  if (!instance) return null;
  return <InstancesDetailPageWithId instanceId={instance} />;
};

const InstancesDetailPageWithId: FC<{ instanceId: string }> = ({
  instanceId,
}) => {
  const { data } = useInstanceDetails({ instanceId });

  if (!data) return null;

  return (
    <div className="flex grow flex-col">
      <Header instanceId={instanceId} />
      <div className="grid grow grid-cols-2 gap-5 p-5">
        <Card className="p-5">
          <Logs />
        </Card>
        <Card>
          <Diagram workflowPath={data.workflow.path} flow={data.flow} />
        </Card>
        <Card className="p-5">
          <Input instanceId={instanceId} />
        </Card>
        <Card className="p-5">
          <Output instanceId={instanceId} />
        </Card>
      </div>
    </div>
  );
};

export default InstancesDetailPage;
