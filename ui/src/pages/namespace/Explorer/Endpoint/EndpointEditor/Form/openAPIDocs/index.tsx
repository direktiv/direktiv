import { Dialog, DialogTrigger } from "~/design/Dialog";
import { FC, useState } from "react";

import Button from "~/design/Button";
import { EndpointFormSchemaType } from "../../schema";
import { MethodsSchemaType } from "~/api/gateway/schema";
import { ModalWrapper } from "~/components/ModalWrapper";
import { OpenAPIDocsEditor } from "./OpenAPIDocsEditor";
import { ScrollText } from "lucide-react";
import { UseFormReturn } from "react-hook-form";
import { useTranslation } from "react-i18next";

type OpenAPIDocsFormProps = {
  form: UseFormReturn<EndpointFormSchemaType>;
  onSave: (value: EndpointFormSchemaType) => void;
};

export const OpenAPIDocsForm: FC<OpenAPIDocsFormProps> = ({ form, onSave }) => {
  const { t } = useTranslation();
  const { handleSubmit, watch } = form;
  const [dialogOpen, setDialogOpen] = useState(false);

  const onSubmit = (configuration: MethodsSchemaType) => {
    setDialogOpen(false);
    form.setValue("connect", configuration.connect);
    form.setValue("delete", configuration.delete);
    form.setValue("get", configuration.get);
    form.setValue("head", configuration.head);
    form.setValue("options", configuration.options);
    form.setValue("patch", configuration.patch);
    form.setValue("post", configuration.post);
    form.setValue("put", configuration.put);
    form.setValue("trace", configuration.trace);
    handleSubmit(onSave)();
  };

  const formId = "openAPIDocsForm";

  return (
    <Dialog
      open={dialogOpen}
      onOpenChange={(isOpen) => {
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
        <OpenAPIDocsEditor
          id={formId}
          defaultValue={{
            connect: watch("connect"),
            delete: watch("delete"),
            get: watch("get"),
            head: watch("head"),
            options: watch("options"),
            patch: watch("patch"),
            post: watch("post"),
            put: watch("put"),
            trace: watch("trace"),
          }}
          onSubmit={onSubmit}
        />
      </ModalWrapper>
    </Dialog>
  );
};
