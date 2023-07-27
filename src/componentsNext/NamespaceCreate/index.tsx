import {
  DialogClose,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "~/design/Dialog";
import { Home, PlusCircle } from "lucide-react";
import { MirrorSchema, MirrorSchemaType } from "~/api/namespaces/schema";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "~/design/Select";
import { SubmitHandler, useForm } from "react-hook-form";
import { Tabs, TabsList, TabsTrigger } from "~/design/Tabs";

import Button from "~/design/Button";
import FormErrors from "~/componentsNext/FormErrors";
import Input from "~/design/Input";
import { fileNameSchema } from "~/api/tree/schema";
import { pages } from "~/util/router/pages";
import { useCreateNamespace } from "~/api/namespaces/mutate/createNamespace";
import { useListNamespaces } from "~/api/namespaces/query/get";
import { useNamespaceActions } from "~/util/store/namespace";
import { useNavigate } from "react-router-dom";
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { z } from "zod";
import { zodResolver } from "@hookform/resolvers/zod";

type FormInput = {
  name: string;
} & MirrorSchemaType;

const mirrorAuthTypes = ["none", "ssh", "token"] as const;

const NamespaceCreate = ({ close }: { close: () => void }) => {
  const { t } = useTranslation();
  const [isMirror, setIsMirror] = useState<boolean>(false);
  const [authType, setAuthType] = useState<string>("none");
  const { data } = useListNamespaces();
  const { setNamespace } = useNamespaceActions();
  const navigate = useNavigate();

  const existingNamespaces = data?.results.map((n) => n.name) || [];

  const nameSchema = fileNameSchema.and(
    z.string().refine((name) => !existingNamespaces.some((n) => n === name), {
      message: t("components.namespaceCreate.nameAlreadyExists"),
    })
  );

  const simpleNamespaceResolver = zodResolver(z.object({ name: nameSchema }));

  const mirrorResolver = zodResolver(
    MirrorSchema.and(z.object({ name: nameSchema }))
  );

  const {
    register,
    handleSubmit,
    formState: { isDirty, errors, isValid, isSubmitted },
  } = useForm<FormInput>({
    resolver: isMirror ? mirrorResolver : simpleNamespaceResolver,
  });

  const { mutate: createNamespace, isLoading } = useCreateNamespace({
    onSuccess: (data) => {
      setNamespace(data.namespace.name);
      navigate(
        pages.explorer.createHref({
          namespace: data.namespace.name,
        })
      );
      close();
    },
  });

  const onSubmit: SubmitHandler<FormInput> = ({
    name,
    ref,
    url,
    passphrase,
    publicKey,
    privateKey,
  }) => {
    createNamespace({
      name,
      mirror: { ref, url, passphrase, publicKey, privateKey },
    });
  };

  // you can not submit if the form has not changed or if there are any errors and
  // you have already submitted the form (errors will first show up after submit)
  const disableSubmit = !isDirty || (isSubmitted && !isValid);

  const formId = `new-namespace`;
  return (
    <>
      <DialogHeader>
        <DialogTitle>
          <Home /> {t("components.namespaceCreate.title")}
        </DialogTitle>
      </DialogHeader>

      <Tabs className="w-[400px]" defaultValue="namespace">
        <TabsList>
          <TabsTrigger value="namespace" onClick={() => setIsMirror(false)}>
            Namespace
          </TabsTrigger>
          <TabsTrigger value="mirror" onClick={() => setIsMirror(true)}>
            Mirror
          </TabsTrigger>
        </TabsList>
      </Tabs>

      <div className="my-3">
        <FormErrors errors={errors} className="mb-5" />
        <form
          id={formId}
          onSubmit={handleSubmit(onSubmit)}
          className="flex flex-col gap-y-5"
        >
          <fieldset className="flex items-center gap-5">
            <label className="w-[90px] text-right text-[14px]" htmlFor="name">
              {t("components.namespaceCreate.nameLabel")}
            </label>
            <Input
              id="name"
              data-testid="new-namespace-name"
              placeholder={t("components.namespaceCreate.placeholder")}
              {...register("name")}
            />
          </fieldset>

          {isMirror && (
            <>
              <fieldset className="flex items-center gap-5">
                <label
                  className="w-[90px] text-right text-[14px]"
                  htmlFor="name"
                >
                  {t("components.namespaceCreate.urlLabel")}
                </label>
                <Input
                  id="name"
                  data-testid="new-namespace-name"
                  placeholder={t("components.namespaceCreate.placeholder")}
                  {...register("url")}
                />
              </fieldset>

              <fieldset className="flex items-center gap-5">
                <label
                  className="w-[90px] text-right text-[14px]"
                  htmlFor="name"
                >
                  {t("components.namespaceCreate.refLabel")}
                </label>
                <Input
                  id="name"
                  data-testid="new-namespace-name"
                  placeholder={t("components.namespaceCreate.placeholder")}
                  {...register("ref")}
                />
              </fieldset>

              <Select value={authType} onValueChange={setAuthType}>
                <SelectTrigger variant="outline" className="ml-[90px]">
                  <SelectValue placeholder="Select Auth Method" />
                </SelectTrigger>
                <SelectContent>
                  {mirrorAuthTypes.map((option) => (
                    <SelectItem
                      key={option}
                      value={option}
                      onClick={() => setAuthType(option)}
                    >
                      {t(`components.namespaceCreate.authType.${option}`)}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>

              {(authType === "token" || authType == "ssh") && (
                <fieldset className="flex items-center gap-5">
                  <label
                    className="w-[90px] text-right text-[14px]"
                    htmlFor="name"
                  >
                    {t("components.namespaceCreate.passphrase")}
                  </label>
                  <Input
                    id="name"
                    data-testid="new-namespace-name"
                    placeholder={t("components.namespaceCreate.placeholder")}
                    {...register("passphrase")}
                  />
                </fieldset>
              )}

              {authType === "ssh" && (
                <>
                  <fieldset className="flex items-center gap-5">
                    <label
                      className="w-[90px] text-right text-[14px]"
                      htmlFor="name"
                    >
                      {t("components.namespaceCreate.publicKey")}
                    </label>
                    <Input
                      id="name"
                      data-testid="new-namespace-name"
                      placeholder={t("components.namespaceCreate.placeholder")}
                      {...register("publicKey")}
                    />
                  </fieldset>

                  <fieldset className="flex items-center gap-5">
                    <label
                      className="w-[90px] text-right text-[14px]"
                      htmlFor="name"
                    >
                      {t("components.namespaceCreate.privateKey")}
                    </label>
                    <Input
                      id="name"
                      data-testid="new-namespace-name"
                      placeholder={t("components.namespaceCreate.placeholder")}
                      {...register("privateKey")}
                    />
                  </fieldset>
                </>
              )}
            </>
          )}
        </form>
      </div>

      <DialogFooter>
        <DialogClose asChild>
          <Button variant="ghost">
            {t("components.namespaceCreate.cancelBtn")}
          </Button>
        </DialogClose>
        <Button
          data-testid="new-namespace-submit"
          type="submit"
          disabled={disableSubmit}
          loading={isLoading}
          form={formId}
        >
          {!isLoading && <PlusCircle />}
          {t("components.namespaceCreate.createBtn")}
        </Button>
      </DialogFooter>
    </>
  );
};

export default NamespaceCreate;
