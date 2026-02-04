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

import Alert from "~/design/Alert";
import Button from "~/design/Button";
import { FC } from "react";
import { InstanceFlowResponseSchema } from "~/api/instances/schema";
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

  const parsedInstanceFlow = InstanceFlowResponseSchema.safeParse(workflowData);

  if (!parsedInstanceFlow.success) {
    return (
      <div className="relative flex grow">
        <div>
          <Alert variant="error" className="grow-0">
            {t("pages.instances.detail.diagram.invalid")}
          </Alert>
        </div>
      </div>
    );
  }

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
        states={parsedInstanceFlow.data}
        orientation="horizontal"
        instanceStatus="pending"
      />
    </div>
  );
};

export default Diagram;
