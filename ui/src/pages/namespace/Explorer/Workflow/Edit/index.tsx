import { FC } from "react";
import WorkflowEditor from "./WorkflowEditor";
import { useFile } from "~/api/files/query/file";
import { useParams } from "@tanstack/react-router";

const WorkflowEditPage: FC = () => {
  const { _splat: path } = useParams({ strict: false });

  const { data } = useFile({ path });

  if (data?.type !== "workflow" || !path) return null;

  // Editor is moved into a separate component to give us a state where
  // data and path is alwawys defined. This makes handling side effects
  // much easier.
  return <WorkflowEditor data={data} />;
};

export default WorkflowEditPage;
