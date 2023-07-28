import Entry from "./Entry";
import { FC } from "react";
import { Logs } from "~/design/Logs";
import { useLogs } from "~/api/logs/query/get";

const LogsPanel: FC<{ instanceId: string; stream: boolean }> = ({
  instanceId,
  stream,
}) => {
  const { data } = useLogs({ instanceId }, { stream });

  if (!data) return null;

  return (
    <Logs linewrap={true} className="grow">
      {data.results.map((logEntry) => (
        <Entry key={logEntry.t} logEntry={logEntry} />
      ))}
    </Logs>
  );
};

export default LogsPanel;
