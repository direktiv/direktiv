import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "~/design/Dialog";
import { Eye, EyeOff, KeyRound, LogIn } from "lucide-react";
import { SubmitHandler, useForm } from "react-hook-form";
import { useApiActions, useApiKey } from "~/util/store/apiKey";
import { useEffect, useState } from "react";

import Button from "~/design/Button";
import FormErrors from "../FormErrors";
import Input from "~/design/Input";
import { InputWithButton } from "~/design/InputWithButton";
import Logo from "~/components/Logo";
import { useAuthenticate } from "~/api/authenticate/mutate/authenticate";
import { useTranslation } from "react-i18next";
import { z } from "zod";
import { zodResolver } from "@hookform/resolvers/zod";

type FormInput = {
  apiKey: string;
};

export const Authdialog = () => {
  const { t } = useTranslation();
  const [showKey, setShowKey] = useState(false);
  const { setApiKey: storeApiKey } = useApiActions();
  const apiKeyFromLocalStorage = useApiKey();

  const {
    register,
    handleSubmit,
    setError,
    formState: { isDirty, errors, isValid, isSubmitted },
  } = useForm<FormInput>({
    resolver: zodResolver(
      z.object({
        apiKey: z.string(),
      })
    ),
    defaultValues: {
      apiKey: apiKeyFromLocalStorage ?? "",
    },
  });

  const { mutate: authenticate, isPending } = useAuthenticate({
    onSuccess: (isKeyCorrect, apiKey) => {
      isKeyCorrect
        ? storeApiKey(apiKey)
        : setError("apiKey", {
            message: t("pages.authenticate.wrongKey"),
          });
    },
  });

  const onSubmit: SubmitHandler<FormInput> = ({ apiKey }) => {
    authenticate(apiKey);
  };

  useEffect(() => {
    /**
     * when this component is rendered and there is an api key in local storage we
     * already know that the key is wrong, so we can delete it and show a message to
     * the user
     */
    if (apiKeyFromLocalStorage) {
      setError("apiKey", {
        message: t("pages.authenticate.wrongOldKey"),
      });
      storeApiKey(null);
    }
  }, [apiKeyFromLocalStorage, setError, storeApiKey, t]);

  const disableSubmit = !isDirty || (isSubmitted && !isValid);
  const formId = `authdialog-form`;

  return (
    <Dialog open={true}>
      <DialogContent
        className="rounded-md bg-white shadow ring-1 ring-gray-5 dark:bg-black dark:ring-gray-dark-5 max-sm:top-20"
        overlayProps={{
          className: "bg-white dark:bg-black",
        }}
      >
        <div className="absolute -top-14 flex w-full justify-center">
          <Logo />
        </div>
        <DialogHeader>
          <DialogTitle>
            <KeyRound /> {t("pages.authenticate.title")}
          </DialogTitle>
        </DialogHeader>
        <div className="my-3">{t("pages.authenticate.description")}</div>
        <div className="mb-3">
          <FormErrors errors={errors} className="mb-3" />
          <form id={formId} onSubmit={handleSubmit(onSubmit)}>
            <fieldset className="flex items-center gap-5">
              <label
                className="w-[90px] text-right text-[14px]"
                htmlFor="apiKey"
              >
                {t("pages.authenticate.apiKey")}
              </label>
              <InputWithButton className="w-full">
                <Input
                  id="apiKey"
                  placeholder={t("pages.authenticate.apiKeyPlaceholder")}
                  {...register("apiKey")}
                  type={showKey ? "text" : "password"}
                />
                <Button
                  icon
                  type="button"
                  variant="ghost"
                  onClick={() => {
                    setShowKey((prev) => !prev);
                  }}
                >
                  {showKey ? <EyeOff /> : <Eye />}
                </Button>
              </InputWithButton>
            </fieldset>
          </form>
        </div>
        <DialogFooter>
          <Button
            type="submit"
            disabled={disableSubmit}
            loading={isPending}
            form={formId}
          >
            {!isPending && <LogIn />}
            {t("pages.authenticate.loginBtn")}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};
