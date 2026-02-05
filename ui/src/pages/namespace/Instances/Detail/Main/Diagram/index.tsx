import {
  InstanceSchemaType,
  WorkflowStatesSchema,
  WorkflowStatesSchemaType,
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

import Alert from "~/design/Alert";
import Button from "~/design/Button";
import { FC } from "react";
import WorkflowDiagram from "~/design/WorkflowDiagram";
import { instanceStatusToDiagramStatus } from "./utils";
import { useTranslation } from "react-i18next";

type DiagramProps = {
  states?: WorkflowStatesSchemaType;
  instanceStatus?: InstanceSchemaType["status"];
  resizable?: boolean;
};

const Diagram: FC<DiagramProps> = ({
  states,
  instanceStatus,
  resizable = false,
}) => {
  const { setMaximizedPanel } = useLogsPreferencesActions();
  const { t } = useTranslation();
  const maximizedPanel = useLogsPreferencesMaximizedPanel();
  const isMaximized = maximizedPanel === "diagram";

  const parsedInstanceFlow = WorkflowStatesSchema.safeParse(states);

  if (!parsedInstanceFlow.success) {
    return (
      <div className="flex grow flex-col items-center justify-center">
        <Alert variant="error" className="grow-0">
          {t("pages.instances.detail.diagram.invalid")}
        </Alert>
      </div>
    );
  }

  return (
    <div className="relative flex grow">
      {resizable && (
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
      )}
      <WorkflowDiagram
        states={parsedInstanceFlow.data}
        orientation="horizontal"
        // todo: is this still needed?
        instanceStatus={instanceStatusToDiagramStatus(instanceStatus)}
      />
    </div>
  );
};

export default Diagram;
