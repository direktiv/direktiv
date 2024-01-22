import { Dialog, DialogContent, DialogTrigger } from "~/design/Dialog";
import { FC, useEffect, useState } from "react";
import { Play, Save } from "lucide-react";

import Button from "~/design/Button";
import { CodeEditor } from "./CodeEditor";
import { Diagram } from "./Diagram";
import { EditorLayoutSwitcher } from "~/components/EditorLayoutSwitcher";
import RunWorkflow from "../components/RunWorkflow";
import { WorkspaceLayout } from "~/components/WorkspaceLayout";
import { useEditorLayout } from "~/util/store/editor";
import { useNamespace } from "~/util/store/namespace";
import { useNamespaceLinting } from "~/api/namespaceLinting/query/useNamespaceLinting";
import { useNodeContent } from "~/api/tree/query/node";
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
      <WorkspaceLayout
        layout={currentLayout}
        diagramComponent={
          <Diagram workflowData={editorContent} layout={currentLayout} />
        }
        editorComponent={
          <CodeEditor
            value={editorContent}
            onValueChange={onEditorContentUpdate}
            // TODO: may remove this
            createdAt={data.node.createdAt}
            error={error}
            hasUnsavedChanges={hasUnsavedChanges}
            onSave={onSave}
          />
        }
      />

      <div className="flex flex-col justify-end gap-4 sm:flex-row sm:items-center">
        <EditorLayoutSwitcher />
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
