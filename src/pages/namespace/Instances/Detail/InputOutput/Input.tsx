import Editor from "~/design/Editor";
import { useInput } from "~/api/instances/query/input";
import { useInstanceId } from "../state/instanceContext";
import { useTheme } from "~/util/store/theme";

const Input = () => {
  const instanceId = useInstanceId();
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
