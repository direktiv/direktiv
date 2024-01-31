import { Dialog, DialogTrigger } from "~/design/Dialog";
import { FC, useState } from "react";
import { PatchRow, TableHeader } from "./Table";
import { PatchSchemaType, ServiceFormSchemaType } from "../../schema";
import { Table, TableBody } from "~/design/Table";
import { UseFormReturn, useFieldArray, useWatch } from "react-hook-form";

import Button from "~/design/Button";
import { Card } from "~/design/Card";
import { ModalWrapper } from "~/pages/namespace/Explorer/Endpoint/EndpointEditor/Form/plugins/components/Modal";
import { PatchItemForm } from "./Item";
import { Plus } from "lucide-react";
import { useTranslation } from "react-i18next";

type PatchesFormProps = {
  form: UseFormReturn<ServiceFormSchemaType>;
};

export const PatchesForm: FC<PatchesFormProps> = ({ form }) => {
  const { control } = form;
  const values = useWatch({ control });
  const { t } = useTranslation();

  const formId = "patchItemForm";

  const emptyPatch: PatchSchemaType = {
    op: "add",
    path: "",
    value: "",
  };

  const [dialogOpen, setDialogOpen] = useState(false);
  const [indexToEdit, setIndexToEdit] = useState<number | undefined>();

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

  const handleSubmit = (item: PatchSchemaType) => {
    if (indexToEdit === undefined) {
      addItem(item);
    } else {
      updateItem(indexToEdit, item);
    }
    setIndexToEdit(undefined);
    setDialogOpen(false);
  };

  const itemCount = fields.length;

  if (!values.patches) return <></>;

  return (
    <Dialog
      open={dialogOpen}
      onOpenChange={(isOpen) => {
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
          <TableBody>
            {fields.map((field, index, srcArray) => (
              <PatchRow
                key={field.id}
                patch={field}
                onClick={() => {
                  setIndexToEdit(index);
                  setDialogOpen(true);
                }}
                onDelete={() => deleteItem(index)}
                onMoveUp={
                  index > 0 ? () => moveItem(index, index - 1) : undefined
                }
                onMoveDown={
                  index < srcArray.length - 1
                    ? () => moveItem(index, index + 1)
                    : undefined
                }
              />
            ))}
          </TableBody>
        </Table>
      </Card>

      {/* todo: modal wrapper should be generic component */}

      <ModalWrapper
        formId={formId}
        title={t("pages.explorer.service.editor.form.patches.modal.title")}
      >
        <PatchItemForm
          formId={formId}
          value={indexToEdit !== undefined ? fields?.[indexToEdit] : emptyPatch}
          onSubmit={handleSubmit}
        />
      </ModalWrapper>
    </Dialog>
  );
};
