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
import {
  FolderTree,
  GitCommit,
  Hexagon,
  Key,
  Palette,
  PlusCircle,
  SquareAsterisk,
  Trash,
} from "lucide-react";
import { SubmitHandler, useForm } from "react-hook-form";
import { Table, TableBody, TableCell, TableRow } from "~/design/Table";
import { Trans, useTranslation } from "react-i18next";
import { useApiActions, useApiKey } from "~/util/store/apiKey";
import { useNamespace, useNamespaceActions } from "~/util/store/namespace";
import { useTheme, useThemeActions } from "~/util/store/theme";

import Button from "~/design/Button";
import { Card } from "~/design/Card";
import Input from "~/design/Input";
import { SecretSchemaType } from "~/api/secrets/schema";
import { Textarea } from "~/design/TextArea";
import { useCreateSecret } from "~/api/secrets/mutate/createSecret";
import { useDeleteSecret } from "~/api/secrets/mutate/deleteSecret";
import { useListNamespaces } from "~/api/namespaces/query/get";
import { useSecrets } from "~/api/secrets/query/get";
import { useVersion } from "~/api/version";
import { z } from "zod";
import { zodResolver } from "@hookform/resolvers/zod";

const SettingsPage: FC = () => {
  const apiKey = useApiKey();
  const { setApiKey } = useApiActions();
  const theme = useTheme();
  const { setTheme } = useThemeActions();
  const { setNamespace } = useNamespaceActions();

  const selectedNamespace = useNamespace();

  const secrets = useSecrets();
  const { data: version, isLoading: isVersionLoading } = useVersion();
  const { data: namespaces, isLoading: isLoadingNamespaces } =
    useListNamespaces();

  const [dialogOpen, setDialogOpen] = useState(false);
  const [deleteSecret, setDeleteSecret] = useState<SecretSchemaType>();
  const [createSecret, setCreateSecret] = useState(false);

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

  type SecretFormInput = {
    name: string;
    value: string;
  };

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

  const { t } = useTranslation();

  const resetDialog = (isOpening: boolean) => {
    if (!isOpening) {
      setDeleteSecret(undefined);
      setCreateSecret(false);
      reset();
    }
    setDialogOpen(isOpening);
  };

  return (
    <div className="flex flex-col space-y-6 p-10">
      <section>
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
                    <Trash /> Delete
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
                    <Button variant="ghost">Cancel</Button>
                  </DialogClose>
                  <Button
                    data-testid="secret-delete-confirm"
                    onClick={() =>
                      deleteSecretMutation({ secret: deleteSecret })
                    }
                    variant="destructive"
                  >
                    Delete
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
                      <Trash /> Create
                    </DialogTitle>
                  </DialogHeader>

                  <fieldset className="flex items-center gap-5">
                    <label
                      className="w-[150px] text-right text-[15px]"
                      htmlFor="name"
                    >
                      Name
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
                      <Button variant="ghost">Cancel</Button>
                    </DialogClose>
                    <Button
                      data-testid="secret-create-submit"
                      type="submit"
                      variant="primary"
                    >
                      Create
                    </Button>
                  </DialogFooter>
                </form>
              </DialogContent>
            )}
          </Card>
        </Dialog>
      </section>

      <section>
        <h3 className="mb-3 flex items-center gap-x-2 font-bold text-gray-10 dark:text-gray-dark-10">
          <Palette className="h-5" />
          {t("pages.settings.theme.title")} {theme}
        </h3>

        <Card className="flex gap-x-3 p-4">
          <div className="flex space-x-5">
            <Button onClick={() => setTheme("dark")}>darkmode</Button>
            <Button className="" onClick={() => setTheme("light")}>
              lightmode
            </Button>
            <Button onClick={() => setTheme(null)}>reset theme</Button>
          </div>
        </Card>
      </section>

      <section>
        <h3 className="mb-3 flex items-center gap-x-2 font-bold text-gray-10 dark:text-gray-dark-10">
          <Hexagon className="h-5" />
          {t("pages.settings.namespace.title")} {selectedNamespace}
        </h3>

        <Card className="flex gap-x-3 p-4">
          <div className="flex space-x-5">
            <Button variant="destructive" onClick={() => setNamespace(null)}>
              reset namespace
            </Button>
          </div>
        </Card>
      </section>

      <section>
        <h3 className="mb-3 flex items-center gap-x-2 font-bold text-gray-10 dark:text-gray-dark-10">
          <Key className="h-5" />
          {t("pages.settings.apiKey.title")} {apiKey}
        </h3>

        <Card className="flex gap-x-3 p-4">
          <div className="flex space-x-5">
            <Button onClick={() => setApiKey("password")}>
              set Api key to password
            </Button>
            <Button variant="destructive" onClick={() => setApiKey(null)}>
              reset api key
            </Button>
          </div>
        </Card>
      </section>

      <section>
        <h3 className="mb-3 flex items-center gap-x-2 font-bold text-gray-10 dark:text-gray-dark-10">
          <GitCommit className="h-5" />
          {t("pages.settings.version.title")}
        </h3>
        <div className="mt-2 ml-2">
          {isVersionLoading ? "Loading version...." : version?.api}
        </div>
      </section>
      <section>
        <h3 className="flex items-center gap-x-2 font-bold text-gray-10 dark:text-gray-dark-10">
          <FolderTree className="h-5" />
          {t("pages.settings.namespacesList.title")}
        </h3>
        <div className="mt-2 ml-2">
          {isLoadingNamespaces
            ? "Loading namespaces"
            : namespaces?.results.map((namespace) => (
                <div key={namespace.name}>{namespace.name}</div>
              ))}
        </div>
      </section>
    </div>
  );
};

export default SettingsPage;
