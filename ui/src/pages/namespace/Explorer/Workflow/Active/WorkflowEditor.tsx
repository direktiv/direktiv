import { Dialog, DialogContent, DialogTrigger } from "~/design/Dialog";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "~/design/Dropdown";
import { FC, useEffect, useState } from "react";
import { GitBranchPlus, Play, Save, Tag, Undo } from "lucide-react";

import Button from "~/design/Button";
import { ButtonBar } from "~/design/ButtonBar";
import { CodeEditor } from "./CodeEditor";
import { Diagram } from "./Diagram";
import { EditorLayoutSwitcher } from "~/components/EditorLayoutSwitcher";
import RunWorkflow from "../components/RunWorkflow";
import { RxChevronDown } from "react-icons/rx";
import { WorkspaceLayout } from "~/components/WorkspaceLayout";
import { pages } from "~/util/router/pages";
import { useCreateRevision } from "~/api/tree/mutate/createRevision";
import { useEditorLayout } from "~/util/store/editor";
import { useNamespace } from "~/util/store/namespace";
import { useNamespaceLinting } from "~/api/namespaceLinting/query/useNamespaceLinting";
import { useNodeContent } from "~/api/tree/query/node";
import { useRevertRevision } from "~/api/tree/mutate/revertRevision";
import { useTranslation } from "react-i18next";
import { useUpdateWorkflow } from "~/api/tree/mutate/updateWorkflow";

export type NodeContentType = ReturnType<typeof useNodeContent>["data"];

const WorkflowEditor: FC<{
  data: NonNullable<NodeContentType>;
  path: string;
}> = ({ data, path }) => {
  const currentLayout = useEditorLayout();
  const { t } = useTranslation();
  const namespace = useNamespace();
  const [error, setError] = useState<string | undefined>();
  const [hasUnsavedChanges, setHasUnsavedChanges] = useState(false);
  const { refetch: updateNotificationBell } = useNamespaceLinting();

  const workflowDataFromServer = atob(data?.source ?? "");

  const { mutate: updateWorkflow, isLoading } = useUpdateWorkflow({
    onError: (error) => {
      error && setError(error);
    },
    onSuccess: () => {
      /**
       * updating a workflow might introduce an uninitialized secret. We need
       * to update the notification bell, to see potential new messages.
       */
      updateNotificationBell();
      setHasUnsavedChanges(false);
    },
  });

  const [editorContent, setEditorContent] = useState(workflowDataFromServer);

  /**
   * When the server state of the content changes, the internal state needs to be updated,
   * to have the editor and diagram up to date. This is important, when the user is reverting
   * to an old revision.
   */
  useEffect(() => {
    setEditorContent(workflowDataFromServer);
  }, [workflowDataFromServer]);

  const { mutate: createRevision } = useCreateRevision();
  const { mutate: revertRevision } = useRevertRevision({
    onSuccess: () => {
      setHasUnsavedChanges(false);
    },
  });

  const onEditorContentUpdate = (newData: string) => {
    setHasUnsavedChanges(workflowDataFromServer !== newData);
    setEditorContent(newData ?? "");
  };

  const onSave = (toSave: string | undefined) => {
    if (toSave) {
      setError(undefined);
      updateWorkflow({
        path,
        fileContent: toSave,
      });
    }
  };

  if (!namespace) return null;

  return (
    <div className="relative flex grow flex-col space-y-4 p-5">
      <h3 className="flex items-center gap-x-2 font-bold">
        <Tag className="h-5" />
        {t("pages.explorer.workflow.headline")}
      </h3>
      <WorkspaceLayout
        layout={currentLayout}
        diagramComponent={
          <Diagram workflowData={editorContent} layout={currentLayout} />
        }
        editorComponent={
          <CodeEditor
            value={editorContent}
            onValueChange={onEditorContentUpdate}
            createdAt={data.revision?.createdAt}
            error={error}
            hasUnsavedChanges={hasUnsavedChanges}
            onSave={onSave}
          />
        }
      />

      <div className="flex flex-col justify-end gap-4 sm:flex-row sm:items-center">
        <EditorLayoutSwitcher />
        <DropdownMenu>
          <ButtonBar>
            <Button
              variant="outline"
              disabled={hasUnsavedChanges}
              onClick={() => {
                createRevision({
                  path,
                  createLink: (revision) =>
                    pages.explorer.createHref({
                      namespace,
                      path,
                      subpage: "workflow-revisions",
                      revision,
                    }),
                });
              }}
              className="grow"
              data-testid="workflow-editor-btn-make-revision"
            >
              <GitBranchPlus />
              {t("pages.explorer.workflow.editor.makeRevision")}
            </Button>
            <DropdownMenuTrigger asChild>
              <Button
                disabled={hasUnsavedChanges}
                variant="outline"
                data-testid="workflow-editor-btn-revision-drop"
              >
                <RxChevronDown />
              </Button>
            </DropdownMenuTrigger>
          </ButtonBar>
          <DropdownMenuContent className="w-60">
            <DropdownMenuItem
              onClick={() => {
                revertRevision({
                  path,
                });
              }}
              data-testid="workflow-editor-btn-revert-revision"
            >
              <Undo className="mr-2 h-4 w-4" />
              {t("pages.explorer.workflow.editor.revertToPrevious")}
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
        <Dialog>
          <DialogTrigger asChild>
            <Button
              variant="outline"
              data-testid="workflow-editor-btn-run"
              disabled={hasUnsavedChanges}
            >
              <Play />
              {t("pages.explorer.workflow.editor.runBtn")}
            </Button>
          </DialogTrigger>
          <DialogContent className="sm:max-w-2xl">
            <RunWorkflow path={path} />
          </DialogContent>
        </Dialog>
        <Button
          variant="outline"
          disabled={isLoading}
          onClick={() => {
            onSave(editorContent);
          }}
          data-testid="workflow-editor-btn-save"
        >
          <Save />
          {t("pages.explorer.workflow.editor.saveBtn")}
        </Button>
      </div>
    </div>
  );
};

export default WorkflowEditor;
