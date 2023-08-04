import Editor from "~/design/Editor";
import Toolbar from "./Toolbar";
import { useInput } from "~/api/instances/query/input";
import { useInstanceId } from "../../store/instanceContext";
import { useTheme } from "~/util/store/theme";

const Input = () => {
  const instanceId = useInstanceId();
  const { data } = useInput({ instanceId });
  const theme = useTheme();

  const workflowInput = atob(data?.data ?? "");

  return (
    <div className="flex grow flex-col gap-5 pb-12">
      <Toolbar copyText={workflowInput} variant="input" />
      <Editor
        value={workflowInput}
        language="json"
        theme={theme ?? undefined}
        options={{ readOnly: true }}
      />
    </div>
  );
};

export default Input;
