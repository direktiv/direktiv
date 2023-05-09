import { GitMerge, Tag } from "lucide-react";
import {
  Table,
  TableBody,
  TableCell,
  TableRow,
} from "../../../../../design/Table";

import { Card } from "../../../../../design/Card";
import CopyButton from "../../../../../design/CopyButton";
import { FC } from "react";
import { Link } from "react-router-dom";
import { pages } from "../../../../../util/router/pages";
// import { useNodeContent } from "../../../../../api/tree/query/get";
import { useNodeRevisions } from "../../../../../api/tree/query/revisions";

const WorkflowRevisionsPage: FC = () => {
  const { path, namespace } = pages.explorer.useParams();

  // const { data } = useNodeContent({
  //   path,
  //   revision,
  // });

  const { data: revisions } = useNodeRevisions({ path });

  if (!namespace) return null;

  return (
    <div className="p-5">
      <Card>
        <Table>
          <TableBody>
            {revisions?.results?.map((rev, i) => {
              const isTag = Math.random() > 0.5; // TODO: figure out if this is a tag
              const Icon = isTag ? GitMerge : Tag;
              return (
                <TableRow key={i}>
                  <TableCell>
                    <div className="flex space-x-3">
                      <Icon aria-hidden="true" className="h-5" />

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
                    </div>
                  </TableCell>
                  <TableCell className="w-0">
                    <CopyButton value={rev.name} buttonProps={{ size: "sm" }} />
                  </TableCell>
                </TableRow>
              );
            })}
          </TableBody>
        </Table>
      </Card>
    </div>
  );
};

export default WorkflowRevisionsPage;
