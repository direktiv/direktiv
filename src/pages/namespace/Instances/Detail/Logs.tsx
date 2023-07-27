import { LogEntry, Logs } from "~/design/Logs";

import { FC } from "react";
import { useLogs } from "~/api/logs/query/get";

const LogsPanel: FC<{ instanceId: string }> = ({ instanceId }) => {
  const data = useLogs({ instanceId });

  if (!data) return null;

  return (
    <Logs linewrap={true} className="grow">
      <LogEntry time="12:34:23">Preparing workflow triggered by api.</LogEntry>
      <LogEntry time="12:34:23">Starting workflow demo.yml.</LogEntry>
      <LogEntry time="12:34:23">
        Running state logic (step:1) -- helloworld
      </LogEntry>
      <LogEntry time="12:34:23">Transforming state data.</LogEntry>
      <LogEntry time="12:34:23" variant="warning">
        Warning: this is a very long line with a warning. this is a very long
        line with a warning. this is a very long line with a warning. this is a
        very long line with a warning. this is a very long line with a warning.
        this is a very long line with a warning. this is a very long line with a
        warning. this is a very long line with a warning. this is a very long
        line with a warning. this is a very long line with a warning. this is a
        very long line with a warning. this is a very long line with a warning.
      </LogEntry>
      <LogEntry time="12:34:23" variant="success">
        Workflow demo.yml completed.
      </LogEntry>
    </Logs>
  );
};

export default LogsPanel;
