import { pages } from "~/util/router/pages";

const WorkflowRevisionsPage = () => {
  const { revision: selectedRevision } = pages.explorer.useParams();
  return <>{selectedRevision}</>;
};

export default WorkflowRevisionsPage;
