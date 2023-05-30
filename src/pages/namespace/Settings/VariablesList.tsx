import { Braces, PlusCircle } from "lucide-react";
import { Dialog, DialogTrigger } from "~/design/Dialog";
import { FC, useEffect, useState } from "react";
import { Table, TableBody } from "~/design/Table";

import Button from "~/design/Button";
import { Card } from "~/design/Card";
import Create from "./variables/Create";
import Delete from "./variables/Delete";
import Edit from "./variables/Edit";
import ItemRow from "./ItemRow";
import { VarSchemaType } from "~/api/vars/schema";
import { useDeleteVar } from "~/api/vars/mutate/deleteVar";
import { useTranslation } from "react-i18next";
import { useVars } from "~/api/vars/query/useVars";

const VariablesList: FC = () => {
  const { t } = useTranslation();
  const [dialogOpen, setDialogOpen] = useState(false);
  const [deleteItem, setDeleteItem] = useState<VarSchemaType>();
  const [editItem, setEditItem] = useState<VarSchemaType>();
  const [createItem, setCreateItem] = useState(false);

  const data = useVars();
  const items = data.data?.variables?.results ?? null;

  const { mutate: deleteVarMutation } = useDeleteVar({
    onSuccess: () => {
      setDialogOpen(false);
    },
  });

  useEffect(() => {
    if (dialogOpen === false) {
      setDeleteItem(undefined);
      setCreateItem(false);
      setEditItem(undefined);
    }
  }, [dialogOpen]);

  return (
    <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
      <div className="mb-3 flex flex-row justify-between">
        <h3 className="flex items-center gap-x-2 font-bold text-gray-10 dark:text-gray-dark-10">
          <Braces className="h-5" />
          {t("pages.settings.variables.list.title")}
        </h3>

        <DialogTrigger
          asChild
          data-testid="variable-create"
          onClick={() => setCreateItem(true)}
        >
          <Button variant="ghost">
            <PlusCircle />
          </Button>
        </DialogTrigger>
      </div>

      <Card>
        <Table>
          <TableBody>
            {items?.map((item, i) => (
              <ItemRow
                item={item}
                key={i}
                onDelete={setDeleteItem}
                onEdit={() => setEditItem(item)}
              />
            ))}
          </TableBody>
        </Table>
      </Card>
      {deleteItem && (
        <Delete
          name={deleteItem.name}
          onConfirm={() => deleteVarMutation({ variable: deleteItem })}
        />
      )}
      {createItem && (
        <Create
          onSuccess={() => {
            setCreateItem(false);
            setDialogOpen(false);
          }}
        />
      )}
      {editItem && (
        <Edit
          item={editItem}
          onSuccess={() => {
            setCreateItem(false);
            setDialogOpen(false);
          }}
        />
      )}
    </Dialog>
  );
};

export default VariablesList;
