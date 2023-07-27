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
    <Editor
      value={workflowOutput}
      language="json"
      theme={theme ?? undefined}
      options={{ readOnly: true }}
    />
  );
};

export default Output;
