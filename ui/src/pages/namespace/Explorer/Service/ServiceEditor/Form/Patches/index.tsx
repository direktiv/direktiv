import { FC, useState } from "react";
import { UseFormReturn, useFieldArray, useWatch } from "react-hook-form";

import { Card } from "~/design/Card";
import { Dialog } from "~/design/Dialog";
import { ServiceFormSchemaType } from "../../schema";
import { Table } from "~/design/Table";
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
          ></TableHeader>
        </Table>
      </Card>
    </Dialog>
  );
};
