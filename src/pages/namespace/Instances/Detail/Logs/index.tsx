import { Bug, Copy, Filter, Maximize2, WrapText } from "lucide-react";
import { Dispatch, FC, SetStateAction } from "react";

import Button from "~/design/Button";
import { ButtonBar } from "~/design/ButtonBar";
import { FiltersObj } from "~/api/logs/query/get";
import Input from "~/design/Input";
import ScrollContainer from "./ScrollContainer";

const LogsPanel: FC<{
  instanceId: string;
  query: FiltersObj;
  setQuery: Dispatch<SetStateAction<FiltersObj>>;
}> = ({ instanceId, query, setQuery }) => (
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
      <Input
        className="h-6"
        placeholder="state name"
        onChange={(e) => {
          let query = {};
          if (e.target.value) {
            query = {
              QUERY: {
                type: "MATCH",
                stateName: e.target.value,
              },
            };
          }

          setQuery(() => ({
            ...query,
          }));
        }}
      />
    </div>

    <ScrollContainer instanceId={instanceId} query={query} />
  </>
);

export default LogsPanel;
