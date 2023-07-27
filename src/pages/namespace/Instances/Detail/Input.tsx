import Editor from "~/design/Editor";
import { FC } from "react";
import { useInput } from "~/api/instances/query/input";
import { useTheme } from "~/util/store/theme";

const Input: FC<{ instanceId: string }> = ({ instanceId }) => {
  const { data } = useInput({ instanceId });
  const theme = useTheme();
  if (!data) return null;

  const workflowInput = atob(data.data);

  return (
    <Editor
      value={workflowInput}
      language="json"
      theme={theme ?? undefined}
      options={{ readOnly: true }}
    />
  );
};

export default Input;
