import { Bug, Copy, Maximize2, Plus, WrapText } from "lucide-react";
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "~/design/Tooltip";
import {
  useLogsPreferencesActions,
  useLogsPreferencesVerboseLogs,
  useLogsPreferencesWordWrap,
} from "~/util/store/logs";

import Button from "~/design/Button";
import { ButtonBar } from "~/design/ButtonBar";
import Input from "~/design/Input";
import ScrollContainer from "./ScrollContainer";
import { Toggle } from "~/design/Toggle";
import { useActions } from "../state/instanceContext";

const LogsPanel = () => {
  const { updateFilterStateName, updateFilterWorkflow } = useActions();
  const wordWrap = useLogsPreferencesWordWrap();
  const verboseLogs = useLogsPreferencesVerboseLogs();
  const { setVerboseLogs, setWordWrap } = useLogsPreferencesActions();
  return (
    <>
      <div className="mb-5 flex gap-x-5">
        <h3 className="grow font-medium">Logs</h3>
        <Input
          className="h-6"
          placeholder="filter by workflow name"
          onChange={(e) => {
            updateFilterWorkflow(e.target.value);
          }}
        />
        <Input
          className="h-6"
          placeholder="filter by state name"
          onChange={(e) => {
            updateFilterStateName(e.target.value);
          }}
        />
        <Button icon variant="outline" size="sm">
          <Plus /> Filter
        </Button>
        <ButtonBar>
          <TooltipProvider>
            <Tooltip>
              <TooltipTrigger asChild>
                <div className="flex grow">
                  <Toggle
                    size="sm"
                    pressed={verboseLogs}
                    onClick={() => {
                      setVerboseLogs(!verboseLogs);
                    }}
                  >
                    <Bug />
                  </Toggle>
                </div>
              </TooltipTrigger>
              <TooltipContent>Verbose Logs</TooltipContent>
            </Tooltip>
            <Tooltip>
              <TooltipTrigger asChild>
                <div className="flex grow">
                  <Toggle
                    size="sm"
                    pressed={wordWrap}
                    onClick={() => {
                      setWordWrap(!wordWrap);
                    }}
                  >
                    <WrapText />
                  </Toggle>
                </div>
              </TooltipTrigger>
              <TooltipContent>Word Wrap</TooltipContent>
            </Tooltip>
            <Tooltip>
              <TooltipTrigger asChild>
                <div className="flex grow">
                  <Button icon variant="outline" size="sm">
                    <Copy />
                  </Button>
                </div>
              </TooltipTrigger>
              <TooltipContent>Copy Logs</TooltipContent>
            </Tooltip>

            <Tooltip>
              <TooltipTrigger asChild>
                <div className="flex grow">
                  <Button icon variant="outline" size="sm">
                    <Maximize2 />
                  </Button>
                </div>
              </TooltipTrigger>
              <TooltipContent>Maximize Logs</TooltipContent>
            </Tooltip>
          </TooltipProvider>
        </ButtonBar>
      </div>
      <ScrollContainer />
    </>
  );
};

export default LogsPanel;
