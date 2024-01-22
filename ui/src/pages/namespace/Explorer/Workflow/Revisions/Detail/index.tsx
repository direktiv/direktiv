import { ArrowLeft, GitMerge, Tag, Undo } from "lucide-react";
import { Dialog, DialogContent, DialogTrigger } from "~/design/Dialog";
import { Popover, PopoverContent, PopoverTrigger } from "~/design/Popover";

import Alert from "~/design/Alert";
import Button from "~/design/Button";
import { Card } from "~/design/Card";
import CopyButton from "~/design/CopyButton";
import { Diagram } from "../../Active/Diagram";
import Editor from "~/design/Editor";
import { EditorLayoutSwitcher } from "~/components/EditorLayoutSwitcher";
import { Link } from "react-router-dom";
import Revert from "../components/Revert";
import { WorkspaceLayout } from "../../../../../../components/WorkspaceLayout";
import { pages } from "~/util/router/pages";
import { useEditorLayout } from "~/util/store/editor";
import { useNamespace } from "~/util/store/namespace";
import { useNodeContent } from "~/api/tree/query/node";
import { useNodeTags } from "~/api/tree/query/tags";
import { useState } from "react";
import { useTheme } from "~/util/store/theme";
import { useTranslation } from "react-i18next";

const WorkflowRevisionsPage = () => {
  const namespace = useNamespace();
  const { t } = useTranslation();
  const theme = useTheme();
  const [dialogOpen, setDialogOpen] = useState(false);
  const currentLayout = useEditorLayout();

  const { revision: selectedRevision, path } = pages.explorer.useParams();
  const { data } = useNodeContent({ path, revision: selectedRevision });
  const { data: tags } = useNodeTags({ path });

  const workflowData = atob(data?.source ?? "");
  const isTag =
    tags?.results?.some((tag) => tag.name === selectedRevision) ?? false;
  const Icon = isTag ? Tag : GitMerge;
  const isLatest = selectedRevision === "latest";

  if (!path || !namespace || !selectedRevision || !workflowData) return null;

  return (
    <div className="flex grow flex-col space-y-4">
      <div className="flex gap-x-4">
        <h3
          className="group flex grow items-center gap-x-2 font-bold"
          data-testid="revisions-detail-title"
        >
          <Icon aria-hidden="true" className="h-5" />
          {selectedRevision}
          <CopyButton
            value={selectedRevision}
            buttonProps={{
              variant: "outline",
              className: "hidden group-hover:inline-flex",
              size: "sm",
            }}
          />
        </h3>
      </div>

      <WorkspaceLayout
        layout={currentLayout}
        diagramComponent={
          <Diagram workflowData={workflowData} layout={currentLayout} />
        }
        editorComponent={
          <Popover>
            <PopoverTrigger asChild>
              <Card className="grow p-4" data-testid="revisions-detail-editor">
                <Editor
                  value={workflowData}
                  theme={theme ?? undefined}
                  options={{ readOnly: true }}
                />
              </Card>
            </PopoverTrigger>
            <PopoverContent asChild>
              <Alert variant="info" className="min-w-max">
                {t(
                  "pages.explorer.tree.workflow.revisions.overview.detail.readOnlyNote"
                )}
              </Alert>
            </PopoverContent>
          </Popover>
        }
      />

      <div className="flex flex-col justify-end gap-4 sm:flex-row sm:items-center">
        <EditorLayoutSwitcher />
        <Button asChild isAnchor variant="outline">
          <Link
            data-testid="revisions-detail-back-link"
            to={pages.explorer.createHref({
              namespace,
              path,
              subpage: "workflow-revisions",
            })}
          >
            <ArrowLeft />
            {t(
              "pages.explorer.tree.workflow.revisions.overview.detail.backBtn"
            )}
          </Link>
        </Button>
        <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
          <DialogTrigger asChild>
            {!isLatest && (
              <Button
                variant="outline"
                data-testid="revisions-detail-revert-btn"
              >
                <Undo />
                {t(
                  "pages.explorer.tree.workflow.revisions.overview.detail.revertBtn"
                )}
              </Button>
            )}
          </DialogTrigger>
          <DialogContent>
            <Revert
              path={path}
              revision={{ name: selectedRevision }}
              close={() => {
                setDialogOpen(false);
              }}
            />
          </DialogContent>
        </Dialog>
      </div>
    </div>
  );
};

export default WorkflowRevisionsPage;
