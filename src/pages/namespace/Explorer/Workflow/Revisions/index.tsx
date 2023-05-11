import { GitMerge, MoreVertical, Tag } from "lucide-react";
import {
  Table,
  TableBody,
  TableCell,
  TableRow,
} from "../../../../../design/Table";

import Badge from "../../../../../design/Badge";
import Button from "../../../../../design/Button";
import { Card } from "../../../../../design/Card";
import CopyButton from "../../../../../design/CopyButton";
import { FC } from "react";
import { Link } from "react-router-dom";
import { faker } from "@faker-js/faker";
import { pages } from "../../../../../util/router/pages";
import { useCreateTag } from "../../../../../api/tree/mutate/createTag";
import { useDeleteRevision } from "../../../../../api/tree/mutate/deleteRevision";
import { useDeleteTag } from "../../../../../api/tree/mutate/deleteTag";
import { useNodeRevisions } from "../../../../../api/tree/query/revisions";
import { useNodeTags } from "../../../../../api/tree/query/tags";

const WorkflowRevisionsPage: FC = () => {
  const { path, namespace } = pages.explorer.useParams();

  const { data: revisions } = useNodeRevisions({ path });
  const { data: tags } = useNodeTags({ path });
  const { mutate: deleteRevision } = useDeleteRevision();
  const { mutate: deleteTag } = useDeleteTag();
  const { mutate: createTag } = useCreateTag();

  if (!namespace) return null;
  if (!path) return null;

  return (
    <div className="p-5">
      <Card className="mb-4 flex gap-x-3 p-4">
        {Array.isArray(tags?.results) &&
          tags?.results?.map((x, i) => <Badge key={i}>{x.name}</Badge>)}
      </Card>
      <Card>
        <Table>
          <TableBody>
            {revisions?.results?.map((rev, i) => {
              const isTag = tags?.results?.some((tag) => tag.name === rev.name);
              const Icon = isTag ? Tag : GitMerge;
              const fakeTag = faker.word.adjective(5);
              return (
                <TableRow key={i} className="group">
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
                  <TableCell className="group flex w-auto justify-end gap-x-3">
                    <CopyButton
                      value={rev.name}
                      buttonProps={{
                        variant: "outline",
                        className: "w-24 hidden group-hover:inline-flex",
                        size: "sm",
                      }}
                    >
                      {(copied) => (copied ? "copied" : "copy")}
                    </CopyButton>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() => {
                        createTag({
                          path,
                          ref: rev.name,
                          tag: fakeTag,
                        });
                      }}
                      icon
                      className="w-[170px]"
                    >
                      Create Tag {fakeTag}
                    </Button>
                    <Button
                      variant="destructive"
                      size="sm"
                      onClick={() => {
                        if (isTag === true) {
                          deleteTag({
                            path,
                            tag: rev.name,
                          });
                        } else {
                          deleteRevision({
                            path,
                            revision: rev.name,
                          });
                        }
                      }}
                      icon
                    >
                      Delete
                    </Button>
                    <Button variant="ghost" size="sm" icon>
                      <MoreVertical />
                    </Button>
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
