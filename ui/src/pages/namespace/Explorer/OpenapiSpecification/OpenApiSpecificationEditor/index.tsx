import { FC, useState } from "react";
import { decode, encode } from "js-base64";
import { jsonToYaml, yamlToJsonOrNull } from "../../utils";
import {
  useSetUnsavedChanges,
  useUnsavedChanges,
} from "../../Workflow/store/unsavedChangesContext";

// import Alert from "~/design/Alert";
import Button from "~/design/Button";
import { CodeEditor } from "../../Workflow/Edit/CodeEditor";
import { FileSchemaType } from "~/api/files/schema";
import { Form } from "react-router-dom";
import { OpenapiSpecificationFormSchema } from "./schema";
import { Save } from "lucide-react";
import { useToast } from "~/design/Toast";
// import { ScrollArea } from "~/design/ScrollArea";
import { useTranslation } from "react-i18next";
import { useUpdateFile } from "~/api/files/mutate/updateFile";

type OpenapiSpecificationEditorProps = {
  data: NonNullable<FileSchemaType>;
};

const OpenapiSpecificationEditor: FC<OpenapiSpecificationEditorProps> = ({
  data,
}) => {
  const { t } = useTranslation();
  const { toast: showToast } = useToast();

  const decodedFileContentFromServer = decode(data.data ?? "");
  const hasUnsavedChanges = useUnsavedChanges();
  const setHasUnsavedChanges = useSetUnsavedChanges();

  const [editorContent, setEditorContent] = useState(
    decodedFileContentFromServer
  );

  // const [error, setError] = useState<string | undefined>();

  const { mutate: updateFile, isPending } = useUpdateFile({
    onError: (errorMessage) => {
      showToast({
        variant: "error",
        title: "Save Error",
        description: errorMessage || "An unknown error occurred",
      });
    },
    onSuccess: () => {
      setHasUnsavedChanges(false);

      showToast({
        variant: "success",
        title: "File Saved",
        description: "Your changes have been saved successfully.",
      });
    },
  });

  // const { mutate: updateFile, isPending } = useUpdateFile({
  //   onError: (error) => {
  //     toast.error(error);
  //   },
  //   onSuccess: () => {
  //     setHasUnsavedChanges(false);
  //     setError(undefined);
  //   },
  // });

  const handleFormSubmit = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    saveContent(editorContent);
  };

  const saveContent = (content: string | undefined) => {
    try {
      const parsedContent = yamlToJsonOrNull(content ?? "");
      const result = OpenapiSpecificationFormSchema.safeParse(parsedContent);
      if (!result.success) {
        showToast({
          variant: "error",
          title: "Parsing Error",
          description: result.error?.issues[0]?.code ?? "Unknown error",
        });
        return;
      }

      updateFile({
        path: data.path,
        payload: { data: encode(jsonToYaml(parsedContent ?? {})) },
      });
    } catch (err) {
      showToast({
        variant: "error",
        title: "Save Failed",
        description: String(err),
      });
    }
  };

  const handleEditorChange = (value: string) => {
    setEditorContent(value);
    setHasUnsavedChanges(true);
  };

  // const saveContent = (content: string | undefined) => {
  //   try {
  //     const parsedContent = yamlToJsonOrNull(content ?? "");
  //     const result = OpenApiBaseFileFormSchema.safeParse(parsedContent);
  //     if (!result.success) {
  //       setError(
  //         `Parsing error: ${result.error?.issues[0]?.code ?? "Unknown error"}`
  //       );
  //       return;
  //     }
  //     setError(undefined);

  //     updateFile({
  //       path: data.path,
  //       payload: { data: encode(jsonToYaml(parsedContent ?? {})) },
  //     });
  //   } catch (error) {
  //     setError(`Failed to save: ${String(error)}`);
  //   }
  // };

  return (
    <Form
      onSubmit={handleFormSubmit}
      className="relative flex-col gap-4 p-5 size-full"
    >
      <div className="flex flex-col gap-4 size-full">
        <div className="grid grow  size-full">
          <div className="flex flex-col gap-5 grow size-full">
            <CodeEditor
              value={editorContent}
              onValueChange={handleEditorChange}
              onSave={saveContent}
              hasUnsavedChanges={hasUnsavedChanges}
              updatedAt={data.updatedAt}
              error={undefined}
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
          {/* {error && (
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
          )} */}
        </div>
      </div>
    </Form>
  );
};

export default OpenapiSpecificationEditor;
