import { Dialog, DialogTrigger } from "~/design/Dialog";
import { FC, useState } from "react";
import { Table, TableBody } from "~/design/Table";
import { UseFormReturn, useFieldArray, useWatch } from "react-hook-form";

import Button from "~/design/Button";
import { Card } from "~/design/Card";
import { ModalWrapper } from "~/pages/namespace/Explorer/Endpoint/EndpointEditor/Form/plugins/components/Modal";
import { PatchForm } from "./Form";
import { Plus } from "lucide-react";
import { ServiceFormSchemaType } from "../../schema";
import { TableHeader } from "./Table";
import { useTranslation } from "react-i18next";

type PatchesFormProps = {
  form: UseFormReturn<ServiceFormSchemaType>;
};

export const PatchesForm: FC<PatchesFormProps> = ({ form }) => {
  const { control } = form;
  const values = useWatch({ control });
  const { t } = useTranslation();

  const [dialogOpen, setDialogOpen] = useState(false);

  const {
    append: addItem,
    remove: deleteItem,
    move: moveItem,
    update: updateItem,
    fields,
  } = useFieldArray({
    control,
    name: "patches",
  });

  const itemCount = fields.length;

  return (
    <Dialog
      open={dialogOpen}
      onOpenChange={(isOpen) => {
        // if (isOpen === false) setEditIndex(undefined);
        setDialogOpen(isOpen);
      }}
    >
      <Card noShadow>
        <Table>
          <TableHeader
            title={t("pages.explorer.service.editor.form.patches.label", {
              count: itemCount,
            })}
          >
            <DialogTrigger asChild>
              <Button icon variant="outline" size="sm">
                <Plus />
                {t("pages.explorer.service.editor.form.patches.addButton")}
              </Button>
            </DialogTrigger>
          </TableHeader>
          <TableBody>{/* todo: render an item for every patch */}</TableBody>
        </Table>
      </Card>

      {/* todo: modal wrapper should be generic component */}

      <ModalWrapper
        title={t("pages.explorer.service.editor.form.patches.modal.title")}
      >
        <PatchForm />
      </ModalWrapper>
    </Dialog>
  );
};
