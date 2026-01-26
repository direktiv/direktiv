import {
  InstanceFlowResponseSchema,
  InstanceSchemaType,
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
import { instanceStatusToDiagramStatus } from "./utils";
import { useInstanceFlow } from "~/api/instances/query/flow";
import { useTranslation } from "react-i18next";

type DiagramProps = {
  instanceId: string;
  status: InstanceSchemaType["status"];
};

const Diagram: FC<DiagramProps> = ({ instanceId, status }) => {
  const { setMaximizedPanel } = useLogsPreferencesActions();
  const { t } = useTranslation();
  const maximizedPanel = useLogsPreferencesMaximizedPanel();
  const isMaximized = maximizedPanel === "diagram";

  const { data } = useInstanceFlow({ instanceId });

  const parsedWorkflow = InstanceFlowResponseSchema.safeParse(data);

  if (!parsedWorkflow.success) {
    // Todo: Decide what kind of error handling is appropriate here
    return null;
  }

  if (data === undefined) return null;

  const dataArray = Array.isArray(data) ? data : Object.values(data.data);
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
        instanceStatus={instanceStatusToDiagramStatus(status)}
      />
    </div>
  );
};

export default Diagram;
