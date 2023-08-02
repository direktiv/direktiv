import { FC, PropsWithChildren } from "react";
import { Maximize2, Minimize2 } from "lucide-react";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "~/design/Tooltip";
import {
  useLogsPreferencesActions,
  useLogsPreferencesMaximizedPanel,
} from "~/util/store/logs";

import Button from "~/design/Button";
import { ButtonBar } from "~/design/ButtonBar";
import CopyButton from "~/design/CopyButton";
import Editor from "~/design/Editor";
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
  const { setMaximizedPanel } = useLogsPreferencesActions();
  const maximizedPanel = useLogsPreferencesMaximizedPanel();
  const isMaximized = maximizedPanel === "input-output";

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
      <ButtonBar className="justify-end">
        <TooltipProvider>
          <Tooltip>
            <TooltipTrigger asChild>
              <div>
                <CopyButton
                  value={workflowOutput}
                  buttonProps={{
                    variant: "outline",
                    size: "sm",
                  }}
                />
              </div>
            </TooltipTrigger>
            <TooltipContent>
              {t("pages.instances.detail.inputOutput.copyOutput")}
            </TooltipContent>
          </Tooltip>
          <Tooltip>
            <TooltipTrigger asChild>
              <div>
                <Button
                  icon
                  size="sm"
                  variant="outline"
                  onClick={() => {
                    setMaximizedPanel(isMaximized ? "none" : "input-output");
                  }}
                >
                  {isMaximized ? <Minimize2 /> : <Maximize2 />}
                </Button>
              </div>
            </TooltipTrigger>
            <TooltipContent>
              {isMaximized
                ? t("pages.instances.detail.inputOutput.minimizeOutput")
                : t("pages.instances.detail.inputOutput.maximizeOutput")}
            </TooltipContent>
          </Tooltip>
        </TooltipProvider>
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
