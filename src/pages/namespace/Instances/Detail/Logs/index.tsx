import { Bug, Copy, Maximize2, WrapText } from "lucide-react";

import Button from "~/design/Button";
import { ButtonBar } from "~/design/ButtonBar";
import Input from "~/design/Input";
import ScrollContainer from "./ScrollContainer";
import { useActions } from "../state/instanceContext";

const LogsPanel = () => {
  const { updateFilterStateName, updateFilterWorkflow } = useActions();

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
        <ButtonBar>
          <Button icon variant="outline" size="sm">
            <Bug />
          </Button>
          <Button icon variant="outline" size="sm">
            <WrapText />
          </Button>
          <Button icon variant="outline" size="sm">
            <Maximize2 />
          </Button>
          <Button icon variant="outline" size="sm">
            <Copy />
          </Button>
        </ButtonBar>
      </div>
      <ScrollContainer />
    </>
  );
};

export default LogsPanel;
