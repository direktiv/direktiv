import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "~/design/Dropdown";
import { GitMerge, MoreVertical, Tag, Trash, Undo } from "lucide-react";
import { TableCell, TableRow } from "~/design/Table";

import Badge from "~/design/Badge";
import Button from "~/design/Button";
import CopyButton from "~/design/CopyButton";
import { DialogTrigger } from "~/design/Dialog";
import { FC } from "react";
import { Link } from "react-router-dom";
import { TrimmedRevisionSchemaType } from "~/api/tree/schema/node";
import { pages } from "~/util/router/pages";
import { useRouter } from "~/api/tree/query/router";
import { useTranslation } from "react-i18next";

const RevisionTableRow: FC<{
  revision: TrimmedRevisionSchemaType;
  isTag: boolean;
  isPrimaryTraffic: boolean; // traffic shaping is configured for this revision (left side of the traffic split)
  isSecondaryTraffic: boolean; // traffic shaping is configured for this revision (right side of the traffic split)
  onDeleteRevClicked: (revision: TrimmedRevisionSchemaType) => void;
  onDeleteTagCLicked: (revision: TrimmedRevisionSchemaType) => void;
  onRevertClicked: (revision: TrimmedRevisionSchemaType) => void;
  onCreateTagClicked: (revision: TrimmedRevisionSchemaType) => void;
}> = ({
  revision,
  isTag,
  isPrimaryTraffic,
  isSecondaryTraffic,
  onDeleteRevClicked,
  onDeleteTagCLicked,
  onRevertClicked,
  onCreateTagClicked,
}) => {
  const { t } = useTranslation();
  const { path, namespace } = pages.explorer.useParams();
  const isLatest = revision.name === "latest";
  const { data: router } = useRouter({ path });
  const Icon = isTag ? Tag : GitMerge;
  if (!namespace) return null;
  return (
    <TableRow className="group" data-testid={`revisions-list-${revision.name}`}>
      <TableCell className="w-0">
        <div className="flex space-x-3">
          <Icon aria-hidden="true" className="h-5" />
          <Link
            to={pages.explorer.createHref({
              namespace,
              path,
              subpage: "workflow-revisions",
              revision: revision.name,
            })}
            data-testid={`workflow-revisions-link-item-${revision.name}`}
          >
            {revision.name}
          </Link>
        </div>
      </TableCell>
      <TableCell className="w-0 justify-start gap-x-3">
        {isPrimaryTraffic && (
          <Badge data-testid="traffic-distribution-primary">
            {t(
              "pages.explorer.tree.workflow.revisions.overview.list.distribution",
              {
                count: router?.routes?.[0]?.weight,
              }
            )}
          </Badge>
        )}
        {isSecondaryTraffic && (
          <Badge data-testid="traffic-distribution-secondary" variant="outline">
            {t(
              "pages.explorer.tree.workflow.revisions.overview.list.distribution",
              {
                count: router?.routes?.[1]?.weight,
              }
            )}
          </Badge>
        )}
      </TableCell>
      <TableCell
        className="flex w-auto justify-end gap-x-3"
        data-testid={`workflow-revisions-item-last-row-${revision.name}`}
      >
        {!isLatest && (
          <CopyButton
            value={revision.name}
            buttonProps={{
              variant: "outline",
              className: "hidden group-hover:inline-flex",
              size: "sm",
            }}
          />
        )}
        {!isLatest && (
          <DropdownMenu>
            <DropdownMenuTrigger
              data-testid={`workflow-revisions-item-menu-trg-${revision.name}`}
              asChild
            >
              <Button variant="ghost" size="sm" icon>
                <MoreVertical />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent
              className="w-44"
              data-testid={`workflow-revisions-item-menu-content-${revision.name}`}
            >
              <DropdownMenuLabel>
                {t(
                  "pages.explorer.tree.workflow.revisions.overview.list.contextMenu.title"
                )}
              </DropdownMenuLabel>
              <DropdownMenuSeparator />
              <DialogTrigger
                className="w-full"
                onClick={() => {
                  if (isTag) {
                    onDeleteTagCLicked(revision);
                  } else {
                    onDeleteRevClicked(revision);
                  }
                }}
                data-testid={`workflow-revisions-trg-delete-dlg-${revision.name}`}
              >
                <DropdownMenuItem>
                  <Trash className="mr-2 h-4 w-4" />
                  {t(
                    "pages.explorer.tree.workflow.revisions.overview.list.contextMenu.delete"
                  )}
                </DropdownMenuItem>
              </DialogTrigger>
              <DialogTrigger
                className="w-full"
                onClick={() => {
                  onCreateTagClicked(revision);
                }}
                data-testid={`workflow-revisions-trg-create-tag-dlg-${revision.name}`}
              >
                <DropdownMenuItem>
                  <Tag className="mr-2 h-4 w-4" />
                  {t(
                    "pages.explorer.tree.workflow.revisions.overview.list.contextMenu.tag"
                  )}
                </DropdownMenuItem>
              </DialogTrigger>
              <DialogTrigger
                className="w-full"
                onClick={() => {
                  onRevertClicked(revision);
                }}
                data-testid={`workflow-revisions-trg-revert-dlg-${revision.name}`}
              >
                <DropdownMenuItem>
                  <Undo className="mr-2 h-4 w-4" />
                  {t(
                    "pages.explorer.tree.workflow.revisions.overview.list.contextMenu.revert"
                  )}
                </DropdownMenuItem>
              </DialogTrigger>
            </DropdownMenuContent>
          </DropdownMenu>
        )}
      </TableCell>
    </TableRow>
  );
};

export default RevisionTableRow;
