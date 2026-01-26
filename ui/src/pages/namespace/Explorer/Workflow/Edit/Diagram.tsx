import {
  InstanceFlowResponseSchema,
  workflowStateSchema,
} from "~/api/instances/schema";
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
import { useFile } from "~/api/files/query/file";
import { useTranslation } from "react-i18next";

type DiagramProps = {
  flow?: string[];
  path?: string;
};

const Diagram: FC<DiagramProps> = ({ path }) => {
  const { setMaximizedPanel } = useLogsPreferencesActions();
  const { t } = useTranslation();
  const maximizedPanel = useLogsPreferencesMaximizedPanel();
  const isMaximized = maximizedPanel === "diagram";

  const { data } = useFile({ path });

  if (data === undefined) return null;

  const workflowData =
    data.type === "workflow" ? { data: data.states } : undefined;

  const parsedWorkflow = InstanceFlowResponseSchema.safeParse(workflowData);

  if (!parsedWorkflow.success) {
    // Todo: Decide what kind of error handling is appropriate here
    return null;
  }

  const parsedWorkflowData = workflowStateSchema.safeParse(workflowData?.data);

  if (!parsedWorkflowData.success) {
    // Todo: Decide what kind of error handling is appropriate here
    return null;
  }

  const dataArray = Array.isArray(parsedWorkflow.data.data)
    ? parsedWorkflow.data.data
    : Object.values(parsedWorkflow.data.data ?? {});

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
        workflow={parsedWorkflow.data}
        flow={flowStatesArray}
        orientation="horizontal"
        instanceStatus="pending"
      />
    </div>
  );
};

export default Diagram;
