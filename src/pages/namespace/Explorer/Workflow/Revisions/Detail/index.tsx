import { Card } from "~/design/Card";
import CopyButton from "~/design/CopyButton";
import Editor from "~/design/Editor";
import { Link } from "react-router-dom";
import { pages } from "~/util/router/pages";
import { useNamespace } from "~/util/store/namespace";
import { useNodeContent } from "~/api/tree/query/node";
import { useTheme } from "~/util/store/theme";

const WorkflowRevisionsPage = () => {
  const namespace = useNamespace();

  const { revision: selectedRevision, path } = pages.explorer.useParams();
  const theme = useTheme();
  const { data } = useNodeContent({
    path,
    revision: selectedRevision,
  });

  const workflowData = data?.revision?.source && atob(data?.revision?.source);

  if (!namespace) return null;
  if (!selectedRevision) return null;
  if (!workflowData) return null;

  const backLink = pages.explorer.createHref({
    namespace,
    path,
    subpage: "workflow-revisions",
  });

  return (
    <div className="flex grow flex-col space-y-4">
      <div className="flex gap-x-4">
        <Link to={backLink}>go back</Link>
        <h3 className="group flex items-center gap-x-2 font-bold text-gray-10 dark:text-gray-dark-10">
          {selectedRevision}
          <CopyButton
            value={selectedRevision}
            buttonProps={{
              variant: "outline",
              className: "hidden group-hover:inline-flex",
              size: "sm",
            }}
          >
            {(copied) => (copied ? "copied" : "copy")}
          </CopyButton>
        </h3>
      </div>
      <Card className="grow p-4">
        <Editor
          value={workflowData}
          theme={theme ?? undefined}
          options={{ readOnly: true }}
        />
      </Card>
    </div>
  );
};

export default WorkflowRevisionsPage;
