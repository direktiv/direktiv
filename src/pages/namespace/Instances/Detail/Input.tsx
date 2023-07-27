import { Card } from "~/design/Card";
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
    <Card className="p-5">
      <Editor
        value={workflowInput}
        language="json"
        theme={theme ?? undefined}
        options={{ readOnly: true }}
      />
    </Card>
  );
};

export default Input;
