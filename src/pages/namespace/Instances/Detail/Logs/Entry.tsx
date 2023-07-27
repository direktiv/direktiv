import { FC } from "react";
import { LogEntry } from "~/design/Logs";
import { LogEntryType } from "~/api/logs/schema";
import moment from "moment";

const Entry: FC<{ logEntry: LogEntryType }> = ({ logEntry }) => {
  const { msg, t, level } = logEntry;

  const time = moment(t).format("HH:mm:ss");

  const isErr = level === "error";

  return (
    <LogEntry variant={isErr ? "error" : undefined} time={time}>
      {msg}
    </LogEntry>
  );
};

export default Entry;
