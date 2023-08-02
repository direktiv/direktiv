import { FC, PropsWithChildren } from "react";

import Button from "~/design/Button";
import { ButtonBar } from "~/design/ButtonBar";
import CopyButton from "~/design/CopyButton";
import Editor from "~/design/Editor";
import { Maximize2 } from "lucide-react";
import { useInstanceId } from "../state/instanceContext";
import { useOutput } from "~/api/instances/query/output";
import { useTheme } from "~/util/store/theme";
import { useTranslation } from "react-i18next";

const Info: FC<PropsWithChildren> = ({ children }) => (
  <div className="flex h-full flex-col items-center justify-center gap-y-5 p-10">
    <span className="text-center text-gray-11">{children}</span>
  </div>
);

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
      <Info>
        {t("pages.instances.detail.inputOutput.output.stillRunningMsg")}
      </Info>
    );
  }

  if (isError) {
    return (
      <Info>{t("pages.instances.detail.inputOutput.output.noOutputMsg")}</Info>
    );
  }

  const workflowOutput = atob(data?.data ?? "");

  return (
    <div className="flex grow flex-col gap-5 pb-12">
      <ButtonBar>
        <CopyButton
          value={workflowOutput}
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
        value={workflowOutput}
        language="json"
        theme={theme ?? undefined}
        options={{ readOnly: true }}
      />
    </div>
  );
};

export default Output;
