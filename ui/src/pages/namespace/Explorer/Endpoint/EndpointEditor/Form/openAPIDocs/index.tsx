import { BookOpen, Plus } from "lucide-react";
import { Dialog, DialogTrigger } from "~/design/Dialog";
import { FC, useState } from "react";

import Button from "~/design/Button";
import { EndpointFormSchemaType } from "../../schema";
import { ModalWrapper } from "~/components/ModalWrapper";
import { UseFormReturn } from "react-hook-form";
import { useTranslation } from "react-i18next";

type OpenAPIDocsFormProps = {
  formControls: UseFormReturn<EndpointFormSchemaType>;
  onSave: (value: EndpointFormSchemaType) => void;
};

export const OpenAPIDocsForm: FC<OpenAPIDocsFormProps> = ({
  formControls,
  onSave,
}) => {
  const { t } = useTranslation();
  const { control, handleSubmit: handleParentSubmit } = formControls;

  const [dialogOpen, setDialogOpen] = useState(false);
  const [editIndex, setEditIndex] = useState<number>();

  const formId = "authPluginForm";

  // const handleSubmit = (configuration: PluginConfigSchema) => {
  //   setDialogOpen(false);
  //   if (editIndex === undefined) {
  //     addPlugin(configuration);
  //   } else {
  //     editPlugin(editIndex, configuration);
  //   }
  //   handleParentSubmit(onSave)();
  //   setEditIndex(undefined);
  // };

  return (
    <Dialog
      open={dialogOpen}
      onOpenChange={(isOpen) => {
        if (isOpen === false) setEditIndex(undefined);
        setDialogOpen(isOpen);
      }}
    >
      <DialogTrigger asChild>
        <Button icon variant="outline">
          <BookOpen />
          {t("pages.explorer.endpoint.editor.form.docs.buttonLabel")}
        </Button>
      </DialogTrigger>
      <ModalWrapper
        formId={formId}
        title={t("pages.explorer.endpoint.editor.form.docs.modal.title")}
        onCancel={() => {
          setDialogOpen(false);
        }}
      ></ModalWrapper>
    </Dialog>
  );
};
