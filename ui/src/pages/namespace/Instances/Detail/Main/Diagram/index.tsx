import { InstanceSchemaType, Workflow } from "~/api/instances/schema";
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
import { State } from "~/design/WorkflowDiagram/types";
import WorkflowDiagram from "~/design/WorkflowDiagram";
import { instanceStatusToDiagramStatus } from "./utils";
import { useFile } from "~/api/files/query/file";
import { useInstanceFlow } from "~/api/instances/query/flow";
import { useTranslation } from "react-i18next";

type DiagramProps = {
  flow?: string[];
  instanceId: string;
  path?: string;
  status: InstanceSchemaType["status"];
};

const Diagram: FC<DiagramProps> = ({ instanceId, path, status }) => {
  const { setMaximizedPanel } = useLogsPreferencesActions();
  const { t } = useTranslation();
  const maximizedPanel = useLogsPreferencesMaximizedPanel();
  const isMaximized = maximizedPanel === "diagram";

  const { data: dataFromInstance } = useInstanceFlow({ instanceId });
  const { data: dataFromWorkflow } = useFile({ path });

  const dataObject =
    instanceId === "null" && dataFromWorkflow?.type === "workflow"
      ? { data: dataFromWorkflow.states || {} }
      : (dataFromInstance ?? { data: {} });

  const workflowData = (dataObject as Workflow) ?? undefined;

  if (!workflowData) return null;

  const dataArray = Array.isArray(dataObject)
    ? dataObject
    : Object.values(dataObject.data);
  const flowStatesArray = dataArray.every((item: State) => item.name)
    ? dataArray.map((item) => item.name)
    : [""];

  return (
    <div className="relative flex grow">
      <TooltipProvider>
        <Tooltip>
          <TooltipTrigger asChild>
            <div className="absolute right-5 top-5 z-50">
              <Button
                data-testid="resizeDiagram"
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
        flow={flowStatesArray}
        orientation="horizontal"
        instanceStatus={instanceStatusToDiagramStatus(status)}
      />
    </div>
  );
};

export default Diagram;
