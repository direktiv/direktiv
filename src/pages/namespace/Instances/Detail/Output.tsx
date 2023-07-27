import { Card } from "~/design/Card";
import Editor from "~/design/Editor";
import { FC } from "react";
import { useOutput } from "~/api/instances/query/output";
import { useTheme } from "~/util/store/theme";

const Output: FC<{ instanceId: string }> = ({ instanceId }) => {
  const { data } = useOutput({ instanceId });
  const theme = useTheme();
  if (!data) return null;

  const workflowOutput = atob(data.data);

  return (
    <Card className="p-5">
      <Editor
        value={workflowOutput}
        language="json"
        theme={theme ?? undefined}
        options={{ readOnly: true }}
      />
    </Card>
  );
};

export default Output;
