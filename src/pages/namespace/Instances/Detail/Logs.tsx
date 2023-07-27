import { LogEntry, Logs } from "~/design/Logs";

const LogsPanel = () => (
  <Logs linewrap={true} className="grow">
    <LogEntry time="12:34:23">Preparing workflow triggered by api.</LogEntry>
    <LogEntry time="12:34:23">Starting workflow demo.yml.</LogEntry>
    <LogEntry time="12:34:23">
      Running state logic (step:1) -- helloworld
    </LogEntry>
    <LogEntry time="12:34:23">Transforming state data.</LogEntry>
    <LogEntry time="12:34:23" variant="warning">
      Warning: this is a very long line with a warning. this is a very long line
      with a warning. this is a very long line with a warning. this is a very
      long line with a warning. this is a very long line with a warning. this is
      a very long line with a warning. this is a very long line with a warning.
      this is a very long line with a warning. this is a very long line with a
      warning. this is a very long line with a warning. this is a very long line
      with a warning. this is a very long line with a warning.
    </LogEntry>
    <LogEntry time="12:34:23" variant="success">
      Workflow demo.yml completed.
    </LogEntry>
  </Logs>
);

export default LogsPanel;
