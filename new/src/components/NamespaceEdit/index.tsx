import {
  DialogClose,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "~/design/Dialog";
import { GitCompare, Home, PlusCircle, Save } from "lucide-react";
import { MirrorFormType, MirrorInfoSchemaType } from "~/api/tree/schema/mirror";
import { SubmitHandler, useForm } from "react-hook-form";
import { Tabs, TabsList, TabsTrigger } from "~/design/Tabs";
import { useEffect, useState } from "react";

import Alert from "~/design/Alert";
import Button from "~/design/Button";
import { Checkbox } from "~/design/Checkbox";
import FormErrors from "~/components/FormErrors";
import FormTypeSelect from "./FormTypeSelect";
import InfoTooltip from "./InfoTooltip";
import Input from "~/design/Input";
import { InputWithButton } from "~/design/InputWithButton";
import { MirrorValidationSchema } from "~/api/tree/schema/mirror/validation";
import { Textarea } from "~/design/TextArea";
import { fileNameSchema } from "~/api/tree/schema/node";
import { pages } from "~/util/router/pages";
import { useCreateNamespace } from "~/api/namespaces/mutate/createNamespace";
import { useListNamespaces } from "~/api/namespaces/query/get";
import { useNamespaceActions } from "~/util/store/namespace";
import { useNavigate } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { useUpdateMirror } from "~/api/tree/mutate/updateMirror";
import { z } from "zod";
import { zodResolver } from "@hookform/resolvers/zod";

type FormInput = {
  name: string;
  formType: MirrorFormType;
  url: string;
  ref: string;
  passphrase: string;
  publicKey: string;
  privateKey: string;
  insecure: boolean;
};

const NamespaceEdit = ({
  mirror,
  close,
}: {
  mirror?: MirrorInfoSchemaType;
  close: () => void;
}) => {
  // note that isMirror is initially redundant with !isNew, but
  // isMirror may change through user interaction.
  const [isMirror, setIsMirror] = useState(!!mirror);
  const isNew = !mirror;
  const { data } = useListNamespaces();
  const { setNamespace } = useNamespaceActions();
  const navigate = useNavigate();
  const { t } = useTranslation();

  const existingNamespaces = data?.results.map((n) => n.name) || [];

  const newNameSchema = fileNameSchema.and(
    z.string().refine((name) => !existingNamespaces.some((n) => n === name), {
      message: t("components.namespaceEdit.nameAlreadyExists"),
    })
  );

  const baseSchema = z.object({ name: isNew ? newNameSchema : z.string() });
  const mirrorSchema = baseSchema.and(MirrorValidationSchema);

  let initialFormType: MirrorFormType = "public";

  if (mirror?.info.url.startsWith("git@")) {
    initialFormType = "keep-ssh";
  } else if (mirror?.info.passphrase) {
    initialFormType = "keep-token";
  }

  const {
    handleSubmit,
    register,
    setValue,
    trigger,
    watch,
    formState: { dirtyFields, errors, isValid, isSubmitted },
  } = useForm<FormInput>({
    resolver: zodResolver(isMirror ? mirrorSchema : baseSchema),
    defaultValues: mirror
      ? {
          formType: initialFormType,
          name: mirror.namespace,
          url: mirror.info.url,
          ref: mirror.info.ref,
          insecure: mirror.info.insecure,
        }
      : {
          formType: initialFormType,
          insecure: false,
        },
  });

  // For some strange reason, useForm's formState.isDirty doesn't react when
  // the field "ref" becomes dirty, even though it is registered in dirtyFields.
  // So as a workaround, we infer isDirty from dirtyFields.
  const isDirty = Object.values(dirtyFields).some((value) => value === true);

  const formType: MirrorFormType = watch("formType");
  const insecure: boolean = watch("insecure");

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

  const { mutate: updateMirror } = useUpdateMirror({
    onSuccess: () => {
      close();
    },
  });

  const onSubmit: SubmitHandler<FormInput> = ({
    name,
    ref,
    url,
    formType,
    passphrase,
    publicKey,
    privateKey,
    insecure,
  }) => {
    if (isNew) {
      return createNamespace({
        name,
        mirror: isMirror
          ? { ref, url, passphrase, publicKey, privateKey, insecure }
          : undefined,
      });
    }

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
        passphrase,
      };
    }
    if (formType === "keep-ssh") {
      updateAuthValues = {
        passphrase: "-",
        publicKey: "-",
        privateKey: "-",
      };
    }
    if (formType === "ssh") {
      updateAuthValues = {
        passphrase,
        publicKey,
        privateKey,
      };
    }

    return updateMirror({
      name,
      mirror: {
        ref,
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
                namespace: mirror?.namespace,
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
                    {...register("ref")}
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
                      {...register("passphrase")}
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
                        {...register("passphrase")}
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
