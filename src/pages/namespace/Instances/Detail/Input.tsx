import { Card } from "~/design/Card";
import { FC } from "react";
import { useInput } from "~/api/instances/query/input";

const Input: FC<{ instanceId: string }> = ({ instanceId }) => {
  const { data } = useInput({ instanceId });
  if (!data) return null;

  const workflowInput = atob(data.data);

  return (
    <Card className="p-5">
      <pre>{workflowInput}</pre>
    </Card>
  );
};

export default Input;
