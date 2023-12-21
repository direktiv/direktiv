import Alert from "~/design/Alert";
import Button from "~/design/Button";
import { Card } from "~/design/Card";
import Editor from "~/design/Editor";
import { EndpointFormSchemaType } from "./schema";
import { FC } from "react";
import { Form } from "./Form";
import FormErrors from "~/componentsNext/FormErrors";
import { RouteSchemeType } from "~/api/gateway/schema";
import { Save } from "lucide-react";
import { ScrollArea } from "~/design/ScrollArea";
import { serializeEndpointFile } from "./utils";
import { stringify } from "json-to-pretty-yaml";
import { useNodeContent } from "~/api/tree/query/node";
import { useTheme } from "~/util/store/theme";
import { useTranslation } from "react-i18next";
import { useUpdateWorkflow } from "~/api/tree/mutate/updateWorkflow";

type NodeContentType = ReturnType<typeof useNodeContent>["data"];

type EndpointEditorProps = {
  path: string;
  data: NonNullable<NodeContentType>;
  route?: RouteSchemeType;
};

const EndpointEditor: FC<EndpointEditorProps> = ({ data, path }) => {
  const { t } = useTranslation();
  const theme = useTheme();
  const endpointFileContent = atob(data.revision?.source ?? "");
  const [endpointConfig, endpointConfigError] =
    serializeEndpointFile(endpointFileContent);
  const { mutate: updateRoute, isLoading } = useUpdateWorkflow();

  const save = (data: EndpointFormSchemaType) => {
    const toSave = stringify(data);
    updateRoute({
      path,
      fileContent: toSave,
    });
  };

  return (
    <Form defaultConfig={endpointConfig}>
      {({
        formControls: {
          formState: { errors },
          handleSubmit,
        },
        formMarkup,
        values,
      }) => {
        const preview = stringify(values);
        const isDirty = preview !== endpointFileContent;

        return (
          <form
            onSubmit={handleSubmit(save)}
            className="relative flex grow flex-col space-y-4 p-5"
          >
            <div className="flex grow">
              <div className="grid grow grid-cols-1 gap-5 lg:grid-cols-2">
                <Card className="p-5 lg:h-[calc(100vh-15rem)] lg:overflow-y-scroll">
                  {!endpointConfig ? (
                    <div className="flex flex-col gap-5">
                      <Alert variant="error">
                        {t(
                          "pages.explorer.endpoint.editor.form.serialisationError"
                        )}
                      </Alert>
                      <ScrollArea className="h-full w-full whitespace-nowrap">
                        <pre className="grow text-sm text-primary-500">
                          {JSON.stringify(endpointConfigError, null, 2)}
                        </pre>
                      </ScrollArea>
                    </div>
                  ) : (
                    <div>
                      <FormErrors errors={errors} className="mb-5" />
                      {formMarkup}
                    </div>
                  )}
                </Card>
                <Card className="flex grow p-4 max-lg:h-[500px]">
                  <Editor
                    value={preview}
                    theme={theme ?? undefined}
                    options={{
                      readOnly: true,
                    }}
                  />
                </Card>
              </div>
            </div>
            <div className="flex flex-col justify-end gap-4 sm:flex-row sm:items-center">
              {isDirty && (
                <div className="text-sm text-gray-8 dark:text-gray-dark-8">
                  <span className="text-center">
                    {t("pages.explorer.workflow.editor.unsavedNote")}
                  </span>
                </div>
              )}
              <Button
                variant={isDirty ? "primary" : "outline"}
                disabled={isLoading}
                type="submit"
              >
                <Save />
                {t("pages.explorer.endpoint.editor.saveBtn")}
              </Button>
            </div>
          </form>
        );
      }}
    </Form>
  );
};

export default EndpointEditor;
