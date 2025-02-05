import { FC, useState } from "react";
import { decode, encode } from "js-base64";
import { jsonToYaml, yamlToJsonOrNull } from "../../utils";
import {
  useSetUnsavedChanges,
  useUnsavedChanges,
} from "../../Workflow/store/unsavedChangesContext";

import Alert from "~/design/Alert";
import Button from "~/design/Button";
import { CodeEditor } from "../../Workflow/Edit/CodeEditor";
import { FileSchemaType } from "~/api/files/schema";
import { Form } from "react-router-dom";
import { OpenApiBaseFileFormSchema } from "./schema";
import { Save } from "lucide-react";
import { ScrollArea } from "~/design/ScrollArea";
import { useTranslation } from "react-i18next";
import { useUpdateFile } from "~/api/files/mutate/updateFile";

type BaseFileEditorProps = {
  data: NonNullable<FileSchemaType>;
};

const BaseFileEditor: FC<BaseFileEditorProps> = ({ data }) => {
  const { t } = useTranslation();
  const decodedFileContentFromServer = decode(data.data ?? "");

  const hasUnsavedChanges = useUnsavedChanges();
  const setHasUnsavedChanges = useSetUnsavedChanges();

  const [editorContent, setEditorContent] = useState(
    decodedFileContentFromServer
  );

  const [error, setError] = useState<string | undefined>();

  const { mutate: updateFile, isPending } = useUpdateFile({
    onError: (error) => {
      error && setError(error);
    },
    onSuccess: () => {
      setHasUnsavedChanges(false);
      setError(undefined);
    },
  });

  const handleFormSubmit = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    saveContent(editorContent);
  };

  const saveContent = (content: string | undefined) => {
    try {
      const parsedContent = yamlToJsonOrNull(content ?? "");
      const result = OpenApiBaseFileFormSchema.safeParse(parsedContent);
      if (!result.success) {
        setError(result.error?.issues[0]?.code ?? "Unknown error");
        return;
      }
      setError(undefined);

      updateFile({
        path: data.path,
        payload: { data: encode(jsonToYaml(parsedContent ?? {})) },
      });
    } catch (error) {
      setError(`Parsing error: ${String(error)}`);
    }
  };

  return (
    <Form onSubmit={handleFormSubmit} className="relative flex-col gap-4 p-5">
      <div className="flex flex-col gap-4">
        <div className="grid grow grid-cols-1 gap-5 lg:grid-cols-2">
          <div className="flex flex-col gap-5 grow min-h-96">
            <CodeEditor
              value={editorContent}
              onValueChange={setEditorContent}
              onSave={saveContent}
              hasUnsavedChanges={hasUnsavedChanges}
              updatedAt={data.updatedAt}
              error={error}
            />
            <div className="flex flex-col justify-end gap-4 sm:flex-row sm:items-center">
              {hasUnsavedChanges && (
                <div className="text-sm text-gray-8 dark:text-gray-dark-8">
                  <span className="text-center" data-testid="unsaved-note">
                    {t("pages.explorer.consumer.editor.unsavedNote")}
                  </span>
                </div>
              )}
              <Button
                variant={hasUnsavedChanges ? "primary" : "outline"}
                disabled={isPending}
                type="submit"
              >
                <Save />
                {t("pages.explorer.consumer.editor.saveBtn")}
              </Button>
            </div>
          </div>
          {error && (
            <div className="flex flex-col gap-5">
              <Alert variant="error">
                {t("pages.explorer.consumer.editor.form.serialisationError")}
                <ScrollArea className="size-full whitespace-nowrap">
                  <pre className="grow text-sm text-primary-500">
                    {JSON.stringify(error, null, 2)}
                  </pre>
                </ScrollArea>
              </Alert>
            </div>
          )}
        </div>
      </div>
    </Form>
  );
};

export default BaseFileEditor;
