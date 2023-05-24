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
import { SubmitHandler, useForm } from "react-hook-form";
import { Table, TableBody, TableCell, TableRow } from "~/design/Table";
import { Trans, useTranslation } from "react-i18next";

import Button from "~/design/Button";
import { Card } from "~/design/Card";
import Input from "~/design/Input";
import { SecretSchemaType } from "~/api/secrets/schema";
import { Textarea } from "~/design/TextArea";
import { useCreateSecret } from "~/api/secrets/mutate/createSecret";
import { useDeleteSecret } from "~/api/secrets/mutate/deleteSecret";
import { useSecrets } from "~/api/secrets/query/get";
import { z } from "zod";
import { zodResolver } from "@hookform/resolvers/zod";

type SecretFormInput = {
  name: string;
  value: string;
};

const SecretsList: FC = () => {
  const { t } = useTranslation();

  const [dialogOpen, setDialogOpen] = useState(false);
  const [deleteSecret, setDeleteSecret] = useState<SecretSchemaType>();
  const [createSecret, setCreateSecret] = useState(false);

  const secrets = useSecrets();

  const { mutate: deleteSecretMutation } = useDeleteSecret({
    onSuccess: () => {
      setDeleteSecret(undefined);
      setDialogOpen(false);
    },
  });

  const { mutate: createSecretMutation } = useCreateSecret({
    onSuccess: () => {
      setCreateSecret(false);
      setDialogOpen(false);
    },
  });

  const onSubmit: SubmitHandler<SecretFormInput> = ({ name, value }) => {
    createSecretMutation({
      name,
      value,
    });
  };

  const { register, handleSubmit, reset } = useForm<SecretFormInput>({
    resolver: zodResolver(
      z.object({
        name: z.string(),
        value: z.string(),
      })
    ),
  });

  const resetDialog = (isOpening: boolean) => {
    if (!isOpening) {
      setDeleteSecret(undefined);
      setCreateSecret(false);
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
          {t("pages.settings.secrets.list.title")}
        </h3>

        <DialogTrigger
          asChild
          data-testid="secret-create"
          onClick={() => setCreateSecret(true)}
        >
          <Button variant="ghost">
            <PlusCircle />
          </Button>
        </DialogTrigger>
      </div>

      <Card>
        <Table>
          <TableBody>
            {secrets.data?.secrets.results.map((secret, i) => (
              <TableRow key={i}>
                <TableCell>{secret.name}</TableCell>
                <TableCell className="w-0">
                  <DialogTrigger
                    asChild
                    data-testid="secret-delete"
                    onClick={() => setDeleteSecret(secret)}
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
        {deleteSecret && (
          <DialogContent>
            <DialogHeader>
              <DialogTitle>
                <Trash /> {t("components.dialog.header.confirm")}
              </DialogTitle>
            </DialogHeader>
            <div className="my-3">
              <Trans
                i18nKey="pages.settings.secrets.delete.description"
                values={{ name: deleteSecret.name }}
              />
            </div>
            <DialogFooter>
              <DialogClose asChild>
                <Button variant="ghost">
                  {t("components.button.label.cancel")}
                </Button>
              </DialogClose>
              <Button
                data-testid="secret-delete-confirm"
                onClick={() => deleteSecretMutation({ secret: deleteSecret })}
                variant="destructive"
              >
                {t("components.button.label.delete")}
              </Button>
            </DialogFooter>
          </DialogContent>
        )}
        {createSecret && (
          <DialogContent>
            <form
              id="create-secret"
              onSubmit={handleSubmit(onSubmit)}
              className="flex flex-col space-y-5"
            >
              <DialogHeader>
                <DialogTitle>
                  <PlusCircle />{" "}
                  {t("pages.settings.secrets.create.description")}
                </DialogTitle>
              </DialogHeader>

              <fieldset className="flex items-center gap-5">
                <label
                  className="w-[150px] text-right text-[15px]"
                  htmlFor="name"
                >
                  {t("pages.settings.secrets.create.name")}
                </label>
                <Input
                  data-testid="new-secret-name"
                  id="name"
                  placeholder="secret-name"
                  {...register("name")}
                />
              </fieldset>

              <fieldset className="flex items-start gap-5">
                <Textarea
                  className="h-96"
                  data-testid="new-workflow-editor"
                  {...register("value")}
                />
              </fieldset>

              <DialogFooter>
                <DialogClose asChild>
                  <Button variant="ghost">
                    {t("components.button.label.cancel")}
                  </Button>
                </DialogClose>
                <Button
                  data-testid="secret-create-submit"
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

export default SecretsList;
