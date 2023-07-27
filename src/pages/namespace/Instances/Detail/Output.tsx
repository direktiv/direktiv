import { Card } from "~/design/Card";
import { FC } from "react";
import { useOutput } from "~/api/instances/query/output";

const Output: FC<{ instanceId: string }> = ({ instanceId }) => {
  const { data } = useOutput({ instanceId });
  if (!data) return null;

  const workflowOutput = atob(data.data);

  return (
    <Card className="p-5">
      <pre>{workflowOutput}</pre>
    </Card>
  );
};

export default Output;
