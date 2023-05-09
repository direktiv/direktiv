import { Copy, CopyCheck, GitMerge, Tag } from "lucide-react";
import { FC, useEffect, useState } from "react";
import {
  Table,
  TableBody,
  TableCell,
  TableRow,
} from "../../../../../design/Table";

import Button from "../../../../../design/Button";
import { Card } from "../../../../../design/Card";
import { Link } from "react-router-dom";
import { pages } from "../../../../../util/router/pages";
// import { useNodeContent } from "../../../../../api/tree/query/get";
import { useNodeRevisions } from "../../../../../api/tree/query/revisions";

const CopyButton: FC<{ value: string }> = ({ value }) => {
  const [copied, setCopied] = useState(false);
  useEffect(() => {
    let timeout: NodeJS.Timeout;
    if (copied === true) {
      timeout = setTimeout(() => {
        setCopied(false);
      }, 1000);
    }
    return () => clearTimeout(timeout);
  }, [copied]);

  return (
    <Button
      size="sm"
      variant="ghost"
      onClick={() => {
        navigator.clipboard.writeText(value);
        setCopied(true);
      }}
    >
      {copied ? (
        <CopyCheck className="text-success-10 dark:text-success-dark-10" />
      ) : (
        <Copy />
      )}
    </Button>
  );
};

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
                    <CopyButton value={rev.name} />
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
