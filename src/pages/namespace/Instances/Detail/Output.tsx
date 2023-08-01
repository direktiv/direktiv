import Editor from "~/design/Editor";
import { FC } from "react";
import { useInstanceId } from "./state/instanceContext";
import { useOutput } from "~/api/instances/query/output";
import { useTheme } from "~/util/store/theme";

const Output: FC<{ instanceIsFinished: boolean }> = ({
  instanceIsFinished,
}) => {
  const instanceId = useInstanceId();
  const { data, isFetched, isError } = useOutput({
    instanceId,
    enabled: instanceIsFinished,
  });
  const theme = useTheme();

  let workflowOutput = "// no data";

  if (!isFetched) {
    return (
      <div className="flex h-full flex-col items-center justify-center gap-y-5 p-10">
        <span className="text-center text-gray-11">
          The workflow is still running
        </span>
      </div>
    );
  }

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
