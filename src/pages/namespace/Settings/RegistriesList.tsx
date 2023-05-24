import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "~/design/Dialog";
import { FC, useState } from "react";
import { PlusCircle, SquareAsterisk, Trash } from "lucide-react";
import {
  RegistryFormSchema,
  RegistryFormSchemaType,
  RegistrySchemaType,
} from "~/api/registries/schema";
import { SubmitHandler, useForm } from "react-hook-form";
import { Table, TableBody, TableCell, TableRow } from "~/design/Table";
import { Trans, useTranslation } from "react-i18next";

import Button from "~/design/Button";
import { Card } from "~/design/Card";
import Input from "~/design/Input";
import { useCreateRegistry } from "~/api/registries/mutate/createRegistry";
import { useDeleteRegistry } from "~/api/registries/mutate/deleteRegistry";
import { useRegistries } from "~/api/registries/query/get";
import { zodResolver } from "@hookform/resolvers/zod";

const RegistriesList: FC = () => {
  const { t } = useTranslation();

  const [dialogOpen, setDialogOpen] = useState(false);
  const [deleteRegistry, setDeleteRegistry] = useState<RegistrySchemaType>();
  const [createRegistry, setCreateRegistry] = useState(false);

  const registries = useRegistries();

  const { mutate: deleteRegistryMutation } = useDeleteRegistry({
    onSuccess: () => {
      setDeleteRegistry(undefined);
      setDialogOpen(false);
    },
  });

  const { mutate: createRegistryMutation } = useCreateRegistry({
    onSuccess: () => {
      setCreateRegistry(false);
      setDialogOpen(false);
    },
  });

  const onSubmit: SubmitHandler<RegistryFormSchemaType> = (data) => {
    createRegistryMutation(data);
  };

  const { register, handleSubmit, reset } = useForm<RegistryFormSchemaType>({
    resolver: zodResolver(RegistryFormSchema),
  });

  const resetDialog = (isOpening: boolean) => {
    if (!isOpening) {
      setDeleteRegistry(undefined);
      setCreateRegistry(false);
      reset();
    }
    setDialogOpen(isOpening);
  };

  return (
    <Dialog
      open={dialogOpen}
      onOpenChange={(isOpening) => resetDialog(isOpening)}
    >
      <div className="mb-3 flex flex-row justify-between">
        <h3 className="flex items-center gap-x-2 font-bold text-gray-10 dark:text-gray-dark-10">
          <SquareAsterisk className="h-5" />
          {t("pages.settings.registries.list.title")}
        </h3>

        <DialogTrigger
          asChild
          data-testid="registry-create"
          onClick={() => setCreateRegistry(true)}
        >
          <Button variant="ghost">
            <PlusCircle />
          </Button>
        </DialogTrigger>
      </div>

      <Card>
        <Table>
          <TableBody>
            {registries.data?.registries.map((item, i) => (
              <TableRow key={i}>
                <TableCell>{item.name}</TableCell>
                <TableCell className="w-0">
                  <DialogTrigger
                    asChild
                    data-testid="registry-delete"
                    onClick={() => {
                      setDeleteRegistry(item);
                    }}
                  >
                    <Button variant="ghost">
                      <Trash />
                    </Button>
                  </DialogTrigger>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
        {deleteRegistry && (
          <DialogContent>
            <DialogHeader>
              <DialogTitle>
                <Trash /> {t("components.dialog.header.confirm")}
              </DialogTitle>
            </DialogHeader>
            <div className="my-3">
              <Trans
                i18nKey="pages.settings.registries.delete.description"
                values={{ name: deleteRegistry.name }}
              />
            </div>
            <DialogFooter>
              <DialogClose asChild>
                <Button variant="ghost">Cancel</Button>
              </DialogClose>
              <Button
                data-testid="registry-delete-confirm"
                onClick={() =>
                  deleteRegistryMutation({ registry: deleteRegistry })
                }
                variant="destructive"
              >
                {t("components.button.label.delete")}
              </Button>
            </DialogFooter>
          </DialogContent>
        )}
        {createRegistry && (
          <DialogContent>
            <form
              id="create-registry"
              onSubmit={handleSubmit(onSubmit)}
              className="flex flex-col space-y-5"
            >
              <DialogHeader>
                <DialogTitle>
                  <PlusCircle />
                  {t("pages.settings.registries.create.description")}
                </DialogTitle>
              </DialogHeader>

              <fieldset className="flex items-center gap-5">
                <label
                  className="w-[150px] text-right text-[15px]"
                  htmlFor="name"
                >
                  {t("pages.settings.registries.create.url")}
                </label>
                <Input
                  data-testid="new-registry-url"
                  placeholder="https://example.com/registry"
                  {...register("url")}
                />
              </fieldset>

              <fieldset className="flex items-center gap-5">
                <label
                  className="w-[150px] text-right text-[15px]"
                  htmlFor="name"
                >
                  {t("pages.settings.registries.create.user")}
                </label>
                <Input
                  className="h-96"
                  data-testid="new-registry-user"
                  placeholder="user-name"
                  {...register("user")}
                />
              </fieldset>

              <fieldset className="flex items-center gap-5">
                <label
                  className="w-[150px] text-right text-[15px]"
                  htmlFor="name"
                >
                  {t("pages.settings.registries.create.password")}
                </label>
                <Input
                  className="h-96"
                  data-testid="new-registry-pwd"
                  type="password"
                  placeholder="password"
                  {...register("password")}
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
        )}
      </Card>
    </Dialog>
  );
};

export default RegistriesList;
