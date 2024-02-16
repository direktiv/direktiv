import { Dialog, DialogContent, DialogTrigger } from "~/design/Dialog";
import { FC, useState } from "react";
import { Play, Save } from "lucide-react";
import { decode, encode } from "js-base64";

import Button from "~/design/Button";
import { CodeEditor } from "./CodeEditor";
import { Diagram } from "./Diagram";
import { EditorLayoutSwitcher } from "~/components/EditorLayoutSwitcher";
import RunWorkflow from "../components/RunWorkflow";
import { WorkspaceLayout } from "~/components/WorkspaceLayout";
import { useEditorLayout } from "~/util/store/editor";
import { useNamespace } from "~/util/store/namespace";
import { useNamespaceLinting } from "~/api/namespaceLinting/query/useNamespaceLinting";
import { useNode } from "~/api/filesTree/query/node";
import { useTranslation } from "react-i18next";
import { useUpdateFile } from "~/api/filesTree/mutate/updateFile";

export type NodeContentType = ReturnType<typeof useNode>["data"];

const WorkflowEditor: FC<{
  data: NonNullable<NodeContentType>;
}> = ({ data }) => {
  const currentLayout = useEditorLayout();
  const { t } = useTranslation();
  const namespace = useNamespace();
  const [error, setError] = useState<string | undefined>();
  const [hasUnsavedChanges, setHasUnsavedChanges] = useState(false);
  const { refetch: updateNotificationBell } = useNamespaceLinting();

  const workflowDataFromServer = decode(data?.file.data ?? "");

  const { mutate, isLoading } = useUpdateFile({
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

  const onEditorContentUpdate = (newData: string) => {
    setHasUnsavedChanges(workflowDataFromServer !== newData);
    setEditorContent(newData ?? "");
  };

  const onSave = (toSave: string | undefined) => {
    if (toSave) {
      setError(undefined);
      mutate({
        node: data.file,
        file: { data: encode(toSave) },
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
            updatedAt={data.file.updatedAt}
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
            <RunWorkflow path={data.file.path} />
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
