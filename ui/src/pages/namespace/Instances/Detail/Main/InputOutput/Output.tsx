import Editor from "~/design/Editor";
import InfoText from "./OutputInfo";
import Toolbar from "./Toolbar";
import { decode } from "js-base64";
import { forwardRef } from "react";
import { prettifyJsonString } from "~/util/helpers";
import { useInstanceId } from "../../store/instanceContext";
import { useOutput } from "~/api/instances/query/output";
import { useTheme } from "~/util/store/theme";
import { useTranslation } from "react-i18next";

const Output = forwardRef<
  HTMLDivElement,
  {
    instanceIsFinished: boolean;
  }
>(({ instanceIsFinished }, ref) => {
  const instanceId = useInstanceId();
  const { t } = useTranslation();
  const { data, isError } = useOutput({
    instanceId,
    enabled: instanceIsFinished,
  });
  const theme = useTheme();

  if (!instanceIsFinished) {
    return (
      <InfoText>
        {t("pages.instances.detail.inputOutput.output.stillRunningMsg")}
      </InfoText>
    );
  }

  if (isError) {
    return (
      <InfoText>
        {t("pages.instances.detail.inputOutput.output.noOutputMsg")}
      </InfoText>
    );
  }

  const workflowOutput = decode(data?.data ?? "");
  const workflowOutputPretty = prettifyJsonString(workflowOutput);

  return (
    <div className="flex grow flex-col gap-5 pb-12" ref={ref}>
      <Toolbar copyText={workflowOutput} variant="output" />
      <Editor
        value={workflowOutputPretty}
        language="json"
        theme={theme ?? undefined}
        options={{ readOnly: true }}
      />
    </div>
  );
});

Output.displayName = "Output";

export default Output;
