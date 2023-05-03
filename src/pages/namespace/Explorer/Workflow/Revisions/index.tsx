import Badge from "../../../../../design/Badge";
import { Card } from "../../../../../design/Card";
import { FC } from "react";
import { Link } from "react-router-dom";
import { pages } from "../../../../../util/router/pages";
import { useNodeContent } from "../../../../../api/tree/query/get";
import { useNodeRevisions } from "../../../../../api/tree/query/revisions";

const WorkflowRevisionsPage: FC = () => {
  const { path, namespace, revision } = pages.explorer.useParams();
  const { data } = useNodeContent({
    path,
    revision,
  });

  const { data: revisions } = useNodeRevisions({ path });

  if (!namespace) return null;

  return (
    <div className="flex flex-col space-y-4 p-4">
      <h1>WorkflowRevisionsPage</h1>
      <Card className="p-4">
        <Badge>{data?.revision?.hash.slice(0, 8)}</Badge>
      </Card>
      <Card className="space-x-2 space-y-2 p-4">
        {revisions?.results?.map((rev, i) => (
          <Badge
            variant={revision === rev.name ? undefined : "outline"}
            key={i}
          >
            <Link
              to={pages.explorer.createHref({
                namespace,
                path,
                subpage: "workflow-revisions",
                revision: rev.name,
              })}
            >
              {rev.name}
            </Link>
          </Badge>
        ))}
      </Card>
      <Card className="p-4">
        {data?.revision?.source && atob(data?.revision?.source)}
      </Card>
    </div>
  );
};

export default WorkflowRevisionsPage;
