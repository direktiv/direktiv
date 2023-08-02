import Button from "~/design/Button";
import { ButtonBar } from "~/design/ButtonBar";
import CopyButton from "~/design/CopyButton";
import Editor from "~/design/Editor";
import { Maximize2 } from "lucide-react";
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
    <div className="flex grow flex-col gap-5 pb-12">
      <ButtonBar className="justify-end">
        <CopyButton
          value={workflowInput}
          buttonProps={{
            variant: "outline",
            size: "sm",
          }}
        />
        <Button icon size="sm" variant="outline">
          <Maximize2 />
        </Button>
      </ButtonBar>
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
