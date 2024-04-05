import { Dialog, DialogTrigger } from "~/design/Dialog";
import { FC, useState } from "react";
import { PatchRow, TableHeader } from "./Table";
import { PatchSchemaType, ServiceFormSchemaType } from "../../schema";
import { Table, TableBody } from "~/design/Table";
import { UseFormReturn, useFieldArray, useWatch } from "react-hook-form";

import Button from "~/design/Button";
import { Card } from "~/design/Card";
import { ModalWrapper } from "~/components/ModalWrapper";
import { PatchItemForm } from "./Item";
import { Plus } from "lucide-react";
import { useTranslation } from "react-i18next";

type PatchesFormProps = {
  form: UseFormReturn<ServiceFormSchemaType>;
  onSave: (value: ServiceFormSchemaType) => void;
};

export const PatchesForm: FC<PatchesFormProps> = ({ form, onSave }) => {
  const { control, handleSubmit: handleParentSubmit } = form;
  const values = useWatch({ control });
  const { t } = useTranslation();

  const formId = "patchItemForm";

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
    handleParentSubmit(onSave)();
    setDialogOpen(false);
  };

  const itemCount = fields.length;

  if (!values.patches) return null;

  return (
    <Dialog
      open={dialogOpen}
      onOpenChange={(isOpen) => {
        if (isOpen === false) setIndexToEdit(undefined);
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

      <ModalWrapper
        formId={formId}
        title={
          indexToEdit !== undefined
            ? t("pages.explorer.service.editor.form.patches.modal.title.edit")
            : t("pages.explorer.service.editor.form.patches.modal.title.create")
        }
        onCancel={() => {
          setDialogOpen(false);
        }}
      >
        <PatchItemForm
          formId={formId}
          value={indexToEdit !== undefined ? fields?.[indexToEdit] : undefined}
          onSubmit={handleSubmit}
        />
      </ModalWrapper>
    </Dialog>
  );
};
