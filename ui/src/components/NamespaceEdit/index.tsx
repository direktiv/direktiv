import {
  DialogClose,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "~/design/Dialog";
import { GitCompare, Home, PlusCircle, Save } from "lucide-react";
import { SubmitHandler, useForm } from "react-hook-form";
import { Tabs, TabsList, TabsTrigger } from "~/design/Tabs";
import { useEffect, useState } from "react";
import { useNamespace, useNamespaceActions } from "~/util/store/namespace";

import Alert from "~/design/Alert";
import Button from "~/design/Button";
import { Checkbox } from "~/design/Checkbox";
import FormErrors from "~/components/FormErrors";
import FormTypeSelect from "./FormTypeSelect";
import InfoTooltip from "./InfoTooltip";
import Input from "~/design/Input";
import { InputWithButton } from "~/design/InputWithButton";
import { MirrorFormType } from "~/api/tree/schema/mirror";
import { MirrorSchemaType } from "~/api/namespaces/schema/namespace";
import { MirrorValidationSchema } from "~/api/tree/schema/mirror/validation";
import { Textarea } from "~/design/TextArea";
import { fileNameSchema } from "~/api/tree/schema/node";
import { pages } from "~/util/router/pages";
import { useCreateNamespace } from "~/api/namespaces/mutate/createNamespace";
import { useListNamespaces } from "~/api/namespaces/query/get";
import { useNavigate } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { useUpdateNamespace } from "~/api/namespaces/mutate/updateNamespace";
import { z } from "zod";
import { zodResolver } from "@hookform/resolvers/zod";

type FormInput = {
  name: string;
  formType: MirrorFormType;
  url: string;
  gitRef: string;
  authToken: string;
  publicKey: string;
  privateKey: string;
  privateKeyPassphrase: string;
  insecure: boolean;
};

/**
 * Form for creating or editing a namespace. Since the namespace name cannot be changed,
 * editing only makes sense to update the mirror definition.
 * @param mirror if present, the form assumes an existing namespace's mirror is edited.
 * If mirror is not present, the form will create a new namespace.
 */
const NamespaceEdit = ({
  mirror,
  close,
}: {
  mirror?: MirrorSchemaType;
  close: () => void;
}) => {
  // note that isMirror is initially redundant with !isNew, but
  // isMirror may change through user interaction.
  const [isMirror, setIsMirror] = useState(!!mirror);
  const isNew = !mirror;
  const { data: namespaces } = useListNamespaces();
  const { setNamespace } = useNamespaceActions();
  const navigate = useNavigate();
  const { t } = useTranslation();
  const namespace = useNamespace();

  const existingNamespaces = namespaces?.data.map((n) => n.name) || [];

  const newNameSchema = fileNameSchema.and(
    z.string().refine((name) => !existingNamespaces.some((n) => n === name), {
      message: t("components.namespaceEdit.nameAlreadyExists"),
    })
  );

  const baseSchema = isNew
    ? z.object({
        name: newNameSchema,
      })
    : z.object({});

  const mirrorSchema = baseSchema.and(MirrorValidationSchema);

  let initialFormType: MirrorFormType = "public";

  if (mirror?.url.startsWith("git@")) {
    initialFormType = "keep-ssh";
  } else if (mirror?.privateKeyPassphrase) {
    initialFormType = "keep-token";
  }

  const {
    handleSubmit,
    register,
    setValue,
    trigger,
    watch,
    formState: { isDirty, errors, isValid, isSubmitted },
  } = useForm<FormInput>({
    resolver: zodResolver(isMirror ? mirrorSchema : baseSchema),
    defaultValues: mirror
      ? {
          formType: initialFormType,
          url: mirror.url,
          gitRef: mirror.gitRef,
          insecure: mirror.insecure,
        }
      : {
          formType: initialFormType,
          insecure: false,
        },
  });

  const formType: MirrorFormType = watch("formType");
  const insecure: boolean = watch("insecure");

  const { mutate: createNamespace, isLoading } = useCreateNamespace({
    onSuccess: (data) => {
      setNamespace(data.data.name);
      navigate(
        pages.explorer.createHref({
          namespace: data.data.name,
        })
      );
      close();
    },
  });

  const { mutate: updateMirror } = useUpdateNamespace({
    onSuccess: () => {
      close();
    },
  });

  const onSubmit: SubmitHandler<FormInput> = ({
    name,
    gitRef,
    url,
    formType,
    authToken,
    publicKey,
    privateKey,
    privateKeyPassphrase,
    insecure,
  }) => {
    if (isNew) {
      return createNamespace({
        name,
        mirror: isMirror
          ? {
              gitRef,
              authToken,
              url,
              publicKey,
              privateKey,
              privateKeyPassphrase,
              insecure,
            }
          : undefined,
      });
    }

    if (!namespace) throw Error("Namespace undefined while updating mirror");

    let updateAuthValues = {};

    if (formType === "public") {
      updateAuthValues = {
        passphrase: "",
        publicKey: "",
        privateKey: "",
      };
    }

    if (formType === "keep-token") {
      updateAuthValues = {
        passphrase: "-",
      };
    }

    if (formType === "token") {
      updateAuthValues = {
        publicKey,
      };
    }

    if (formType === "keep-ssh") {
      updateAuthValues = {
        publicKey: "-",
        privateKey: "-",
        privateKeyPassphrase: "-",
      };
    }

    if (formType === "ssh") {
      updateAuthValues = {
        publicKey,
        privateKey,
        privateKeyPassphrase,
      };
    }

    return updateMirror({
      namespace,
      mirror: {
        gitRef,
        url,
        ...updateAuthValues,
        insecure,
      },
    });
  };

  // you can not submit if the form has not changed or if there are any errors and
  // you have already submitted the form (errors will first show up after submit)
  const disableSubmit = !isDirty || (isSubmitted && !isValid);

  // if the form has errors, we need to re-validate when isMirror or formType
  // has been changed, after useForm has updated the resolver.
  useEffect(() => {
    if (isSubmitted && !isValid) {
      trigger();
    }
  }, [isMirror, formType, isSubmitted, isValid, trigger]);

  const formId = "new-namespace";
  return (
    <>
      <DialogHeader>
        <DialogTitle>
          {isNew ? <Home /> : <GitCompare />}
          {isNew
            ? t("components.namespaceEdit.title.new")
            : t("components.namespaceEdit.title.edit", {
                namespace,
              })}
        </DialogTitle>
      </DialogHeader>

      {isNew && (
        <Tabs className="mt-2 sm:w-[400px]" defaultValue="namespace">
          <TabsList variant="boxed">
            <TabsTrigger
              variant="boxed"
              value="namespace"
              onClick={() => setIsMirror(false)}
            >
              {t("components.namespaceEdit.tab.namespace")}
            </TabsTrigger>
            <TabsTrigger
              variant="boxed"
              value="mirror"
              onClick={() => setIsMirror(true)}
            >
              {t("components.namespaceEdit.tab.mirror")}
            </TabsTrigger>
          </TabsList>
        </Tabs>
      )}

      <div className="mt-1 mb-3">
        <FormErrors errors={errors} className="mb-5" />
        <form
          id={formId}
          onSubmit={handleSubmit(onSubmit)}
          className="flex flex-col gap-y-5"
        >
          {isNew && (
            <fieldset className="flex items-center gap-5">
              <label
                className="w-[112px] overflow-hidden text-right text-[14px]"
                htmlFor="name"
              >
                {t("components.namespaceEdit.label.name")}
              </label>
              <Input
                id="name"
                data-testid="new-namespace-name"
                placeholder={t("components.namespaceEdit.placeholder.name")}
                {...register("name")}
              />
            </fieldset>
          )}

          {isMirror && (
            <>
              <fieldset className="flex items-center gap-5">
                <label
                  className="w-[112px] flex-row overflow-hidden text-right text-[14px]"
                  htmlFor="url"
                >
                  {t("components.namespaceEdit.label.url")}
                </label>
                <InputWithButton>
                  <Input
                    id="url"
                    data-testid="new-namespace-url"
                    placeholder={t(
                      formType === "ssh"
                        ? "components.namespaceEdit.placeholder.gitUrl"
                        : "components.namespaceEdit.placeholder.httpUrl"
                    )}
                    {...register("url")}
                  />
                  <InfoTooltip>
                    {t("components.namespaceEdit.tooltip.url")}
                  </InfoTooltip>
                </InputWithButton>
              </fieldset>

              <fieldset className="flex items-center gap-5">
                <label
                  className="w-[112px] overflow-hidden text-right text-[14px]"
                  htmlFor="ref"
                >
                  {t("components.namespaceEdit.label.ref")}
                </label>
                <InputWithButton>
                  <Input
                    id="ref"
                    data-testid="new-namespace-ref"
                    placeholder={t("components.namespaceEdit.placeholder.ref")}
                    {...register("gitRef")}
                  />
                  <InfoTooltip>
                    {t("components.namespaceEdit.tooltip.ref")}
                  </InfoTooltip>
                </InputWithButton>
              </fieldset>

              <fieldset className="flex items-center gap-5">
                <label
                  className="w-[112px] overflow-hidden text-right text-[14px]"
                  htmlFor="formType"
                >
                  {t("components.namespaceEdit.label.formType")}
                </label>
                <FormTypeSelect
                  value={formType}
                  storedValue={initialFormType}
                  isNew={isNew}
                  onValueChange={(value: MirrorFormType) =>
                    setValue("formType", value, { shouldDirty: true })
                  }
                />
              </fieldset>

              {!isNew && formType.startsWith("keep") && (
                <Alert variant="info" className="text-sm">
                  {t("components.namespaceEdit.formTypeMessage.keep")}
                </Alert>
              )}

              {!isNew && !formType.startsWith("keep") && (
                <Alert variant="info" className="text-sm">
                  {t("components.namespaceEdit.formTypeMessage.replace")}
                </Alert>
              )}

              {formType === "token" && (
                <fieldset className="flex items-center gap-5">
                  <label
                    className="w-[112px] overflow-hidden text-right text-[14px]"
                    htmlFor="token"
                  >
                    {t("components.namespaceEdit.label.token")}
                  </label>
                  <InputWithButton>
                    <Textarea
                      id="token"
                      data-testid="new-namespace-token"
                      placeholder={t(
                        "components.namespaceEdit.placeholder.token"
                      )}
                      {...register("authToken")}
                    />
                    <InfoTooltip>
                      {t("components.namespaceEdit.tooltip.token")}
                    </InfoTooltip>
                  </InputWithButton>
                </fieldset>
              )}

              {formType === "ssh" && (
                <>
                  <fieldset className="flex items-center gap-5">
                    <label
                      className="w-[112px] overflow-hidden text-right text-[14px]"
                      htmlFor="passphrase"
                    >
                      {t("components.namespaceEdit.label.passphrase")}
                    </label>
                    <InputWithButton>
                      <Textarea
                        id="passphrase"
                        data-testid="new-namespace-passphrase"
                        placeholder={t(
                          "components.namespaceEdit.placeholder.passphrase"
                        )}
                        {...register("privateKeyPassphrase")}
                      />
                      <InfoTooltip>
                        {t("components.namespaceEdit.tooltip.passphrase")}
                      </InfoTooltip>
                    </InputWithButton>
                  </fieldset>
                  <fieldset className="flex items-center gap-5">
                    <label
                      className="w-[112px] overflow-hidden text-right text-[14px]"
                      htmlFor="public-key"
                    >
                      {t("components.namespaceEdit.label.publicKey")}
                    </label>
                    <InputWithButton>
                      <Textarea
                        id="public-key"
                        data-testid="new-namespace-pubkey"
                        placeholder={t(
                          "components.namespaceEdit.placeholder.publicKey"
                        )}
                        {...register("publicKey")}
                      />
                      <InfoTooltip>
                        {t("components.namespaceEdit.tooltip.publicKey")}
                      </InfoTooltip>
                    </InputWithButton>
                  </fieldset>

                  <fieldset className="flex items-center gap-5">
                    <label
                      className="w-[112px] overflow-hidden text-right text-[14px]"
                      htmlFor="private-key"
                    >
                      {t("components.namespaceEdit.label.privateKey")}
                    </label>
                    <InputWithButton>
                      <Textarea
                        id="private-key"
                        data-testid="new-namespace-privkey"
                        placeholder={t(
                          "components.namespaceEdit.placeholder.privateKey"
                        )}
                        {...register("privateKey")}
                      />
                      <InfoTooltip>
                        {t("components.namespaceEdit.tooltip.privateKey")}
                      </InfoTooltip>
                    </InputWithButton>
                  </fieldset>
                </>
              )}

              <fieldset className="flex items-center justify-between gap-5">
                <label className="pl-5 text-[14px]" htmlFor="insecure">
                  {t("components.namespaceEdit.label.insecure")}
                </label>
                <div className="flex gap-5 pr-2">
                  <Checkbox
                    id="insecure"
                    checked={insecure}
                    onCheckedChange={() =>
                      setValue("insecure", !insecure, { shouldDirty: true })
                    }
                  />
                  <InfoTooltip>
                    {t("components.namespaceEdit.tooltip.insecure")}
                  </InfoTooltip>
                </div>
              </fieldset>
            </>
          )}
        </form>
      </div>

      <DialogFooter>
        <DialogClose asChild>
          <Button variant="ghost">
            {t("components.namespaceEdit.cancelBtn")}
          </Button>
        </DialogClose>
        <Button
          data-testid="new-namespace-submit"
          type="submit"
          disabled={disableSubmit}
          loading={isLoading}
          form={formId}
        >
          {!isLoading && (isNew ? <PlusCircle /> : <Save />)}
          {isNew
            ? t("components.namespaceEdit.submitBtn.create")
            : t("components.namespaceEdit.submitBtn.save")}
        </Button>
      </DialogFooter>
    </>
  );
};

export default NamespaceEdit;
