import { Braces, PlusCircle, Trash } from "lucide-react";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "~/design/Dialog";
import { FC, useEffect, useState } from "react";
import { SubmitHandler, useForm } from "react-hook-form";
import { Table, TableBody, TableCell, TableRow } from "~/design/Table";
import { Trans, useTranslation } from "react-i18next";
import {
  VarFormSchema,
  VarFormSchemaType,
  VarSchemaType,
} from "~/api/vars/schema";

import Button from "~/design/Button";
import { Card } from "~/design/Card";
import Input from "~/design/Input";
import { Textarea } from "~/design/TextArea";
import { useCreateVar } from "~/api/vars/mutate/createVar";
import { useDeleteVar } from "~/api/vars/mutate/deleteVar";
import { useVars } from "~/api/vars/query/get";
import { zodResolver } from "@hookform/resolvers/zod";

// TODO: Componentize this? Then type needs to be more universal
type ItemRowProps = {
  item: VarSchemaType;
  onDelete: (item: VarSchemaType) => void;
};

const ItemRow = ({ item, onDelete }: ItemRowProps) => (
  <TableRow>
    <TableCell>{item.name}</TableCell>
    <TableCell className="w-0">
      <DialogTrigger
        asChild
        data-testid="registry-delete"
        onClick={() => onDelete(item)}
      >
        <Button variant="ghost">
          <Trash />
        </Button>
      </DialogTrigger>
    </TableCell>
  </TableRow>
);

type DeleteProps = {
  name: string;
  onConfirm: () => void;
};

const Delete = ({ name, onConfirm }: DeleteProps) => {
  const { t } = useTranslation();

  return (
    <DialogContent>
      <DialogHeader>
        <DialogTitle>
          <Trash /> {t("components.dialog.header.confirm")}
        </DialogTitle>
      </DialogHeader>
      <div className="my-3">
        <Trans
          i18nKey="pages.settings.registries.delete.description"
          values={{ name }}
        />
      </div>
      <DialogFooter>
        <DialogClose asChild>
          <Button variant="ghost">{t("components.button.label.cancel")}</Button>
        </DialogClose>
        <Button
          data-testid="registry-delete-confirm"
          onClick={onConfirm}
          variant="destructive"
        >
          {t("components.button.label.delete")}
        </Button>
      </DialogFooter>
    </DialogContent>
  );
};

type createProps = { onSuccess: () => void };

const Create = ({ onSuccess }: createProps) => {
  const { t } = useTranslation();

  const { register, handleSubmit } = useForm<VarFormSchemaType>({
    resolver: zodResolver(VarFormSchema),
  });

  const { mutate: createVarMutation } = useCreateVar({
    onSuccess,
  });

  const onSubmit: SubmitHandler<VarFormSchemaType> = (data) => {
    createVarMutation(data);
  };

  return (
    <DialogContent>
      <form
        id="create-registry"
        onSubmit={handleSubmit(onSubmit)}
        className="flex flex-col space-y-5"
      >
        <DialogHeader>
          <DialogTitle>
            <PlusCircle />
            {t("pages.settings.variables.create.description")}
          </DialogTitle>
        </DialogHeader>

        <fieldset className="flex items-center gap-5">
          <label className="w-[150px] text-right text-[15px]" htmlFor="name">
            {t("pages.settings.variables.create.name")}
          </label>
          <Input
            data-testid="new-variable-name"
            placeholder="https://example.com/registry"
            {...register("name")}
          />
        </fieldset>

        <fieldset className="flex items-start gap-5">
          <Textarea
            className="h-96"
            data-testid="new-workflow-editor"
            {...register("content")}
          />
        </fieldset>

        <DialogFooter>
          <DialogClose asChild>
            <Button variant="ghost">
              {t("components.button.label.cancel")}
            </Button>
          </DialogClose>
          <Button
            data-testid="registry-create-submit"
            type="submit"
            variant="primary"
          >
            {t("components.button.label.create")}
          </Button>
        </DialogFooter>
      </form>
    </DialogContent>
  );
};

const VarsList: FC = () => {
  const { t } = useTranslation();
  const [dialogOpen, setDialogOpen] = useState(false);
  const [deleteItem, setDeleteItem] = useState<VarSchemaType>();
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
              <ItemRow item={item} key={i} onDelete={setDeleteItem} />
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
    </Dialog>
  );
};

export default VarsList;
