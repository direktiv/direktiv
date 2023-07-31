import { Bug, Copy, Filter, Maximize2, WrapText } from "lucide-react";

import Button from "~/design/Button";
import { ButtonBar } from "~/design/ButtonBar";
import { FC } from "react";
import Input from "~/design/Input";
import ScrollContainer from "./ScrollContainer";

const LogsPanel: FC<{ instanceId: string }> = ({ instanceId }) => (
  <>
    <div className="mb-5 flex gap-x-5">
      <h3 className="grow font-medium">Logs</h3>
      <ButtonBar>
        <Button icon variant="outline" size="sm">
          <Bug />
        </Button>
        <Button icon variant="outline" size="sm">
          <WrapText />
        </Button>
        <Button icon variant="outline" size="sm">
          <Filter />
        </Button>
        <Button icon variant="outline" size="sm">
          <Maximize2 />
        </Button>
        <Button icon variant="outline" size="sm">
          <Copy />
        </Button>
      </ButtonBar>
      <Input className="h-6" placeholder="workflow name" />
      <Input className="h-6" placeholder="state name" />
    </div>

    <ScrollContainer instanceId={instanceId} />
  </>
);

export default LogsPanel;
