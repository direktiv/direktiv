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
import { FC } from "react";
import WorkflowDiagram from "~/design/WorkflowDiagram";
import { useNodeContent } from "~/api/tree/query/node";
import { useTranslation } from "react-i18next";

const Diagram: FC<{ workflowPath: string; flow: string[] }> = ({
  workflowPath,
  flow,
}) => {
  const { data } = useNodeContent({ path: workflowPath });
  const { setMaximizedPanel } = useLogsPreferencesActions();
  const { t } = useTranslation();
  const maximizedPanel = useLogsPreferencesMaximizedPanel();
  const isMaximized = maximizedPanel === "diagram";

  if (!data) return null;

  const workflowData = atob(data.revision?.source ?? "");

  return (
    <div className="relative flex grow">
      <TooltipProvider>
        <Tooltip>
          <TooltipTrigger asChild>
            <div className="absolute right-5 top-5 z-50">
              <Button
                icon
                size="sm"
                variant="outline"
                onClick={() => {
                  setMaximizedPanel(isMaximized ? "none" : "diagram");
                }}
              >
                {isMaximized ? <Minimize2 /> : <Maximize2 />}
              </Button>
            </div>
          </TooltipTrigger>
          <TooltipContent>
            {isMaximized
              ? t("pages.instances.detail.diagram.minimizeInput")
              : t("pages.instances.detail.diagram.maximizeInput")}
          </TooltipContent>
        </Tooltip>
      </TooltipProvider>
      <WorkflowDiagram
        workflow={workflowData}
        flow={flow}
        orientation="horizontal"
        instanceStatus="complete"
      />
    </div>
  );
};

export default Diagram;
