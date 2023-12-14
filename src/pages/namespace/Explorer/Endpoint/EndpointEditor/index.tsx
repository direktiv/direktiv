import { EndpointFormSchemaType, serializeEndpointFile } from "../utils";

import Alert from "~/design/Alert";
import Button from "~/design/Button";
import { Card } from "~/design/Card";
import Editor from "~/design/Editor";
import EndpointPreview from "../EndpointPreview";
import { FC } from "react";
import { Form } from "./Form";
import FormErrors from "~/componentsNext/FormErrors";
import { RouteSchemeType } from "~/api/gateway/schema";
import { Save } from "lucide-react";
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
  const endpointConfig = serializeEndpointFile(endpointFileContent);
  const { mutate: updateRoute, isLoading } = useUpdateWorkflow();

  const save = (data: EndpointFormSchemaType) => {
    const toSave = stringify(data);
    updateRoute({
      path,
      fileContent: toSave,
    });
  };

  return (
    <Form endpointConfig={endpointConfig}>
      {({
        formControls: {
          formState: { isDirty, errors },
          handleSubmit,
        },
        formMarkup,
        values,
      }) => {
        const preview = stringify(values);
        return (
          <form
            onSubmit={handleSubmit(save)}
            className="relative flex grow flex-col space-y-4 p-5"
          >
            <Card className="flex grow flex-col p-4">
              <div className="grow">
                <div className="grid grow grid-cols-2 gap-3">
                  {!endpointConfig ? (
                    <Alert variant="error">
                      {t(
                        "pages.explorer.endpoint.editor.form.serialisationError"
                      )}
                    </Alert>
                  ) : (
                    <div>
                      <FormErrors errors={errors} className="mb-5" />
                      {formMarkup}
                    </div>
                  )}
                  <div className="grid h-full grid-rows-2 gap-3">
                    <Card className="h-[400px] p-5">
                      <Editor
                        value={endpointFileContent}
                        theme={theme ?? undefined}
                        options={{
                          readOnly: true,
                        }}
                      />
                    </Card>
                    <Card className="h-[400px] p-5">
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
              </div>
              <div className="flex justify-end gap-2 pt-2 text-sm text-gray-8 dark:text-gray-dark-8">
                {isDirty && (
                  <span className="text-center">
                    {t("pages.explorer.workflow.editor.unsavedNote")}
                  </span>
                )}
              </div>
            </Card>
            <div className="flex flex-col justify-end gap-4 sm:flex-row sm:items-center">
              <EndpointPreview fileContent={stringify(endpointConfig)} />
              <Button variant="outline" disabled={isLoading} type="submit">
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
