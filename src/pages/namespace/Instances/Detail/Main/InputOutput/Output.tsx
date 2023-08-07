import Editor from "~/design/Editor";
import { FC } from "react";
import InfoText from "./OutputInfo";
import Toolbar from "./Toolbar";
import { useInstanceId } from "../../store/instanceContext";
import { useOutput } from "~/api/instances/query/output";
import { useTheme } from "~/util/store/theme";
import { useTranslation } from "react-i18next";

const Output: FC<{ instanceIsFinished: boolean }> = ({
  instanceIsFinished,
}) => {
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

  const workflowOutput = atob(data?.data ?? "");

  return (
    <div className="flex grow flex-col gap-5 pb-12">
      <Toolbar copyText={workflowOutput} variant="output" />
      <Editor
        value={workflowOutput}
        language="json"
        theme={theme ?? undefined}
        options={{ readOnly: true }}
      />
    </div>
  );
};

export default Output;
