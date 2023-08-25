import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "~/design/Dialog";
import { Eye, EyeOff, KeyRound, LogIn } from "lucide-react";
import { SubmitHandler, useForm } from "react-hook-form";

import Button from "~/design/Button";
import FormErrors from "../FormErrors";
import Input from "~/design/Input";
import { InputWithButton } from "~/design/InputWithButton";
import Logo from "~/design/Logo";
import { useApiActions } from "~/util/store/apiKey";
import { useAuthenticate } from "~/api/authenticate/mutate/authenticate";
import { useState } from "react";
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
  const { mutate: authenticate, isLoading } = useAuthenticate({
    onSuccess: (isKeyCorrect, apiKey) => {
      if (isKeyCorrect) {
        storeApiKey(apiKey);
      }
    },
  });

  const {
    register,
    handleSubmit,
    formState: { isDirty, errors, isValid, isSubmitted },
  } = useForm<FormInput>({
    resolver: zodResolver(
      z.object({
        apiKey: z.string(),
      })
    ),
  });

  const onSubmit: SubmitHandler<FormInput> = ({ apiKey }) => {
    authenticate(apiKey);
  };

  // TODO: check if there is an old, and show the user if its wrong
  // TODO: delete old one if not needed anymore

  const disableSubmit = !isDirty || (isSubmitted && !isValid);
  const formId = `authdialog-form`;

  return (
    <Dialog open={true}>
      <DialogContent className="max-sm:top-20">
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
            loading={isLoading}
            form={formId}
          >
            {!isLoading && <LogIn />}
            {t("pages.authenticate.loginBtn")}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};
