import { Braces, PlusCircle } from "lucide-react";
import { Dialog, DialogTrigger } from "~/design/Dialog";
import { FC, useEffect, useState } from "react";
import { Table, TableBody } from "~/design/Table";

import Button from "~/design/Button";
import { Card } from "~/design/Card";
import Create from "./Create";
import Delete from "./Delete";
import Edit from "./Edit";
import EmptyList from "../compopnents/EmptyList";
import ItemRow from "../compopnents/ItemRow";
import { VarSchemaType } from "~/api/variables/schema";
import { useDeleteVar } from "~/api/variables/mutate/deleteVariable";
import { useTranslation } from "react-i18next";
import { useVars } from "~/api/variables/query/useVariables";

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
        {items?.length ? (
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
        ) : (
          <EmptyList>{t("pages.settings.variables.list.empty")}</EmptyList>
        )}
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
            setDialogOpen(false);
          }}
        />
      )}
      {editItem && (
        <Edit
          item={editItem}
          onSuccess={() => {
            setDialogOpen(false);
          }}
        />
      )}
    </Dialog>
  );
};

export default VariablesList;
