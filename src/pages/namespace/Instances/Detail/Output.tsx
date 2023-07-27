import Editor from "~/design/Editor";
import { FC } from "react";
import { useOutput } from "~/api/instances/query/output";
import { useTheme } from "~/util/store/theme";

const Output: FC<{ instanceId: string }> = ({ instanceId }) => {
  const { data, isFetched, isError } = useOutput({ instanceId });
  const theme = useTheme();

  if (!isFetched) return null;

  let workflowOutput = "// no data";

  if (isError) {
    workflowOutput = "// no output data was resolved";
  }

  if (data) {
    workflowOutput = atob(data.data);
  }

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
