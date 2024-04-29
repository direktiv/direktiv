import { decode, encode } from "js-base64";

import Alert from "~/design/Alert";
import Button from "~/design/Button";
import { Card } from "~/design/Card";
import Editor from "~/design/Editor";
import { FC } from "react";
import { FileSchemaType } from "~/api/files/schema";
import { Form } from "./Form";
import FormErrors from "~/components/FormErrors";
import NavigationBlocker from "~/components/NavigationBlocker";
import { Save } from "lucide-react";
import { ScrollArea } from "~/design/ScrollArea";
import { ServiceFormSchemaType } from "./schema";
import { jsonToYaml } from "../../utils";
import { serializeServiceFile } from "./utils";
import { useTheme } from "~/util/store/theme";
import { useTranslation } from "react-i18next";
import { useUpdateFile } from "~/api/files/mutate/updateFile";

type ServiceEditorProps = {
  data: NonNullable<FileSchemaType>;
};

const ServiceEditor: FC<ServiceEditorProps> = ({ data }) => {
  const { t } = useTranslation();
  const theme = useTheme();

  const fileContentFromServer = decode(data.data ?? "");

  const [serviceConfig, serviceConfigError] = serializeServiceFile(
    fileContentFromServer
  );

  const { mutate: updateService, isPending } = useUpdateFile({});

  const save = (value: ServiceFormSchemaType) => {
    const toSave = jsonToYaml(value);
    updateService({
      path: data.path,
      payload: { data: encode(toSave) },
    });
  };

  return (
    <Form defaultConfig={serviceConfig} onSave={save}>
      {({
        formControls: {
          formState: { errors },
          handleSubmit,
        },
        formMarkup,
        values,
      }) => {
        const preview = jsonToYaml(values);

        const filehasChanged = preview === fileContentFromServer;
        const isDirty = !serviceConfigError && !filehasChanged;
        const disableButton = isPending || !!serviceConfigError;

        return (
          <form
            onSubmit={handleSubmit(save)}
            className="relative flex-col gap-4 p-5"
          >
            {isDirty && <NavigationBlocker />}
            <div className="flex flex-col gap-4">
              <div className="grid grow grid-cols-1 gap-5 lg:grid-cols-2">
                <Card className="p-5 lg:h-[calc(100vh-15.5rem)] lg:overflow-y-scroll">
                  {serviceConfigError ? (
                    <div className="flex flex-col gap-5">
                      <Alert variant="error">
                        {t(
                          "pages.explorer.service.editor.form.serialisationError"
                        )}
                      </Alert>
                      <ScrollArea className="h-full w-full whitespace-nowrap">
                        <pre className="grow text-sm text-primary-500">
                          {JSON.stringify(serviceConfigError, null, 2)}
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
                    <span className="text-center" data-testid="unsaved-note">
                      {t("pages.explorer.service.editor.unsavedNote")}
                    </span>
                  </div>
                )}
                <Button
                  variant={isDirty ? "primary" : "outline"}
                  disabled={disableButton}
                  type="submit"
                >
                  <Save />
                  {t("pages.explorer.service.editor.saveBtn")}
                </Button>
              </div>
            </div>
          </form>
        );
      }}
    </Form>
  );
};

export default ServiceEditor;
