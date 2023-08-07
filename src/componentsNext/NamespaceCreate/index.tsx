import {
  DialogClose,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "~/design/Dialog";
import { Home, PlusCircle } from "lucide-react";
import {
  MirrorFormSchema,
  MirrorFormSchemaType,
  MirrorSshFormSchema,
  MirrorSshFormSchemaType,
  MirrorTokenFormSchema,
  MirrorTokenFormSchemaType,
} from "~/api/namespaces/schema";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "~/design/Select";
import { SubmitHandler, useForm } from "react-hook-form";
import { Tabs, TabsList, TabsTrigger } from "~/design/Tabs";
import { useEffect, useState } from "react";

import Button from "~/design/Button";
import FormErrors from "~/componentsNext/FormErrors";
import InfoTooltip from "./InfoTooltip";
import Input from "~/design/Input";
import { InputWithButton } from "~/design/InputWithButton";
import { Textarea } from "~/design/TextArea";
import { fileNameSchema } from "~/api/tree/schema";
import { pages } from "~/util/router/pages";
import { useCreateNamespace } from "~/api/namespaces/mutate/createNamespace";
import { useListNamespaces } from "~/api/namespaces/query/get";
import { useNamespaceActions } from "~/util/store/namespace";
import { useNavigate } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { z } from "zod";
import { zodResolver } from "@hookform/resolvers/zod";

type FormInput = {
  name: string;
} & MirrorFormSchemaType &
  MirrorTokenFormSchemaType &
  MirrorSshFormSchemaType;

const mirrorAuthTypes = ["none", "ssh", "token"] as const;

type MirrorAuthType = (typeof mirrorAuthTypes)[number];

const NamespaceCreate = ({ close }: { close: () => void }) => {
  const { t } = useTranslation();
  const [isMirror, setIsMirror] = useState<boolean>(false);
  const [authType, setAuthType] = useState<MirrorAuthType>("none");
  const { data } = useListNamespaces();
  const { setNamespace } = useNamespaceActions();
  const navigate = useNavigate();

  const existingNamespaces = data?.results.map((n) => n.name) || [];

  const nameSchema = fileNameSchema.and(
    z.string().refine((name) => !existingNamespaces.some((n) => n === name), {
      message: t("components.namespaceCreate.nameAlreadyExists"),
    })
  );

  const baseSchema = z.object({ name: nameSchema });

  const getResolver = (isMirror: boolean, authType: MirrorAuthType) => {
    if (!isMirror) {
      return zodResolver(baseSchema);
    }
    if (authType === "token") {
      return zodResolver(baseSchema.and(MirrorTokenFormSchema));
    }
    if (authType === "ssh") {
      return zodResolver(baseSchema.and(MirrorSshFormSchema));
    }
    return zodResolver(baseSchema.and(MirrorFormSchema));
  };

  const {
    register,
    handleSubmit,
    trigger,
    formState: { isDirty, errors, isValid, isSubmitted },
  } = useForm<FormInput>({
    resolver: getResolver(isMirror, authType),
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

  // if the form has errors, we need to re-validate when isMirror or authType
  // has been changed, after useForm has updated the resolver.
  useEffect(() => {
    if (isSubmitted && !isValid) {
      trigger();
    }
  }, [isMirror, authType, isSubmitted, isValid, trigger]);

  const formId = `new-namespace`;
  return (
    <>
      <DialogHeader>
        <DialogTitle>
          <Home /> {t("components.namespaceCreate.title")}
        </DialogTitle>
      </DialogHeader>

      <Tabs className="mt-2 sm:w-[400px]" defaultValue="namespace">
        <TabsList variant="boxed">
          <TabsTrigger
            variant="boxed"
            value="namespace"
            onClick={() => setIsMirror(false)}
          >
            {t("components.namespaceCreate.tab.namespace")}
          </TabsTrigger>
          <TabsTrigger
            variant="boxed"
            value="mirror"
            onClick={() => setIsMirror(true)}
          >
            {t("components.namespaceCreate.tab.mirror")}
          </TabsTrigger>
        </TabsList>
      </Tabs>

      <div className="mt-1 mb-3">
        <FormErrors errors={errors} className="mb-5" />
        <form
          id={formId}
          onSubmit={handleSubmit(onSubmit)}
          className="flex flex-col gap-y-5"
        >
          <fieldset className="flex items-center gap-5">
            <label
              className="w-[112px] overflow-hidden text-right text-[14px]"
              htmlFor="name"
            >
              {t("components.namespaceCreate.label.name")}
            </label>
            <Input
              id="name"
              data-testid="new-namespace-name"
              placeholder={t("components.namespaceCreate.placeholder.name")}
              {...register("name")}
            />
          </fieldset>

          {isMirror && (
            <>
              <fieldset className="flex items-center gap-5">
                <label
                  className="w-[112px] flex-row overflow-hidden text-right text-[14px]"
                  htmlFor="url"
                >
                  {t("components.namespaceCreate.label.url")}
                </label>
                <InputWithButton>
                  <Input
                    id="url"
                    data-testid="new-namespace-url"
                    placeholder={t(
                      authType === "ssh"
                        ? "components.namespaceCreate.placeholder.gitUrl"
                        : "components.namespaceCreate.placeholder.httpUrl"
                    )}
                    {...register("url")}
                  />
                  <InfoTooltip>
                    {t("components.namespaceCreate.tooltip.url")}
                  </InfoTooltip>
                </InputWithButton>
              </fieldset>

              <fieldset className="flex items-center gap-5">
                <label
                  className="w-[112px] overflow-hidden text-right text-[14px]"
                  htmlFor="ref"
                >
                  {t("components.namespaceCreate.label.ref")}
                </label>
                <InputWithButton>
                  <Input
                    id="ref"
                    data-testid="new-namespace-ref"
                    placeholder={t(
                      "components.namespaceCreate.placeholder.ref"
                    )}
                    {...register("ref")}
                  />
                  <InfoTooltip>
                    {t("components.namespaceCreate.tooltip.ref")}
                  </InfoTooltip>
                </InputWithButton>
              </fieldset>

              <fieldset className="flex items-center gap-5">
                <label
                  className="w-[112px] overflow-hidden text-right text-[14px]"
                  htmlFor="ref"
                >
                  {t("components.namespaceCreate.label.authType")}
                </label>
                <Select
                  value={authType}
                  onValueChange={(value: MirrorAuthType) => setAuthType(value)}
                >
                  <SelectTrigger variant="outline" className="w-full">
                    <SelectValue
                      placeholder={t(
                        "components.namespaceCreate.placeholder.authType"
                      )}
                    />
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
              </fieldset>

              {authType === "token" && (
                <fieldset className="flex items-center gap-5">
                  <label
                    className="w-[112px] overflow-hidden text-right text-[14px]"
                    htmlFor="token"
                  >
                    {t("components.namespaceCreate.label.token")}
                  </label>
                  <InputWithButton>
                    <Textarea
                      id="token"
                      data-testid="new-namespace-token"
                      placeholder={t(
                        "components.namespaceCreate.placeholder.token"
                      )}
                      {...register("passphrase")}
                    />
                    <InfoTooltip>
                      {t("components.namespaceCreate.tooltip.token")}
                    </InfoTooltip>
                  </InputWithButton>
                </fieldset>
              )}

              {authType === "ssh" && (
                <>
                  <fieldset className="flex items-center gap-5">
                    <label
                      className="w-[112px] overflow-hidden text-right text-[14px]"
                      htmlFor="passphrase"
                    >
                      {t("components.namespaceCreate.label.passphrase")}
                    </label>
                    <InputWithButton>
                      <Textarea
                        id="passphrase"
                        data-testid="new-namespace-passphrase"
                        placeholder={t(
                          "components.namespaceCreate.placeholder.passphrase"
                        )}
                        {...register("passphrase")}
                      />
                      <InfoTooltip>
                        {t("components.namespaceCreate.tooltip.passphrase")}
                      </InfoTooltip>
                    </InputWithButton>
                  </fieldset>
                  <fieldset className="flex items-center gap-5">
                    <label
                      className="w-[112px] overflow-hidden text-right text-[14px]"
                      htmlFor="public-key"
                    >
                      {t("components.namespaceCreate.label.publicKey")}
                    </label>
                    <InputWithButton>
                      <Textarea
                        id="public-key"
                        data-testid="new-namespace-pubkey"
                        placeholder={t(
                          "components.namespaceCreate.placeholder.publicKey"
                        )}
                        {...register("publicKey")}
                      />
                      <InfoTooltip>
                        {t("components.namespaceCreate.tooltip.publicKey")}
                      </InfoTooltip>
                    </InputWithButton>
                  </fieldset>

                  <fieldset className="flex items-center gap-5">
                    <label
                      className="w-[112px] overflow-hidden text-right text-[14px]"
                      htmlFor="private-key"
                    >
                      {t("components.namespaceCreate.label.privateKey")}
                    </label>
                    <InputWithButton>
                      <Textarea
                        id="private-key"
                        data-testid="new-namespace-privkey"
                        placeholder={t(
                          "components.namespaceCreate.placeholder.privateKey"
                        )}
                        {...register("privateKey")}
                      />
                      <InfoTooltip>
                        {t("components.namespaceCreate.tooltip.privateKey")}
                      </InfoTooltip>
                    </InputWithButton>
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
