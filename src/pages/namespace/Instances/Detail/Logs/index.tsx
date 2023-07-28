import Entry from "./Entry";
import { FC } from "react";
import { Logs } from "~/design/Logs";
import { useLogs } from "~/api/logs/query/get";

const LogsPanel: FC<{ instanceId: string }> = ({ instanceId }) => {
  const { data } = useLogs({ instanceId }, { streaming: true });

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
