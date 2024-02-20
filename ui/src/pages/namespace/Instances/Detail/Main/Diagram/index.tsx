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
import { InstanceSchemaType } from "~/api/instances/schema";
import WorkflowDiagram from "~/design/WorkflowDiagram";
import { decode } from "js-base64";
import { instanceStatusToDiagramStatus } from "./utils";
import { useNode } from "~/api/files/query/node";
import { useTranslation } from "react-i18next";

type DiagramProps = {
  workflowPath: string;
  flow: string[];
  status: InstanceSchemaType["status"];
};

const Diagram: FC<DiagramProps> = ({ workflowPath, flow, status }) => {
  const { data } = useNode({ path: workflowPath });
  const { setMaximizedPanel } = useLogsPreferencesActions();
  const { t } = useTranslation();
  const maximizedPanel = useLogsPreferencesMaximizedPanel();
  const isMaximized = maximizedPanel === "diagram";

  if (!data) return null;

  const workflowData = decode(data.data ?? "");

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
                className="bg-white dark:bg-black"
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
        instanceStatus={instanceStatusToDiagramStatus(status)}
      />
    </div>
  );
};

export default Diagram;
