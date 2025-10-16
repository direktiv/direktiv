import { Dialog, DialogContent, DialogTrigger } from "~/design/Dialog";
import { FC, useState } from "react";
import { FileSchemaType, SaveFileResponseSchemaType } from "~/api/files/schema";
import { Play, Save } from "lucide-react";
import { decode, encode } from "js-base64";
import { updateValidationCache, useSha1Hash } from "~/api/validate/utils";
import {
  useSetUnsavedChanges,
  useUnsavedChanges,
} from "../store/unsavedChangesContext";

import Button from "~/design/Button";
import { CodeEditor } from "./CodeEditor";
import RunWorkflow from "../components/RunWorkflow";
import { useNamespace } from "~/util/store/namespace";
import { useNotifications } from "~/api/notifications/query/get";
import { useTranslation } from "react-i18next";
import useTsWorkflowLibs from "~/hooks/useTsWorkflowLibs";
import { useUpdateFile } from "~/api/files/mutate/updateFile";
import { useValidate } from "~/api/validate/get";

const WorkflowEditor: FC<{
  data: NonNullable<FileSchemaType>;
}> = ({ data }) => {
  const { t } = useTranslation();
  const namespace = useNamespace();
  const [error, setError] = useState<string | undefined>();
  const { refetch: updateNotificationBell } = useNotifications();

  const hasUnsavedChanges = useUnsavedChanges();
  const setHasUnsavedChanges = useSetUnsavedChanges();

  const tsLibs = useTsWorkflowLibs(true);

  const workflowDataFromServer = decode(data?.data ?? "");
  const [editorContent, setEditorContent] = useState(workflowDataFromServer);
  const hash = useSha1Hash(workflowDataFromServer);
  const { data: markers } = useValidate({ hash });

  const onEditorContentUpdate = (newData: string) => {
    setHasUnsavedChanges(workflowDataFromServer !== newData);
    setEditorContent(newData ?? "");
  };

  const { mutate: updateFile, isPending } = useUpdateFile({
    onError: (error) => {
      error && setError(error);
    },
    onSuccess: async (data: SaveFileResponseSchemaType) => {
      await updateValidationCache(data);
      /**
       * updating a workflow might introduce an uninitialized secret. We need
       * to update the notification bell, to see potential new messages.
       */
      updateNotificationBell();
      setHasUnsavedChanges(false);
    },
  });

  if (!data) throw Error("data (file) is undefined in Editor");

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
      <CodeEditor
        value={editorContent}
        onValueChange={onEditorContentUpdate}
        updatedAt={data.updatedAt}
        error={error}
        hasUnsavedChanges={hasUnsavedChanges}
        onSave={onSave}
        language="typescript"
        tsLibs={tsLibs}
        markers={markers}
      />
      <div className="flex flex-col justify-end gap-4 sm:flex-row sm:items-center">
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
          disabled={isPending}
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
