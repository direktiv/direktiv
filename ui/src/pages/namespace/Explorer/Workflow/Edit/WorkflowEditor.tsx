import { Dialog, DialogContent, DialogTrigger } from "~/design/Dialog";
import { FC, useState } from "react";
import { Play, Save } from "lucide-react";
import { decode, encode } from "js-base64";
import {
  useSetUnsavedChanges,
  useUnsavedChanges,
} from "../store/unsavedChangesContext";

import Button from "~/design/Button";
import { CodeEditor } from "./CodeEditor";
import { Diagram } from "./Diagram";
import { EditorLayoutSwitcher } from "~/components/EditorLayoutSwitcher";
import { FileSchemaType } from "~/api/files/schema";
import RunWorkflow from "../components/RunWorkflow";
import { WorkspaceLayout } from "~/components/WorkspaceLayout";
import { useEditorLayout } from "~/util/store/editor";
import { useNamespace } from "~/util/store/namespace";
import { useNotifications } from "~/api/notifications/query/get";
import { useTranslation } from "react-i18next";
import { useUpdateFile } from "~/api/files/mutate/updateFile";

const WorkflowEditor: FC<{
  data: NonNullable<FileSchemaType>;
}> = ({ data }) => {
  const currentLayout = useEditorLayout();
  const { t } = useTranslation();
  const namespace = useNamespace();
  const [error, setError] = useState<string | undefined>();
  const { refetch: updateNotificationBell } = useNotifications();

  const hasUnsavedChanges = useUnsavedChanges();
  const setHasUnsavedChanges = useSetUnsavedChanges();

  const workflowDataFromServer = decode(data?.data ?? "");

  const { mutate: updateFile, isLoading } = useUpdateFile({
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
      updateFile({
        path: data.path,
        payload: { data: encode(toSave) },
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
            updatedAt={data.updatedAt}
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
            <RunWorkflow path={data.path} />
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
