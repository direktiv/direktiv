import Editor from "~/design/Editor";
import Toolbar from "./Toolbar";
import { decode } from "js-base64";
import { prettifyJsonString } from "~/util/helpers";
import { useInput } from "~/api/instances/query/input";
import { useInstanceId } from "../../store/instanceContext";
import { useTheme } from "~/util/store/theme";

const Input = () => {
  const instanceId = useInstanceId();
  const { data } = useInput({ instanceId });
  const theme = useTheme();

  const workflowInput = decode(data?.data ?? "");
  const workflowInputPretty = prettifyJsonString(workflowInput);

  return (
    <div className="flex grow flex-col gap-5 pb-12">
      <Toolbar copyText={workflowInputPretty} variant="input" />
      <Editor
        value={workflowInputPretty}
        language="json"
        theme={theme ?? undefined}
        options={{ readOnly: true }}
      />
    </div>
  );
};

export default Input;
