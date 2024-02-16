import { FC } from "react";
import WorkflowEditor from "./WorkflowEditor";
import { pages } from "~/util/router/pages";
import { useNode } from "~/api/filesTree/query/node";

const WorkflowOverviewPage: FC = () => {
  const { path } = pages.explorer.useParams();
  const { data } = useNode({ path });
  if (!data || !path) return null;

  // Editor is moved into a separate component to give us a state where
  // data and path is alwawys defined. This makes handling side effects
  // much easier.
  return <WorkflowEditor data={data} />;
};

export default WorkflowOverviewPage;
