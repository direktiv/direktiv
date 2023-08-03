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
import { useInput } from "~/api/instances/query/input";
import { useInstanceId } from "../state/instanceContext";
import { useTheme } from "~/util/store/theme";
import { useTranslation } from "react-i18next";

const Input = () => {
  const instanceId = useInstanceId();
  const { t } = useTranslation();
  const { data } = useInput({ instanceId });
  const theme = useTheme();
  const { setMaximizedPanel } = useLogsPreferencesActions();
  const maximizedPanel = useLogsPreferencesMaximizedPanel();
  const isMaximized = maximizedPanel === "input-output";

  const workflowInput = atob(data?.data ?? "");

  return (
    <div className="flex grow flex-col gap-5 pb-12">
      <ButtonBar className="justify-end">
        <TooltipProvider>
          <Tooltip>
            <TooltipTrigger asChild>
              <div>
                <CopyButton
                  value={workflowInput}
                  buttonProps={{
                    variant: "outline",
                    size: "sm",
                  }}
                />
              </div>
            </TooltipTrigger>
            <TooltipContent>
              {t("pages.instances.detail.inputOutput.copyInput")}
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
                ? t("pages.instances.detail.inputOutput.minimizeInput")
                : t("pages.instances.detail.inputOutput.maximizeInput")}
            </TooltipContent>
          </Tooltip>
        </TooltipProvider>
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
