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
import { FC } from "react";
import { useTranslation } from "react-i18next";

const Toolbar: FC<{ copyText?: string; variant: "input" | "output" }> = ({
  copyText,
  variant,
}) => {
  const { t } = useTranslation();
  const { setMaximizedPanel } = useLogsPreferencesActions();
  const maximizedPanel = useLogsPreferencesMaximizedPanel();
  const isMaximized = maximizedPanel === "input-output";

  return (
    <ButtonBar className="justify-end">
      <TooltipProvider>
        <Tooltip>
          <TooltipTrigger asChild>
            <div>
              <CopyButton
                value={copyText ?? ""}
                buttonProps={{
                  variant: "outline",
                  size: "sm",
                  disabled: !copyText,
                }}
              />
            </div>
          </TooltipTrigger>
          <TooltipContent>
            {t(
              variant === "input"
                ? "pages.instances.detail.inputOutput.copyInput"
                : "pages.instances.detail.inputOutput.copyOutput"
            )}
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
            {isMaximized &&
              variant === "input" &&
              t("pages.instances.detail.inputOutput.minimizeInput")}
            {isMaximized &&
              variant === "output" &&
              t("pages.instances.detail.inputOutput.minimizeOutput")}
            {!isMaximized &&
              variant === "input" &&
              t("pages.instances.detail.inputOutput.maximizeInput")}
            {!isMaximized &&
              variant === "output" &&
              t("pages.instances.detail.inputOutput.maximizeOutput")}
          </TooltipContent>
        </Tooltip>
      </TooltipProvider>
    </ButtonBar>
  );
};

export default Toolbar;
