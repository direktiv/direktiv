import { FC } from "react";
import ScrollContainer from "./ScrollContainer";

const LogsPanel: FC<{ instanceId: string }> = ({ instanceId }) => (
  <ScrollContainer instanceId={instanceId} />
);

export default LogsPanel;
