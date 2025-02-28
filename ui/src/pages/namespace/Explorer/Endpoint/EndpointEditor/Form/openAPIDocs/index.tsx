import { Dialog, DialogTrigger } from "~/design/Dialog";
import { FC, useState } from "react";
import { UseFormReturn, useForm } from "react-hook-form";
import { jsonToYaml, yamlToJsonOrNull } from "~/pages/namespace/Explorer/utils";

import Button from "~/design/Button";
import { Card } from "~/design/Card";
import Editor from "~/design/Editor";
import { EndpointFormSchemaType } from "../../schema";
import { MethodsSchema } from "~/api/gateway/schema";
import { ModalWrapper } from "~/components/ModalWrapper";
import { ScrollText } from "lucide-react";
import { useTheme } from "~/util/store/theme";
import { useTranslation } from "react-i18next";
import { z } from "zod";
import { zodResolver } from "@hookform/resolvers/zod";

type OpenAPIDocsFormProps = {
  form: UseFormReturn<EndpointFormSchemaType>;
  onSave: (value: EndpointFormSchemaType) => void;
};

const FormSchema = z.object({
  editor: MethodsSchema,
});

type FormSchemaType = z.infer<typeof FormSchema>;

export const OpenAPIDocsForm: FC<OpenAPIDocsFormProps> = ({ form, onSave }) => {
  const theme = useTheme();
  const { t } = useTranslation();
  const { handleSubmit: handleParentSubmit, watch: getParentValues } = form;

  const {
    handleSubmit,
    getValues,
    setValue,
    reset,
    watch,
    formState: { errors },
  } = useForm<FormSchemaType>({
    resolver: zodResolver(FormSchema),
    defaultValues: {
      editor: {
        connect: getParentValues("connect"),
        delete: getParentValues("delete"),
        get: getParentValues("get"),
        head: getParentValues("head"),
        options: getParentValues("options"),
        patch: getParentValues("patch"),
        post: getParentValues("post"),
        put: getParentValues("put"),
        trace: getParentValues("trace"),
      },
    },
  });

  const [dialogOpen, setDialogOpen] = useState(false);

  const formId = "openAPIDocsForm";

  const onSubmit = (configuration: FormSchemaType) => {
    setDialogOpen(false);
    form.setValue("connect", configuration.editor.connect);
    form.setValue("delete", configuration.editor.delete);
    form.setValue("get", configuration.editor.get);
    form.setValue("head", configuration.editor.head);
    form.setValue("options", configuration.editor.options);
    form.setValue("patch", configuration.editor.patch);
    form.setValue("post", configuration.editor.post);
    form.setValue("put", configuration.editor.put);
    form.setValue("trace", configuration.editor.trace);
    handleParentSubmit(onSave)();
  };

  return (
    <Dialog
      open={dialogOpen}
      onOpenChange={(isOpen) => {
        // TODO: check what happens when user presses cancel, changes the  methods and reopens the dialog
        if (isOpen === false) {
          // TODO: is this needed?
          reset();
        }
        setDialogOpen(isOpen);
      }}
    >
      <DialogTrigger asChild>
        <Button icon variant="outline">
          <ScrollText />
          {t("pages.explorer.endpoint.editor.form.docs.buttonLabel")}
        </Button>
      </DialogTrigger>
      <ModalWrapper
        size="lg"
        formId={formId}
        title={t("pages.explorer.endpoint.editor.form.docs.modal.title")}
        onCancel={() => {
          setDialogOpen(false);
        }}
      >
        <form id={formId} onSubmit={handleSubmit(onSubmit)}>
          <div className="flex gap-5 p-5">
            <pre className="text-xs h-96 overflow-y-scroll">
              This Form:
              {JSON.stringify(watch("editor"), null, 2)}
            </pre>
            <pre className="text-xs h-96 overflow-y-scroll">
              Parent Form:
              {JSON.stringify(
                {
                  connect: getParentValues("connect"),
                  delete: getParentValues("delete"),
                  get: getParentValues("get"),
                  head: getParentValues("head"),
                  options: getParentValues("options"),
                  patch: getParentValues("patch"),
                  post: getParentValues("post"),
                  put: getParentValues("put"),
                  trace: getParentValues("trace"),
                },
                null,
                2
              )}
            </pre>
            <pre className="text-xs h-96 overflow-y-scroll">
              ERRORS: {JSON.stringify(errors, null, 2)}
            </pre>
          </div>

          <Card className="h-96 w-full p-4" noShadow background="weight-1">
            <Editor
              defaultValue={jsonToYaml(getValues("editor"))}
              onChange={(newDocs) => {
                if (newDocs !== undefined) {
                  const docsAsJson = yamlToJsonOrNull(newDocs) ?? {};
                  setValue("editor", docsAsJson);
                }
              }}
              theme={theme ?? undefined}
            />
          </Card>
        </form>
      </ModalWrapper>
    </Dialog>
  );
};
