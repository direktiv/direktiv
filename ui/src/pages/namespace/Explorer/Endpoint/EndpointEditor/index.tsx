import { compareYamlStructure, jsonToYaml } from "../../utils";
import { decode, encode } from "js-base64";

import Alert from "~/design/Alert";
import Button from "~/design/Button";
import { Card } from "~/design/Card";
import Editor from "~/design/Editor";
import { EndpointFormSchemaType } from "./schema";
import { FC } from "react";
import { Form } from "./Form";
import FormErrors from "~/components/FormErrors";
import { RouteSchemaType } from "~/api/gateway/schema";
import { Save } from "lucide-react";
import { ScrollArea } from "~/design/ScrollArea";
import { serializeEndpointFile } from "./utils";
import { useNode } from "~/api/filesTree/query/node";
import { useTheme } from "~/util/store/theme";
import { useTranslation } from "react-i18next";
import { useUpdateFile } from "~/api/filesTree/mutate/updateFile";

type NodeContentType = ReturnType<typeof useNode>["data"];

type EndpointEditorProps = {
  data: NonNullable<NodeContentType>;
  route?: RouteSchemaType;
};

const EndpointEditor: FC<EndpointEditorProps> = ({ data }) => {
  const { t } = useTranslation();
  const theme = useTheme();
  const fileContentFromServer = decode(data.file.data ?? "");
  const [endpointConfig, endpointConfigError] = serializeEndpointFile(
    fileContentFromServer
  );
  const { mutate: updateRoute, isLoading } = useUpdateFile();

  const save = (value: EndpointFormSchemaType) => {
    const toSave = jsonToYaml(value);
    updateRoute({
      node: data.file,
      file: { data: encode(toSave) },
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
        const preview = jsonToYaml(values);
        const filehasChanged = compareYamlStructure(
          preview,
          fileContentFromServer
        );
        const isDirty = !endpointConfigError && !filehasChanged;
        const disableButton = isLoading || !!endpointConfigError;

        return (
          <form
            onSubmit={handleSubmit(save)}
            className="relative flex-col gap-4 p-5"
          >
            <div className="flex flex-col gap-4">
              <div className="grid grow grid-cols-1 gap-5 lg:grid-cols-2">
                <Card className="p-5 lg:h-[calc(100vh-15.5rem)] lg:overflow-y-scroll">
                  {endpointConfigError ? (
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
              <div className="flex flex-col justify-end gap-4 sm:flex-row sm:items-center">
                {isDirty && (
                  <div className="text-sm text-gray-8 dark:text-gray-dark-8">
                    <span className="text-center">
                      {t("pages.explorer.endpoint.editor.unsavedNote")}
                    </span>
                  </div>
                )}
                <Button
                  variant={isDirty ? "primary" : "outline"}
                  disabled={disableButton}
                  type="submit"
                >
                  <Save />
                  {t("pages.explorer.endpoint.editor.saveBtn")}
                </Button>
              </div>
            </div>
          </form>
        );
      }}
    </Form>
  );
};

export default EndpointEditor;
